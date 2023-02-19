package gateway

import (
	"encoding/hex"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// getApprovalHandler
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
			log.Debugf("action(%v) statusCode(%v) %v", "getApproval", statusCode, requestContext.generateRequestDetail())
		} else {
			log.Errorw("action(%v) statusCode(%v) %v", "getApproval", statusCode, requestContext.generateRequestDetail())
		}
	}()

	requestContext = newRequestContext(r)
	if requestContext.bucketName == "" {
		errorDescription = InvalidBucketName
		return
	}

	if err = requestContext.verifySignature(); err != nil {
		errorDescription = SignatureDoesNotMatch
		log.Errorw("failed to verify signature", "error", err)
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
