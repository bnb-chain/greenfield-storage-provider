package gater

import (
	"encoding/hex"
	"io"
	"net/http"

	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/golang/protobuf/proto"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/modular/p2p/p2pnode"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

// getApprovalHandler handles the get create bucket/object approval request.
// Before create bucket/object to the greenfield, the user should the primary
// SP whether willing serve for the user to manage the bucket/object.
// SP checks the user's account if it has the permission to operate, and send
// the request to approver that running the SP approval's Strategy.
func (g *GateModular) getApprovalHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                  error
		reqCtx               *RequestContext
		approvalMsg          []byte
		createBucketApproval = storagetypes.MsgCreateBucket{}
		createObjectApproval = storagetypes.MsgCreateObject{}
		authorized           bool
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

	reqCtx, err = NewRequestContext(r)
	if err != nil {
		return
	}

	approvalType := reqCtx.vars["action"]
	approvalMsg, err = hex.DecodeString(r.Header.Get(model.GnfdUnsignedApprovalMsgHeader))
	if err != nil {
		log.Errorw("failed to parse approval header", "approval_type", approvalType,
			"approval", r.Header.Get(model.GnfdUnsignedApprovalMsgHeader))
		err = ErrDecodeMsg
		return
	}

	switch approvalType {
	case createBucketApprovalAction:
		if err = storagetypes.ModuleCdc.UnmarshalJSON(approvalMsg, &createBucketApproval); err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to unmarshal approval", "approval",
				r.Header.Get(model.GnfdUnsignedApprovalMsgHeader), "error", err)
			err = ErrDecodeMsg
			return
		}
		if err = createBucketApproval.ValidateBasic(); err != nil {
			log.Errorw("failed to basic check bucket approval msg", "bucket_approval_msg",
				createBucketApproval, "error", err)
			err = ErrValidateMsg
			return
		}
		if reqCtx.NeedVerifyAuthorizer() {
			authorized, err = g.baseApp.GfSpClient().VerifyAuthorize(
				reqCtx.Context(), coremodule.AuthOpAskCreateBucketApproval,
				reqCtx.Account(), createBucketApproval.GetBucketName(), "")
			if err != nil {
				log.CtxErrorw(reqCtx.Context(), "failed to verify authorize", "error", err)
				return
			}
			if !authorized {
				log.CtxErrorw(reqCtx.Context(), "no permission to operator")
				err = ErrNoPermission
				return
			}
		}
		task := &gfsptask.GfSpCreateBucketApprovalTask{}
		task.InitApprovalCreateBucketTask(&createBucketApproval, g.baseApp.TaskPriority(task))
		var approvalTask coretask.ApprovalCreateBucketTask
		approved, approvalTask, err = g.baseApp.GfSpClient().AskCreateBucketApproval(reqCtx.Context(), task)
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
		w.Header().Set(model.GnfdSignedApprovalMsgHeader, hex.EncodeToString(sdktypes.MustSortJSON(bz)))
	case createObjectApprovalAction:
		if err = storagetypes.ModuleCdc.UnmarshalJSON(approvalMsg, &createObjectApproval); err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to unmarshal approval", "approval",
				r.Header.Get(model.GnfdUnsignedApprovalMsgHeader), "error", err)
			err = ErrDecodeMsg
			return
		}
		if err = createObjectApproval.ValidateBasic(); err != nil {
			log.Errorw("failed to basic check approval msg", "object_approval_msg",
				createObjectApproval, "error", err)
			err = ErrValidateMsg
			return
		}
		if reqCtx.NeedVerifyAuthorizer() {
			authorized, err = g.baseApp.GfSpClient().VerifyAuthorize(
				reqCtx.Context(), coremodule.AuthOpAskCreateObjectApproval,
				reqCtx.Account(), createObjectApproval.GetBucketName(),
				createObjectApproval.GetObjectName())
			if err != nil {
				log.CtxErrorw(reqCtx.Context(), "failed to verify authorize", "error", err)
				return
			}
			if !authorized {
				log.CtxErrorw(reqCtx.Context(), "no permission to operator")
				err = ErrNoPermission
				return
			}
		}
		task := &gfsptask.GfSpCreateObjectApprovalTask{}
		task.InitApprovalCreateObjectTask(&createObjectApproval, g.baseApp.TaskPriority(task))
		var approvedTask coretask.ApprovalCreateObjectTask
		approved, approvedTask, err = g.baseApp.GfSpClient().AskCreateObjectApproval(r.Context(), task)
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
		w.Header().Set(model.GnfdSignedApprovalMsgHeader, hex.EncodeToString(sdktypes.MustSortJSON(bz)))
	default:
		err = ErrUnsupportedRequestType
		return
	}
	log.CtxDebugw(reqCtx.Context(), "succeed to ask approval")
	return
}

// challengeHandler handles get challenge piece info request. Current only greenfield
// validator can challenge piece is store correctly. The challenge piece info includes:
// the challenged piece data, all piece hashes and the integrity hash. The challenger
// can verify the info whether are correct by comparing with the greenfield info.
func (g *GateModular) challengeHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		reqCtx     *RequestContext
		authorized bool
		integrity  []byte
		checksums  [][]byte
		data       []byte
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

	reqCtx, err = NewRequestContext(r)
	if err != nil {
		return
	}
	objectID, err := util.StringToUint64(reqCtx.request.Header.Get(model.GnfdObjectIDHeader))
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse object id", "object_id",
			reqCtx.request.Header.Get(model.GnfdObjectIDHeader))
		err = ErrInvalidHeader
		return
	}
	objectInfo, err := g.baseApp.Consensus().QueryObjectInfoByID(reqCtx.Context(),
		reqCtx.request.Header.Get(model.GnfdObjectIDHeader))
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get object info from consensus", "error", err)
		err = ErrConsensus
		return
	}
	if reqCtx.NeedVerifyAuthorizer() {
		authorized, err = g.baseApp.GfSpClient().VerifyAuthorize(reqCtx.Context(),
			coremodule.AuthOpTypeChallengePiece, reqCtx.Account(), objectInfo.GetBucketName(),
			objectInfo.GetObjectName())
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to verify authorize", "error", err)
			return
		}
		if !authorized {
			log.CtxErrorw(reqCtx.Context(), "no permission to operator")
			err = ErrNoPermission
			return
		}
	}

	bucketInfo, err := g.baseApp.Consensus().QueryBucketInfo(reqCtx.Context(), objectInfo.GetBucketName())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get bucket info from consensus", "error", err)
		err = ErrConsensus
		return
	}
	redundancyIdx, err := util.StringToInt32(reqCtx.request.Header.Get(model.GnfdRedundancyIndexHeader))
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse redundancy idx", "redundancy_idx",
			reqCtx.request.Header.Get(model.GnfdRedundancyIndexHeader))
		err = ErrInvalidHeader
		return
	}
	segmentIdx, err := util.StringToUint32(reqCtx.request.Header.Get(model.GnfdPieceIndexHeader))
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse segment_idx", "segment_idx",
			reqCtx.request.Header.Get(model.GnfdPieceIndexHeader))
		err = ErrInvalidHeader
		return
	}
	parms, err := g.baseApp.Consensus().QueryStorageParams(reqCtx.Context())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get storage params", "error", err)
		return
	}
	var pieceSize uint64
	if redundancyIdx < 0 {
		pieceSize = uint64(g.baseApp.PieceOp().SegmentSize(objectInfo.GetPayloadSize(),
			segmentIdx, parms.VersionedParams.GetMaxSegmentSize()))
	} else {
		pieceSize = uint64(g.baseApp.PieceOp().PieceSize(objectInfo.GetPayloadSize(),
			segmentIdx, parms.VersionedParams.GetMaxSegmentSize(),
			parms.VersionedParams.GetRedundantDataChunkNum()))
	}
	task := &gfsptask.GfSpChallengePieceTask{}
	task.InitChallengePieceTask(objectInfo, bucketInfo, g.baseApp.TaskPriority(task), reqCtx.Account(),
		redundancyIdx, segmentIdx, g.baseApp.TaskTimeout(task, pieceSize), g.baseApp.TaskMaxRetry(task))
	ctx := log.WithValue(reqCtx.Context(), log.CtxKeyTask, task.Key().String())
	integrity, checksums, data, err = g.baseApp.GfSpClient().GetChallengeInfo(reqCtx.Context(), task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to challenge piece", "error", err)
		return
	}
	w.Header().Set(model.GnfdObjectIDHeader, util.Uint64ToString(objectID))
	w.Header().Set(model.GnfdIntegrityHashHeader, hex.EncodeToString(integrity))
	w.Header().Set(model.GnfdPieceHashHeader, util.BytesSliceToString(checksums))
	w.Write(data)
	log.CtxDebugw(reqCtx.Context(), "succeed to challenge piece")
}

// replicateHandler handles the replicate piece from primary SP request. The Primary
// replicates the piece data one by one, and will ask the integrity hash and the
// signature to seal object on greenfield.
func (g *GateModular) replicateHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err           error
		reqCtx        *RequestContext
		approvalMsg   []byte
		receiveMsg    []byte
		data          []byte
		integrity     []byte
		signature     []byte
		currentHeight uint64
		approval      = gfsptask.GfSpReplicatePieceApprovalTask{}
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
	// ignore the error, because the replicate request only between SPs, the request
	// verification is by signature of the ReceivePieceTask
	reqCtx, _ = NewRequestContext(r)

	approvalMsg, err = hex.DecodeString(r.Header.Get(model.GnfdReplicatePieceApprovalHeader))
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse replicate piece approval header",
			"approval", r.Header.Get(model.GnfdReceiveMsgHeader))
		err = ErrDecodeMsg
		return
	}
	err = proto.Unmarshal(approvalMsg, &approval)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal replicate piece approval header",
			"receive", r.Header.Get(model.GnfdReceiveMsgHeader))
		err = ErrDecodeMsg
		return
	}
	if approval.GetApprovedSpOperatorAddress() != g.baseApp.OperateAddress() {
		log.CtxErrorw(reqCtx.Context(), "failed to verify replicate piece approval, sp mismatch")
		err = ErrMismatchSp
		return
	}
	err = p2pnode.VerifySignature(g.baseApp.OperateAddress(), approval.GetSignBytes(), approval.GetApprovedSignature())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to verify replicate piece approval signature")
		err = ErrSignature
		return
	}
	currentHeight, err = g.baseApp.Consensus().CurrentHeight(reqCtx.Context())
	if err != nil {
		// ignore the system's inner error,let the request go
		log.CtxErrorw(reqCtx.Context(), "failed to get current block height")
	} else if currentHeight > approval.GetExpiredHeight() {
		log.CtxErrorw(reqCtx.Context(), "replicate piece approval expired")
		err = ErrApprovalExpired
		return
	}

	receiveMsg, err = hex.DecodeString(r.Header.Get(model.GnfdReceiveMsgHeader))
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse receive header",
			"receive", r.Header.Get(model.GnfdReceiveMsgHeader))
		err = ErrDecodeMsg
		return
	}
	receiveTask := gfsptask.GfSpReceivePieceTask{}
	err = proto.Unmarshal(receiveMsg, &receiveTask)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal receive header",
			"receive", r.Header.Get(model.GnfdReceiveMsgHeader))
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
	data, err = io.ReadAll(r.Body)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to read replicate piece data", "error", err)
		err = ErrExceptionStream
		return
	}
	if receiveTask.GetPieceIdx() > 0 {
		err = g.baseApp.GfSpClient().ReplicatePiece(reqCtx.Context(), &receiveTask, data)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to receive piece", "error", err)
			return
		}
	} else {
		integrity, signature, err = g.baseApp.GfSpClient().DoneReplicatePiece(reqCtx.Context(), &receiveTask)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to done receive piece", "error", err)
			return
		}
		w.Header().Set(model.GnfdIntegrityHashHeader, hex.EncodeToString(integrity))
		w.Header().Set(model.GnfdIntegrityHashSignatureHeader, hex.EncodeToString(signature))
	}
	log.CtxDebugw(reqCtx.Context(), "succeed to replicate piece")
}
