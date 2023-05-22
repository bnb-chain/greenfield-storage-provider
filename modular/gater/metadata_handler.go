package gater

import (
	"bytes"
	"encoding/base64"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/modular/retriever/types"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield/types/s3util"
	"net/http"
	"net/url"
	"strings"

	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/ethereum/go-ethereum/common"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// getUserBucketsHandler handle get object request
func (g *GateModular) getUserBucketsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		b      bytes.Buffer
		reqCtx *RequestContext
	)
	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get user buckets", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, err = NewRequestContext(r)
	if err != nil {
		return
	}

	if ok := common.IsHexAddress(r.Header.Get(model.GnfdUserAddressHeader)); !ok {
		log.Errorw("failed to check account id", "account_id", reqCtx.account, "error", err)
		err = ErrInvalidHeader
		return
	}

	resp, err := g.baseApp.GfSpClient().GetUserBuckets(reqCtx.Context(), r.Header.Get(model.GnfdUserAddressHeader))
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get user buckets", "error", err)
		return
	}

	grpcResponse := &types.GfSpGetUserBucketsResponse{
		Buckets: resp,
	}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get user buckets", "error", err)
		return
	}

	w.Header().Set(model.ContentTypeHeader, model.ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// listObjectsByBucketNameHandler handle list objects by bucket name request
func (g *GateModular) listObjectsByBucketNameHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                      error
		b                        bytes.Buffer
		maxKeys                  uint64
		reqCtx                   *RequestContext
		ok                       bool
		requestBucketName        string
		requestMaxKeys           string
		requestStartAfter        string
		requestContinuationToken string
		requestDelimiter         string
		requestPrefix            string
		continuationToken        string
		decodedContinuationToken []byte
		queryParams              url.Values
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to list objects by bucket name", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, err = NewRequestContext(r)
	if err != nil {
		return
	}

	queryParams = reqCtx.request.URL.Query()
	requestBucketName = reqCtx.bucketName
	requestMaxKeys = queryParams.Get(model.ListObjectsMaxKeysQuery)
	requestStartAfter = queryParams.Get(model.ListObjectsStartAfterQuery)
	requestContinuationToken = queryParams.Get(model.ListObjectsContinuationTokenQuery)
	requestDelimiter = queryParams.Get(model.ListObjectsDelimiterQuery)
	requestPrefix = queryParams.Get(model.ListObjectsPrefixQuery)

	if err = s3util.CheckValidBucketName(requestBucketName); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to check bucket name", "bucket_name", requestBucketName, "error", err)
		return
	}

	if requestMaxKeys != "" {
		if maxKeys, err = util.StringToUint64(requestMaxKeys); err != nil || maxKeys == 0 {
			log.CtxErrorw(reqCtx.Context(), "failed to parse or check maxKeys", "max_keys", requestMaxKeys, "error", err)
			err = ErrInvalidQuery
			return
		}
	}

	if requestStartAfter != "" {
		if err = s3util.CheckValidObjectName(requestStartAfter); err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to check startAfter", "start_after", requestStartAfter, "error", err)
			err = ErrInvalidQuery
			return
		}
	}

	if requestContinuationToken != "" {
		decodedContinuationToken, err = base64.StdEncoding.DecodeString(requestContinuationToken)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to check requestContinuationToken", "continuation_token", requestContinuationToken, "error", err)
			err = ErrInvalidQuery
			return
		}
		continuationToken = string(decodedContinuationToken)

		if err = s3util.CheckValidObjectName(continuationToken); err != nil {
			log.Errorw("failed to check requestContinuationToken", "continuation_token", continuationToken, "error", err)
			err = ErrInvalidQuery
			return
		}

		if !strings.HasPrefix(continuationToken, requestPrefix) {
			log.Errorw("failed to check requestContinuationToken", "continuation_token", continuationToken, "prefix", requestPrefix, "error", err)
			err = ErrInvalidQuery
			return
		}
	}

	if ok = checkValidObjectPrefix(requestPrefix); !ok {
		log.CtxErrorw(reqCtx.Context(), "failed to check requestPrefix", "prefix", requestPrefix, "error", err)
		err = ErrInvalidQuery
		return
	}

	if requestContinuationToken == "" {
		continuationToken = requestStartAfter
	}

	objects,
		keyCount,
		maxKeys,
		isTruncated,
		nextContinuationToken,
		name,
		prefix,
		delimiter,
		commonPrefixes,
		continuationToken,
		err :=
		g.baseApp.GfSpClient().ListObjectsByBucketName(
			reqCtx.Context(),
			requestBucketName,
			"",
			maxKeys,
			requestStartAfter,
			continuationToken,
			requestDelimiter,
			requestPrefix)
	if err != nil {
		log.Errorf("failed to list objects by bucket name", "error", err)
		return
	}

	grpcResponse := &types.GfSpListObjectsByBucketNameResponse{
		Objects:               objects,
		KeyCount:              keyCount,
		MaxKeys:               maxKeys,
		IsTruncated:           isTruncated,
		NextContinuationToken: nextContinuationToken,
		Name:                  name,
		Prefix:                prefix,
		Delimiter:             delimiter,
		CommonPrefixes:        commonPrefixes,
		ContinuationToken:     continuationToken,
	}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.Errorf("failed to list objects by bucket name", "error", err)
		return
	}

	w.Header().Set(model.ContentTypeHeader, model.ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// getObjectMetaHandler handle get object metadata request
func (g *GateModular) getObjectMetaHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		b      bytes.Buffer
		reqCtx *RequestContext
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get object meta", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, err = NewRequestContext(r)
	if err != nil {
		return
	}

	if err = s3util.CheckValidBucketName(reqCtx.bucketName); err != nil {
		log.Errorw("failed to check bucket name", "bucket_name", reqCtx.bucketName, "error", err)
		return
	}

	if err = s3util.CheckValidObjectName(reqCtx.objectName); err != nil {
		log.Errorw("failed to check object name", "object_name", reqCtx.objectName, "error", err)
		return
	}

	resp, err := g.baseApp.GfSpClient().GetObjectMeta(reqCtx.Context(), reqCtx.objectName, reqCtx.bucketName, true)
	if err != nil {
		log.Errorf("failed to get object meta", "error", err)
		return
	}

	grpcResponse := &types.GfSpGetObjectMetaResponse{
		Object: resp,
	}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.Errorf("failed to get object meta", "error", err)
		return
	}

	w.Header().Set(model.ContentTypeHeader, model.ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// getBucketMetaHandler handle get bucket metadata request
func (g *GateModular) getBucketMetaHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		b      bytes.Buffer
		reqCtx *RequestContext
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get bucket meta", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, err = NewRequestContext(r)
	if err != nil {
		return
	}

	if err = s3util.CheckValidBucketName(reqCtx.bucketName); err != nil {
		log.Errorw("failed to check bucket name", "bucket_name", reqCtx.bucketName, "error", err)
		return
	}

	bucket, streamRecord, err := g.baseApp.GfSpClient().GetBucketMeta(reqCtx.Context(), reqCtx.bucketName, true)
	if err != nil {
		log.Errorf("failed to get bucket metadata", "error", err)
		return
	}

	grpcResponse := &types.GfSpGetBucketMetaResponse{
		Bucket:       bucket,
		StreamRecord: streamRecord,
	}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.Errorf("failed to get bucket metadata", "error", err)
		return
	}

	w.Header().Set(model.ContentTypeHeader, model.ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}
