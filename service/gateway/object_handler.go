package gateway

import (
	"github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
	"net/http"
)

func (g *GatewayService) putObjectTxHandler(w http.ResponseWriter, r *http.Request) {
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
	if reqCtx.object == "" {
		errorDescription = InvalidKey
		return
	}
	if err := reqCtx.verifySign(); err != nil {
		errorDescription = SignatureDoesNotMatch
		return
	}
	if err := reqCtx.verifyAuth(g.retriever); err != nil {
		errorDescription = UnauthorizedAccess
		return
	}

	var opt = &putObjectTxOption{
		reqCtx:   reqCtx,
		debugDir: g.config.DebugDir,
	}
	info, err := g.uploader.putObjectTx(reqCtx.object, opt)
	if err != nil {
		if err == errors.ErrDuplicateObject {
			errorDescription = ObjectAlreadyExists
			return
		}
		// else common.ErrInternalError
		errorDescription = InternalError
		return
	}
	// succeed ack
	w.Header().Set(RequestIDHeader, reqCtx.requestID)
	w.Header().Set(TransactionHashHeader, info.txHash)
}

func (g *GatewayService) putObjectHandler(w http.ResponseWriter, r *http.Request) {
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
	if reqCtx.object == "" {
		errorDescription = InvalidKey
		return
	}

	if err := reqCtx.verifySign(); err != nil {
		errorDescription = SignatureDoesNotMatch
		return
	}
	if err := reqCtx.verifyAuth(g.retriever); err != nil {
		errorDescription = UnauthorizedAccess
		return
	}

	var opt = &putObjectOption{
		reqCtx:   reqCtx,
		debugDir: g.config.DebugDir,
	}
	info, err := g.uploader.putObject(reqCtx.object, r.Body, opt)
	if err != nil {
		if err == errors.ErrObjectTxNotExist {
			errorDescription = ObjectTxNotFound
			return
		}
		// else common.ErrInternalError
		errorDescription = InternalError
		return
	}
	w.Header().Set(RequestIDHeader, reqCtx.requestID)
	w.Header().Set(ETagHeader, info.eTag)
}

func (g *GatewayService) getObjectHandler(w http.ResponseWriter, r *http.Request) {
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
	if reqCtx.object == "" {
		errorDescription = InvalidKey
		return
	}

	if err := reqCtx.verifySign(); err != nil {
		errorDescription = SignatureDoesNotMatch
		return
	}
	if err := reqCtx.verifyAuth(g.retriever); err != nil {
		errorDescription = UnauthorizedAccess
		return
	}

	var opt = &getObjectOption{
		reqCtx:   reqCtx,
		debugDir: g.config.DebugDir,
	}
	err = g.downloader.getObject(reqCtx.object, w, opt)
	if err != nil {
		if err == errors.ErrObjectNotExist {
			errorDescription = NoSuchKey
			return
		}
		// else common.ErrInternalError
		errorDescription = InternalError
		return
	}
	return
}
