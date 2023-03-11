package gateway

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/bnb-chain/greenfield/types/s3util"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader/types"
	uploadertypes "github.com/bnb-chain/greenfield-storage-provider/service/uploader/types"
)

// getObjectHandler handle get object request
func (g *Gateway) getObjectHandler(w http.ResponseWriter, r *http.Request) {
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
	)

	reqContext = newRequestContext(r)
	defer func() {
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

	if g.downloader == nil {
		log.Errorw("failed to get object due to not config downloader")
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
	if err = g.checkAuthorization(reqContext, addr); err != nil {
		log.Errorw("failed to check authorization", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}

	isRange, rangeStart, rangeEnd = parseRange(reqContext.request.Header.Get(model.RangeHeader))

	if rangeStart > 0 && rangeEnd > 0 && rangeStart > rangeEnd {
		errDescription = InvalidRange
		return
	}

	req := &types.DownloaderObjectRequest{
		BucketInfo:  reqContext.bucketInfo,
		ObjectInfo:  reqContext.objectInfo,
		UserAddress: addr.String(),
		IsRange:     isRange,
		RangeStart:  rangeStart,
		RangeEnd:    rangeEnd,
	}
	ctx := log.Context(context.Background(), req)
	stream, err := g.downloader.DownloaderObject(ctx, req)
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
			errDescription = makeErrorDescription(err)
			return
		}

		if readN = len(resp.Data); readN == 0 {
			log.Errorw("failed to download due to return empty data", "response", resp)
			continue
		}
		if resp.IsValidRange {
			statusCode = http.StatusPartialContent
			w.WriteHeader(statusCode)
			generateContentRangeHeader(w, rangeStart, rangeEnd)
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

// putObjectHandler handle put object request
func (g *Gateway) putObjectHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		errDescription *errorDescription
		reqContext     *requestContext
		addr           sdk.AccAddress
		size           int
		readN          int
		buf            = make([]byte, model.StreamBufSize)
		hashBuf        = make([]byte, model.StreamBufSize)
		md5Hash        = md5.New()
	)

	reqContext = newRequestContext(r)
	defer func() {
		if errDescription != nil {
			_ = errDescription.errorResponse(w, reqContext)
		}
		if errDescription != nil && errDescription.statusCode != http.StatusOK {
			log.Errorf("action(%v) statusCode(%v) %v", putObjectRouterName, errDescription.statusCode, reqContext.generateRequestDetail())
		} else {
			log.Infof("action(%v) statusCode(200) %v", putObjectRouterName, reqContext.generateRequestDetail())
		}
	}()

	if g.uploader == nil {
		log.Errorw("failed to put object due to not config uploader")
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
	// TODO: maybe tx_hash will be used in the future
	_, _ = hex.DecodeString(reqContext.request.Header.Get(model.GnfdTransactionHashHeader))

	if addr, err = reqContext.verifySignature(); err != nil {
		log.Errorw("failed to verify signature", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	if err = g.checkAuthorization(reqContext, addr); err != nil {
		log.Errorw("failed to check authorization", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}

	stream, err := g.uploader.UploadObject(context.Background())
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
			req := &uploadertypes.UploadObjectRequest{
				ObjectInfo: reqContext.objectInfo,
				Payload:    buf[:readN],
			}
			if err := stream.Send(req); err != nil {
				log.Errorw("failed to put object failed due to stream send error", "error", err)
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
