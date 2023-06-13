package gater

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/bnb-chain/greenfield/types/s3util"
	payment_types "github.com/bnb-chain/greenfield/x/payment/types"
	permission_types "github.com/bnb-chain/greenfield/x/permission/types"
	storage_types "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/ethereum/go-ethereum/common"

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

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}

	if ok := common.IsHexAddress(r.Header.Get(GnfdUserAddressHeader)); !ok {
		log.Errorw("failed to check account id", "account_id", reqCtx.account, "error", err)
		err = ErrInvalidHeader
		return
	}

	resp, err := g.baseApp.GfSpClient().GetUserBuckets(reqCtx.Context(), r.Header.Get(GnfdUserAddressHeader), true)
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

	reqCtx, err = NewRequestContext(r, g)
	if err != nil {
		return
	}

	queryParams = reqCtx.request.URL.Query()
	requestBucketName = reqCtx.bucketName
	requestMaxKeys = queryParams.Get(ListObjectsMaxKeysQuery)
	requestStartAfter = queryParams.Get(ListObjectsStartAfterQuery)
	requestContinuationToken = queryParams.Get(ListObjectsContinuationTokenQuery)
	requestDelimiter = queryParams.Get(ListObjectsDelimiterQuery)
	requestPrefix = queryParams.Get(ListObjectsPrefixQuery)

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
			true)
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

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get object meta", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, err = NewRequestContext(r, g)
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

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get bucket meta", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, err = NewRequestContext(r, g)
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

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to verify permission", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, err = NewRequestContext(r, g)
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

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to get group list", reqCtx.String())
			MakeErrorResponse(w, err)
		}
	}()

	reqCtx, err = NewRequestContext(r, g)
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

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to list objects by ids", reqCtx.String())
			MakeErrorResponse(w, err)
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

	defer func() {
		reqCtx.Cancel()
		if err != nil {
			reqCtx.SetError(gfsperrors.MakeGfSpError(err))
			log.CtxErrorw(reqCtx.Context(), "failed to list buckets by ids", reqCtx.String())
			MakeErrorResponse(w, err)
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
		err                     error
		b                       bytes.Buffer
		reqCtx                  *RequestContext
		requestLimit            string
		requestCreateAt         string
		requestPrimarySpAddress string
		limit                   int64
		createAt                int64
		buckets                 []*types.Bucket
		queryParams             url.Values
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
	requestPrimarySpAddress = queryParams.Get(PrimarySpAddressQuery)

	if limit, err = util.StringToInt64(requestLimit); err != nil || limit <= 0 {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check limit", "limit", requestLimit, "error", err)
		err = ErrInvalidQuery
		return
	}

	if limit > MaximumListObjectsAndBucketsSize {
		limit = MaximumListObjectsAndBucketsSize
	}

	if createAt, err = util.StringToInt64(requestCreateAt); err != nil || createAt < 0 {
		log.CtxErrorw(reqCtx.Context(), "failed to parse or check create at", "create-at", requestLimit, "error", err)
		err = ErrInvalidQuery
		return
	}

	buckets, err = g.baseApp.GfSpClient().ListExpiredBucketsBySp(reqCtx.Context(), createAt, requestPrimarySpAddress, limit)
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
