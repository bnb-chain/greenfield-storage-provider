package gater

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	commonhash "github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-common/go/redundancy"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/modular/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/modular/executor"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

const MaxSpRequestExpiryAgeInSec int32 = 1000

// getApprovalHandler handles the get create bucket/object approval request.
// Before create bucket/object to the greenfield, the user should the primary
// SP whether willing serve for the user to manage the bucket/object.
// SP checks the user's account if it has the permission to operate, and send
// the request to approver that running the SP approval's strategy.
func (g *GateModular) getApprovalHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                   error
		reqCtx                *RequestContext
		approvalMsg           []byte
		createBucketApproval  = storagetypes.MsgCreateBucket{}
		migrateBucketApproval = storagetypes.MsgMigrateBucket{}
		createObjectApproval  = storagetypes.MsgCreateObject{}
		fingerprint           []byte
		spInfo                *sptypes.StorageProvider
		authenticated         bool
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
			metrics.ReqCounter.WithLabelValues(GatewayFailureGetApproval).Inc()
			metrics.ReqTime.WithLabelValues(GatewayFailureGetApproval).Observe(time.Since(startTime).Seconds())
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(startTime).Seconds())
			metrics.ReqCounter.WithLabelValues(GatewaySuccessGetApproval).Inc()
			metrics.ReqTime.WithLabelValues(GatewaySuccessGetApproval).Observe(time.Since(startTime).Seconds())
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
	fingerprint = commonhash.GenerateChecksum(approvalMsg)

	switch approvalType {
	case createBucketApprovalAction:
		// check sp status firstly
		spInfo, err = g.baseApp.Consensus().QuerySP(reqCtx.Context(), g.baseApp.OperatorAddress())
		if err != nil {
			log.Errorw("failed to query sp by operator address", "operator_address", g.baseApp.OperatorAddress(),
				"error", err)
			return
		}
		spStatus := spInfo.GetStatus()
		if spStatus != sptypes.STATUS_IN_SERVICE && !fromSpMaintenanceAcct(spStatus, spInfo.MaintenanceAddress, reqCtx.account) {
			log.Errorw("sp is not in service status", "operator_address", g.baseApp.OperatorAddress(),
				"sp_status", spStatus, "sp_id", spInfo.GetId(), "endpoint", spInfo.GetEndpoint(), "request_acct", reqCtx.account)
			err = ErrSPUnavailable
			return
		}
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
		startVerifyAuthentication := time.Now()
		authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(reqCtx.Context(), coremodule.AuthOpAskCreateBucketApproval,
			reqCtx.Account(), createBucketApproval.GetBucketName(), "")
		metrics.PerfApprovalTime.WithLabelValues("gateway_create_bucket_auth_cost").Observe(time.Since(startVerifyAuthentication).Seconds())
		metrics.PerfApprovalTime.WithLabelValues("gateway_create_bucket_auth_end").Observe(time.Since(startTime).Seconds())
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to verify authentication", "error", err)
			return
		}
		if !authenticated {
			log.CtxError(reqCtx.Context(), "no permission to operate")
			err = ErrNoPermission
			return
		}
		task := &gfsptask.GfSpCreateBucketApprovalTask{}
		task.InitApprovalCreateBucketTask(reqCtx.Account(), &createBucketApproval, fingerprint, g.baseApp.TaskPriority(task))
		var approvalTask coretask.ApprovalCreateBucketTask
		startAskCreateBucketApproval := time.Now()
		authenticated, approvalTask, err = g.baseApp.GfSpClient().AskCreateBucketApproval(reqCtx.Context(), task)
		metrics.PerfApprovalTime.WithLabelValues("gateway_create_bucket_ask_approval_cost").Observe(time.Since(startAskCreateBucketApproval).Seconds())
		metrics.PerfApprovalTime.WithLabelValues("gateway_create_bucket_ask_approval_end").Observe(time.Since(startTime).Seconds())
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to ask create bucket approval", "error", err)
			return
		}
		if !authenticated {
			log.CtxErrorw(reqCtx.Context(), "refuse the ask create bucket approval")
			err = ErrRefuseApproval
			return
		}
		bz := storagetypes.ModuleCdc.MustMarshalJSON(approvalTask.GetCreateBucketInfo())
		w.Header().Set(GnfdSignedApprovalMsgHeader, hex.EncodeToString(sdktypes.MustSortJSON(bz)))
	case migrateBucketApprovalAction:
		if err = storagetypes.ModuleCdc.UnmarshalJSON(approvalMsg, &migrateBucketApproval); err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to unmarshal approval", "approval",
				r.Header.Get(GnfdUnsignedApprovalMsgHeader), "error", err)
			err = ErrDecodeMsg
			return
		}
		if err = migrateBucketApproval.ValidateBasic(); err != nil {
			log.Errorw("failed to basic check migrate bucket approval msg", "bucket_approval_msg",
				migrateBucketApproval, "error", err)
			err = ErrValidateMsg
			return
		}
		authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(
			reqCtx.Context(), coremodule.AuthOpAskMigrateBucketApproval,
			reqCtx.Account(), migrateBucketApproval.GetBucketName(), "")
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to verify authentication", "error", err)
			return
		}
		if !authenticated {
			log.CtxErrorw(reqCtx.Context(), "no permission to operate")
			err = ErrNoPermission
			return
		}
		task := &gfsptask.GfSpMigrateBucketApprovalTask{}
		task.InitApprovalMigrateBucketTask(&migrateBucketApproval, g.baseApp.TaskPriority(task))
		var approvalTask coretask.ApprovalMigrateBucketTask
		authenticated, approvalTask, err = g.baseApp.GfSpClient().AskMigrateBucketApproval(reqCtx.Context(), task)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to ask migrate bucket approval", "error", err)
			return
		}
		if !authenticated {
			log.CtxErrorw(reqCtx.Context(), "refuse the ask migrate bucket approval")
			err = ErrRefuseApproval
			return
		}
		bz := storagetypes.ModuleCdc.MustMarshalJSON(approvalTask.GetMigrateBucketInfo())
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
		if err = g.checkSPAndBucketStatus(reqCtx.Context(), createObjectApproval.GetBucketName(), createObjectApproval.Creator); err != nil {
			log.Errorw("create object approval failed to check sp and bucket status", "error", err)
			return
		}

		startVerifyAuthentication := time.Now()
		authenticated, err = g.baseApp.GfSpClient().VerifyAuthentication(
			reqCtx.Context(), coremodule.AuthOpAskCreateObjectApproval,
			reqCtx.Account(), createObjectApproval.GetBucketName(),
			createObjectApproval.GetObjectName())
		metrics.PerfApprovalTime.WithLabelValues("gateway_create_object_auth_cost").Observe(time.Since(startVerifyAuthentication).Seconds())
		metrics.PerfApprovalTime.WithLabelValues("gateway_create_object_auth_end").Observe(time.Since(startTime).Seconds())
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to verify authentication", "error", err)
			return
		}
		if !authenticated {
			log.CtxErrorw(reqCtx.Context(), "no permission to operate")
			err = ErrNoPermission
			return
		}
		task := &gfsptask.GfSpCreateObjectApprovalTask{}
		task.InitApprovalCreateObjectTask(reqCtx.Account(), &createObjectApproval, fingerprint, g.baseApp.TaskPriority(task))
		var approvedTask coretask.ApprovalCreateObjectTask
		startAskCreateObjectApproval := time.Now()
		authenticated, approvedTask, err = g.baseApp.GfSpClient().AskCreateObjectApproval(r.Context(), task)
		metrics.PerfApprovalTime.WithLabelValues("gateway_create_object_ask_approval_cost").Observe(time.Since(startAskCreateObjectApproval).Seconds())
		metrics.PerfApprovalTime.WithLabelValues("gateway_create_object_ask_approval_end").Observe(time.Since(startTime).Seconds())
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to ask object approval", "error", err)
			return
		}
		if !authenticated {
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
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
			metrics.ReqCounter.WithLabelValues(GatewayFailureGetChallengeInfo).Inc()
			metrics.ReqTime.WithLabelValues(GatewayFailureGetChallengeInfo).Observe(time.Since(startTime).Seconds())
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
			metrics.ReqCounter.WithLabelValues(GatewaySuccessGetChallengeInfo).Inc()
			metrics.ReqTime.WithLabelValues(GatewaySuccessGetChallengeInfo).Observe(time.Since(startTime).Seconds())
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
			err = ErrConsensusWithDetail("failed to get object info from consensus, error: " + err.Error())
		}
		return
	}
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

	getBucketTime := time.Now()
	bucketInfo, err := g.baseApp.Consensus().QueryBucketInfo(reqCtx.Context(), objectInfo.GetBucketName())
	metrics.PerfChallengeTimeHistogram.WithLabelValues("challenge_get_bucket_time").Observe(time.Since(getBucketTime).Seconds())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket info from consensus", "error", err)
		err = ErrConsensusWithDetail("failed to get bucket info from consensus, error: " + err.Error())
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
	params, err := g.baseApp.Consensus().QueryStorageParamsByTimestamp(reqCtx.Context(), objectInfo.GetCreateAt())
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
	_, _ = w.Write(data)
	metrics.ReqPieceSize.WithLabelValues(GatewayChallengePieceSize).Observe(float64(len(data)))
	log.CtxDebug(reqCtx.Context(), "succeed to get challenge info")
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
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(receivePieceStartTime).Seconds())
			metrics.ReqCounter.WithLabelValues(GatewayFailureReplicatePiece).Inc()
			metrics.ReqTime.WithLabelValues(GatewayFailureReplicatePiece).Observe(time.Since(receivePieceStartTime).Seconds())
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(receivePieceStartTime).Seconds())
			metrics.ReqCounter.WithLabelValues(GatewaySuccessReplicatePiece).Inc()
			metrics.ReqTime.WithLabelValues(GatewaySuccessReplicatePiece).Observe(time.Since(receivePieceStartTime).Seconds())
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
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

	// verify receive task signature
	signatureAddr, err := reqCtx.verifyTaskSignature(receiveTask.GetSignBytes(), receiveTask.GetSignature())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to verify receive task", "error", err, "task", receiveTask)
		err = ErrSignature
		return
	}

	// check if the update time of the task has expired
	taskUpdateTime := receiveTask.GetUpdateTime()
	timeDifference := time.Duration(time.Now().Unix()-taskUpdateTime) * time.Second
	if int32(timeDifference.Seconds()) > MaxSpRequestExpiryAgeInSec {
		log.CtxErrorw(reqCtx.Context(), "the update time of receive task has exceeded the expiration time", "object", receiveTask.ObjectInfo.ObjectName)
		err = ErrTaskMsgExpired
		return
	}

	err = g.checkReplicatePermission(reqCtx.Context(), receiveTask, signatureAddr.String())
	if err != nil {
		return
	}

	if receiveTask.GetObjectInfo() == nil || int(receiveTask.GetRedundancyIdx()) >= len(receiveTask.GetObjectInfo().GetChecksums()) {
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
	if !receiveTask.GetFinished() {
		metrics.ReqPieceSize.WithLabelValues(GatewayReplicatePieceSize).Observe(float64(len(data)))
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
	log.CtxDebug(reqCtx.Context(), "succeed to replicate piece")
}

func (g *GateModular) checkReplicatePermission(ctx context.Context, receiveTask gfsptask.GfSpReceivePieceTask, signatureAddr string) error {
	// check if the request account is the primary SP of the object of the receiving task
	bucketInfo, err := g.baseApp.Consensus().QueryBucketInfo(ctx, receiveTask.GetObjectInfo().BucketName)
	if err != nil {
		err = ErrConsensusWithDetail("QueryBucketInfo error: " + err.Error())
		return err
	}

	gvg, err := g.baseApp.GfSpClient().GetGlobalVirtualGroup(ctx, bucketInfo.Id.Uint64(), receiveTask.GetGlobalVirtualGroupID())
	if err != nil {
		return ErrConsensusWithDetail("QueryGVGInfo error: " + err.Error())
	}

	// judge if sender is primary sp
	primarySp, err := g.baseApp.Consensus().QuerySPByID(ctx, gvg.PrimarySpId)
	if err != nil {
		return ErrConsensusWithDetail("QuerySPInfo error: " + err.Error())
	}

	if primarySp.GetOperatorAccAddress().String() != signatureAddr {
		log.CtxErrorw(ctx, "primary sp mismatch", "expect",
			primarySp.GetOperatorAccAddress().String(), "current", signatureAddr)
		return ErrPrimaryMismatch
	}

	// judge if myself is secondary sp
	spID, err := g.getSPID()
	if err != nil {
		return ErrConsensusWithDetail("getSPID error: " + err.Error())
	}

	expectSecondarySPID := gvg.GetSecondarySpIds()[int(receiveTask.GetRedundancyIdx())]
	if expectSecondarySPID != spID {
		log.CtxErrorw(ctx, "secondary sp mismatch", "expect", expectSecondarySPID, "current", spID)
		return ErrSecondaryMismatch
	}

	return nil
}

// getRecoverDataHandler handles the query for recovery request from secondary SP or primary SP.
// if it is used to recovery primary SP and the handler SP is the corresponding secondary,
// it returns the EC piece data stored in the secondary SP for the requested object.
// if it is used to recovery secondary SP and the handler is the corresponding primary SP,
// it directly returns the EC piece data of the secondary SP.
func (g *GateModular) getRecoverDataHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		reqCtx      *RequestContext
		recoveryMsg []byte
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			reqCtx.SetHTTPCode(int(gfsperrors.MakeGfSpError(err).GetHttpStatusCode()))
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
			metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
		} else {
			reqCtx.SetHTTPCode(http.StatusOK)
			metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
			metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(startTime).Seconds())
		}
		log.CtxDebugw(reqCtx.Context(), reqCtx.String())
	}()

	// ignore the error, because the recovery request only between SPs, the request
	// verification is by signature of the RecoveryTask
	reqCtx, _ = NewRequestContext(r, g)

	recoveryMsg, err = hex.DecodeString(r.Header.Get(GnfdRecoveryMsgHeader))
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse recovery header",
			"header", r.Header.Get(GnfdRecoveryMsgHeader))
		err = ErrDecodeMsg
		return
	}

	recoveryTask := gfsptask.GfSpRecoverPieceTask{}
	err = json.Unmarshal(recoveryMsg, &recoveryTask)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal recovery msg header",
			"header", r.Header.Get(GnfdRecoveryMsgHeader))
		err = ErrDecodeMsg
		return
	}

	// verify recovery task signature
	signatureAddr, err := reqCtx.verifyTaskSignature(recoveryTask.GetSignBytes(), recoveryTask.GetSignature())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to verify recovery task", "error", err, "task", recoveryTask)
		err = ErrSignature
		return
	}

	// check if the update time of the task has expired
	taskUpdateTime := recoveryTask.GetUpdateTime()
	timeDifference := time.Duration(time.Now().Unix()-taskUpdateTime) * time.Second
	if int32(timeDifference.Seconds()) > MaxSpRequestExpiryAgeInSec {
		log.CtxErrorw(reqCtx.Context(), "the update time of recovery task has exceeded the expiration time", "object", recoveryTask.ObjectInfo.ObjectName)
		err = ErrTaskMsgExpired
		return
	}

	// handle request from primary SP to recovery segment {
	objectInfo := recoveryTask.GetObjectInfo()
	if objectInfo == nil {
		log.CtxErrorw(reqCtx.Context(), "recovery task params error")
		err = ErrInvalidHeader
		return
	}

	if objectInfo.RedundancyType != storagetypes.REDUNDANCY_EC_TYPE {
		log.CtxErrorw(reqCtx.Context(), "recovery redundancy type only support EC")
		err = ErrRecoveryRedundancyType
		return
	}

	chainObjectInfo, bucketInfo, params, err := g.getObjectChainMeta(reqCtx.Context(), objectInfo.ObjectName, objectInfo.BucketName)
	if err != nil {
		err = ErrInvalidHeader
		return
	}

	redundancyIdx := recoveryTask.EcIdx
	maxRedundancyIndex := int32(params.GetRedundantParityChunkNum()+params.GetRedundantDataChunkNum()) - 1
	if redundancyIdx < int32(RecoveryMinEcIndex) || redundancyIdx > maxRedundancyIndex {
		err = executor.ErrRecoveryPieceIndex
		return
	}

	var pieceData []byte

	spID, err := g.getSPID()
	if err != nil {
		err = ErrConsensusWithDetail("getSPID error: " + err.Error())
		return
	}
	bucketSPID, err := util.GetBucketPrimarySPID(reqCtx.Context(), g.baseApp.Consensus(), bucketInfo)
	if err != nil {
		err = ErrConsensusWithDetail("GetBucketPrimarySPID error: " + err.Error())
		return
	}

	if redundancyIdx >= 0 && spID == bucketSPID {
		// get segment piece data from primary SP
		pieceData, err = g.getRecoverSegment(reqCtx.Context(), chainObjectInfo, bucketInfo, recoveryTask, params, signatureAddr)
		if err != nil {
			return
		}
	} else {
		// get EC piece data from secondary SP
		pieceData, err = g.getRecoverPiece(reqCtx.Context(), chainObjectInfo, bucketInfo, recoveryTask, params, signatureAddr)
		if err != nil {
			return
		}
	}

	_, err = w.Write(pieceData)
	if err != nil {
		err = ErrReplyData
		log.CtxErrorw(reqCtx.Context(), "failed to reply the recovery data", "error", err)
		return
	}
	log.CtxDebugw(reqCtx.Context(), "succeed to get one ec piece data")
}

// getRecoverPiece get EC piece data from handler and return
func (g *GateModular) getRecoverPiece(ctx context.Context, objectInfo *storagetypes.ObjectInfo, bucketInfo *storagetypes.BucketInfo,
	recoveryTask gfsptask.GfSpRecoverPieceTask, params *storagetypes.Params, signatureAddr sdktypes.AccAddress) ([]byte, error) {
	var err error

	bucketSPID, err := util.GetBucketPrimarySPID(ctx, g.baseApp.Consensus(), bucketInfo)
	if err != nil {
		return nil, err
	}
	primarySp, err := g.baseApp.Consensus().QuerySPByID(ctx, bucketSPID)
	if err != nil {
		return nil, err
	}

	// the primary sp of the object should be consistent with task signature
	gvg, err := g.baseApp.GfSpClient().GetGlobalVirtualGroup(ctx, bucketInfo.Id.Uint64(), objectInfo.LocalVirtualGroupId)
	if err != nil {
		return nil, err
	}

	sspAddress := make([]string, 0)
	for _, sspID := range gvg.SecondarySpIds {
		sp, queryErr := g.baseApp.Consensus().QuerySPByID(ctx, sspID)
		if queryErr != nil {
			return nil, queryErr
		}
		sspAddress = append(sspAddress, sp.OperatorAddress)
	}

	// if myself is secondary, the sender of the request can be both of  the primary SP or the secondary SP of the gvg
	if primarySp.OperatorAddress != signatureAddr.String() {
		log.CtxDebug(ctx, "recovery request not come from primary sp", "secondary sp", signatureAddr.String())
		// judge if the sender is not one of the secondary SP
		isRequestFromSecondary := false
		var taskECIndex int32
		for idx, sspAddr := range sspAddress {
			if sspAddr == signatureAddr.String() {
				isRequestFromSecondary = true
				taskECIndex = int32(idx)
			}
		}
		if !isRequestFromSecondary || (isRequestFromSecondary && taskECIndex != recoveryTask.GetEcIdx()) {
			log.CtxErrorw(ctx, "recovery request not come from the correct secondary sp")
			err = ErrRecoverySP
			return nil, err
		}
	}

	isOneOfSecondary := false
	var ECIndex int
	for idx, spAddr := range sspAddress {
		if spAddr == g.baseApp.OperatorAddress() {
			isOneOfSecondary = true
			ECIndex = idx
			break
		}
	}

	// if the handler SP is not one of the secondary SP of the task object, return err
	if !isOneOfSecondary {
		return nil, ErrRecoverySP
	}

	ECPieceSize := g.baseApp.PieceOp().ECPieceSize(objectInfo.PayloadSize, recoveryTask.GetSegmentIdx(),
		params.GetMaxSegmentSize(), params.GetRedundantDataChunkNum())
	ECPieceKey := g.baseApp.PieceOp().ECPieceKey(recoveryTask.GetObjectInfo().Id.Uint64(), recoveryTask.GetSegmentIdx(), uint32(ECIndex))

	pieceTask := &gfsptask.GfSpDownloadPieceTask{}
	// no need to check quota when recovering primary SP segment data
	pieceTask.InitDownloadPieceTask(objectInfo, bucketInfo, params, g.baseApp.TaskPriority(pieceTask),
		false, primarySp.OperatorAddress, uint64(ECPieceSize), ECPieceKey, 0, uint64(ECPieceSize),
		g.baseApp.TaskTimeout(pieceTask, uint64(pieceTask.GetSize())), g.baseApp.TaskMaxRetry(pieceTask))

	pieceData, err := g.baseApp.GfSpClient().GetPiece(ctx, pieceTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to download piece", "error", err)
		return nil, downloader.ErrPieceStoreWithDetail("failed to download piece, error: " + err.Error())
	}

	return pieceData, nil
}

// getRecoverSegment get segment piece data from handler and return the EC piece data based on recovery task info
func (g *GateModular) getRecoverSegment(ctx context.Context, objectInfo *storagetypes.ObjectInfo, bucketInfo *storagetypes.BucketInfo,
	recoveryTask gfsptask.GfSpRecoverPieceTask, params *storagetypes.Params, signatureAddr sdktypes.AccAddress) ([]byte, error) {
	ECPieceSize := g.baseApp.PieceOp().ECPieceSize(objectInfo.PayloadSize, recoveryTask.GetSegmentIdx(),
		params.GetMaxSegmentSize(), params.GetRedundantDataChunkNum())

	// if the handler is not the primary SP of the object, return error
	// TODO get sp id from config
	spID, err := g.getSPID()
	if err != nil {
		return nil, ErrConsensusWithDetail("getSPID error: " + err.Error())
	}
	bucketSPID, err := util.GetBucketPrimarySPID(ctx, g.baseApp.Consensus(), bucketInfo)
	if err != nil {
		return nil, ErrConsensusWithDetail("GetBucketPrimarySPID error: " + err.Error())
	}

	if bucketSPID != spID {
		log.CtxErrorw(ctx, "it is not the right the primary SP to handle secondary SP recovery", "actual_sp_id", spID,
			"expected_sp_id", bucketSPID)
		return nil, ErrRecoverySP
	}
	gvg, err := g.baseApp.GfSpClient().GetGlobalVirtualGroup(ctx, bucketInfo.Id.Uint64(), objectInfo.LocalVirtualGroupId)
	if err != nil {
		return nil, ErrRecoverySP
	}
	// if the sender is not one of the secondarySp, return err
	isOneOfSecondary := false
	var ECIndex int32
	for idx, sspId := range gvg.GetSecondarySpIds() {
		ssp, err := g.baseApp.Consensus().QuerySPByID(ctx, sspId)
		if err != nil {
			return nil, ErrConsensusWithDetail("QuerySPByID error: " + err.Error())
		}
		if ssp.OperatorAddress == signatureAddr.String() {
			isOneOfSecondary = true
			ECIndex = int32(idx)
		}
	}
	redundancyIdx := recoveryTask.EcIdx
	if !isOneOfSecondary || ECIndex != recoveryTask.EcIdx {
		return nil, ErrRecoverySP
	}
	pieceTask := &gfsptask.GfSpDownloadPieceTask{}
	segmentPieceKey := g.baseApp.PieceOp().SegmentPieceKey(recoveryTask.GetObjectInfo().Id.Uint64(),
		recoveryTask.GetSegmentIdx())

	// if recovery data chunk, just download the data part of segment in primarySP
	// no need to check quota when recovering primary SP or secondary SP data
	bucketPrimarySp, err := g.baseApp.Consensus().QuerySPByID(ctx, bucketSPID)
	if err != nil {
		return nil, err
	}
	if redundancyIdx < int32(params.GetRedundantDataChunkNum())-1 {
		pieceOffset := int64(redundancyIdx) * ECPieceSize

		pieceTask.InitDownloadPieceTask(objectInfo, bucketInfo, params, g.baseApp.TaskPriority(pieceTask),
			false, bucketPrimarySp.OperatorAddress, uint64(ECPieceSize), segmentPieceKey, uint64(pieceOffset), uint64(ECPieceSize),
			g.baseApp.TaskTimeout(pieceTask, uint64(pieceTask.GetSize())), g.baseApp.TaskMaxRetry(pieceTask))

		pieceData, err := g.baseApp.GfSpClient().GetPiece(ctx, pieceTask)
		if err != nil {
			log.CtxErrorw(ctx, "failed to download piece", "error", err)
			return nil, downloader.ErrPieceStoreWithDetail("failed to download piece, error: " + err.Error())
		}
		return pieceData, nil
	}

	// if recovery parity chunk, need to download the total segment and ec encode it
	segmentPieceSize := g.baseApp.PieceOp().SegmentPieceSize(objectInfo.PayloadSize, recoveryTask.GetSegmentIdx(), params.GetMaxSegmentSize())
	pieceTask.InitDownloadPieceTask(objectInfo, bucketInfo, params, g.baseApp.TaskPriority(pieceTask),
		false, bucketPrimarySp.OperatorAddress, uint64(ECPieceSize), segmentPieceKey, 0, uint64(segmentPieceSize),
		g.baseApp.TaskTimeout(pieceTask, uint64(pieceTask.GetSize())), g.baseApp.TaskMaxRetry(pieceTask))

	segmentData, err := g.baseApp.GfSpClient().GetPiece(ctx, pieceTask)
	if err != nil {
		log.CtxErrorw(ctx, "failed to download piece", "error", err)
		return nil, executor.ErrPieceStoreWithDetail("failed to download piece, error: " + err.Error())
	}

	ecData, err := redundancy.EncodeRawSegment(segmentData, int(params.GetRedundantDataChunkNum()),
		int(params.GetRedundantParityChunkNum()))
	if err != nil {
		log.CtxErrorw(ctx, "failed to ec encode data when recovering secondary SP", "error", err)
		return nil, err
	}

	return ecData[redundancyIdx], nil
}
