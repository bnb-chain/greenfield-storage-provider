package gateway

import (
	"context"
	"encoding/hex"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	btypes "github.com/bnb-chain/greenfield/x/bridge/types"
	types "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func isAdminRouter(routerName string) bool {
	return routerName == approvalRouterName || routerName == challengeRouterName
}

// getApprovalHandler handle create bucket or create object approval
func (g *Gateway) getApprovalHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err              error
		errorDescription *errorDescription
		requestContext   *requestContext
		addr             sdk.AccAddress
		statusCode       = http.StatusOK
	)

	requestContext = newRequestContext(r)
	defer func() {
		if errorDescription != nil {
			statusCode = errorDescription.statusCode
			_ = errorDescription.errorResponse(w, requestContext)
		}
		if statusCode == http.StatusOK {
			log.Infof("action(%v) statusCode(%v) %v", "getApproval", statusCode, requestContext.generateRequestDetail())
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "getApproval", statusCode, requestContext.generateRequestDetail())
		}
	}()

	if addr, err = requestContext.verifySignature(); err != nil {
		log.Infow("failed to verify signature", "error", err)
		errorDescription = SignatureNotMatch
		return
	}
	if err = g.checkAuthorization(requestContext, addr); err != nil {
		log.Warnw("failed to check authorization", "error", err)
		errorDescription = UnauthorizedAccess
		return
	}

	actionName := requestContext.vars["action"]
	approvalMsg, err := hex.DecodeString(r.Header.Get(model.GnfdUnsignedApprovalMsgHeader))
	if err != nil {
		log.Warnw("invalid approval", "approval", r.Header.Get(model.GnfdUnsignedApprovalMsgHeader))
		errorDescription = InvalidHeader
	}

	switch actionName {
	case createBucketApprovalAction:
		var (
			msg               = types.MsgCreateBucket{}
			approvalSignature []byte
		)

		btypes.ModuleCdc.MustUnmarshalJSON(approvalMsg, &msg)
		// TODO: to config it
		msg.PrimarySpApproval.ExpiredHeight = 10
		approvalSignature, err = g.signer.SignBucketApproval(context.Background(), &msg)
		if err != nil {
			log.Warnw("failed to sign create bucket approval", "error", err)
			errorDescription = InternalError
			return
		}
		msg.PrimarySpApproval.Sig = approvalSignature
		bz := btypes.ModuleCdc.MustMarshalJSON(&msg)
		w.Header().Set(model.GnfdSignedApprovalMsgHeader, hex.EncodeToString(sdk.MustSortJSON(bz)))
	case createObjectApprovalAction:
		var (
			msg               = types.MsgCreateObject{}
			approvalSignature []byte
		)
		btypes.ModuleCdc.MustUnmarshalJSON(approvalMsg, &msg)
		// TODO: to config it
		msg.PrimarySpApproval.ExpiredHeight = 10
		approvalSignature, err = g.signer.SignObjectApproval(context.Background(), &msg)
		if err != nil {
			log.Warnw("failed to sign create object approval", "error", err)
			errorDescription = InternalError
			return
		}
		msg.PrimarySpApproval.Sig = approvalSignature
		bz := btypes.ModuleCdc.MustMarshalJSON(&msg)
		w.Header().Set(model.GnfdSignedApprovalMsgHeader, hex.EncodeToString(sdk.MustSortJSON(bz)))
	default:
		log.Warnw("not implement approval", "action", actionName)
		errorDescription = NotImplementedError
		return
	}
}

// challengeHandler handle challenge request
func (g *Gateway) challengeHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err              error
		errorDescription *errorDescription
		requestContext   *requestContext
		addr             sdk.AccAddress

		objectID         uint64
		challengePrimary bool
		segmentIdx       uint32
		ecIdx            uint32
		redundancyType   ptypes.RedundancyType
		spAddress        string
		statusCode       = http.StatusOK
	)

	requestContext = newRequestContext(r)
	defer func() {
		if errorDescription != nil {
			statusCode = errorDescription.statusCode
			_ = errorDescription.errorResponse(w, requestContext)
		}
		if statusCode == http.StatusOK {
			log.Infof("action(%v) statusCode(%v) %v", "challenge", statusCode, requestContext.generateRequestDetail())
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "challenge", statusCode, requestContext.generateRequestDetail())
		}
	}()

	if addr, err = requestContext.verifySignature(); err != nil {
		log.Warnw("failed to verify signature", "error", err)
		errorDescription = SignatureNotMatch
		return
	}
	if err = g.checkAuthorization(requestContext, addr); err != nil {
		log.Warnw("failed to check authorization", "error", err)
		errorDescription = UnauthorizedAccess
		return
	}

	if objectID, err = util.HeaderToUint64(requestContext.request.Header.Get(model.GnfdObjectIDHeader)); err != nil {
		log.Warnw("invalid object id", "object_id", requestContext.request.Header.Get(model.GnfdObjectIDHeader))
		errorDescription = InvalidHeader
		return
	}

	redundancyIndex, err := util.HeaderToInt64(requestContext.request.Header.Get(model.GnfdRedundancyIndexHeader))
	if err != nil {
		log.Warnw("invalid redundancy index", "redundancy_index", requestContext.request.Header.Get(model.GnfdRedundancyIndexHeader))
		errorDescription = InvalidHeader
		return
	}
	if redundancyIndex < 0 {
		challengePrimary = true
	} else {
		ecIdx = uint32(redundancyIndex)
	}
	if segmentIdx, err = util.HeaderToUint32(requestContext.request.Header.Get(model.GnfdPieceIndexHeader)); err != nil {
		log.Warnw("invalid segment idx", "segment_idx", requestContext.request.Header.Get(model.GnfdPieceIndexHeader))
		errorDescription = InvalidHeader
		return
	}
	spAddress = g.config.StorageProvider

	req := &stypes.ChallengeServiceChallengePieceRequest{
		TraceId:               requestContext.requestID,
		ObjectId:              objectID,
		ChallengePrimaryPiece: challengePrimary,
		SegmentIdx:            segmentIdx,
		EcIdx:                 ecIdx,
		RedundancyType:        redundancyType,
		StorageProviderId:     spAddress,
	}
	ctx := log.Context(context.Background(), req)
	resp, err := g.challenge.ChallengePiece(ctx, req)
	if err != nil {
		log.Warnf("failed to challenge", "error", err)
		errorDescription = InternalError
		return
	}
	w.Header().Set(model.GnfdRequestIDHeader, requestContext.requestID)
	w.Header().Set(model.GnfdObjectIDHeader, util.Uint64ToHeader(objectID))
	w.Header().Set(model.GnfdIntegrityHashHeader, hex.EncodeToString(resp.IntegrityHash))
	w.Header().Set(model.GnfdPieceHashHeader, util.EncodePieceHash(resp.PieceHash))
	w.Write(resp.PieceData)
}
