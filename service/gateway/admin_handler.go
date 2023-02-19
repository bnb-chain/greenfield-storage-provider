package gateway

import (
	"context"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util"
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
		statusCode := http.StatusOK
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
		log.Infow("failed to verify signature", "error", err)
		errorDescription = SignatureDoesNotMatch
		return
	}

	req := &stypes.UploaderServiceGetApprovalRequest{
		TraceId: requestContext.requestID,
		Bucket:  requestContext.bucketName,
		Object:  requestContext.objectName,
		Action:  requestContext.actionName,
	}

	ctx := log.Context(context.Background(), req)
	resp, err := g.uploader.GetApproval(ctx, req)
	if err != nil {
		log.Warnf("failed to get approval", "error", err)
		errorDescription = InternalError
		return
	}
	w.Header().Set(model.GnfdRequestIDHeader, requestContext.requestID)
	w.Header().Set(model.GnfdPreSignatureHeader, hex.EncodeToString(resp.PreSignature))

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
		statusCode := http.StatusOK
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

	requestContext = newRequestContext(r)
	if err = requestContext.verifySignature(); err != nil {
		log.Warnw("failed to verify signature", "error", err)
		errorDescription = SignatureDoesNotMatch
		return
	}

	if objectID, err = util.HeaderToUint64(requestContext.request.Header.Get(model.GnfdObjectIDHeader)); err != nil {
		log.Warnw("invalid object id", "object_id", requestContext.request.Header.Get(model.GnfdObjectIDHeader))
		errorDescription = InvalidHeader
		return
	}
	if challengePrimary, err = util.HeaderToBool(requestContext.request.Header.Get(model.GnfdIsChallengePrimaryHeader)); err != nil {
		log.Warnw("invalid challenge primary", "challenge_primary", requestContext.request.Header.Get(model.GnfdIsChallengePrimaryHeader))
		errorDescription = InvalidHeader
		return
	}
	if segmentIdx, err = util.HeaderToUint32(requestContext.request.Header.Get(model.GnfdPieceIndexHeader)); err != nil {
		log.Warnw("invalid segment idx", "segment_idx", requestContext.request.Header.Get(model.GnfdPieceIndexHeader))
		errorDescription = InvalidHeader
		return
	}
	ecIdx, _ = util.HeaderToUint32(requestContext.request.Header.Get(model.GnfdECIndexHeader))
	redundancyType = util.HeaderToRedundancyType(requestContext.request.Header.Get(model.GnfdRedundancyTypeHeader))
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
	w.Header().Set(model.GnfdRequestIDHeader, requestContext.requestID)
	w.Header().Set(model.GnfdObjectIDHeader, util.Uint64ToHeader(objectID))
	w.Header().Set(model.GnfdIntegrityHashHeader, hex.EncodeToString(resp.IntegrityHash))
	w.Header().Set(model.GnfdPieceHashHeader, util.EncodePieceHash(resp.PieceHash))
	w.Write(resp.PieceData)
}
