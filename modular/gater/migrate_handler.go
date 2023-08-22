package gater

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	permissiontypes "github.com/bnb-chain/greenfield/x/permission/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	modelgateway "github.com/bnb-chain/greenfield-storage-provider/model/gateway"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield/types/common"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

// dest sp receive migrate gvg notify from src sp.
func (g *GateModular) notifyMigrateSwapOutHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		reqCtx     *RequestContext
		swapOutMsg []byte
	)
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(err)
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, _ = NewRequestContext(r, g)
	migrateSwapOutHeader := r.Header.Get(GnfdMigrateSwapOutMsgHeader)
	if swapOutMsg, err = hex.DecodeString(migrateSwapOutHeader); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse migrate swap out header", "error", err)
		err = ErrDecodeMsg
		return
	}
	swapOut := virtualgrouptypes.MsgSwapOut{}
	if err = json.Unmarshal(swapOutMsg, &swapOut); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal migrate swap out msg", "error", err)
		err = ErrDecodeMsg
		return
	}
	if err = g.baseApp.GfSpClient().NotifyMigrateSwapOut(reqCtx.Context(), &swapOut); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to notify migrate swap out", "swap_out", swapOut, "error", err)
		err = ErrNotifySwapOutWithDetail("failed to notify migrate swap out, swap_out: " + swapOut.String() + ",error: " + err.Error())
		return
	}
}

func (g *GateModular) checkMigratePieceAuth(reqCtx *RequestContext, migrateGVGHeader string) (bool, error) {
	var (
		err           error
		migrateGVGMsg []byte
	)
	migrateGVGMsg, err = hex.DecodeString(migrateGVGHeader)
	if err != nil {
		log.Errorw("failed to parse migrate gvg header", "migrate_gvg_header", migrateGVGHeader, "error", err)
		return false, ErrDecodeMsg
	}
	migrateGVG := gfsptask.GfSpMigrateGVGTask{}
	err = json.Unmarshal(migrateGVGMsg, &migrateGVG)
	if err != nil {
		log.Errorw("failed to unmarshal migrate gvg msg", "error", err)
		return false, ErrDecodeMsg
	}
	if migrateGVG.GetExpireTime() < time.Now().Unix() {
		log.Errorw("failed to check migrate gvg expire time", "gvg_task", migrateGVG)
		return false, ErrNoPermission
	}
	destSPAddr, err := reqCtx.verifyTaskSignature(migrateGVG.GetSignBytes(), migrateGVG.GetSignature())
	if err != nil {
		log.Errorw("failed to verify task signature", "gvg_task", migrateGVG, "error", err)
		return false, err
	}
	sp, err := g.spCachePool.QuerySPByAddress(destSPAddr.String())
	if err != nil {
		log.Errorw("failed to query sp", "gvg_task", migrateGVG, "dest_sp_addr", destSPAddr.String(), "error", err)
		return false, err
	}
	effect, err := g.baseApp.GfSpClient().VerifyMigrateGVGPermission(reqCtx.Context(), migrateGVG.GetBucketID(), migrateGVG.GetSrcGvg().GetId(), sp.GetId())
	if effect == nil || err != nil {
		log.Errorw("failed to verify migrate gvg permission", "gvg_bucket_id", migrateGVG.GetBucketID(),
			"src_gvg_id", migrateGVG.GetSrcGvg().GetId(), "dest_sp", sp.GetId(), "effect", effect, "error", err)
		return false, err
	}
	if *effect == permissiontypes.EFFECT_ALLOW {
		return true, nil
	}
	return false, nil
}

// checkMigrateBucketQuotaAuth parse migrateBucketMsgHeader to GfSpBucketMigrationStatus and check if request has permission
// only used by preMigrateBucketHandler and postMigrateBucketHandler
func (g *GateModular) checkMigrateBucketQuotaAuth(reqCtx *RequestContext, migrateBucketMsgHeader string) (bool, *gfsptask.GfSpBucketMigrationStatus, error) {
	allow, migrateBucketStatus, sp, err := g.verifySignatureAndSP(reqCtx, migrateBucketMsgHeader)
	if !allow {
		return allow, migrateBucketStatus, err
	}

	effect, err := g.baseApp.GfSpClient().VerifyMigrateGVGPermission(reqCtx.Context(), migrateBucketStatus.GetBucketId(), 0 /*not used*/, sp.GetId())
	if effect == nil || err != nil {
		// metadate serive support cancel & completion migrate bucket
		if strings.Contains(err.Error(), "the bucket is not in migration status") {
			return true, migrateBucketStatus, nil
		}
		log.Errorw("failed to verify migrate gvg permission", "bucket_migration", migrateBucketStatus,
			"dest_sp", sp.GetId(), "effect", effect, "error", err)
		return false, migrateBucketStatus, err
	}
	if *effect == permissiontypes.EFFECT_ALLOW {
		return true, migrateBucketStatus, nil
	}
	return false, migrateBucketStatus, nil
}

// verifySignatureAndSP parse migrateBucketMsgHeader to GfSpBucketMigrationStatus and check if request has permission
func (g *GateModular) verifySignatureAndSP(reqCtx *RequestContext, migrateBucketMsgHeader string) (bool, *gfsptask.GfSpBucketMigrationStatus, *sptypes.StorageProvider, error) {
	var (
		err              error
		migrateBucketMsg []byte
		sp               *sptypes.StorageProvider
	)
	migrateBucketMsg, err = hex.DecodeString(migrateBucketMsgHeader)
	if err != nil {
		log.Errorw("failed to parse migrate gvg header", "migrate_gvg_header", migrateBucketMsg, "error", err)
		return false, &gfsptask.GfSpBucketMigrationStatus{}, sp, ErrDecodeMsg
	}
	migrateBucketStatus := gfsptask.GfSpBucketMigrationStatus{}
	if err = json.Unmarshal(migrateBucketMsg, &migrateBucketStatus); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal post migrate bucket msg", "error", err)
		err = ErrDecodeMsg
		return false, &migrateBucketStatus, sp, err
	}
	if migrateBucketStatus.GetExpireTime() < time.Now().Unix() {
		log.Errorw("failed to check migrate gvg expire time", "bucket_migration", migrateBucketStatus)
		return false, &migrateBucketStatus, sp, ErrNoPermission
	}
	destSPAddr, err := reqCtx.verifyTaskSignature(migrateBucketStatus.GetSignBytes(), migrateBucketStatus.GetSignature())
	if err != nil {
		log.Errorw("failed to verify task signature", "bucket_migration", migrateBucketStatus, "error", err)
		return false, &migrateBucketStatus, sp, err
	}
	sp, err = g.spCachePool.QuerySPByAddress(destSPAddr.String())
	if err != nil {
		log.Errorw("failed to query sp", "bucket_migration", migrateBucketStatus, "dest_sp_addr", destSPAddr.String(), "error", err)
		return false, &migrateBucketStatus, sp, err
	}
	return true, &migrateBucketStatus, sp, nil
}

// migratePieceHandler handles migrate piece request between SPs which is used in SP exiting case.
func (g *GateModular) migratePieceHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err             error
		allowMigrate    bool
		reqCtx          *RequestContext
		migratePieceMsg []byte
		pieceKey        string
		pieceSize       int64
		pieceData       []byte
	)
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(err)
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, _ = NewRequestContext(r, g)
	migratePieceHeader := r.Header.Get(GnfdMigratePieceMsgHeader)
	migratePieceMsg, err = hex.DecodeString(migratePieceHeader)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse migrate piece header", "error", err)
		err = ErrDecodeMsg
		return
	}

	migratePiece := gfsptask.GfSpMigratePieceTask{}
	err = json.Unmarshal(migratePieceMsg, &migratePiece)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal migrate piece msg", "error", err)
		err = ErrDecodeMsg
		return
	}

	if allowMigrate, err = g.checkMigratePieceAuth(reqCtx, r.Header.Get(GnfdMigrateGVGMsgHeader)); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to check migrate piece auth", "migrate_piece", migratePiece, "error", err)
		return
	}
	if !allowMigrate {
		log.CtxErrorw(reqCtx.Context(), "no permission to migrate piece", "migrate_piece", migratePiece)
		err = ErrNoPermission
		return
	}

	objectInfo := migratePiece.GetObjectInfo()
	if objectInfo == nil {
		log.CtxError(reqCtx.Context(), "failed to get migrate piece object info due to has no object info")
		err = ErrInvalidHeader
		return
	}
	chainObjectInfo, bucketInfo, params, err := g.getObjectChainMeta(reqCtx.Context(), objectInfo.GetObjectName(), objectInfo.GetBucketName())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object on chain meta", "error", err)
		err = ErrInvalidHeader
		return
	}

	redundancyNumber := int32(migratePiece.GetStorageParams().GetRedundantDataChunkNum()+migratePiece.GetStorageParams().GetRedundantParityChunkNum()) - 1
	objectID := migratePiece.GetObjectInfo().Id.Uint64()
	segmentIdx := migratePiece.GetSegmentIdx()
	redundancyIdx := migratePiece.GetRedundancyIdx()
	if redundancyIdx < piecestore.PrimarySPRedundancyIndex || redundancyIdx > redundancyNumber {
		err = ErrInvalidRedundancyIndex
		return
	}
	if redundancyIdx == piecestore.PrimarySPRedundancyIndex {
		pieceKey = g.baseApp.PieceOp().SegmentPieceKey(objectID, segmentIdx)
		pieceSize = g.baseApp.PieceOp().SegmentPieceSize(migratePiece.ObjectInfo.GetPayloadSize(),
			segmentIdx, migratePiece.GetStorageParams().GetMaxSegmentSize())
	} else {
		pieceKey = g.baseApp.PieceOp().ECPieceKey(objectID, segmentIdx, uint32(redundancyIdx))
		pieceSize = g.baseApp.PieceOp().ECPieceSize(migratePiece.ObjectInfo.GetPayloadSize(), segmentIdx,
			migratePiece.GetStorageParams().GetMaxSegmentSize(), migratePiece.GetStorageParams().GetRedundantDataChunkNum())
	}

	pieceTask := &gfsptask.GfSpDownloadPieceTask{}
	pieceTask.InitDownloadPieceTask(chainObjectInfo, bucketInfo, params, coretask.DefaultSmallerPriority, false, bucketInfo.Owner,
		uint64(pieceSize), pieceKey, 0, uint64(pieceSize),
		g.baseApp.TaskTimeout(pieceTask, objectInfo.GetPayloadSize()), g.baseApp.TaskMaxRetry(pieceTask))
	pieceData, err = g.baseApp.GfSpClient().GetPiece(reqCtx.Context(), pieceTask)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to download segment piece", "piece_key", pieceKey, "error", err)
		return
	}

	_, err = w.Write(pieceData)
	if err != nil {
		err = ErrReplyData
		log.CtxErrorw(reqCtx.Context(), "failed to reply the migrate data", "error", err)
		return
	}
	log.CtxInfow(reqCtx.Context(), "succeed to migrate one piece", "object_id", objectID, "segment_piece_index",
		segmentIdx, "redundancy_index", redundancyIdx)
}

// getSecondaryBlsMigrationBucketApprovalHandler handles the bucket migration approval request for secondarySP using bls
func (g *GateModular) getSecondaryBlsMigrationBucketApprovalHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                        error
		reqCtx                     *RequestContext
		migrationBucketApprovalMsg []byte
	)
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(err)
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, _ = NewRequestContext(r, g)
	migrationBucketApprovalHeader := r.Header.Get(GnfdSecondarySPMigrationBucketMsgHeader)
	migrationBucketApprovalMsg, err = hex.DecodeString(migrationBucketApprovalHeader)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse secondary migration bucket approval header", "error", err)
		err = ErrDecodeMsg
		return
	}

	signDoc := &storagetypes.SecondarySpMigrationBucketSignDoc{}
	if err = storagetypes.ModuleCdc.UnmarshalJSON(migrationBucketApprovalMsg, signDoc); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal migration bucket approval msg", "error", err)
		err = ErrDecodeMsg
		return
	}
	signature, err := g.baseApp.GfSpClient().SignSecondarySPMigrationBucket(reqCtx.Context(), signDoc)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to sign secondary sp migration bucket", "error", err)
		err = ErrMigrateApprovalWithDetail("failed to sign secondary sp migration bucket, error: " + err.Error())
		return
	}
	w.Header().Set(GnfdSecondarySPMigrationBucketApprovalHeader, hex.EncodeToString(signature))
	log.CtxInfow(reqCtx.Context(), "succeed to sign secondary sp migration bucket approval", "bucket_id",
		signDoc.BucketId.String())
}

func (g *GateModular) getSwapOutApproval(w http.ResponseWriter, r *http.Request) {
	var (
		err                error
		reqCtx             *RequestContext
		swapOutApprovalMsg []byte
	)
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(err)
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, _ = NewRequestContext(r, g)
	swapOutApprovalHeader := r.Header.Get(GnfdUnsignedApprovalMsgHeader)
	swapOutApprovalMsg, err = hex.DecodeString(swapOutApprovalHeader)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse swap out approval header", "error", err)
		err = ErrDecodeMsg
		return
	}

	swapOutApproval := &virtualgrouptypes.MsgSwapOut{}
	if err = virtualgrouptypes.ModuleCdc.UnmarshalJSON(swapOutApprovalMsg, swapOutApproval); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal swap out approval msg", "error", err)
		err = ErrDecodeMsg
		return
	}
	swapOutApproval.SuccessorSpApproval = &common.Approval{ExpiredHeight: 100}

	if err = swapOutApproval.ValidateBasic(); err != nil {
		log.Errorw("failed to basic check approval msg", "swap_out_approval", swapOutApproval, "error", err)
		err = ErrValidateMsg
		return
	}

	log.Infow("get swap out approval", "msg", swapOutApproval)
	signature, err := g.baseApp.GfSpClient().SignSwapOut(reqCtx.Context(), swapOutApproval)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to sign swap out", "error", err)
		err = ErrMigrateApprovalWithDetail("failed to sign swap out, error: " + err.Error())
		return
	}
	swapOutApproval.SuccessorSpApproval.Sig = signature
	bz := storagetypes.ModuleCdc.MustMarshalJSON(swapOutApproval)
	w.Header().Set(GnfdSignedApprovalMsgHeader, hex.EncodeToString(sdktypes.MustSortJSON(bz)))
	log.CtxInfow(reqCtx.Context(), "succeed to sign swap out approval", "swap_out", swapOutApproval.String())
}

// getLatestBucketQuotaHandler handles the query quota request for bucket migrate
func (g *GateModular) getLatestBucketQuotaHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                 error
		reqCtx              *RequestContext
		bucketID            uint64
		migrateBucketStatus *gfsptask.GfSpBucketMigrationStatus
		allowMigrate        bool
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(err)
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, _ = NewRequestContext(r, g)
	if allowMigrate, migrateBucketStatus, _, err = g.verifySignatureAndSP(reqCtx, r.Header.Get(GnfdMigrateBucketMsgHeader)); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to check migrate bucket auth", "migrate_bucket", migrateBucketStatus, "error", err)
		return
	}
	if !allowMigrate {
		log.CtxErrorw(reqCtx.Context(), "no permission to migrate piece", "migrate_bucket", migrateBucketStatus)
		err = ErrNoPermission
		return
	}

	bucketID = migrateBucketStatus.GetBucketId()

	quota, err := g.baseApp.GfSpClient().GetLatestBucketReadQuota(
		reqCtx.Context(), bucketID)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket read quota", "bucket_id",
			bucketID, "error", err)
		return
	}

	bz, err := quota.Marshal()
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to marshal", "bucket_id",
			bucketID, "error", err)
		return
	}

	w.Header().Set(GnfdSignedApprovalMsgHeader, hex.EncodeToString(bz))
	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)

	log.CtxInfow(reqCtx.Context(), "succeed to get latest bucket quota", "bucket_id",
		bucketID, "quota", quota)
}

// preMigrateBucketHandler handles the prepare request for bucket migrate
func (g *GateModular) preMigrateBucketHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                 error
		reqCtx              *RequestContext
		bucketID            uint64
		bucketSize          uint64
		migrateBucketStatus *gfsptask.GfSpBucketMigrationStatus
		allowMigrate        bool
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(err)
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, _ = NewRequestContext(r, g)
	if allowMigrate, migrateBucketStatus, err = g.checkMigrateBucketQuotaAuth(reqCtx, r.Header.Get(GnfdMigrateBucketMsgHeader)); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to check migrate bucket auth", "migrate_bucket", migrateBucketStatus, "error", err)
		return
	}
	if !allowMigrate {
		log.CtxErrorw(reqCtx.Context(), "no permission to migrate piece", "migrate_bucket", migrateBucketStatus)
		err = ErrNoPermission
		return
	}

	bucketID = migrateBucketStatus.GetBucketId()

	// get bucket quota and check, lock quota
	bucketSize, err = g.getBucketTotalSize(reqCtx.Context(), bucketID)
	if err != nil {
		return
	}

	quota, err := g.baseApp.GfSpClient().GetLatestBucketReadQuota(
		reqCtx.Context(), bucketID)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket read quota", "bucket_id",
			bucketID, "error", err)
		return
	}

	// reduce quota, sqldb
	err = g.baseApp.GfSpClient().DeductQuotaForBucketMigrate(
		reqCtx.Context(), bucketID, bucketSize, quota.GetMonth())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket read quota", "bucket_id",
			bucketID, "error", err)
		return
	}

	bz, err := quota.Marshal()
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to marshal", "bucket_id",
			bucketID, "error", err)
		return
	}

	err = g.baseApp.GfSpClient().NotifyPreMigrateBucket(reqCtx.Context(), bucketID)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to pre migrate bucket, the bucket may already notified", "bucket_id",
			bucketID, "error", err)
		return
	}

	w.Header().Set(GnfdSignedApprovalMsgHeader, hex.EncodeToString(bz))
	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)

	log.CtxInfow(reqCtx.Context(), "succeed to pre bucket migrate and deduct quota", "bucket_id",
		bucketID, "quota", quota, "bucket_quota_size", bucketSize)
}

// postMigrateBucketHandler notifying the source sp about the completion of migration bucket
func (g *GateModular) postMigrateBucketHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                 error
		reqCtx              *RequestContext
		bucketID            uint64
		bucketSize          uint64
		migrateBucketStatus *gfsptask.GfSpBucketMigrationStatus
		allowMigrate        bool
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(err)
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, _ = NewRequestContext(r, g)
	if allowMigrate, migrateBucketStatus, err = g.checkMigrateBucketQuotaAuth(reqCtx, r.Header.Get(GnfdMigrateBucketMsgHeader)); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to check migrate bucket auth", "migrate_bucket", migrateBucketStatus, "error", err)
		return
	}
	if !allowMigrate {
		log.CtxErrorw(reqCtx.Context(), "no permission to migrate piece", "migrate_bucket", migrateBucketStatus)
		err = ErrNoPermission
		return
	}

	bucketID = migrateBucketStatus.GetBucketId()

	quota, err := g.baseApp.GfSpClient().GetLatestBucketReadQuota(
		reqCtx.Context(), bucketID)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket read quota", "bucket_id",
			bucketID, "error", err)
		return
	}
	err = g.baseApp.GfSpClient().NotifyPostMigrateBucket(reqCtx.Context(), bucketID)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "post migrate bucket error, the bucket may already notified", "bucket_id",
			bucketID, "error", err)
		return
	}

	if migrateBucketStatus.GetFinished() {
		bz, err := quota.Marshal()
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to marshal", "bucket_id",
				bucketID, "error", err)
			return
		}
		w.Header().Set(GnfdSignedApprovalMsgHeader, hex.EncodeToString(bz))
	} else {
		// get bucket quota and check TODO month check
		bucketSize, err = g.getBucketTotalSize(reqCtx.Context(), bucketID)
		if err != nil {
			return
		}
		migratedBytes := migrateBucketStatus.GetMigratedBytesSize()
		extraQuota := bucketSize - migratedBytes
		if migratedBytes > bucketSize {
			log.CtxErrorw(reqCtx.Context(), "failed to recoup extra quota to user", "error", err)
			return
		}
		quotaUpdateErr := g.baseApp.GfSpClient().RecoupQuota(reqCtx.Context(), migrateBucketStatus.GetBucketId(), extraQuota, quota.GetMonth())
		// no need to return the db error to user
		if quotaUpdateErr != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to recoup extra quota to user", "error", err)
		}
		log.CtxDebugw(reqCtx.Context(), "succeed to recoup extra quota to user", "extra_quote", extraQuota)
	}

	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)

	log.CtxInfow(reqCtx.Context(), "succeed to post bucket migrate", "bucket_id",
		bucketID, "quota", quota, "postMigrateBucket", migrateBucketStatus)
}

// sufficientQuotaForBucketMigrationHandler check if the source SP node has sufficient quota for bucket migration at approval phase
func (g *GateModular) sufficientQuotaForBucketMigrationHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                 error
		reqCtx              *RequestContext
		bucketID            uint64
		bucketSize          uint64
		migrateBucketStatus *gfsptask.GfSpBucketMigrationStatus
		allowMigrate        bool
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(err)
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			modelgateway.MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, _ = NewRequestContext(r, g)
	if allowMigrate, migrateBucketStatus, _, err = g.verifySignatureAndSP(reqCtx, r.Header.Get(GnfdMigrateBucketMsgHeader)); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to check migrate bucket auth", "migrate_bucket", migrateBucketStatus, "error", err)
		return
	}
	if !allowMigrate {
		log.CtxErrorw(reqCtx.Context(), "no permission to migrate piece", "migrate_bucket", migrateBucketStatus)
		err = ErrNoPermission
		return
	}

	bucketID = migrateBucketStatus.GetBucketId()

	quota, err := g.baseApp.GfSpClient().GetLatestBucketReadQuota(
		reqCtx.Context(), bucketID)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket read quota", "bucket_id",
			bucketID, "error", err)
		return
	}

	// get bucket quota and check
	bucketSize, err = g.getBucketTotalSize(reqCtx.Context(), bucketID)
	if err != nil {
		return
	}

	// empty bucket will approval
	if quota.FreeQuotaSize > bucketSize || bucketSize == 0 {
		w.Header().Set(GnfdSignedApprovalMsgHeader, "true")
	} else {
		w.Header().Set(GnfdSignedApprovalMsgHeader, "false")
	}

	w.Header().Set(ContentTypeHeader, ContentTypeXMLHeaderValue)

	log.CtxInfow(reqCtx.Context(), "succeed to check bucket migrate quota", "bucket_id",
		bucketID, "quota", quota, "bucketSize", bucketSize)
}
