package gateway

import (
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
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
		requestContext   *requestContext
	)

	defer func() {
		statusCode := 200
		if errorDescription != nil {
			statusCode = errorDescription.statusCode
			_ = errorDescription.errorResponse(w, requestContext)
		}
		if statusCode == 200 {
			log.Debugf("action(%v) statusCode(%v) %v", "putObjectTx", statusCode, generateRequestDetail(requestContext))
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "putObjectTx", statusCode, generateRequestDetail(requestContext))
		}
	}()

	requestContext = newRequestContext(r)
	if requestContext.bucketName == "" {
		errorDescription = InvalidBucketName
		return
	}
	if requestContext.objectName == "" {
		errorDescription = InvalidKey
		return
	}
	if err := requestContext.verifySign(); err != nil {
		errorDescription = SignatureDoesNotMatch
		return
	}
	if err := requestContext.verifyAuth(g.retriever); err != nil {
		errorDescription = UnauthorizedAccess
		return
	}

	// todo: check more params
	sizeStr := requestContext.r.Header.Get(model.GnfdContentLengthHeader)
	sizeInt, _ := strconv.Atoi(sizeStr)
	isPrivate, _ := strconv.ParseBool(requestContext.r.Header.Get(model.GnfdIsPrivateHeader))
	option := &putObjectTxOption{
		requestContext: requestContext,
		objectSize:     uint64(sizeInt),
		contentType:    requestContext.r.Header.Get(model.GnfdContentTypeHeader),
		checksum:       []byte(requestContext.r.Header.Get(model.GnfdChecksumHeader)),
		isPrivate:      isPrivate,
		redundancyType: requestContext.r.Header.Get(model.GnfdRedundancyTypeHeader),
	}
	info, err := g.uploadProcessor.putObjectTx(requestContext.objectName, option)
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
	w.Header().Set(model.GnfdRequestIDHeader, requestContext.requestID)
	w.Header().Set(model.GnfdTransactionHashHeader, hex.EncodeToString(info.txHash))
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
		requestContext   *requestContext
	)

	defer func() {
		statusCode := 200
		if errorDescription != nil {
			statusCode = errorDescription.statusCode
			_ = errorDescription.errorResponse(w, requestContext)
		}
		if statusCode == 200 {
			log.Debugf("action(%v) statusCode(%v) %v", "putObject", statusCode, generateRequestDetail(requestContext))
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "putObject", statusCode, generateRequestDetail(requestContext))
		}
	}()

	requestContext = newRequestContext(r)
	if requestContext.bucketName == "" {
		errorDescription = InvalidBucketName
		return
	}
	if requestContext.objectName == "" {
		errorDescription = InvalidKey
		return
	}

	if err := requestContext.verifySign(); err != nil {
		errorDescription = SignatureDoesNotMatch
		return
	}
	if err := requestContext.verifyAuth(g.retriever); err != nil {
		errorDescription = UnauthorizedAccess
		return
	}
	txHash, err := hex.DecodeString(requestContext.r.Header.Get(model.GnfdTransactionHashHeader))
	if err != nil && len(txHash) != hash.LengthHash {
		errorDescription = InvalidTxHash
		return
	}

	option := &putObjectOption{
		requestContext: requestContext,
		txHash:         txHash,
	}

	info, err := g.uploadProcessor.putObject(requestContext.objectName, r.Body, option)
	if err != nil {
		if err == errors.ErrObjectTxNotExist {
			errorDescription = ObjectTxNotFound
			return
		}
		// else common.ErrInternalError
		errorDescription = InternalError
		return
	}
	w.Header().Set(model.GnfdRequestIDHeader, requestContext.requestID)
	w.Header().Set(model.ETagHeader, info.eTag)
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
		requestContext   *requestContext
	)

	defer func() {
		statusCode := 200
		if errorDescription != nil {
			statusCode = errorDescription.statusCode
			_ = errorDescription.errorResponse(w, requestContext)
		}
		if statusCode == 200 {
			log.Debugf("action(%v) statusCode(%v) %v", "getObject", statusCode, generateRequestDetail(requestContext))
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "getObject", statusCode, generateRequestDetail(requestContext))
		}
	}()

	requestContext = newRequestContext(r)
	if requestContext.bucketName == "" {
		errorDescription = InvalidBucketName
		return
	}
	if requestContext.objectName == "" {
		errorDescription = InvalidKey
		return
	}

	if err := requestContext.verifySign(); err != nil {
		errorDescription = SignatureDoesNotMatch
		return
	}
	if err := requestContext.verifyAuth(g.retriever); err != nil {
		errorDescription = UnauthorizedAccess
		return
	}

	option := &getObjectOption{
		requestContext: requestContext,
	}
	err = g.downloadProcessor.getObject(requestContext.objectName, w, option)
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

// putObjectV2Handler handle put object request v2.
func (g *Gateway) putObjectV2Handler(w http.ResponseWriter, r *http.Request) {
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
			log.Debugf("action(%v) statusCode(%v) %v", "putObjectV2", statusCode, generateRequestDetail(requestContext))
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "putObjectV2", statusCode, generateRequestDetail(requestContext))
		}
	}()

	requestContext = newRequestContext(r)
	if requestContext.bucketName == "" {
		errorDescription = InvalidBucketName
		return
	}
	if requestContext.objectName == "" {
		errorDescription = InvalidKey
		return
	}

	if err := requestContext.verifySign(); err != nil {
		errorDescription = SignatureDoesNotMatch
		return
	}
	if err := requestContext.verifyAuth(g.retriever); err != nil {
		errorDescription = UnauthorizedAccess
		return
	}

	txHash, err := hex.DecodeString(requestContext.r.Header.Get(model.GnfdTransactionHashHeader))
	if err != nil && len(txHash) != hash.LengthHash {
		errorDescription = InvalidTxHash
		return
	}

	sizeStr := requestContext.r.Header.Get(model.ContentLengthHeader)
	sizeInt, _ := strconv.Atoi(sizeStr)
	option := &putObjectOption{
		requestContext: requestContext,
		txHash:         txHash,
		size:           uint64(sizeInt),
		redundancyType: requestContext.r.Header.Get(model.GnfdRedundancyTypeHeader),
	}

	info, err := g.uploadProcessor.putObjectV2(requestContext.objectName, r.Body, option)
	if err != nil {
		if err == errors.ErrObjectTxNotExist {
			errorDescription = ObjectTxNotFound
			return
		}
		if err == errors.ErrObjectIsEmpty {
			errorDescription = InvalidPayload
			return
		}
		// else common.ErrInternalError
		errorDescription = InternalError
		return
	}
	w.Header().Set(model.GnfdRequestIDHeader, requestContext.requestID)
	w.Header().Set(model.ETagHeader, info.eTag)
}
