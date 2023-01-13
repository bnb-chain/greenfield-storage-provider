package gateway

import (
	"encoding/hex"
	"net/http"

	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// getAuthenticationHandler
func (g *Gateway) getAuthenticationHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err              error
		errorDescription *errorDescription
		reqCtx           *requestContext
		opt              *getAuthenticationOption
	)

	defer func() {
		statusCode := 200
		if errorDescription != nil {
			statusCode = errorDescription.statusCode
			_ = errorDescription.errorResponse(w, reqCtx)
		}
		if statusCode == 200 {
			log.Infof("action(%v) statusCode(%v) %v", "getAuthentication", statusCode, generateRequestDetail(reqCtx))
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "getAuthentication", statusCode, generateRequestDetail(reqCtx))
		}
	}()

	reqCtx = newRequestContext(r)
	if reqCtx.bucket == "" {
		errorDescription = InvalidBucketName
		return
	}

	if err := reqCtx.verifySign(); err != nil {
		errorDescription = SignatureDoesNotMatch
		return
	}

	opt = &getAuthenticationOption{
		reqCtx: reqCtx,
	}
	info, err := g.uploadProcessor.getAuthentication(opt)
	if err != nil {
		errorDescription = InternalError
		return
	}
	w.Header().Set(BFSRequestIDHeader, reqCtx.requestID)
	w.Header().Set(BFSPreSignatureHeader, hex.EncodeToString(info.preSignature))
}
