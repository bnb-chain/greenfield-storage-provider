package gateway

import (
	"encoding/hex"
	"net/http"

	"github.com/bnb-chain/inscription-storage-provider/model"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// getAuthenticationHandler
func (g *Gateway) getAuthenticationHandler(w http.ResponseWriter, r *http.Request) {
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
			log.Infof("action(%v) statusCode(%v) %v", "getAuthentication", statusCode, generateRequestDetail(requestContext))
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "getAuthentication", statusCode, generateRequestDetail(requestContext))
		}
	}()

	requestContext = newRequestContext(r)
	if requestContext.bucketName == "" {
		errorDescription = InvalidBucketName
		return
	}

	if err := requestContext.verifySign(); err != nil {
		errorDescription = SignatureDoesNotMatch
		return
	}

	option := &getAuthenticationOption{
		requestContext: requestContext,
	}
	info, err := g.uploadProcessor.getAuthentication(option)
	if err != nil {
		errorDescription = InternalError
		return
	}
	w.Header().Set(model.BFSRequestIDHeader, requestContext.requestID)
	w.Header().Set(model.BFSPreSignatureHeader, hex.EncodeToString(info.preSignature))
}
