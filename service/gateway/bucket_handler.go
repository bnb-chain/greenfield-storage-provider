package gateway

import (
	"github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
	"net/http"
)

func (g *GatewayService) createBucketHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err              error
		errorDescription *errorDescription
		reqCtx           *requestContext
	)

	defer func() {
		statusCode := 200
		if errorDescription != nil {
			statusCode = errorDescription.statusCode
			_ = errorDescription.errorResponse(w, reqCtx)
		}
		if statusCode == 200 {
			log.Infof("action(%v) statusCode(%v) %v", "createBucket", statusCode, generateRequestDetail(reqCtx))
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "createBucket", statusCode, generateRequestDetail(reqCtx))
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

	var opt = &createBucketOption{
		reqCtx:   reqCtx,
		debugDir: g.config.DebugDir,
	}
	err = g.chain.createBucket(reqCtx.bucket, opt)
	if err != nil {
		if err == errors.ErrDuplicateBucket {
			errorDescription = BucketAlreadyExists
			return
		}
		// else common.ErrInternalError
		errorDescription = InternalError
	}
}
