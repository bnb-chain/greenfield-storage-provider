package gateway

import (
	"context"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

const (
	approvalRouterName  = "GetApproval"
	challengeRouterName = "Challenge"
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
	)

	defer func() {
		statusCode := 200
		if errorDescription != nil {
			statusCode = errorDescription.statusCode
			_ = errorDescription.errorResponse(w, requestContext)
		}
		if statusCode == 200 {
			log.Infof("action(%v) statusCode(%v) %v", "getApproval", statusCode, requestContext.generateRequestDetail())
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "getApproval", statusCode, requestContext.generateRequestDetail())
		}
	}()

	requestContext = newRequestContext(r)
	requestContext.actionName = requestContext.vars["action"]
	requestContext.bucketName = r.Header.Get(model.GnfdResourceHeader)
	fields := strings.Split(requestContext.bucketName, "/")
	if len(fields) >= 2 {
		requestContext.bucketName = fields[0]
		requestContext.objectName = strings.Join(fields[1:], "/")
	}
	if requestContext.bucketName == "" {
		errorDescription = InvalidBucketName
		return
	}

	if err = requestContext.verifySignature(); err != nil {
		errorDescription = SignatureDoesNotMatch
		log.Infow("failed to verify signature", "error", err)
		return
	}

	option := &getApprovalOption{
		requestContext: requestContext,
	}
	info, err := g.uploadProcessor.getApproval(option)
	if err != nil {
		errorDescription = InternalError
		return
	}
	w.Header().Set(model.GnfdRequestIDHeader, requestContext.requestID)
	w.Header().Set(model.GnfdPreSignatureHeader, hex.EncodeToString(info.preSignature))
}

// challengeHandler handle challenge request
func (g *Gateway) challengeHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err              error
		errorDescription *errorDescription
		requestContext   *requestContext

		objectID         uint64
		challengePrimary bool
		segmentIdx       uint32
		ecIdx            uint32
		redundancyType   ptypes.RedundancyType
		spAddress        string
	)

	defer func() {
		statusCode := 200
		if errorDescription != nil {
			statusCode = errorDescription.statusCode
			_ = errorDescription.errorResponse(w, requestContext)
		}
		if statusCode == 200 {
			log.Infof("action(%v) statusCode(%v) %v", "challenge", statusCode, requestContext.generateRequestDetail())
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "challenge", statusCode, requestContext.generateRequestDetail())
		}
	}()

	requestContext = newRequestContext(r)
	if err = requestContext.verifySignature(); err != nil {
		log.Infow("failed to verify signature", "error", err)
		errorDescription = SignatureDoesNotMatch
		return
	}

	if objectID, err = headerToUint64(requestContext.request.Header.Get(model.GnfdObjectIDHeader)); err != nil {
		log.Warnf("invalid object id", "object_id", requestContext.request.Header.Get(model.GnfdObjectIDHeader))
		errorDescription = InvalidHeader
		return
	}
	if challengePrimary, err = headerToBool(requestContext.request.Header.Get(model.GnfdIsChallengePrimaryHeader)); err != nil {
		log.Warnf("invalid challenge primary", "challenge_primary", requestContext.request.Header.Get(model.GnfdIsChallengePrimaryHeader))
		errorDescription = InvalidHeader
		return
	}
	if segmentIdx, err = headerToUint32(requestContext.request.Header.Get(model.GnfdPieceIndexHeader)); err != nil {
		log.Warnf("invalid segment idx", "segment_idx", requestContext.request.Header.Get(model.GnfdPieceIndexHeader))
		errorDescription = InvalidHeader
		return
	}
	if ecIdx, err = headerToUint32(requestContext.request.Header.Get(model.GnfdECIndexHeader)); err != nil {
		log.Warnf("invalid ec idx", "ec_idx", requestContext.request.Header.Get(model.GnfdECIndexHeader))
		errorDescription = InvalidHeader
		return
	}
	redundancyType = headerToRedundancyType(requestContext.request.Header.Get(model.GnfdRedundancyTypeHeader))
	spAddress = requestContext.request.Header.Get(model.GnfdStorageProviderHeader)

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
	pieceHashList := make([]string, len(resp.PieceHash))
	for _, h := range resp.PieceHash {
		pieceHashList = append(pieceHashList, hex.EncodeToString(h))
	}
	w.Header().Set(model.GnfdRequestIDHeader, requestContext.requestID)
	w.Header().Set(model.GnfdObjectIDHeader, uint64ToHeader(objectID))
	w.Header().Set(model.GnfdIntegrityHashHeader, hex.EncodeToString(resp.IntegrityHash))
	w.Header().Set(model.GnfdPieceHashHeader, stringSliceToHeader(pieceHashList))
	w.Write(resp.PieceData)
}
