package gateway

import (
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// createBucketHandler handle create bucket request, include steps:
// 1.check request params validation;
// 2.check request signature;
// 3.forward createBucket metadata to blockchain by chainClient.
func (g *Gateway) createBucketHandler(w http.ResponseWriter, r *http.Request) {
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
			log.Infof("action(%v) statusCode(%v) %v", "createBucket", statusCode, generateRequestDetail(requestContext))
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "createBucket", statusCode, generateRequestDetail(requestContext))
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

	option := &createBucketOption{
		requestContext: requestContext,
	}
	err = g.chain.createBucket(requestContext.bucketName, option)
	if err != nil {
		if err == errors.ErrDuplicateBucket {
			errorDescription = BucketAlreadyExists
			return
		}
		// else common.ErrInternalError
		errorDescription = InternalError
	}
}
