package gater

import (
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	permissiontypes "github.com/bnb-chain/greenfield/x/permission/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield/types/common"
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
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
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
	sp, err := g.baseApp.GfSpClient().QuerySPByOperatorAddress(reqCtx.Context(), destSPAddr.String())
	if err != nil {
		log.Errorw("failed to query sp", "gvg_task", migrateGVG, "dest_sp_addr", destSPAddr.String(), "error", err)
		return false, err
	}
	effect, err := g.baseApp.GfSpClient().VerifyMigrateGVGPermission(reqCtx.Context(), migrateGVG.GetBucketID(), migrateGVG.GetSrcGvg().GetId(), sp.GetId())
	if effect == nil || err != nil {
		log.Errorw("failed to verify migrate gvg permission", "gvg_task", migrateGVG, "dest_sp", sp, "effect", effect, "error", err)
		return false, err
	}
	if *effect == permissiontypes.EFFECT_ALLOW {
		return true, nil
	}
	return false, nil
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
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
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
	pieceTask.InitDownloadPieceTask(chainObjectInfo, bucketInfo, params, coretask.DefaultSmallerPriority, migratePiece.GetIsBucketMigrate(), bucketInfo.Owner,
		uint64(pieceSize), pieceKey, 0, uint64(pieceSize),
		g.baseApp.TaskTimeout(pieceTask, objectInfo.GetPayloadSize()), g.baseApp.TaskMaxRetry(pieceTask))
	pieceData, err = g.baseApp.GfSpClient().GetPiece(reqCtx.Context(), pieceTask)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to download segment piece", "piece_key", pieceKey, "error", err)
		return
	}

	_, _ = w.Write(pieceData)
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
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
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
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
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
