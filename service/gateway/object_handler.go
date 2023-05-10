package gateway

import (
	"context"
	"crypto/md5"
	"io"
	"net/http"
	"net/url"

	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield/types/s3util"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader/types"
	metatypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
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

	if addr, err = gateway.verifySignature(reqContext); err != nil {
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
	stream, err := gateway.downloader.GetObject(ctx, req)
	if err != nil {
		log.Errorf("failed to get object", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}
	w.Header().Set(model.GnfdRequestIDHeader, reqContext.requestID)
	w.Header().Set(model.ContentTypeHeader, reqContext.objectInfo.GetContentType())
	if !isRange {
		w.Header().Set(model.ContentLengthHeader, util.Uint64ToString(reqContext.objectInfo.GetPayloadSize()))
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

	if addr, err = gateway.verifySignature(reqContext); err != nil {
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
	w.Header().Set(model.GnfdRequestIDHeader, reqContext.requestID)
	// Greenfield has an integrity hash, so there is no need for an etag
	// w.Header().Set(model.ETagHeader, hex.EncodeToString(md5Hash.Sum(nil)))

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
}

// getObjectByUniversalEndpointHandler handles the get object request sent by universal endpoint
func (gateway *Gateway) getObjectByUniversalEndpointHandler(w http.ResponseWriter, r *http.Request, isDownload bool) {
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
		redirectUrl    string
	)

	reqContext = newRequestContext(r)
	defer func() {
		cancel()
		if errDescription != nil {
			statusCode = errDescription.statusCode
			_ = errDescription.errorResponse(w, reqContext)
		}

		var getObjectByUniversalEndpointName string
		if isDownload {
			getObjectByUniversalEndpointName = downloadObjectByUniversalEndpointName
		} else {
			getObjectByUniversalEndpointName = viewObjectByUniversalEndpointName
		}

		if statusCode == http.StatusOK || statusCode == http.StatusPartialContent {
			log.Infof("action(%v) statusCode(%v) %v", getObjectByUniversalEndpointName, statusCode, reqContext.generateRequestDetail())
		} else {
			log.Errorf("action(%v) statusCode(%v) %v", getObjectByUniversalEndpointName, statusCode, reqContext.generateRequestDetail())
		}
	}()

	if gateway.downloader == nil {
		log.Error("failed to get object by universal endpoint due to not config downloader")
		errDescription = NotExistComponentError
		return
	}

	escapedObjectName, err := url.PathUnescape(reqContext.objectName)
	if err != nil {
		log.Errorw("failed to unescape object name ", "object_name", reqContext.objectName, "error", err)
		errDescription = InvalidKey
		return
	}

	if err = s3util.CheckValidBucketName(reqContext.bucketName); err != nil {
		log.Errorw("failed to check bucket name", "bucket_name", reqContext.bucketName, "error", err)
		errDescription = InvalidBucketName
		return
	}
	if err = s3util.CheckValidObjectName(escapedObjectName); err != nil {
		log.Errorw("failed to check object name", "object_name", escapedObjectName, "error", err)
		errDescription = InvalidKey
		return
	}

	getBucketInfoReq := &metatypes.GetBucketByBucketNameRequest{
		BucketName: reqContext.bucketName,
		IsFullList: true,
	}

	getBucketInfoRes, err := gateway.metadata.GetBucketByBucketName(ctx, getBucketInfoReq)
	if err != nil || getBucketInfoRes == nil || getBucketInfoRes.GetBucket() == nil || getBucketInfoRes.GetBucket().GetBucketInfo() == nil {
		log.Errorw("failed to check bucket info", "bucket_name", reqContext.bucketName, "error", err)
		errDescription = InvalidKey
		return
	}

	bucketPrimarySpAddress := getBucketInfoRes.GetBucket().GetBucketInfo().GetPrimarySpAddress()

	//if bucket not in the current sp, 302 redirect to the sp that contains the bucket
	if bucketPrimarySpAddress != gateway.config.SpOperatorAddress {
		log.Debugw("primary sp address not matched ",
			"bucketPrimarySpAddress", bucketPrimarySpAddress, "gateway.config.SpOperatorAddress", gateway.config.SpOperatorAddress,
		)

		endpoint, err := gateway.downloader.GetEndpointBySpAddress(ctx, bucketPrimarySpAddress)

		if err != nil || endpoint == "" {
			log.Errorw("failed to get endpoint by address ", "sp_address", reqContext.bucketName, "error", err)
			errDescription = InvalidAddress
			return
		}

		if isDownload {
			redirectUrl = endpoint + "/download/" + reqContext.bucketName + "/" + reqContext.objectName
		} else {
			redirectUrl = endpoint + "/view/" + reqContext.bucketName + "/" + reqContext.objectName
		}

		log.Debugw("getting redirect url:", "redirectUrl", redirectUrl)

		http.Redirect(w, r, redirectUrl, 302)
		return
	}

	// In first phase, do not provide universal endpoint for private object
	getObjectInfoReq := &metatypes.GetObjectByObjectNameAndBucketNameRequest{
		ObjectName: escapedObjectName,
		BucketName: reqContext.bucketName,
		IsFullList: false,
	}

	getObjectInfoRes, err := gateway.metadata.GetObjectByObjectNameAndBucketName(ctx, getObjectInfoReq)
	if err != nil || getObjectInfoRes == nil || getObjectInfoRes.GetObject() == nil || getObjectInfoRes.GetObject().GetObjectInfo() == nil {
		log.Errorw("failed to check object info", "object_name", escapedObjectName, "error", err)
		errDescription = InvalidKey
		return
	}

	if getObjectInfoRes.GetObject().GetObjectInfo().GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED {
		log.Errorw("object is not sealed",
			"status", getObjectInfoRes.GetObject().GetObjectInfo().GetObjectStatus())
		errDescription = InvalidObjectState
		return
	}

	getObjectReq := &types.GetObjectRequest{
		BucketInfo:  getBucketInfoRes.GetBucket().GetBucketInfo(),
		ObjectInfo:  getObjectInfoRes.GetObject().GetObjectInfo(),
		UserAddress: addr.String(),
		IsRange:     isRange,
		RangeStart:  uint64(rangeStart),
		RangeEnd:    uint64(rangeEnd),
	}

	stream, err := gateway.downloader.GetObject(ctx, getObjectReq)
	if err != nil {
		log.Errorf("failed to get object", "error", err)
		errDescription = makeErrorDescription(err)
		return
	}

	if isDownload {
		w.Header().Set(model.ContentDispositionHeader, model.ContentDispositionAttachmentValue+"; filename=\""+escapedObjectName+"\"")
	} else {
		w.Header().Set(model.ContentDispositionHeader, model.ContentDispositionInlineValue)
	}
	w.Header().Set(model.GnfdRequestIDHeader, reqContext.requestID)

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
}

// downloadObjectByUniversalEndpointHandler handles the download object request sent by universal endpoint
func (gateway *Gateway) downloadObjectByUniversalEndpointHandler(w http.ResponseWriter, r *http.Request) {
	gateway.getObjectByUniversalEndpointHandler(w, r, true)
}

// viewObjectByUniversalEndpointHandler handles the view object request sent by universal endpoint
func (gateway *Gateway) viewObjectByUniversalEndpointHandler(w http.ResponseWriter, r *http.Request) {
	gateway.getObjectByUniversalEndpointHandler(w, r, false)
}
