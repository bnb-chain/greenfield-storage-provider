package gateway

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	metatypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
	"github.com/bnb-chain/greenfield/types/s3util"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"io"
	"net/http"

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
	// TODO: maybe tx_hash will be used in the future
	_, _ = hex.DecodeString(reqContext.request.Header.Get(model.GnfdTransactionHashHeader))

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

// getObjectByUniversalEndpointHandler handles the get object request sent by universal endpoint
func (gateway *Gateway) getObjectByUniversalEndpointHandler(w http.ResponseWriter, r *http.Request) {
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
	log.Debugw("getting bucketPrimarySpAddress:", "bucketPrimarySpAddress", bucketPrimarySpAddress)

	//if bucket not in the current sp, 302 redirect to the sp that contains the bucket
	if bucketPrimarySpAddress != gateway.config.SpOperatorAddress {
		log.Debugw("primary sp address not matched ",
			"bucketPrimarySpAddress", bucketPrimarySpAddress, "gateway.config.SpOperatorAddress", gateway.config.SpOperatorAddress,
		)
		getSpByAddressReq := &types.GetSpByAddressRequest{
			BucketName: bucketPrimarySpAddress,
		}

		getSpByAddressRes, err := gateway.downloader.GetSpByAddress(ctx, getSpByAddressReq)

		if err != nil || getSpByAddressRes == nil {
			log.Errorw("failed to get sp by address ", "sp_address", reqContext.bucketName, "error", err)
			errDescription = InvalidKey
			return
		}
		redirectUrl := getSpByAddressRes.Endpoint + "/download/" + reqContext.bucketName + "/" + reqContext.objectName

		log.Debugw("getting redirect url:", "redirectUrl", redirectUrl)

		http.Redirect(w, r, redirectUrl, 302)
		return
	}

	// In first phase, do not provide universal endpoint for private object
	getObjectInfoReq := &metatypes.GetObjectByObjectNameAndBucketNameRequest{
		ObjectName: reqContext.objectName,
		BucketName: reqContext.bucketName,
		IsFullList: false,
	}

	getObjectInfoRes, err := gateway.metadata.GetObjectByObjectNameAndBucketName(ctx, getObjectInfoReq)
	if err != nil || getObjectInfoRes == nil || getObjectInfoRes.GetObject() == nil || getObjectInfoRes.GetObject().GetObjectInfo() == nil {
		log.Errorw("failed to check object info", "object_name", reqContext.objectName, "error", err)
		errDescription = InvalidKey
		return
	}

	log.Debugw("getting bucketInfo and objectInfo:",
		"getBucketInfoRes.Bucket.BucketInfo.Id.Uint64()", getBucketInfoRes.GetBucket().GetBucketInfo().Id.Uint64(), "getBucketInfoRes.GetBucket().GetBucketInfo().GetChargedReadQuota()", getBucketInfoRes.GetBucket().GetBucketInfo().GetChargedReadQuota(),
		"getObjectInfoRes.GetObject().GetObjectInfo().Id.Uint64()", getObjectInfoRes.GetObject().GetObjectInfo().Id.Uint64(), "getObjectInfoRes.GetObject().GetObjectInfo().GetPayloadSize()", getObjectInfoRes.GetObject().GetObjectInfo().GetPayloadSize(),
	)

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
