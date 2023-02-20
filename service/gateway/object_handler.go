package gateway

import (
	"context"
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
			log.Debugf("action(%v) statusCode(%v) %v", "putObjectTx", statusCode, requestContext.generateRequestDetail())
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "putObjectTx", statusCode, requestContext.generateRequestDetail())
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

	// todo: check more params
	sizeStr := requestContext.request.Header.Get(model.GnfdContentLengthHeader)
	sizeInt, _ := strconv.Atoi(sizeStr)
	isPrivate, _ := strconv.ParseBool(requestContext.request.Header.Get(model.GnfdIsPrivateHeader))
	option := &putObjectTxOption{
		requestContext: requestContext,
		objectSize:     uint64(sizeInt),
		contentType:    requestContext.request.Header.Get(model.GnfdContentTypeHeader),
		checksum:       []byte(requestContext.request.Header.Get(model.GnfdChecksumHeader)),
		isPrivate:      isPrivate,
		redundancyType: requestContext.request.Header.Get(model.GnfdRedundancyTypeHeader),
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
			log.Debugf("action(%v) statusCode(%v) %v", "putObject", statusCode, requestContext.generateRequestDetail())
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "putObject", statusCode, requestContext.generateRequestDetail())
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

	txHash, err := hex.DecodeString(requestContext.request.Header.Get(model.GnfdTransactionHashHeader))
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
			//TODO:: update read quota
			log.Debugf("action(%v) statusCode(%v) %v", "getObject", statusCode, requestContext.generateRequestDetail())
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "getObject", statusCode, requestContext.generateRequestDetail())
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

	_, bktExist, objSerStatus, tokenEnough, spBkt, bucketId, readQuota, ownObj, err := g.chain.AuthDownloadObjectWithAccount(
		context.Background(),
		requestContext.bucketName,
		requestContext.objectName,
		r.Header.Get(model.GnfdAddress),
		g.config.SpId,
	)
	if err != nil {
		errorDescription = InternalError
		return
	}
	if !bktExist || !objSerStatus || !tokenEnough || !spBkt || !ownObj {
		errorDescription = AccessDenied
		return
	}
	//TODO:: query read quota enough
	_, _ = bucketId, readQuota

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
			log.Debugf("action(%v) statusCode(%v) %v", "putObjectV2", statusCode, requestContext.generateRequestDetail())
		} else {
			log.Warnf("action(%v) statusCode(%v) %v", "putObjectV2", statusCode, requestContext.generateRequestDetail())
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

	txHash, err := hex.DecodeString(requestContext.request.Header.Get(model.GnfdTransactionHashHeader))
	if err != nil && len(txHash) != hash.LengthHash {
		errorDescription = InvalidTxHash
		return
	}

	sizeStr := requestContext.request.Header.Get(model.ContentLengthHeader)
	sizeInt, _ := strconv.Atoi(sizeStr)
	option := &putObjectOption{
		requestContext: requestContext,
		txHash:         txHash,
		size:           uint64(sizeInt),
		redundancyType: requestContext.request.Header.Get(model.GnfdRedundancyTypeHeader),
	}

	_, bktExist, objInitStatue, tokenEnough, spBkt, ownObj, err := g.chain.AuthUploadObjectWithAccount(
		context.Background(),
		requestContext.bucketName,
		requestContext.objectName,
		r.Header.Get(model.GnfdAddress),
		g.config.SpId,
	)
	if err != nil {
		errorDescription = InternalError
		return
	}
	if !bktExist || !objInitStatue || !tokenEnough || !spBkt || !ownObj {
		errorDescription = AccessDenied
		return
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
