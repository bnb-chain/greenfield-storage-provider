package gateway

import (
	"context"
	"io"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// getObjectHandler handle get object request
func (g *Gateway) getObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err              error
		errorDescription *errorDescription
		requestContext   *requestContext
		addr             sdk.AccAddress

		isRange    bool
		rangeStart int64
		rangeEnd   int64

		readN, writeN int
		size          int
		statusCode    = http.StatusOK
	)

	requestContext = newRequestContext(r)
	defer func() {
		if errorDescription != nil {
			statusCode = errorDescription.statusCode
			_ = errorDescription.errorResponse(w, requestContext)
		}
		if statusCode == http.StatusOK || statusCode == http.StatusPartialContent {
			log.Infof("action(%v) statusCode(%v) %v", getObjectRouterName, statusCode, requestContext.generateRequestDetail())
		} else {
			log.Errorf("action(%v) statusCode(%v) %v", getObjectRouterName, statusCode, requestContext.generateRequestDetail())
		}
	}()

	// TODO: polish it by using common lib
	if requestContext.bucketName == "" {
		errorDescription = InvalidBucketName
		return
	}
	if requestContext.objectName == "" {
		errorDescription = InvalidKey
		return
	}

	if addr, err = requestContext.verifySignature(); err != nil {
		log.Errorw("failed to verify signature", "error", err)
		errorDescription = SignatureNotMatch
		return
	}
	if err = g.checkAuthorization(requestContext, addr); err != nil {
		log.Errorw("failed to check authorization", "error", err)
		errorDescription = UnauthorizedAccess
		return
	}

	isRange, rangeStart, rangeEnd = parseRange(requestContext.request.Header.Get(model.RangeHeader))

	if rangeStart > 0 && rangeEnd > 0 && rangeStart > rangeEnd {
		errorDescription = InvalidRange
		return
	}

	req := &stypes.DownloaderServiceDownloaderObjectRequest{
		TraceId:    requestContext.requestID,
		BucketName: requestContext.bucketName,
		ObjectName: requestContext.objectName,
		IsRange:    isRange,
		RangeStart: rangeStart,
		RangeEnd:   rangeEnd,
	}
	ctx := log.Context(context.Background(), req)
	stream, err := g.downloader.DownloaderObject(ctx, req)
	if err != nil {
		log.Errorf("failed to get object", "error", err)
		errorDescription = InternalError
		return
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorw("failed to read stream", "error", err)
			errorDescription = InternalError
			return
		}
		if resp.ErrMessage != nil && resp.ErrMessage.ErrCode != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
			log.Errorw("failed to read stream", "error", err)
			errorDescription = InternalError
			return
		}
		if readN = len(resp.Data); readN == 0 {
			log.Errorw("download return empty data", "response", resp)
			continue
		}
		if resp.IsValidRange {
			statusCode = http.StatusPartialContent
			w.WriteHeader(statusCode)
			generateContentRangeHeader(w, rangeStart, rangeEnd)
		}
		if writeN, err = w.Write(resp.Data); err != nil {
			log.Errorw("failed to read stream", "error", err)
			errorDescription = InternalError
			return
		}
		if readN != writeN {
			log.Errorw("failed to read stream", "error", err)
			errorDescription = InternalError
			return
		}
		size = size + writeN
	}
	w.Header().Set(model.GnfdRequestIDHeader, requestContext.requestID)
}

// putObjectHandler handle put object request
func (g *Gateway) putObjectHandler(w http.ResponseWriter, r *http.Request) {
	//var (
	//	err              error
	//	errorDescription *errorDescription
	//	requestContext   *requestContext
	//	addr             sdk.AccAddress
	//
	//	buf      = make([]byte, 65536)
	//	readN    int
	//	size     uint64
	//	hashBuf  = make([]byte, 65536)
	//	md5Hash  = md5.New()
	//	md5Value string
	//	objectID uint64
	//)
	//
	//requestContext = newRequestContext(r)
	//defer func() {
	//	if errorDescription != nil {
	//		_ = errorDescription.errorResponse(w, requestContext)
	//	}
	//	if errorDescription != nil && errorDescription.statusCode != http.StatusOK {
	//		log.Errorf("action(%v) statusCode(%v) %v", putObjectRouterName, errorDescription.statusCode, requestContext.generateRequestDetail())
	//	} else {
	//		log.Infof("action(%v) statusCode(200) %v", putObjectRouterName, requestContext.generateRequestDetail())
	//	}
	//}()
	//
	//// TODO: polish it by using common lib
	//if requestContext.bucketName == "" {
	//	errorDescription = InvalidBucketName
	//	return
	//}
	//if requestContext.objectName == "" {
	//	errorDescription = InvalidKey
	//	return
	//}
	//
	//if addr, err = requestContext.verifySignature(); err != nil {
	//	log.Errorw("failed to verify signature", "error", err)
	//	errorDescription = SignatureNotMatch
	//	return
	//}
	//if err = g.checkAuthorization(requestContext, addr); err != nil {
	//	log.Errorw("failed to check authorization", "error", err)
	//	errorDescription = UnauthorizedAccess
	//	return
	//}
	//
	//txHash, err := hex.DecodeString(requestContext.request.Header.Get(model.GnfdTransactionHashHeader))
	//if err != nil && len(txHash) != hash.LengthHash {
	//	log.Errorw("failed to parse tx_hash", "tx_hash", requestContext.request.Header.Get(model.GnfdTransactionHashHeader))
	//	errorDescription = InvalidTxHash
	//	return
	//}
	//objectSize, _ := util.HeaderToUint64(requestContext.request.Header.Get(model.ContentLengthHeader))
	//
	//stream, err := g.uploader.UploadPayload(context.Background())
	//if err != nil {
	//	log.Errorf("failed to put object", "error", err)
	//	errorDescription = InternalError
	//	return
	//}
	//for {
	//	readN, err = r.Body.Read(buf)
	//	if err != nil && err != io.EOF {
	//		log.Errorw("put object failed, due to reader error", "error", err)
	//		errorDescription = InternalError
	//		return
	//	}
	//	if readN > 0 {
	//		req := &stypes.UploaderServiceUploadPayloadRequest{
	//			TraceId:        requestContext.requestID,
	//			TxHash:         txHash,
	//			PayloadData:    buf[:readN],
	//			BucketName:     requestContext.bucketName,
	//			ObjectName:     requestContext.objectName,
	//			ObjectSize:     objectSize,
	//			RedundancyType: util.HeaderToRedundancyType(requestContext.request.Header.Get(model.GnfdRedundancyTypeHeader)),
	//		}
	//		if err := stream.Send(req); err != nil {
	//			log.Errorw("put object failed, due to stream send error", "error", err)
	//			errorDescription = InternalError
	//			return
	//		}
	//		size += uint64(readN)
	//		copy(hashBuf, buf[:readN])
	//		md5Hash.Write(hashBuf[:readN])
	//	}
	//	if err == io.EOF {
	//		if size == 0 {
	//			log.Errorw("put object failed, due to payload is empty")
	//			errorDescription = InvalidPayload
	//			return
	//		}
	//		resp, err := stream.CloseAndRecv()
	//		if err != nil {
	//			log.Errorw("put object failed, due to stream close", "error", err)
	//			errorDescription = InternalError
	//			return
	//		}
	//		if errMsg := resp.GetErrMessage(); errMsg != nil && errMsg.ErrCode != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
	//			log.Errorw("failed to receive grpc response error", "error", resp.ErrMessage)
	//			errorDescription = InternalError
	//			return
	//		}
	//		objectID = resp.GetObjectId()
	//		break
	//	}
	//}
	//md5Value = hex.EncodeToString(md5Hash.Sum(nil))
	//w.Header().Set(model.GnfdRequestIDHeader, requestContext.requestID)
	//w.Header().Set(model.ETagHeader, md5Value)
	//w.Header().Set(model.GnfdObjectIDHeader, util.Uint64ToHeader(objectID))
}
