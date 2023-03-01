package gateway

//import (
//	"context"
//	"encoding/hex"
//	"math"
//	"net/http"
//
//	"github.com/bnb-chain/greenfield-storage-provider/model"
//	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
//	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
//	"github.com/bnb-chain/greenfield-storage-provider/util"
//	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
//	"github.com/bnb-chain/greenfield/x/storage/types"
//	sdk "github.com/cosmos/cosmos-sdk/types"
//)
//
//// getApprovalHandler handle create bucket or create object approval
//func (g *Gateway) getApprovalHandler(w http.ResponseWriter, r *http.Request) {
//	var (
//		err              error
//		errorDescription *errorDescription
//		requestContext   *requestContext
//		addr             sdk.AccAddress
//	)
//
//	requestContext = newRequestContext(r)
//	defer func() {
//		if errorDescription != nil {
//			_ = errorDescription.errorResponse(w, requestContext)
//		}
//		if errorDescription != nil && errorDescription.statusCode != http.StatusOK {
//			log.Errorf("action(%v) statusCode(%v) %v", approvalRouterName, errorDescription.statusCode, requestContext.generateRequestDetail())
//		} else {
//			log.Infof("action(%v) statusCode(200) %v", approvalRouterName, requestContext.generateRequestDetail())
//		}
//	}()
//
//	if addr, err = requestContext.verifySignature(); err != nil {
//		log.Errorw("failed to verify signature", "error", err)
//		errorDescription = SignatureNotMatch
//		return
//	}
//	if err = g.checkAuthorization(requestContext, addr); err != nil {
//		log.Errorw("failed to check authorization", "error", err)
//		errorDescription = UnauthorizedAccess
//		return
//	}
//
//	actionName := requestContext.vars["action"]
//	approvalMsg, err := hex.DecodeString(r.Header.Get(model.GnfdUnsignedApprovalMsgHeader))
//	if err != nil {
//		log.Errorw("invalid approval", "approval", r.Header.Get(model.GnfdUnsignedApprovalMsgHeader))
//		errorDescription = InvalidHeader
//		return
//	}
//
//	switch actionName {
//	case createBucketApprovalAction:
//		var (
//			msg               = types.MsgCreateBucket{}
//			approvalSignature []byte
//		)
//		if types.ModuleCdc.UnmarshalJSON(approvalMsg, &msg) != nil {
//			log.Errorw("invalid approval", "approval", r.Header.Get(model.GnfdUnsignedApprovalMsgHeader))
//			errorDescription = InvalidHeader
//			return
//		}
//
//		// TODO: to config it
//		msg.PrimarySpApproval = &types.Approval{ExpiredHeight: math.MaxUint64}
//		approvalSignature, err = g.signer.SignBucketApproval(context.Background(), &msg)
//		if err != nil {
//			log.Errorw("failed to sign create bucket approval", "error", err)
//			errorDescription = InternalError
//			return
//		}
//		msg.PrimarySpApproval.Sig = approvalSignature
//		bz := types.ModuleCdc.MustMarshalJSON(&msg)
//		w.Header().Set(model.GnfdSignedApprovalMsgHeader, hex.EncodeToString(sdk.MustSortJSON(bz)))
//	case createObjectApprovalAction:
//		var (
//			msg               = types.MsgCreateObject{}
//			approvalSignature []byte
//		)
//		if types.ModuleCdc.UnmarshalJSON(approvalMsg, &msg) != nil {
//			log.Errorw("invalid approval", "approval", r.Header.Get(model.GnfdUnsignedApprovalMsgHeader))
//			errorDescription = InvalidHeader
//			return
//		}
//		// TODO: to config it
//		msg.PrimarySpApproval = &types.Approval{ExpiredHeight: math.MaxUint64}
//		approvalSignature, err = g.signer.SignObjectApproval(context.Background(), &msg)
//		if err != nil {
//			log.Errorw("failed to sign create object approval", "error", err)
//			errorDescription = InternalError
//			return
//		}
//		msg.PrimarySpApproval.Sig = approvalSignature
//		bz := types.ModuleCdc.MustMarshalJSON(&msg)
//		w.Header().Set(model.GnfdSignedApprovalMsgHeader, hex.EncodeToString(sdk.MustSortJSON(bz)))
//	default:
//		log.Errorw("not implement approval", "action", actionName)
//		errorDescription = NotImplementedError
//		return
//	}
//	w.Header().Set(model.GnfdRequestIDHeader, requestContext.requestID)
//}
//
//// challengeHandler handle challenge request
//func (g *Gateway) challengeHandler(w http.ResponseWriter, r *http.Request) {
//	var (
//		err              error
//		errorDescription *errorDescription
//		requestContext   *requestContext
//		addr             sdk.AccAddress
//
//		objectID         uint64
//		challengePrimary bool
//		segmentIdx       uint32
//		ecIdx            uint32
//		redundancyType   ptypes.RedundancyType
//		spAddress        string
//	)
//
//	requestContext = newRequestContext(r)
//	defer func() {
//		if errorDescription != nil {
//			_ = errorDescription.errorResponse(w, requestContext)
//		}
//		if errorDescription != nil && errorDescription.statusCode != http.StatusOK {
//			log.Errorf("action(%v) statusCode(%v) %v", challengeRouterName, errorDescription.statusCode, requestContext.generateRequestDetail())
//		} else {
//			log.Infof("action(%v) statusCode(200) %v", challengeRouterName, requestContext.generateRequestDetail())
//		}
//	}()
//
//	if addr, err = requestContext.verifySignature(); err != nil {
//		log.Errorw("failed to verify signature", "error", err)
//		errorDescription = SignatureNotMatch
//		return
//	}
//	if err = g.checkAuthorization(requestContext, addr); err != nil {
//		log.Errorw("failed to check authorization", "error", err)
//		errorDescription = UnauthorizedAccess
//		return
//	}
//
//	if objectID, err = util.StringToUin64(requestContext.request.Header.Get(model.GnfdObjectIDHeader)); err != nil {
//		log.Errorw("invalid object id", "object_id", requestContext.request.Header.Get(model.GnfdObjectIDHeader))
//		errorDescription = InvalidHeader
//		return
//	}
//
//	redundancyIndex, err := util.HeaderToInt64(requestContext.request.Header.Get(model.GnfdRedundancyIndexHeader))
//	if err != nil {
//		log.Errorw("invalid redundancy index", "redundancy_index", requestContext.request.Header.Get(model.GnfdRedundancyIndexHeader))
//		errorDescription = InvalidHeader
//		return
//	}
//	if redundancyIndex < 0 {
//		challengePrimary = true
//	} else {
//		ecIdx = uint32(redundancyIndex)
//	}
//	if segmentIdx, err = util.StringToUint32(requestContext.request.Header.Get(model.GnfdPieceIndexHeader)); err != nil {
//		log.Errorw("invalid segment idx", "segment_idx", requestContext.request.Header.Get(model.GnfdPieceIndexHeader))
//		errorDescription = InvalidHeader
//		return
//	}
//	spAddress = g.config.StorageProvider
//
//	req := &stypes.ChallengeServiceChallengePieceRequest{
//		TraceId:               requestContext.requestID,
//		ObjectId:              objectID,
//		ChallengePrimaryPiece: challengePrimary,
//		SegmentIdx:            segmentIdx,
//		EcIdx:                 ecIdx,
//		RedundancyType:        redundancyType,
//		StorageProviderId:     spAddress,
//	}
//	ctx := log.Context(context.Background(), req)
//	resp, err := g.challenge.ChallengePiece(ctx, req)
//	if err != nil {
//		log.Errorf("failed to challenge", "error", err)
//		errorDescription = InternalError
//		return
//	}
//	w.Header().Set(model.GnfdRequestIDHeader, requestContext.requestID)
//	w.Header().Set(model.GnfdObjectIDHeader, util.Uint64ToString(objectID))
//	w.Header().Set(model.GnfdIntegrityHashHeader, hex.EncodeToString(resp.IntegrityHash))
//	w.Header().Set(model.GnfdPieceHashHeader, util.BytesSliceToString(resp.PieceHash))
//	w.Write(resp.PieceData)
//}
