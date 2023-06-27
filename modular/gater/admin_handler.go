package gater

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/modular/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

// getApprovalHandler handles the get create bucket/object approval request.
// Before create bucket/object to the greenfield, the user should the primary
// SP whether willing serve for the user to manage the bucket/object.
// SP checks the user's account if it has the permission to operate, and send
// the request to approver that running the SP approval's strategy.
func (g *GateModular) getApprovalHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                  error
		reqCtx               *RequestContext
		approvalMsg          []byte
		createBucketApproval = storagetypes.MsgCreateBucket{}
		createObjectApproval = storagetypes.MsgCreateObject{}
		authenticated        bool
		approved             bool
	)
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHttpCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHttpCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}

	approvalType := reqCtx.vars["action"]
	approvalMsg, err = hex.DecodeString(r.Header.Get(GnfdUnsignedApprovalMsgHeader))
	if err != nil {
		log.Errorw("failed to parse approval header", "approval_type", approvalType,
			"approval", r.Header.Get(GnfdUnsignedApprovalMsgHeader))
		err = ErrDecodeMsg
		return
	}

	switch approvalType {
	case createBucketApprovalAction:
		if err = storagetypes.ModuleCdc.UnmarshalJSON(approvalMsg, &createBucketApproval); err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to unmarshal approval", "approval",
				r.Header.Get(GnfdUnsignedApprovalMsgHeader), "error", err)
			err = ErrDecodeMsg
			return
		}
		if err = createBucketApproval.ValidateBasic(); err != nil {
			log.Errorw("failed to basic check bucket approval msg", "bucket_approval_msg",
				createBucketApproval, "error", err)
			err = ErrValidateMsg
			return
		}
		if reqCtx.NeedVerifyAuthentication() {
			startVerifyAuthentication := time.Now()
			authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(
				reqCtx.Context(), coremodule.AuthOpAskCreateBucketApproval,
				reqCtx.Account(), createBucketApproval.GetBucketName(), "")
			metrics.PerfGetApprovalTimeHistogram.WithLabelValues("verify_authorize").Observe(time.Since(startVerifyAuthentication).Seconds())
			if err != nil {
				log.CtxErrorw(reqCtx.Context(), "failed to verify authentication", "error", err)
				return
			}
			if !authenticated {
				log.CtxErrorw(reqCtx.Context(), "no permission to operate")
				err = ErrNoPermission
				return
			}
		}
		task := &gfsptask.GfSpCreateBucketApprovalTask{}
		task.InitApprovalCreateBucketTask(&createBucketApproval, g.baseApp.TaskPriority(task))
		var approvalTask coretask.ApprovalCreateBucketTask
		startAskCreateBucketApproval := time.Now()
		approved, approvalTask, err = g.baseApp.GfSpClient().AskCreateBucketApproval(reqCtx.Context(), task)
		metrics.PerfGetApprovalTimeHistogram.WithLabelValues("ask_create_bucket_approval").Observe(time.Since(startAskCreateBucketApproval).Seconds())
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to ask create bucket approval", "error", err)
			return
		}
		if !approved {
			log.CtxErrorw(reqCtx.Context(), "refuse the ask create bucket approval")
			err = ErrRefuseApproval
			return
		}
		bz := storagetypes.ModuleCdc.MustMarshalJSON(approvalTask.GetCreateBucketInfo())
		w.Header().Set(GnfdSignedApprovalMsgHeader, hex.EncodeToString(sdktypes.MustSortJSON(bz)))
	case createObjectApprovalAction:
		if err = storagetypes.ModuleCdc.UnmarshalJSON(approvalMsg, &createObjectApproval); err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to unmarshal approval", "approval",
				r.Header.Get(GnfdUnsignedApprovalMsgHeader), "error", err)
			err = ErrDecodeMsg
			return
		}
		if err = createObjectApproval.ValidateBasic(); err != nil {
			log.Errorw("failed to basic check approval msg", "object_approval_msg",
				createObjectApproval, "error", err)
			err = ErrValidateMsg
			return
		}
		if reqCtx.NeedVerifyAuthentication() {
			startVerifyAuthentication := time.Now()
			authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(
				reqCtx.Context(), coremodule.AuthOpAskCreateObjectApproval,
				reqCtx.Account(), createObjectApproval.GetBucketName(),
				createObjectApproval.GetObjectName())
			metrics.PerfGetApprovalTimeHistogram.WithLabelValues("verify_authorize").Observe(time.Since(startVerifyAuthentication).Seconds())
			if err != nil {
				log.CtxErrorw(reqCtx.Context(), "failed to verify authentication", "error", err)
				return
			}
			if !authenticated {
				log.CtxErrorw(reqCtx.Context(), "no permission to operate")
				err = ErrNoPermission
				return
			}
		}
		task := &gfsptask.GfSpCreateObjectApprovalTask{}
		task.InitApprovalCreateObjectTask(&createObjectApproval, g.baseApp.TaskPriority(task))
		var approvedTask coretask.ApprovalCreateObjectTask
		startAskCreateObjectApproval := time.Now()
		approved, approvedTask, err = g.baseApp.GfSpClient().AskCreateObjectApproval(r.Context(), task)
		metrics.PerfGetApprovalTimeHistogram.WithLabelValues("ask_create_object_approval").Observe(time.Since(startAskCreateObjectApproval).Seconds())
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to ask object approval", "error", err)
			return
		}
		if !approved {
			log.CtxErrorw(reqCtx.Context(), "refuse the ask create object approval")
			err = ErrRefuseApproval
			return
		}
		bz := storagetypes.ModuleCdc.MustMarshalJSON(approvedTask.GetCreateObjectInfo())
		w.Header().Set(GnfdSignedApprovalMsgHeader, hex.EncodeToString(sdktypes.MustSortJSON(bz)))
	default:
		err = ErrUnsupportedRequestType
		return
	}
	log.CtxDebugw(reqCtx.Context(), "succeed to ask approval")
}

// getChallengeInfoHandler handles get challenge piece info request. Current only greenfield
// validator can challenge piece is store correctly. The challenge piece info includes:
// the challenged piece data, all piece hashes and the integrity hash. The challenger
// can verify the info whether are correct by comparing with the greenfield info.
func (g *GateModular) getChallengeInfoHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		reqCtx        *RequestContext
		authenticated bool
		integrity     []byte
		checksums     [][]byte
		data          []byte
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHttpCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHttpCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
		metrics.PerfChallengeTimeHistogram.WithLabelValues("challenge_total_time").Observe(time.Since(startTime).Seconds())
	}()

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}
	objectID, err := util.StringToUint64(reqCtx.request.Header.Get(GnfdObjectIDHeader))
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse object id", "object_id",
			reqCtx.request.Header.Get(GnfdObjectIDHeader))
		err = ErrInvalidHeader
		return
	}

	getObjectTime := time.Now()
	objectInfo, err := g.baseApp.Consensus().QueryObjectInfoByID(reqCtx.Context(),
		reqCtx.request.Header.Get(GnfdObjectIDHeader))
	metrics.PerfChallengeTimeHistogram.WithLabelValues("challenge_get_object_time").Observe(time.Since(getObjectTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		if strings.Contains(err.Error(), "No such object") {
			err = ErrNoSuchObject
		} else {
			err = ErrConsensus
		}
		return
	}
	if reqCtx.NeedVerifyAuthentication() {
		authTime := time.Now()
		authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(),
			coremodule.AuthOpTypeGetChallengePieceInfo, reqCtx.Account(), objectInfo.GetBucketName(),
			objectInfo.GetObjectName())
		metrics.PerfChallengeTimeHistogram.WithLabelValues("challenge_auth_time").Observe(time.Since(authTime).Seconds())
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to verify authentication", "error", err)
			return
		}
		if !authenticated {
			log.CtxErrorw(reqCtx.Context(), "failed to get challenge info due to no permission")
			err = ErrNoPermission
			return
		}
	}

	getBucketTime := time.Now()
	bucketInfo, err := g.baseApp.Consensus().QueryBucketInfo(reqCtx.Context(), objectInfo.GetBucketName())
	metrics.PerfChallengeTimeHistogram.WithLabelValues("challenge_get_bucket_time").Observe(time.Since(getBucketTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket info from consensus", "error", err)
		err = ErrConsensus
		return
	}
	redundancyIdx, err := util.StringToInt32(reqCtx.request.Header.Get(GnfdRedundancyIndexHeader))
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse redundancy index", "redundancy_idx",
			reqCtx.request.Header.Get(GnfdRedundancyIndexHeader))
		err = ErrInvalidHeader
		return
	}
	segmentIdx, err := util.StringToUint32(reqCtx.request.Header.Get(GnfdPieceIndexHeader))
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse segment index", "segment_idx",
			reqCtx.request.Header.Get(GnfdPieceIndexHeader))
		err = ErrInvalidHeader
		return
	}
	getParamTime := time.Now()
	params, err := g.baseApp.Consensus().QueryStorageParamsByTimestamp(
		reqCtx.Context(), objectInfo.GetCreateAt())
	metrics.PerfChallengeTimeHistogram.WithLabelValues("challenge_get_param_time").Observe(time.Since(getParamTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params", "error", err)
		return
	}
	var pieceSize uint64
	if redundancyIdx < 0 {
		pieceSize = uint64(g.baseApp.PieceOp().SegmentPieceSize(objectInfo.GetPayloadSize(),
			segmentIdx, params.VersionedParams.GetMaxSegmentSize()))
	} else {
		pieceSize = uint64(g.baseApp.PieceOp().ECPieceSize(objectInfo.GetPayloadSize(),
			segmentIdx, params.VersionedParams.GetMaxSegmentSize(),
			params.VersionedParams.GetRedundantDataChunkNum()))
	}
	task := &gfsptask.GfSpChallengePieceTask{}
	task.InitChallengePieceTask(objectInfo, bucketInfo, params, g.baseApp.TaskPriority(task), reqCtx.Account(),
		redundancyIdx, segmentIdx, g.baseApp.TaskTimeout(task, pieceSize), g.baseApp.TaskMaxRetry(task))
	ctx := log.WithValue(reqCtx.Context(), log.CtxKeyTask, task.Key().String())
	getChallengeInfoTime := time.Now()
	integrity, checksums, data, err = g.baseApp.GfSpClient().GetChallengeInfo(reqCtx.Context(), task)
	metrics.PerfChallengeTimeHistogram.WithLabelValues("challenge_get_info_time").Observe(time.Since(getChallengeInfoTime).Seconds())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get challenge info", "error", err)
		return
	}
	w.Header().Set(GnfdObjectIDHeader, util.Uint64ToString(objectID))
	w.Header().Set(GnfdIntegrityHashHeader, hex.EncodeToString(integrity))
	w.Header().Set(GnfdPieceHashHeader, util.BytesSliceToString(checksums))
	w.Write(data)
	log.CtxDebugw(reqCtx.Context(), "succeed to get challenge info")
}

// replicateHandler handles the replicate piece from primary SP request. The Primary
// replicates the piece data one by one, and will ask the integrity hash and the
// signature to seal object on greenfield.
func (g *GateModular) replicateHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		reqCtx     *RequestContext
		receiveMsg []byte
		data       []byte
		integrity  []byte
		signature  []byte
	)
	receivePieceStartTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHttpCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHttpCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
		metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_total_time").Observe(time.Since(receivePieceStartTime).Seconds())
	}()
	// ignore the error, because the replicate request only between SPs, the request
	// verification is by signature of the ReceivePieceTask
	reqCtx, _ = NewRequestContext(r, g)
	decodeTime := time.Now()
	receiveMsg, err = hex.DecodeString(r.Header.Get(GnfdReceiveMsgHeader))
	metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_decode_task_time").Observe(time.Since(decodeTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse receive header",
			"receive", r.Header.Get(GnfdReceiveMsgHeader))
		err = ErrDecodeMsg
		return
	}
	receiveTask := gfsptask.GfSpReceivePieceTask{}
	err = json.Unmarshal(receiveMsg, &receiveTask)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal receive header",
			"receive", r.Header.Get(GnfdReceiveMsgHeader))
		err = ErrDecodeMsg
		return
	}
	if receiveTask.GetObjectInfo() == nil ||
		int(receiveTask.GetReplicateIdx()) >=
			len(receiveTask.GetObjectInfo().GetChecksums()) {
		log.CtxErrorw(reqCtx.Context(), "receive task params error")
		err = ErrInvalidHeader
		return
	}
	readDataTime := time.Now()
	data, err = io.ReadAll(r.Body)
	metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_read_piece_time").Observe(time.Since(readDataTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to read replicate piece data", "error", err)
		err = ErrExceptionStream
		return
	}
	if receiveTask.GetPieceIdx() >= 0 {
		handlePieceTime := time.Now()
		err = g.baseApp.GfSpClient().ReplicatePiece(reqCtx.Context(), &receiveTask, data)
		metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_receive_data_time").Observe(time.Since(handlePieceTime).Seconds())
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to receive piece", "error", err)
			return
		}
	} else {
		donePieceTime := time.Now()
		integrity, signature, err = g.baseApp.GfSpClient().DoneReplicatePiece(reqCtx.Context(), &receiveTask)
		metrics.PerfReceivePieceTimeHistogram.WithLabelValues("receive_piece_done_time").Observe(time.Since(donePieceTime).Seconds())
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to done receive piece", "error", err)
			return
		}
		w.Header().Set(GnfdIntegrityHashHeader, hex.EncodeToString(integrity))
		w.Header().Set(GnfdIntegrityHashSignatureHeader, hex.EncodeToString(signature))
	}
	log.CtxDebugw(reqCtx.Context(), "succeed to replicate piece")
}

const MigrateMinEcIndex = -1

// transferObjectHandler handles the transfer data request between SPs which is used in SP exiting case.
// First, gateway should verify Authorization header to ensure the requests are from correct SPs.
// Second, retrieve and get data from downloader module including: PrimarySP and SecondarySP pieces
// Third, transfer data to client which is a selected SP.
func (g *GateModular) transferObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		reqCtx     *RequestContext
		migrateMsg []byte
		pieceData  []byte
	)
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(err)
			reqCtx.SetHttpCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		} else {
			reqCtx.SetHttpCode(http.StatusOK)
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	reqCtx, _ = NewRequestContext(r, g)

	migrateHeader := r.Header.Get(GnfdMigratePieceMsgHeader)
	migrateMsg, err = hex.DecodeString(migrateHeader)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse migrate piece header", "header", migrateHeader)
		err = ErrDecodeMsg
		return
	}

	migratePieceTask := gfsptask.GfSpMigratePieceTask{}
	err = json.Unmarshal(migrateMsg, &migratePieceTask)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal migrate piece task msg", "header", migrateHeader)
		err = ErrDecodeMsg
		return
	}

	taskSig := migratePieceTask.GetSignature()
	sigAddr, pk, err := RecoverAddr(crypto.Keccak256(migratePieceTask.GetSignBytes()), taskSig)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get migrate piece task address", "error", err)
		err = ErrSignature
		return
	}

	if !secp256k1.VerifySignature(pk.Bytes(), crypto.Keccak256(migratePieceTask.GetSignBytes()), taskSig[:len(taskSig)-1]) {
		log.CtxError(reqCtx.Context(), "failed to verify migrate piece task signature")
		err = ErrSignature
		return
	}

	objectInfo := migratePieceTask.GetObjectInfo()
	if objectInfo == nil {
		log.CtxError(reqCtx.Context(), "failed to get migrate piece object info")
		err = ErrInvalidHeader
		return
	}

	isPrimaryHeader := r.Header.Get(GnfdIsPrimaryHeader)
	if isPrimaryHeader == "" {
		log.CtxError(reqCtx.Context(), "migrate piece is primary header is empty")
	}
	isPrimary, err := strconv.ParseBool(isPrimaryHeader)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse migrate piece is primary header", "header",
			isPrimaryHeader, "error", err)
	}

	chainObjectInfo, bucketInfo, params, err := getObjectChainMeta(reqCtx, g.baseApp, objectInfo.GetObjectName(), objectInfo.GetBucketName())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object on chain meta", "error", err)
		err = ErrInvalidHeader
		return
	}
	redundancyIdx := migratePieceTask.EcIdx
	maxRedundancyIdx := int32(params.GetRedundantDataChunkNum()+params.GetRedundantParityChunkNum()) - 1
	if redundancyIdx < int32(MigrateMinEcIndex) || redundancyIdx > maxRedundancyIdx {
		// TODO: customize an error to return to sp
		return
	}

	// if isPrimary is equal to true, we should migrate pieces from primary SP
	if isPrimary {
		pieceData, err = g.getSegmentPiece(reqCtx.Context(), chainObjectInfo, bucketInfo, params, migratePieceTask, sigAddr)
		if err != nil {
			return
		}
	} else { // in this case, we should migrate pieces from secondary SP
		pieceData, err = g.getECPiece(reqCtx.Context(), chainObjectInfo, bucketInfo, params, migratePieceTask, sigAddr)
		if err != nil {
			return
		}
	}

	w.Write(pieceData)
	log.CtxDebug(reqCtx.Context(), "succeed to migrate one piece")
}

func (g *GateModular) getSegmentPiece(ctx context.Context, objectInfo *storagetypes.ObjectInfo, bucketInfo *storagetypes.BucketInfo,
	params *storagetypes.Params, migratePieceTask gfsptask.GfSpMigratePieceTask, sigAddr sdktypes.AccAddress) ([]byte, error) {
	ECPieceSize := g.baseApp.PieceOp().ECPieceSize(objectInfo.PayloadSize, migratePieceTask.GetSegmentIdx(),
		params.GetMaxSegmentSize(), params.GetRedundantDataChunkNum())
	pieceTask := &gfsptask.GfSpDownloadPieceTask{}
	segmentPieceKey := g.baseApp.PieceOp().SegmentPieceKey(migratePieceTask.GetObjectInfo().Id.Uint64(), migratePieceTask.GetSegmentIdx())
	segmentPieceSize := g.baseApp.PieceOp().SegmentPieceSize(objectInfo.PayloadSize, migratePieceTask.GetSegmentIdx(), params.GetMaxSegmentSize())
	// TODO: refine how to get userAddress, mock temporarily
	mockAddress := "mock"
	pieceTask.InitDownloadPieceTask(objectInfo, bucketInfo, params, g.baseApp.TaskPriority(pieceTask),
		false, mockAddress, uint64(ECPieceSize), segmentPieceKey, 0, uint64(segmentPieceSize),
		g.baseApp.TaskTimeout(pieceTask, uint64(pieceTask.GetSize())), g.baseApp.TaskMaxRetry(pieceTask))
	segmentPieceData, err := g.baseApp.GfSpClient().GetPiece(ctx, pieceTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to download segment piece data", "error", err)
		return nil, downloader.ErrPieceStore
	}
	return segmentPieceData, nil
}

func (g *GateModular) getECPiece(ctx context.Context, objectInfo *storagetypes.ObjectInfo, bucketInfo *storagetypes.BucketInfo,
	params *storagetypes.Params, migratePieceTask gfsptask.GfSpMigratePieceTask, sigAddr sdktypes.AccAddress) ([]byte, error) {
	ecPieceSize := g.baseApp.PieceOp().ECPieceSize(objectInfo.PayloadSize, migratePieceTask.GetSegmentIdx(),
		params.GetMaxSegmentSize(), params.GetRedundantDataChunkNum())
	// TODO: should refine how to get ec index
	ecIndex := 0
	ecPieceKey := g.baseApp.PieceOp().ECPieceKey(migratePieceTask.GetObjectInfo().Id.Uint64(),
		migratePieceTask.GetSegmentIdx(), uint32(ecIndex))
	pieceTask := &gfsptask.GfSpDownloadPieceTask{}
	// TODO: refine how to get userAddress, mock temporarily
	mockAddress := "mock"
	pieceTask.InitDownloadPieceTask(objectInfo, bucketInfo, params, g.baseApp.TaskPriority(pieceTask), false, mockAddress,
		uint64(ecPieceSize), ecPieceKey, 0, uint64(ecPieceSize), g.baseApp.TaskTimeout(pieceTask, uint64(pieceTask.GetSize())),
		g.baseApp.TaskMaxRetry(pieceTask))
	ecPieceData, err := g.baseApp.GfSpClient().GetPiece(ctx, pieceTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to download ec piece data", "error", err)
		return nil, downloader.ErrPieceStore
	}
	return ecPieceData, nil
}

// TODO: this function should be deleted in the future
func getObjectChainMeta(reqCtx *RequestContext, baseApp *gfspapp.GfSpBaseApp, objectName, bucketName string) (*storagetypes.ObjectInfo, *storagetypes.BucketInfo, *storagetypes.Params, error) {
	objectInfo, err := baseApp.Consensus().QueryObjectInfo(reqCtx.Context(), bucketName, objectName)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		return nil, nil, nil, ErrConsensus
	}

	bucketInfo, err := baseApp.Consensus().QueryBucketInfo(reqCtx.Context(), objectInfo.GetBucketName())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket info from consensus", "error", err)
		return nil, nil, nil, ErrConsensus
	}

	params, err := baseApp.Consensus().QueryStorageParamsByTimestamp(
		reqCtx.Context(), objectInfo.GetCreateAt())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params", "error", err)
		return nil, nil, nil, ErrConsensus
	}

	return objectInfo, bucketInfo, params, nil
}
