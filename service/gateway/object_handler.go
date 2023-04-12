package gateway

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/bnb-chain/greenfield/types/s3util"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader/types"
	uploadertypes "github.com/bnb-chain/greenfield-storage-provider/service/uploader/types"
)

// getObjectHandler handles the get object request
func (gateway *Gateway) getObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		errDescription *errorDescription
		reqContext     *requestContext
		addr           sdk.AccAddress
		isRange        bool
		rangeStart     int64
		rangeEnd       int64
		readN, writeN  int
		size           int
		statusCode     = http.StatusOK
		ctx, cancel    = context.WithCancel(context.Background())
	)

	reqContext = newRequestContext(r)
	defer func() {
		cancel()
		if errDescription != nil {
			statusCode = errDescription.statusCode
			_ = errDescription.errorResponse(w, reqContext)
		}
		if statusCode == http.StatusOK || statusCode == http.StatusPartialContent {
			log.Infof("action(%v) statusCode(%v) %v", getObjectRouterName, statusCode, reqContext.generateRequestDetail())
		} else {
			log.Errorf("action(%v) statusCode(%v) %v", getObjectRouterName, statusCode, reqContext.generateRequestDetail())
		}
	}()

	if gateway.downloader == nil {
		log.Error("failed to get object due to not config downloader")
		errDescription = NotExistComponentError
		return
	}

	if err = s3util.CheckValidBucketName(reqContext.bucketName); err != nil {
		log.Errorw("failed to check bucket name", "bucket_name", reqContext.bucketName, "error", err)
		errDescription = InvalidBucketName
		return
	}
	if err = s3util.CheckValidObjectName(reqContext.objectName); err != nil {
		log.Errorw("failed to check object name", "object_name", reqContext.objectName, "error", err)
		errDescription = InvalidKey
		return
	}

	if addr, err = reqContext.verifySignature(); err != nil {
		log.Errorw("failed to verify signature", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	if err = gateway.checkAuthorization(reqContext, addr); err != nil {
		log.Errorw("failed to check authorization", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}

	isRange, rangeStart, rangeEnd = parseRange(reqContext.request.Header.Get(model.RangeHeader))
	if isRange && (rangeEnd < 0 || rangeEnd >= int64(reqContext.objectInfo.GetPayloadSize())) {
		rangeEnd = int64(reqContext.objectInfo.GetPayloadSize()) - 1
	}
	if isRange && (rangeStart < 0 || rangeEnd < 0 || rangeStart > rangeEnd) {
		errDescription = InvalidRange
		return
	}

	req := &types.GetObjectRequest{
		BucketInfo:  reqContext.bucketInfo,
		ObjectInfo:  reqContext.objectInfo,
		UserAddress: addr.String(),
		IsRange:     isRange,
		RangeStart:  uint64(rangeStart),
		RangeEnd:    uint64(rangeEnd),
	}
	// ctx := log.Context(context.Background(), req)
	stream, err := gateway.downloader.GetObject(ctx, req)
	if err != nil {
		log.Errorf("failed to get object", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorw("failed to read stream", "error", err)
			errDescription = makeErrorDescription(merrors.GRPCErrorToInnerError(err))
			return
		}

		if readN = len(resp.Data); readN == 0 {
			log.Errorw("failed to get object due to return empty data", "response", resp)
			continue
		}
		if isRange {
			statusCode = http.StatusPartialContent
			w.WriteHeader(statusCode)
			makeContentRangeHeader(w, rangeStart, rangeEnd)
		}
		if writeN, err = w.Write(resp.Data); err != nil {
			log.Errorw("failed to read stream", "error", err)
			errDescription = makeErrorDescription(err)
			return
		}
		if readN != writeN {
			log.Errorw("failed to read stream", "error", err)
			errDescription = InternalError
			return
		}
		size = size + writeN
	}
	w.Header().Set(model.GnfdRequestIDHeader, reqContext.requestID)
}

// putObjectHandler handles the put object request
func (gateway *Gateway) putObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		errDescription *errorDescription
		reqContext     *requestContext
		addr           sdk.AccAddress
		size           int
		readN          int
		buf            = make([]byte, model.DefaultStreamBufSize)
		hashBuf        = make([]byte, model.DefaultStreamBufSize)
		md5Hash        = md5.New()
		ctx, cancel    = context.WithCancel(context.Background())
	)

	reqContext = newRequestContext(r)
	defer func() {
		cancel()
		if errDescription != nil {
			_ = errDescription.errorResponse(w, reqContext)
		}
		if errDescription != nil && errDescription.statusCode != http.StatusOK {
			log.Errorf("action(%v) statusCode(%v) %v", putObjectRouterName, errDescription.statusCode, reqContext.generateRequestDetail())
		} else {
			log.Infof("action(%v) statusCode(200) %v", putObjectRouterName, reqContext.generateRequestDetail())
		}
	}()

	if gateway.uploader == nil {
		log.Error("failed to put object due to not config uploader")
		errDescription = NotExistComponentError
		return
	}

	if err = s3util.CheckValidBucketName(reqContext.bucketName); err != nil {
		log.Errorw("failed to check bucket name", "bucket_name", reqContext.bucketName, "error", err)
		errDescription = InvalidBucketName
		return
	}
	if err = s3util.CheckValidObjectName(reqContext.objectName); err != nil {
		log.Errorw("failed to check object name", "object_name", reqContext.objectName, "error", err)
		errDescription = InvalidKey
		return
	}

	if addr, err = reqContext.verifySignature(); err != nil {
		log.Errorw("failed to verify signature", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	if err = gateway.checkAuthorization(reqContext, addr); err != nil {
		log.Errorw("failed to check authorization", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}

	if reqContext.objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED {
		log.Errorw("failed to auth due to object status is not created",
			"object_status", reqContext.objectInfo.GetObjectStatus())
		err = merrors.ErrCheckObjectCreated
		errDescription = makeErrorDescription(err)
		return
	}

	stream, err := gateway.uploader.PutObject(ctx)
	if err != nil {
		log.Errorf("failed to put object", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	for {
		readN, err = r.Body.Read(buf)
		if err != nil && err != io.EOF {
			log.Errorw("failed to put object due to reader error", "error", err)
			errDescription = InternalError
			return
		}
		if readN > 0 {
			req := &uploadertypes.PutObjectRequest{
				ObjectInfo: reqContext.objectInfo,
				Payload:    buf[:readN],
			}
			if err := stream.Send(req); err != nil {
				log.Errorw("failed to put object due to stream send error", "error", err)
				errDescription = InternalError
				return
			}
			size += readN
			copy(hashBuf, buf[:readN])
			md5Hash.Write(hashBuf[:readN])
		}
		if err == io.EOF {
			if size == 0 {
				log.Errorw("failed to put object due to payload is empty")
				errDescription = InvalidPayload
				return
			}
			_, err = stream.CloseAndRecv()
			if err != nil {
				log.Errorw("failed to put object due to stream close", "error", err)
				errDescription = makeErrorDescription(err)
				return
			}
			// succeed to put object
			break
		}
	}

	w.Header().Set(model.GnfdRequestIDHeader, reqContext.requestID)
	w.Header().Set(model.ETagHeader, hex.EncodeToString(md5Hash.Sum(nil)))
}
