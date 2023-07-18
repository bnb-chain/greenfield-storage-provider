package gater

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield/types/resource"
	resource_types "github.com/bnb-chain/greenfield/types/resource"
	"github.com/bnb-chain/greenfield/types/s3util"
	payment_types "github.com/bnb-chain/greenfield/x/payment/types"
	permission_types "github.com/bnb-chain/greenfield/x/permission/types"
	storage_types "github.com/bnb-chain/greenfield/x/storage/types"
	virtual_types "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

const (
	MaximumGetGroupListLimit         = 1000
	MaximumGetGroupListOffset        = 100000
	MaximumListObjectsAndBucketsSize = 1000
	DefaultGetGroupListLimit         = 50
	DefaultGetGroupListOffset        = 0
	HandlerSuccess                   = "success"
	HandlerFailure                   = "failure"
	HandlerLevel                     = "handler"
)

func MetadataHandlerFailureMetrics(err error, startTime time.Time, handlerName string) {
	gfspErr := gfsperrors.MakeGfSpError(err)
	code := gfspErr.HttpStatusCode
	metrics.ReqCounter.WithLabelValues(GatewayTotalFailure).Inc()
	metrics.ReqTime.WithLabelValues(GatewayTotalFailure).Observe(time.Since(startTime).Seconds())
	metrics.MetadataReqTime.WithLabelValues(HandlerFailure, HandlerLevel, handlerName, strconv.Itoa(int(code))).Observe(time.Since(startTime).Seconds())
}

func MetadataHandlerSuccessMetrics(startTime time.Time, handlerName string) {
	metrics.ReqCounter.WithLabelValues(GatewayTotalSuccess).Inc()
	metrics.ReqTime.WithLabelValues(GatewayTotalSuccess).Observe(time.Since(startTime).Seconds())
	metrics.MetadataReqTime.WithLabelValues(HandlerSuccess, HandlerLevel, handlerName, strconv.Itoa(http.StatusOK)).Observe(time.Since(startTime).Seconds())
}

// getUserBucketsHandler handle get object request
func (g *GateModular) getUserBucketsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		requestIncludeRemoved string
		includedRemoved       bool
		queryParams           url.Values
		err                   error
		b                     bytes.Buffer
		reqCtx                *RequestContext
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		handlerName := mux.CurrentRoute(r).GetName()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get user buckets", reqCtx.String())
			MakeErrorResponse(w, err)
			MetadataHandlerFailureMetrics(err, startTime, handlerName)
		} else {
			MetadataHandlerSuccessMetrics(startTime, handlerName)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	requestIncludeRemoved = queryParams.Get(ListObjectsIncludeRemovedQuery)

	if ok := common.IsHexAddress(r.Header.Get(GnfdUserAddressHeader)); !ok {
		log.Errorw("failed to check account id", "account_id", reqCtx.account, "error", err)
		err = ErrInvalidHeader
		return
	}

	if requestIncludeRemoved != "" {
		includedRemoved, err = strconv.ParseBool(requestIncludeRemoved)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to parse requestRemoved", "request_include_removed", requestIncludeRemoved, "error", err)
			err = ErrInvalidQuery
			return
		}
	}

	resp, err := g.baseApp.GfSpClient().GetUserBuckets(reqCtx.Context(), r.Header.Get(GnfdUserAddressHeader), includedRemoved)
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

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
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
		requestIncludeRemoved    string
		continuationToken        string
		includedRemoved          bool
		decodedContinuationToken []byte
		queryParams              url.Values
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		handlerName := mux.CurrentRoute(r).GetName()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to list objects by bucket name", reqCtx.String())
			MakeErrorResponse(w, err)
			MetadataHandlerFailureMetrics(err, startTime, handlerName)
		} else {
			MetadataHandlerSuccessMetrics(startTime, handlerName)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	requestBucketName = reqCtx.bucketName
	requestMaxKeys = queryParams.Get(ListObjectsMaxKeysQuery)
	requestStartAfter = queryParams.Get(ListObjectsStartAfterQuery)
	requestContinuationToken = queryParams.Get(ListObjectsContinuationTokenQuery)
	requestDelimiter = queryParams.Get(ListObjectsDelimiterQuery)
	requestPrefix = queryParams.Get(ListObjectsPrefixQuery)
	requestIncludeRemoved = queryParams.Get(ListObjectsIncludeRemovedQuery)

	if requestDelimiter != "" && requestDelimiter != "/" {
		log.CtxErrorw(reqCtx.Context(), "failed to check delimiter", "delimiter", requestDelimiter, "error", err)
		err = ErrInvalidQuery
		return
	}

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

	if requestIncludeRemoved != "" {
		includedRemoved, err = strconv.ParseBool(requestIncludeRemoved)
		if err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to parse requestRemoved", "request_include_removed", requestIncludeRemoved, "error", err)
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
			requestPrefix,
			includedRemoved)
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

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// getObjectMetaHandler handle get object metadata request
func (g *GateModular) getObjectMetaHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		b      bytes.Buffer
		reqCtx *RequestContext
	)

	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		handlerName := mux.CurrentRoute(r).GetName()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get object meta", reqCtx.String())
			MakeErrorResponse(w, err)
			MetadataHandlerFailureMetrics(err, startTime, handlerName)
		} else {
			MetadataHandlerSuccessMetrics(startTime, handlerName)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

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

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// getBucketMetaHandler handle get bucket metadata request
func (g *GateModular) getBucketMetaHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		b      bytes.Buffer
		reqCtx *RequestContext
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		handlerName := mux.CurrentRoute(r).GetName()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get bucket meta", reqCtx.String())
			MakeErrorResponse(w, err)
			MetadataHandlerFailureMetrics(err, startTime, handlerName)
		} else {
			MetadataHandlerSuccessMetrics(startTime, handlerName)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

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

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// verifyPermissionHandler handle verify permission request
func (g *GateModular) verifyPermissionHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		operator    string
		objectName  string
		actionType  string
		action      int
		b           bytes.Buffer
		queryParams url.Values
		effect      *permission_types.Effect
		reqCtx      *RequestContext
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		handlerName := mux.CurrentRoute(r).GetName()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to verify permission", reqCtx.String())
			MakeErrorResponse(w, err)
			MetadataHandlerFailureMetrics(err, startTime, handlerName)
		} else {
			MetadataHandlerSuccessMetrics(startTime, handlerName)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)
	if err != nil {
		return
	}

	queryParams = reqCtx.request.URL.Query()
	objectName = queryParams.Get(VerifyPermissionObjectQuery)
	operator = reqCtx.vars[VerifyPermissionOperator]
	actionType = reqCtx.vars[VerifyPermissionActionType]

	if err = s3util.CheckValidBucketName(reqCtx.bucketName); err != nil {
		log.Errorw("failed to check bucket name", "bucket_name", reqCtx.bucketName, "error", err)
		return
	}

	if objectName != "" {
		if err = s3util.CheckValidObjectName(objectName); err != nil {
			log.Errorw("failed to check object name", "object_name", objectName, "error", err)
			return
		}
	}

	if ok := common.IsHexAddress(operator); !ok {
		log.Errorw("failed to check operator", "operator", operator, "error", err)
		return
	}

	if actionType == "" {
		log.Errorw("failed to check action type", "action_type", actionType, "error", err)
		return
	}

	action, err = strconv.Atoi(actionType)
	if err != nil {
		log.Errorw("failed to check action type", "action_type", actionType, "error", err)
		return
	}

	effect, err = g.baseApp.GfSpClient().VerifyPermission(reqCtx.Context(), operator, reqCtx.bucketName, objectName, permission_types.ActionType(action))
	if err != nil {
		log.Errorf("failed to verify permission", "error", err)
		return
	}

	grpcResponse := &storage_types.QueryVerifyPermissionResponse{Effect: *effect}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.Errorf("failed to verify permission", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// getGroupListHandler handle get group list request
func (g *GateModular) getGroupListHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		b           bytes.Buffer
		queryParams url.Values
		limitStr    string
		offsetStr   string
		name        string
		prefix      string
		sourceType  string
		limit       int64
		offset      int64
		count       int64
		groups      []*types.Group
		reqCtx      *RequestContext
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		handlerName := mux.CurrentRoute(r).GetName()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get group list", reqCtx.String())
			MakeErrorResponse(w, err)
			MetadataHandlerFailureMetrics(err, startTime, handlerName)
		} else {
			MetadataHandlerSuccessMetrics(startTime, handlerName)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)
	if err != nil {
		return
	}

	queryParams = reqCtx.request.URL.Query()
	sourceType = queryParams.Get(GetGroupListSourceTypeQuery)
	limitStr = queryParams.Get(GetGroupListLimitQuery)
	offsetStr = queryParams.Get(GetGroupListOffsetQuery)
	name = queryParams.Get(GetGroupListNameQuery)
	prefix = queryParams.Get(GetGroupListPrefixQuery)

	if name == "" {
		log.CtxErrorw(reqCtx.Context(), "failed to check name", "name", name, "error", err)
		err = ErrInvalidQuery
		return
	}

	if prefix == "" {
		log.CtxErrorw(reqCtx.Context(), "failed to check prefix", "prefix", name, "error", err)
		err = ErrInvalidQuery
		return
	}

	if limitStr != "" {
		if limit, err = util.StringToInt64(limitStr); err != nil || limit <= 0 {
			log.CtxErrorw(reqCtx.Context(), "failed to parse or check limit", "limit", limitStr, "error", err)
			err = ErrInvalidQuery
			return
		}
		if limit > 1000 {
			limit = MaximumGetGroupListLimit
		}
	} else {
		limit = DefaultGetGroupListLimit
	}

	if offsetStr != "" {
		if offset, err = util.StringToInt64(offsetStr); err != nil || offset < 0 || offset > MaximumGetGroupListOffset {
			log.CtxErrorw(reqCtx.Context(), "failed to parse or check offset", "offset", offsetStr, "error", err)
			err = ErrInvalidQuery
			return
		}
	} else {
		offset = DefaultGetGroupListOffset
	}

	if sourceType != "" {
		if _, ok := storage_types.SourceType_value[sourceType]; !ok {
			log.CtxErrorw(reqCtx.Context(), "failed to parse or check source type", "source-type", sourceType, "error", err)
			err = ErrInvalidQuery
			return
		}
	}

	groups, count, err = g.baseApp.GfSpClient().GetGroupList(reqCtx.Context(), name, prefix, sourceType, limit, offset, false)
	if err != nil {
		log.Errorf("failed to get group list", "error", err)
		return
	}

	grpcResponse := &types.GfSpGetGroupListResponse{
		Groups: groups,
		Count:  count,
	}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.Errorf("failed to get group list", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// listObjectsByObjectIDHandler list objects by object ids
func (g *GateModular) listObjectsByObjectIDHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		buf         bytes.Buffer
		objects     map[uint64]*types.Object
		objectIDMap map[uint64]bool
		ok          bool
		objectIDs   bsdb.ObjectIDs
		reqCtx      *RequestContext
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		handlerName := mux.CurrentRoute(r).GetName()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to list objects by ids", reqCtx.String())
			MakeErrorResponse(w, err)
			MetadataHandlerFailureMetrics(err, startTime, handlerName)
		} else {
			MetadataHandlerSuccessMetrics(startTime, handlerName)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	err = json.NewDecoder(r.Body).Decode(&objectIDs)
	if err != nil {
		log.Errorf("failed to parse object ids", "error", err)
		err = ErrInvalidQuery
		return
	}

	if len(objectIDs.IDs) == 0 || len(objectIDs.IDs) > MaximumListObjectsAndBucketsSize {
		log.Errorf("failed to check ids", "error", err)
		err = ErrInvalidQuery
		return
	}

	objectIDMap = make(map[uint64]bool)
	for _, id := range objectIDs.IDs {
		if _, ok = objectIDMap[id]; ok {
			// repeat id keys in request
			log.Errorf("failed to check ids", "error", err)
			err = ErrInvalidQuery
			return
		}
		objectIDMap[id] = true
	}

	objects, err = g.baseApp.GfSpClient().ListObjectsByObjectID(reqCtx.Context(), objectIDs.IDs, false)
	if err != nil {
		log.Errorf("failed to list objects by ids", "error", err)
		return
	}
	grpcResponse := &types.GfSpListObjectsByObjectIDResponse{Objects: objects}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&buf, grpcResponse); err != nil {
		log.Errorf("failed to list objects by ids", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(buf.Bytes())
}

// listBucketsByBucketIDHandler list buckets by bucket ids
func (g *GateModular) listBucketsByBucketIDHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		buf         bytes.Buffer
		buckets     map[uint64]*types.Bucket
		bucketIDMap map[uint64]bool
		ok          bool
		bucketIDs   bsdb.BucketIDs
		reqCtx      *RequestContext
	)
	startTime := time.Now()
	defer func() {
		reqCtx.Cancel()
		handlerName := mux.CurrentRoute(r).GetName()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to list buckets by ids", reqCtx.String())
			MakeErrorResponse(w, err)
			MetadataHandlerFailureMetrics(err, startTime, handlerName)
		} else {
			MetadataHandlerSuccessMetrics(startTime, handlerName)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	err = json.NewDecoder(r.Body).Decode(&bucketIDs)
	if err != nil {
		log.Errorf("failed to parse bucket ids", "error", err)
		err = ErrInvalidQuery
		return
	}

	if len(bucketIDs.IDs) == 0 || len(bucketIDs.IDs) > MaximumListObjectsAndBucketsSize {
		log.Errorf("failed to check ids", "error", err)
		err = ErrInvalidQuery
		return
	}

	bucketIDMap = make(map[uint64]bool)
	for _, id := range bucketIDs.IDs {
		if _, ok = bucketIDMap[id]; ok {
			// repeat id keys in request
			log.Errorf("failed to check ids", "error", err)
			err = ErrInvalidQuery
			return
		}
		bucketIDMap[id] = true
	}

	buckets, err = g.baseApp.GfSpClient().ListBucketsByBucketID(reqCtx.Context(), bucketIDs.IDs, false)
	if err != nil {
		log.Errorf("failed to list buckets by ids", "error", err)
		return
	}
	grpcResponse := &types.GfSpListBucketsByBucketIDResponse{Buckets: buckets}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&buf, grpcResponse); err != nil {
		log.Errorf("failed to list buckets by ids", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(buf.Bytes())
}

// getPaymentByBucketIDHandler get payment by bucket id
func (g *GateModular) getPaymentByBucketIDHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		buf         bytes.Buffer
		payment     *payment_types.StreamRecord
		bucketIDStr string
		bucketID    int64
		queryParams url.Values
		reqCtx      *RequestContext
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get payment by bucket id", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	bucketIDStr = queryParams.Get(BucketIDQuery)

	if bucketID, err = util.StringToInt64(bucketIDStr); err != nil || bucketID < 0 {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check bucket id", "bucket-id", bucketIDStr, "error", err)
		err = ErrInvalidQuery
		return
	}

	payment, err = g.baseApp.GfSpClient().GetPaymentByBucketID(reqCtx.Context(), bucketID, false)
	if err != nil {
		log.Errorf("failed to get payment by bucket id", "error", err)
		return
	}

	grpcResponse := &types.GfSpGetPaymentByBucketIDResponse{StreamRecord: payment}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&buf, grpcResponse); err != nil {
		log.Errorf("failed to get payment by bucket id", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(buf.Bytes())
}

// getPaymentByBucketNameHandler handle get payment by bucket name request
func (g *GateModular) getPaymentByBucketNameHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		b          bytes.Buffer
		reqCtx     *RequestContext
		bucketName string
		payment    *payment_types.StreamRecord
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get payment by bucket name", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	bucketName = reqCtx.bucketName

	if err = s3util.CheckValidBucketName(bucketName); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to check bucket name", "bucket-name", bucketName, "error", err)
		return
	}

	payment, err = g.baseApp.GfSpClient().GetPaymentByBucketName(reqCtx.Context(), bucketName, false)
	if err != nil {
		log.Errorf("failed to get payment by bucket name", "error", err)
		return
	}

	grpcResponse := &types.GfSpGetPaymentByBucketNameResponse{StreamRecord: payment}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.Errorf("failed to get payment by bucket name", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// getBucketByBucketNameHandler handle get bucket by bucket name request
func (g *GateModular) getBucketByBucketNameHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		b          bytes.Buffer
		reqCtx     *RequestContext
		bucketName string
		bucket     *types.Bucket
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get bucket by bucket name", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	bucketName = reqCtx.bucketName

	if err = s3util.CheckValidBucketName(bucketName); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to check bucket name", "bucket-name", bucketName, "error", err)
		return
	}

	bucket, err = g.baseApp.GfSpClient().GetBucketByBucketName(reqCtx.Context(), bucketName, false)
	if err != nil {
		log.Errorf("failed to get bucket by bucket name", "error", err)
		return
	}

	grpcResponse := &types.GfSpGetBucketByBucketNameResponse{Bucket: bucket}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.Errorf("failed to get bucket by bucket name", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// getBucketByBucketIDHandler handle get bucket by bucket id
func (g *GateModular) getBucketByBucketIDHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		buf         bytes.Buffer
		bucket      *types.Bucket
		bucketIDStr string
		bucketID    int64
		queryParams url.Values
		reqCtx      *RequestContext
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get bucket by bucket id", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	bucketIDStr = queryParams.Get(BucketIDQuery)

	if bucketID, err = util.StringToInt64(bucketIDStr); err != nil || bucketID < 0 {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check bucket id", "bucket-id", bucketIDStr, "error", err)
		err = ErrInvalidQuery
		return
	}

	bucket, err = g.baseApp.GfSpClient().GetBucketByBucketID(reqCtx.Context(), bucketID, false)
	if err != nil {
		log.Errorf("failed to get bucket by bucket id", "error", err)
		return
	}

	grpcResponse := &types.GfSpGetBucketByBucketIDResponse{Bucket: bucket}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&buf, grpcResponse); err != nil {
		log.Errorf("failed to get bucket by bucket id", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(buf.Bytes())
}

// listDeletedObjectsByBlockNumberRangeHandler handle list deleted objects info by a block number range request
func (g *GateModular) listDeletedObjectsByBlockNumberRangeHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                        error
		b                          bytes.Buffer
		reqCtx                     *RequestContext
		requestSpOperatorAddress   string
		requestStartBlockNumberStr string
		requestEndBlockNumberStr   string
		startBlockNumber           uint64
		endBlockNumber             uint64
		block                      uint64
		objects                    []*types.Object
		queryParams                url.Values
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to list deleted objects by block number range", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	requestSpOperatorAddress = queryParams.Get(SpOperatorAddressQuery)
	requestStartBlockNumberStr = queryParams.Get(StartBlockNumberQuery)
	requestEndBlockNumberStr = queryParams.Get(EndBlockNumberQuery)

	if ok := common.IsHexAddress(requestSpOperatorAddress); !ok {
		log.Errorw("failed to check operator", "sp-operator-address", requestSpOperatorAddress, "error", err)
		return
	}

	if startBlockNumber, err = util.StringToUint64(requestStartBlockNumberStr); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check start block number", "start-block-number", requestStartBlockNumberStr, "error", err)
		err = ErrInvalidQuery
		return
	}

	if endBlockNumber, err = util.StringToUint64(requestEndBlockNumberStr); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check end block number", "end-block-number", requestEndBlockNumberStr, "error", err)
		err = ErrInvalidQuery
		return
	}

	objects, block, err = g.baseApp.GfSpClient().ListDeletedObjectsByBlockNumberRange(reqCtx.Context(), requestSpOperatorAddress, startBlockNumber, endBlockNumber, true)
	if err != nil {
		log.Errorf("failed to list deleted objects by block number range", "error", err)
		return
	}

	grpcResponse := &types.GfSpListDeletedObjectsByBlockNumberRangeResponse{
		Objects:        objects,
		EndBlockNumber: int64(block),
	}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.Errorf("failed to list deleted objects by block number range", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// getUserBucketsCountHandler handle get user bucket count request
func (g *GateModular) getUserBucketsCountHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err    error
		b      bytes.Buffer
		reqCtx *RequestContext
		count  int64
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get user buckets count", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	if ok := common.IsHexAddress(r.Header.Get(GnfdUserAddressHeader)); !ok {
		log.Errorw("failed to check X-Gnfd-User-Address", "X-Gnfd-User-Address", reqCtx.account, "error", err)
		err = ErrInvalidHeader
		return
	}

	count, err = g.baseApp.GfSpClient().GetUserBucketsCount(reqCtx.Context(), r.Header.Get(GnfdUserAddressHeader), true)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get user buckets count", "error", err)
		return
	}

	grpcResponse := &types.GfSpGetUserBucketsCountResponse{Count: count}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get user buckets count", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// listExpiredBucketsBySpHandler handle list buckets that are expired by specific sp
func (g *GateModular) listExpiredBucketsBySpHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                error
		b                  bytes.Buffer
		reqCtx             *RequestContext
		requestLimit       string
		requestCreateAt    string
		requestPrimarySpID string
		limit              int64
		createAt           int64
		spID               uint32
		buckets            []*types.Bucket
		queryParams        url.Values
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to list expired buckets by sp", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	requestLimit = queryParams.Get(LimitQuery)
	requestCreateAt = queryParams.Get(CreateAtQuery)
	requestPrimarySpID = queryParams.Get(PrimarySpIDQuery)

	if limit, err = util.StringToInt64(requestLimit); err != nil || limit <= 0 {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check limit", "limit", requestLimit, "error", err)
		err = ErrInvalidQuery
		return
	}

	if limit > MaximumListObjectsAndBucketsSize {
		limit = MaximumListObjectsAndBucketsSize
	}

	if createAt, err = util.StringToInt64(requestCreateAt); err != nil || createAt < 0 {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check create at", "create-at", requestCreateAt, "error", err)
		err = ErrInvalidQuery
		return
	}

	if spID, err = util.StringToUint32(requestPrimarySpID); err != nil || createAt < 0 {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check sp id", "sp-id", requestPrimarySpID, "error", err)
		err = ErrInvalidQuery
		return
	}

	buckets, err = g.baseApp.GfSpClient().ListExpiredBucketsBySp(reqCtx.Context(), createAt, spID, limit)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list expired buckets by sp", "error", err)
		return
	}

	grpcResponse := &types.GfSpListExpiredBucketsBySpResponse{Buckets: buckets}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list expired buckets by sp", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// verifyPermissionByIDHandler handle verify permission by id request
func (g *GateModular) verifyPermissionByIDHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                 error
		requestOperator     string
		requestResourceID   string
		requestResourceType string
		requestActionType   string
		resourceID          uint64
		resourceType        uint32
		actionType          uint32
		ok                  bool
		b                   bytes.Buffer
		queryParams         url.Values
		effect              *permission_types.Effect
		reqCtx              *RequestContext
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to verify permission by id", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	requestResourceID = queryParams.Get(ResourceIDQuery)
	requestResourceType = queryParams.Get(ResourceTypeQuery)
	requestOperator = queryParams.Get(VerifyPermissionOperator)
	requestActionType = queryParams.Get(VerifyPermissionActionType)

	if ok = common.IsHexAddress(requestOperator); !ok {
		log.Errorw("failed to check operator", "operator", requestOperator, "error", err)
		return
	}

	if resourceID, err = util.StringToUint64(requestResourceID); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check resource id", "resource-id", requestResourceID, "error", err)
		err = ErrInvalidQuery
		return
	}

	// 0: "RESOURCE_TYPE_UNSPECIFIED", RESOURCE_TYPE_UNSPECIFIED is not considered in this request
	if resourceType, err = util.StringToUint32(requestResourceType); err != nil || resourceType == 0 {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check source type", "resource-type", requestResourceType, "error", err)
		err = ErrInvalidQuery
		return
	}

	if _, ok = resource_types.ResourceType_name[int32(resourceType)]; !ok {
		log.CtxErrorw(reqCtx.Context(), "failed to check source type", "resource-type", resourceType, "error", err)
		err = ErrInvalidQuery
		return
	}

	if actionType, err = util.StringToUint32(requestActionType); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse action type", "action-type", requestActionType, "error", err)
		err = ErrInvalidQuery
		return
	}

	if _, ok = permission_types.ActionType_name[int32(actionType)]; !ok {
		log.CtxErrorw(reqCtx.Context(), "failed to check action type", "action-type", actionType, "error", err)
		err = ErrInvalidQuery
		return
	}

	effect, err = g.baseApp.GfSpClient().VerifyPermissionByID(reqCtx.Context(), requestOperator, resource.ResourceType(resourceType), resourceID, permission_types.ActionType(actionType))
	if err != nil {
		log.Errorf("failed to verify permission by id", "error", err)
		return
	}

	grpcResponse := &storage_types.QueryVerifyPermissionResponse{Effect: *effect}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.Errorf("failed to verify permission by id", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// listVirtualGroupFamiliesBySpIDHandler list virtual group families by sp id
func (g *GateModular) listVirtualGroupFamiliesBySpIDHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		b           bytes.Buffer
		reqCtx      *RequestContext
		requestSpID string
		spID        uint32
		families    []*virtual_types.GlobalVirtualGroupFamily
		queryParams url.Values
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to list virtual group families by sp id", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	requestSpID = queryParams.Get(SpIDQuery)

	if spID, err = util.StringToUint32(requestSpID); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check sp id", "sp-id", requestSpID, "error", err)
		err = ErrInvalidQuery
		return
	}

	families, err = g.baseApp.GfSpClient().ListVirtualGroupFamiliesSpID(reqCtx.Context(), spID)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list virtual group families by sp id", "error", err)
		return
	}

	grpcResponse := &types.GfSpListVirtualGroupFamiliesBySpIDResponse{GlobalVirtualGroupFamilies: families}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list virtual group families by sp id", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// getVirtualGroupFamilyHandler get virtual group families by vgf id
func (g *GateModular) getVirtualGroupFamilyHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err          error
		b            bytes.Buffer
		reqCtx       *RequestContext
		requestVgfID string
		vgfID        uint32
		family       *virtual_types.GlobalVirtualGroupFamily
		queryParams  url.Values
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get virtual group families by vgf id", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	requestVgfID = queryParams.Get(VgfIDQuery)

	if vgfID, err = util.StringToUint32(requestVgfID); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check vgf id", "vgf-id", requestVgfID, "error", err)
		err = ErrInvalidQuery
		return
	}

	family, err = g.baseApp.GfSpClient().GetVirtualGroupFamily(reqCtx.Context(), vgfID)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get virtual group families by vgf id", "error", err)
		return
	}

	grpcResponse := &types.GfSpGetVirtualGroupFamilyResponse{Vgf: family}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get virtual group families by vgf id", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// getGlobalVirtualGroupByGvgIDHandler get global virtual group by gvg id
func (g *GateModular) getGlobalVirtualGroupByGvgIDHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err          error
		b            bytes.Buffer
		reqCtx       *RequestContext
		requestGvgID string
		gvgID        uint32
		group        *virtual_types.GlobalVirtualGroup
		queryParams  url.Values
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get global virtual group by gvg id", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	requestGvgID = queryParams.Get(GvgIDQuery)

	if gvgID, err = util.StringToUint32(requestGvgID); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check gvg id", "gvg-id", requestGvgID, "error", err)
		err = ErrInvalidQuery
		return
	}

	group, err = g.baseApp.GfSpClient().GetGlobalVirtualGroupByGvgID(reqCtx.Context(), gvgID)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get global virtual group by gvg id", "error", err)
		return
	}

	grpcResponse := &types.GfSpGetGlobalVirtualGroupByGvgIDResponse{GlobalVirtualGroup: group}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get global virtual group by gvg id", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// getGlobalVirtualGroupHandler get global virtual group by lvg id and bucket id
func (g *GateModular) getGlobalVirtualGroupHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err          error
		b            bytes.Buffer
		reqCtx       *RequestContext
		requestLvgID string
		bucketIDStr  string
		bucketID     uint64
		lvgID        uint32
		group        *virtual_types.GlobalVirtualGroup
		queryParams  url.Values
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get global virtual group by lvg id and bucket id", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	requestLvgID = queryParams.Get(LvgIDQuery)
	bucketIDStr = queryParams.Get(BucketIDQuery)

	if bucketID, err = util.StringToUint64(bucketIDStr); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check bucket id", "bucket-id", bucketIDStr, "error", err)
		err = ErrInvalidQuery
		return
	}

	if lvgID, err = util.StringToUint32(requestLvgID); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check lvg id", "lvg-id", requestLvgID, "error", err)
		err = ErrInvalidQuery
		return
	}

	group, err = g.baseApp.GfSpClient().GetGlobalVirtualGroup(reqCtx.Context(), bucketID, lvgID)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get global virtual group by lvg id and bucket id", "error", err)
		return
	}

	grpcResponse := &types.GfSpGetGlobalVirtualGroupResponse{Gvg: group}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to get global virtual group by lvg id and bucket id", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// listGlobalVirtualGroupsBySecondarySPHandler list global virtual group by secondary sp id
func (g *GateModular) listGlobalVirtualGroupsBySecondarySPHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		b           bytes.Buffer
		reqCtx      *RequestContext
		requestSpID string
		spID        uint32
		groups      []*virtual_types.GlobalVirtualGroup
		queryParams url.Values
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to list global virtual group by secondary sp id", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	requestSpID = queryParams.Get(SpIDQuery)

	if spID, err = util.StringToUint32(requestSpID); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check sp id", "sp-id", requestSpID, "error", err)
		err = ErrInvalidQuery
		return
	}

	groups, err = g.baseApp.GfSpClient().ListGlobalVirtualGroupsBySecondarySP(reqCtx.Context(), spID)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list global virtual group by secondary sp id", "error", err)
		return
	}

	grpcResponse := &types.GfSpListGlobalVirtualGroupsBySecondarySPResponse{Groups: groups}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list global virtual group by secondary sp id", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// listGlobalVirtualGroupsByBucketHandler list global virtual group by bucket id
func (g *GateModular) listGlobalVirtualGroupsByBucketHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		b           bytes.Buffer
		reqCtx      *RequestContext
		bucketIDStr string
		bucketID    uint64
		groups      []*virtual_types.GlobalVirtualGroup
		queryParams url.Values
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to list global virtual group by bucket id", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	bucketIDStr = queryParams.Get(BucketIDQuery)

	if bucketID, err = util.StringToUint64(bucketIDStr); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check bucket id", "bucket-id", bucketIDStr, "error", err)
		err = ErrInvalidQuery
		return
	}

	groups, err = g.baseApp.GfSpClient().ListGlobalVirtualGroupsByBucket(reqCtx.Context(), bucketID)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list global virtual group by bucket id", "error", err)
		return
	}

	grpcResponse := &types.GfSpListGlobalVirtualGroupsByBucketResponse{Groups: groups}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list global virtual group by bucket id", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// listObjectsInGVGAndBucketHandler list objects by gvg and bucket id
func (g *GateModular) listObjectsInGVGAndBucketHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err               error
		b                 bytes.Buffer
		reqCtx            *RequestContext
		requestGvgID      string
		gvgID             uint32
		bucketIDStr       string
		bucketID          uint64
		limit             uint32
		limitStr          string
		requestStartAfter string
		startAfter        uint64
		objects           []*types.Object
		queryParams       url.Values
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to list objects by gvg and bucket id", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	requestGvgID = queryParams.Get(GvgIDQuery)
	bucketIDStr = queryParams.Get(BucketIDQuery)
	requestStartAfter = queryParams.Get(ListObjectsStartAfterQuery)
	limitStr = queryParams.Get(GetGroupListLimitQuery)

	if bucketID, err = util.StringToUint64(bucketIDStr); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check bucket id", "bucket-id", bucketIDStr, "error", err)
		err = ErrInvalidQuery
		return
	}

	if gvgID, err = util.StringToUint32(requestGvgID); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check gvg id", "gvg-id", requestGvgID, "error", err)
		err = ErrInvalidQuery
		return
	}

	if requestStartAfter != "" {
		if startAfter, err = util.StringToUint64(requestStartAfter); err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to parse or check gvg id", "start-after", requestStartAfter, "error", err)
			err = ErrInvalidQuery
			return
		}
	}

	if limitStr != "" {
		if limit, err = util.StringToUint32(limitStr); err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to parse or check limit", "limit", limitStr, "error", err)
			err = ErrInvalidQuery
			return
		}
	}

	objects, err = g.baseApp.GfSpClient().ListObjectsInGVGAndBucket(reqCtx.Context(), gvgID, bucketID, startAfter, limit)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list objects by gvg and bucket id", "error", err)
		return
	}

	grpcResponse := &types.GfSpListObjectsInGVGResponse{Objects: objects}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list objects by gvg and bucket id", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// listObjectsInGVGHandler list objects by gvg id
func (g *GateModular) listObjectsInGVGHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err               error
		b                 bytes.Buffer
		reqCtx            *RequestContext
		requestGvgID      string
		gvgID             uint32
		limit             uint32
		limitStr          string
		requestStartAfter string
		startAfter        uint64
		objects           []*types.Object
		queryParams       url.Values
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to list objects by gvg id", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	requestGvgID = queryParams.Get(GvgIDQuery)
	requestStartAfter = queryParams.Get(ListObjectsStartAfterQuery)
	limitStr = queryParams.Get(GetGroupListLimitQuery)

	if gvgID, err = util.StringToUint32(requestGvgID); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check gvg id", "gvg-id", requestGvgID, "error", err)
		err = ErrInvalidQuery
		return
	}

	if startAfter, err = util.StringToUint64(requestStartAfter); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check gvg id", "start-after", requestStartAfter, "error", err)
		err = ErrInvalidQuery
		return
	}

	if limitStr != "" {
		if limit, err = util.StringToUint32(limitStr); err != nil {
			log.CtxErrorw(reqCtx.Context(), "failed to parse or check limit", "limit", limitStr, "error", err)
			err = ErrInvalidQuery
			return
		}
	}

	objects, err = g.baseApp.GfSpClient().ListObjectsInGVG(reqCtx.Context(), gvgID, startAfter, limit)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list objects by gvg id", "error", err)
		return
	}

	grpcResponse := &types.GfSpListObjectsInGVGResponse{Objects: objects}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list objects by gvg id", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// listMigrateBucketEventsHandler list migrate bucket events
func (g *GateModular) listMigrateBucketEventsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		b              bytes.Buffer
		reqCtx         *RequestContext
		requestSpID    string
		requestBlockID string
		spID           uint32
		blockID        uint64
		events         []*types.ListMigrateBucketEvents
		queryParams    url.Values
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to list migrate bucket events", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	requestSpID = queryParams.Get(SpIDQuery)
	requestBlockID = queryParams.Get(BlockIDQuery)

	if spID, err = util.StringToUint32(requestSpID); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check sp id", "sp-id", requestSpID, "error", err)
		err = ErrInvalidQuery
		return
	}

	if blockID, err = util.StringToUint64(requestBlockID); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check block id", "block-id", requestBlockID, "error", err)
		err = ErrInvalidQuery
		return
	}

	events, err = g.baseApp.GfSpClient().ListMigrateBucketEvents(reqCtx.Context(), blockID, spID)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list migrate bucket events", "error", err)
		return
	}

	grpcResponse := &types.GfSpListMigrateBucketEventsResponse{Events: events}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list migrate bucket events", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// listSwapOutEventsHandler list swap out events
func (g *GateModular) listSwapOutEventsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		b              bytes.Buffer
		reqCtx         *RequestContext
		requestSpID    string
		requestBlockID string
		spID           uint32
		blockID        uint64
		events         []*types.ListSwapOutEvents
		queryParams    url.Values
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to list swap out events", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	requestSpID = queryParams.Get(SpIDQuery)
	requestBlockID = queryParams.Get(BlockIDQuery)

	if spID, err = util.StringToUint32(requestSpID); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check sp id", "sp-id", requestSpID, "error", err)
		err = ErrInvalidQuery
		return
	}

	if blockID, err = util.StringToUint64(requestBlockID); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check block id", "block-id", requestBlockID, "error", err)
		err = ErrInvalidQuery
		return
	}

	events, err = g.baseApp.GfSpClient().ListSwapOutEvents(reqCtx.Context(), blockID, spID)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list swap out events", "error", err)
		return
	}

	grpcResponse := &types.GfSpListSwapOutEventsResponse{Events: events}
	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list swap out events", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// listSpExitEventsHandler list sp exit events
func (g *GateModular) listSpExitEventsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err                    error
		b                      bytes.Buffer
		reqCtx                 *RequestContext
		requestOperatorAddress string
		requestBlockID         string
		blockID                uint64
		events                 *types.ListSpExitEvents
		queryParams            url.Values
	)

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to list migrate bucket events", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, _ = NewRequestContext(r, g)

	queryParams = reqCtx.request.URL.Query()
	requestOperatorAddress = queryParams.Get(OperatorAddressQuery)
	requestBlockID = queryParams.Get(BlockIDQuery)

	if blockID, err = util.StringToUint64(requestBlockID); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check block id", "block-id", requestBlockID, "error", err)
		err = ErrInvalidQuery
		return
	}

	if ok := common.IsHexAddress(requestOperatorAddress); !ok {
		log.Errorw("failed to check operator", "operator-address", requestOperatorAddress, "error", err)
		return
	}

	events, err = g.baseApp.GfSpClient().ListSpExitEvents(reqCtx.Context(), blockID, requestOperatorAddress)
	if err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list sp exit events", "error", err)
		return
	}

	grpcResponse := &types.GfSpListSpExitEventsResponse{Events: events}

	m := jsonpb.Marshaler{EmitDefaults: true, OrigName: true, EnumsAsInts: true}
	if err = m.Marshal(&b, grpcResponse); err != nil {
		log.CtxErrorw(reqCtx.Context(), "failed to list sp exit events", "error", err)
		return
	}

	w.Header().Set(ContentTypeHeader, ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}
