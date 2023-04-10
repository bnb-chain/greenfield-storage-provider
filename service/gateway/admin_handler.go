package gateway

import (
	"context"
	"encoding/hex"
	"net/http"

	"github.com/bnb-chain/greenfield/x/storage/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

// getApprovalHandler handles the bucket create or object create approval
func (gateway *Gateway) getApprovalHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		errDescription *errorDescription
		reqContext     *requestContext
		addr           sdktypes.AccAddress
	)

	reqContext = newRequestContext(r)
	defer func() {
		if errDescription != nil {
			_ = errDescription.errorResponse(w, reqContext)
		}
		if errDescription != nil && errDescription.statusCode != http.StatusOK {
			log.Errorf("action(%v) statusCode(%v) %v", approvalRouterName, errDescription.statusCode, reqContext.generateRequestDetail())
		} else {
			log.Infof("action(%v) statusCode(200) %v", approvalRouterName, reqContext.generateRequestDetail())
		}
	}()

	if gateway.signer == nil {
		log.Error("failed to get approval due to not config signer")
		errDescription = NotExistComponentError
		return
	}

	if addr, err = reqContext.verifySignature(); err != nil {
		log.Errorw("failed to verify signature", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	if err = gateway.checkAuthorization(reqContext, addr); err != nil {
		log.Errorw("failed to check authorization", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}

	actionName := reqContext.vars["action"]
	approvalMsg, err := hex.DecodeString(r.Header.Get(model.GnfdUnsignedApprovalMsgHeader))
	if err != nil {
		log.Errorw("failed to parse approval header", "approval", r.Header.Get(model.GnfdUnsignedApprovalMsgHeader))
		errDescription = InvalidHeader
		return
	}

	currentHeight, err := gateway.chain.GetCurrentHeight(context.Background())
	if err != nil {
		log.Errorw("failed to query current height", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}

	switch actionName {
	case createBucketApprovalAction:
		var (
			msg               = types.MsgCreateBucket{}
			approvalSignature []byte
		)
		if types.ModuleCdc.UnmarshalJSON(approvalMsg, &msg) != nil {
			log.Errorw("failed to unmarshal approval", "approval", r.Header.Get(model.GnfdUnsignedApprovalMsgHeader))
			errDescription = InvalidHeader
			return
		}
		if err = msg.ValidateBasic(); err != nil {
			log.Errorw("failed to basic check", "bucket_msg", msg, "error", err)
			errDescription = InvalidHeader
			return
		}
		msg.PrimarySpApproval = &types.Approval{ExpiredHeight: currentHeight + model.DefaultTimeoutHeight}
		approvalSignature, err = gateway.signer.SignBucketApproval(context.Background(), &msg)
		if err != nil {
			log.Errorw("failed to sign create bucket approval", "error", err)
			errDescription = makeErrorDescription(err)
			return
		}
		msg.PrimarySpApproval.Sig = approvalSignature
		bz := types.ModuleCdc.MustMarshalJSON(&msg)
		w.Header().Set(model.GnfdSignedApprovalMsgHeader, hex.EncodeToString(sdktypes.MustSortJSON(bz)))
	case createObjectApprovalAction:
		var (
			msg               = types.MsgCreateObject{}
			approvalSignature []byte
		)
		if types.ModuleCdc.UnmarshalJSON(approvalMsg, &msg) != nil {
			log.Errorw("failed to unmarshal approval", "approval", r.Header.Get(model.GnfdUnsignedApprovalMsgHeader))
			errDescription = InvalidHeader
			return
		}
		if err = msg.ValidateBasic(); err != nil {
			log.Errorw("failed to basic check", "object_msg", msg, "error", err)
			errDescription = InvalidHeader
			return
		}
		msg.PrimarySpApproval = &types.Approval{ExpiredHeight: currentHeight + model.DefaultTimeoutHeight}
		approvalSignature, err = gateway.signer.SignObjectApproval(context.Background(), &msg)
		if err != nil {
			log.Errorw("failed to sign create object approval", "error", err)
			errDescription = makeErrorDescription(err)
			return
		}
		msg.PrimarySpApproval.Sig = approvalSignature
		bz := types.ModuleCdc.MustMarshalJSON(&msg)
		w.Header().Set(model.GnfdSignedApprovalMsgHeader, hex.EncodeToString(sdktypes.MustSortJSON(bz)))
	default:
		log.Errorw("failed to get approval due to unimplemented approval type", "action", actionName)
		errDescription = NotImplementedError
		return
	}
	w.Header().Set(model.GnfdRequestIDHeader, reqContext.requestID)
}

// challengeHandler handles the challenge request
func (gateway *Gateway) challengeHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		errDescription *errorDescription
		reqContext     *requestContext
		addr           sdktypes.AccAddress
		objectID       uint64
		redundancyIdx  int32
		segmentIdx     uint32
	)

	reqContext = newRequestContext(r)
	defer func() {
		if errDescription != nil {
			_ = errDescription.errorResponse(w, reqContext)
		}
		if errDescription != nil && errDescription.statusCode != http.StatusOK {
			log.Errorf("action(%v) statusCode(%v) %v", challengeRouterName, errDescription.statusCode, reqContext.generateRequestDetail())
		} else {
			log.Infof("action(%v) statusCode(200) %v", challengeRouterName, reqContext.generateRequestDetail())
		}
	}()

	if gateway.challenge == nil {
		log.Errorw("failed to get challenge due to not config challenge")
		errDescription = NotExistComponentError
		return
	}

	if addr, err = reqContext.verifySignature(); err != nil {
		log.Errorw("failed to verify signature", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	if err = gateway.checkAuthorization(reqContext, addr); err != nil {
		log.Errorw("failed to check authorization", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}

	if objectID, err = util.StringToUint64(reqContext.request.Header.Get(model.GnfdObjectIDHeader)); err != nil {
		log.Errorw("failed to parse object_id", "object_id", reqContext.request.Header.Get(model.GnfdObjectIDHeader))
		errDescription = InvalidHeader
		return
	}

	if redundancyIdx, err = util.StringToInt32(reqContext.request.Header.Get(model.GnfdRedundancyIndexHeader)); err != nil {
		log.Errorw("failed to parse redundancy_idx", "redundancy_idx", reqContext.request.Header.Get(model.GnfdRedundancyIndexHeader))
		errDescription = InvalidHeader
		return
	}
	if segmentIdx, err = util.StringToUint32(reqContext.request.Header.Get(model.GnfdPieceIndexHeader)); err != nil {
		log.Errorw("failed to parse segment_idx", "segment_idx", reqContext.request.Header.Get(model.GnfdPieceIndexHeader))
		errDescription = InvalidHeader
		return
	}
	integrityHash, pieceHash, pieceData, err := gateway.challenge.ChallengePiece(context.Background(), reqContext.objectInfo, redundancyIdx, segmentIdx)
	if err != nil {
		log.Errorf("failed to challenge piece", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	w.Header().Set(model.GnfdRequestIDHeader, reqContext.requestID)
	w.Header().Set(model.GnfdObjectIDHeader, util.Uint64ToString(objectID))
	w.Header().Set(model.GnfdIntegrityHashHeader, hex.EncodeToString(integrityHash))
	w.Header().Set(model.GnfdPieceHashHeader, util.BytesSliceToString(pieceHash))
	w.Write(pieceData)
}
