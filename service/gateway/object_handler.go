package gateway

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// getObjectHandler handle get object request
func (g *Gateway) getObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err              error
		readN, writeN    int
		size             int
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
			log.Infof("action(%v) statusCode(%v) %v", "getObject", statusCode, requestContext.generateRequestDetail())
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

	if err = requestContext.verifySignature(); err != nil {
		log.Infow("failed to verify signature", "error", err)
		errorDescription = SignatureDoesNotMatch
		return
	}
	if err = requestContext.verifyAuth(g.retriever); err != nil {
		errorDescription = UnauthorizedAccess
		return
	}

	req := &stypes.DownloaderServiceDownloaderObjectRequest{
		TraceId:    requestContext.requestID,
		BucketName: requestContext.bucketName,
		ObjectName: requestContext.objectName,
	}
	ctx := log.Context(context.Background(), req)
	stream, err := g.downloader.DownloaderObject(ctx, req)
	if err != nil {
		log.Warnf("failed to get object", "error", err)
		errorDescription = InternalError
		return
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Warnw("failed to read stream", "error", err)
			errorDescription = InternalError
			return
		}
		if res.ErrMessage != nil && res.ErrMessage.ErrCode != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
			log.Warnw("failed to read stream", "error", err)
			errorDescription = InternalError
			return
		}
		if readN = len(res.Data); readN == 0 {
			log.Warnw("download return empty data", "response", res)
			continue
		}
		if writeN, err = w.Write(res.Data); err != nil {
			log.Warnw("failed to read stream", "error", err)
			errorDescription = InternalError
			return
		}
		if readN != writeN {
			log.Warnw("failed to read stream", "error", err)
			errorDescription = InternalError
			return
		}
		size = size + writeN
	}
}

// putObjectHandler handle put object request
func (g *Gateway) putObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err              error
		errorDescription *errorDescription
		requestContext   *requestContext

		buf      = make([]byte, 65536)
		readN    int
		size     uint64
		hashBuf  = make([]byte, 65536)
		md5Hash  = md5.New()
		md5Value string
		objectID uint64
	)

	defer func() {
		statusCode := 200
		if errorDescription != nil {
			statusCode = errorDescription.statusCode
			_ = errorDescription.errorResponse(w, requestContext)
		}
		if statusCode == 200 {
			log.Infof("action(%v) statusCode(%v) %v", "putObject", statusCode, requestContext.generateRequestDetail())
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

	if err = requestContext.verifySignature(); err != nil {
		log.Warnw("failed to verify signature", "error", err)
		errorDescription = SignatureDoesNotMatch
		return
	}
	if err = requestContext.verifyAuth(g.retriever); err != nil {
		errorDescription = UnauthorizedAccess
		return
	}

	txHash, err := hex.DecodeString(requestContext.request.Header.Get(model.GnfdTransactionHashHeader))
	if err != nil && len(txHash) != hash.LengthHash {
		errorDescription = InvalidTxHash
		return
	}
	objectSize, _ := util.HeaderToUint64(requestContext.request.Header.Get(model.ContentLengthHeader))

	stream, err := g.uploader.UploadPayload(context.Background())
	if err != nil {
		log.Warnf("failed to put object", "error", err)
		errorDescription = InternalError
		return
	}
	for {
		readN, err = r.Body.Read(buf)
		if err != nil && err != io.EOF {
			log.Warnw("put object failed, due to reader", "err", err)
			errorDescription = InternalError
			return
		}
		if readN > 0 {
			req := &stypes.UploaderServiceUploadPayloadRequest{
				TraceId:        requestContext.requestID,
				TxHash:         txHash,
				PayloadData:    buf[:readN],
				BucketName:     requestContext.bucketName,
				ObjectName:     requestContext.objectName,
				ObjectSize:     objectSize,
				RedundancyType: util.HeaderToRedundancyType(requestContext.request.Header.Get(model.GnfdRedundancyTypeHeader)),
			}
			if err := stream.Send(req); err != nil {
				log.Warnw("put object failed, due to stream send", "err", err)
				errorDescription = InternalError
				return
			}
			size += uint64(readN)
			copy(hashBuf, buf[:readN])
			md5Hash.Write(hashBuf[:readN])
		}
		if err == io.EOF {
			if size == 0 {
				log.Warnw("put object failed, due to payload is empty")
				errorDescription = InvalidPayload
				return
			}
			resp, err := stream.CloseAndRecv()
			if err != nil {
				log.Warnw("put object failed, due to stream close", "err", err)
				errorDescription = InternalError
				return
			}
			if errMsg := resp.GetErrMessage(); errMsg != nil && errMsg.ErrCode != stypes.ErrCode_ERR_CODE_SUCCESS_UNSPECIFIED {
				log.Warnw("failed to grpc", "err", resp.ErrMessage)
				errorDescription = InternalError
				return
			}
			objectID = resp.GetObjectId()
			break
		}
	}
	md5Value = hex.EncodeToString(md5Hash.Sum(nil))
	w.Header().Set(model.GnfdRequestIDHeader, requestContext.requestID)
	w.Header().Set(model.ETagHeader, md5Value)
	w.Header().Set(model.GnfdObjectIDHeader, util.Uint64ToHeader(objectID))
}
