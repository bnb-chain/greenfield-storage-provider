package gater

import (
	"encoding/hex"
	"io"
	"net/http"

	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func (g *GateModular) getApprovalHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                  error
		authVerified         = true
		allowApproval        = true
		accAddress           sdk.AccAddress
		approvalCreateBucket coretask.ApprovalCreateBucketTask
		approvalCreateObject coretask.ApprovalCreateObjectTask
		reqCtx               = NewRequestContext(r)
	)
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to ask approval", "req_info", reqCtx.String())
			MakeErrorResponse(w, gfsperrors.MakeGfSpError(err))
		}
	}()
	if reqCtx.NeedVerifySignature() {
		accAddress, err = reqCtx.VerifySignature()
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to verify signature", "error", err)
			return
		}
	}
	actionName := reqCtx.vars["action"]
	approvalMsg, err := hex.DecodeString(r.Header.Get(model.GnfdUnsignedApprovalMsgHeader))
	if err != nil {
		log.Errorw("failed to parse approval header", "approval", r.Header.Get(model.GnfdUnsignedApprovalMsgHeader))
		err = ErrDecodeMsg
		return
	}
	switch actionName {
	case createBucketApprovalAction:
		createBucketApproval := storagetypes.MsgCreateBucket{}
		if err = storagetypes.ModuleCdc.UnmarshalJSON(approvalMsg, &createBucketApproval); err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to unmarshal approval", "approval", r.Header.Get(model.GnfdUnsignedApprovalMsgHeader), "error", err)
			err = ErrDecodeMsg
			return
		}
		if err = createBucketApproval.ValidateBasic(); err != nil {
			log.Errorw("failed to basic check", "bucket_approval_msg", createBucketApproval, "error", err)
			err = ErrValidateMsg
			return
		}
		if reqCtx.NeedVerifySignature() {
			authVerified, err = g.baseApp.GfSpClient().VerifyAuthorize(reqCtx.Context(),
				coremodule.AuthOpAskCreateBucketApproval, accAddress.String(),
				createBucketApproval.GetBucketName(), "")
			if err != nil {
				log.CtxErrorw(reqCtx.Context(), "failed to verify authorize", "error", err)
				return
			}
			if !authVerified {
				log.CtxErrorw(reqCtx.Context(), "no permission to operator")
				err = ErrNoPermission
				return
			}
		}
		task := &gfsptask.GfSpCreateBucketApprovalTask{}
		task.InitApprovalCreateBucketTask(&createBucketApproval, g.baseApp.TaskPriority(task))
		allowApproval, approvalCreateBucket, err = g.baseApp.GfSpClient().AskCreateBucketApproval(r.Context(), task)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to ask create bucket approval", "error", err)
			return
		}
		if !allowApproval {
			log.CtxErrorw(reqCtx.Context(), "refuse the ask create bucket approval")
			err = ErrRefuseApproval
			return
		}
		bz := storagetypes.ModuleCdc.MustMarshalJSON(approvalCreateBucket.GetCreateBucketInfo())
		w.Header().Set(model.GnfdSignedApprovalMsgHeader, hex.EncodeToString(sdktypes.MustSortJSON(bz)))
	case createObjectApprovalAction:
		createObjectApproval := storagetypes.MsgCreateObject{}
		if err = storagetypes.ModuleCdc.UnmarshalJSON(approvalMsg, &createObjectApproval); err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to unmarshal approval", "approval",
				r.Header.Get(model.GnfdUnsignedApprovalMsgHeader), "error", err)
			err = ErrDecodeMsg
			return
		}
		if err = createObjectApproval.ValidateBasic(); err != nil {
			log.Errorw("failed to basic check", "object_approval_msg",
				createObjectApproval, "error", err)
			err = ErrValidateMsg
			return
		}
		if reqCtx.NeedVerifySignature() {
			authVerified, err = g.baseApp.GfSpClient().VerifyAuthorize(reqCtx.Context(),
				coremodule.AuthOpAskCreateObjectApproval, accAddress.String(),
				createObjectApproval.GetBucketName(), createObjectApproval.GetObjectName())
			if err != nil {
				log.CtxErrorw(reqCtx.Context(), "failed to verify authorize", "error", err)
				return
			}
			if !authVerified {
				log.CtxErrorw(reqCtx.Context(), "no permission to operator")
				err = ErrNoPermission
				return
			}
		}
		task := &gfsptask.GfSpCreateObjectApprovalTask{}
		task.InitApprovalCreateObjectTask(&createObjectApproval, g.baseApp.TaskPriority(task))
		allowApproval, approvalCreateObject, err = g.baseApp.GfSpClient().AskCreateObjectApproval(r.Context(), task)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to ask object approval", "error", err)
			return
		}
		if !allowApproval {
			log.CtxErrorw(reqCtx.Context(), "refuse the ask create object approval")
			err = ErrRefuseApproval
			return
		}
		bz := storagetypes.ModuleCdc.MustMarshalJSON(approvalCreateObject.GetCreateObjectInfo())
		w.Header().Set(model.GnfdSignedApprovalMsgHeader, hex.EncodeToString(sdktypes.MustSortJSON(bz)))
	default:
		err = ErrUnsupportedRequestType
	}
	return
}

func (g *GateModular) challengeHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		accAddress sdk.AccAddress
		reqCtx     = NewRequestContext(r)
	)
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to challenge piece", "req_info", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()
	if reqCtx.NeedVerifySignature() {
		accAddress, err = reqCtx.VerifySignature()
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to verify signature", "error", err)
			return
		}
	}
	objectID, err := util.StringToUint64(reqCtx.request.Header.Get(model.GnfdObjectIDHeader))
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse object_id", "object_id",
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
	if reqCtx.NeedVerifySignature() {
		verified, err := g.baseApp.GfSpClient().VerifyAuthorize(reqCtx.Context(),
			coremodule.AuthOpTypeChallengePiece, accAddress.String(), objectInfo.GetBucketName(),
			objectInfo.GetObjectName())
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to verify authorize", "error", err)
			return
		}
		if !verified {
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
		log.CtxErrorw(reqCtx.Context(), "failed to parse redundancy_idx", "redundancy_idx",
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
	task := &gfsptask.GfSpChallengePieceTask{}
	task.InitChallengePieceTask(objectInfo, bucketInfo, g.baseApp.TaskPriority(task), accAddress.String(),
		redundancyIdx, segmentIdx, g.baseApp.TaskTimeout(task), g.baseApp.TaskMaxRetry(task))
	ctx := log.WithValue(reqCtx.Context(), log.CtxKeyTask, task.Key().String())
	integrity, checksums, data, err := g.baseApp.GfSpClient().GetChallengeInfo(reqCtx.Context(), task)
	if err != nil {
		log.CtxErrorw(ctx, "failed to challenge piece", "error", err)
		return
	}
	w.Header().Set(model.GnfdObjectIDHeader, util.Uint64ToString(objectID))
	w.Header().Set(model.GnfdIntegrityHashHeader, hex.EncodeToString(integrity))
	w.Header().Set(model.GnfdPieceHashHeader, util.BytesSliceToString(checksums))
	w.Write(data)
}

func (g *GateModular) replicateHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err       error
		integrity []byte
		signature []byte
		reqCtx    = NewRequestContext(r)
	)
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to challenge piece", "req_info", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	approvalMsg, err := hex.DecodeString(r.Header.Get(model.GnfdReplicatePieceApprovalHeader))
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse replicate piece approval header", "approval",
			r.Header.Get(model.GnfdReceiveMsgHeader))
		err = ErrDecodeMsg
		return
	}
	approval := gfsptask.GfSpReplicatePieceApprovalTask{}
	err = proto.Unmarshal(approvalMsg, &approval)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal replicate piece approval header", "receive",
			r.Header.Get(model.GnfdReceiveMsgHeader))
		err = ErrDecodeMsg
		return
	}
	if approval.GetApprovedSpOperatorAddress() != g.baseApp.OperateAddress() {
		log.CtxErrorw(reqCtx.Context(), "failed to receive piece data, sp mismatch")
		err = ErrMisMatchSp
		return
	}
	err = p2pnode.VerifySignature(g.baseApp.OperateAddress(), approval.GetSignBytes(), approval.GetApprovedSignature())
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to verify replicate piece approval signature")
		err = ErrSignature
		return
	}

	receiveMsg, err := hex.DecodeString(r.Header.Get(model.GnfdReceiveMsgHeader))
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse receive header", "receive",
			r.Header.Get(model.GnfdReceiveMsgHeader))
		err = ErrDecodeMsg
		return
	}
	receiveTask := gfsptask.GfSpReceivePieceTask{}
	err = proto.Unmarshal(receiveMsg, &receiveTask)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to unmarshal receive header", "receive",
			r.Header.Get(model.GnfdReceiveMsgHeader))
		err = ErrDecodeMsg
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "stream exception", "error", err)
		err = ErrExceptionStream
		return
	}
	if receiveTask.GetPieceSize() > 0 {
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

}
