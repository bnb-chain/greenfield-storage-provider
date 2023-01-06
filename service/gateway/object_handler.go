package gateway

import (
	"net/http"
	"strconv"

	"github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

// putObjectTxHandler handle put object tx request, include steps:
// 1.check request params validation;
// 2.check request signature;
// 3.check account acl;
// 4.put object tx by uploaderClient.
func (g *Gateway) putObjectTxHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err              error
		errorDescription *errorDescription
		reqCtx           *requestContext
		opt              *putObjectTxOption
	)

	defer func() {
		statusCode := 200
		if errorDescription != nil {
			statusCode = errorDescription.statusCode
			_ = errorDescription.errorResponse(w, reqCtx)
		}
		if statusCode == 200 {
			log.Infof("action(%v) statusCode(%v) %v", "putObjectTx", statusCode, generateRequestDetail(reqCtx))
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "putObjectTx", statusCode, generateRequestDetail(reqCtx))
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

	// todo: check more params
	sizeStr := reqCtx.r.Header.Get(BFSContentLengthHeader)
	sizeInt, _ := strconv.Atoi(sizeStr)
	isPrivate, _ := strconv.ParseBool(reqCtx.r.Header.Get(BFSIsPrivateHeader))
	opt = &putObjectTxOption{
		reqCtx:      reqCtx,
		size:        uint64(sizeInt),
		contentType: reqCtx.r.Header.Get(BFSContentTypeHeader),
		checksum:    []byte(reqCtx.r.Header.Get(BFSChecksumHeader)),
		isPrivate:   isPrivate,
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
	w.Header().Set(BFSRequestIDHeader, reqCtx.requestID)
	w.Header().Set(BFSTransactionHashHeader, info.txHash)
}

// putObjectHandler handle put object request, include steps:
// 1.check request params validation;
// 2.check request signature;
// 3.check account acl;
// 4.put object data by uploaderClient.
func (g *Gateway) putObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err              error
		errorDescription *errorDescription
		reqCtx           *requestContext
		opt              *putObjectOption
	)

	defer func() {
		statusCode := 200
		if errorDescription != nil {
			statusCode = errorDescription.statusCode
			_ = errorDescription.errorResponse(w, reqCtx)
		}
		if statusCode == 200 {
			log.Infof("action(%v) statusCode(%v) %v", "putObject", statusCode, generateRequestDetail(reqCtx))
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "putObject", statusCode, generateRequestDetail(reqCtx))
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

	opt = &putObjectOption{
		reqCtx: reqCtx,
		txHash: []byte(reqCtx.r.Header.Get(BFSTransactionHashHeader)),
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
	w.Header().Set(BFSRequestIDHeader, reqCtx.requestID)
	w.Header().Set(ETagHeader, info.eTag)
}

// getObjectHandler handle get object request, include steps:
// 1.check request params validation;
// 2.check request signature;
// 3.check account acl;
// 4.get object data by downloaderClient.
func (g *Gateway) getObjectHandler(w http.ResponseWriter, r *http.Request) {
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
			log.Infof("action(%v) statusCode(%v) %v", "getObject", statusCode, generateRequestDetail(reqCtx))
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "getObject", statusCode, generateRequestDetail(reqCtx))
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
		reqCtx: reqCtx,
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
