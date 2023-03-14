package gateway

import (
	"bytes"
	"context"
	"net/http"
	"strconv"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	metatypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
)

const (
	HTTPHeaderAddress = "X-Gnfd-Address"
)

// getUserBucketsHandler handle get object request
func (gateway *Gateway) getUserBucketsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		b              bytes.Buffer
		errDescription *errorDescription
		reqContext     *requestContext
	)

	reqContext = newRequestContext(r)
	defer func() {
		if errDescription != nil {
			_ = errDescription.errorJSONResponse(w, reqContext)
		}
		if errDescription != nil && errDescription.statusCode != http.StatusOK {
			log.Errorf("action(%v) statusCode(%v) %v", getUserBucketsRouterName, errDescription.statusCode, reqContext.generateRequestDetail())
		} else {
			log.Infof("action(%v) statusCode(200) %v", getUserBucketsRouterName, reqContext.generateRequestDetail())
		}
	}()

	if gateway.metadata == nil {
		log.Errorw("failed to get user buckets due to not config metadata")
		errDescription = NotExistComponentError
		return
	}

	req := &metatypes.MetadataServiceGetUserBucketsRequest{
		AccountId: r.Header.Get(HTTPHeaderAddress),
	}
	ctx := log.Context(context.Background(), req)
	resp, err := gateway.metadata.GetUserBuckets(ctx, req)
	if err != nil {
		log.Errorf("failed to get user buckets", "error", err)
		return
	}

	m := jsonpb.Marshaler{EmitDefaults: true}
	if err = m.Marshal(&b, resp); err != nil {
		log.Errorf("failed to get user buckets", "error", err)
		return
	}

	w.Header().Set(model.ContentTypeHeader, model.ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// listObjectsByBucketNameHandler handle list objects by bucket name request
func (gateway *Gateway) listObjectsByBucketNameHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		b              bytes.Buffer
		errDescription *errorDescription
		reqContext     *requestContext
	)

	reqContext = newRequestContext(r)
	defer func() {
		if errDescription != nil {
			_ = errDescription.errorJSONResponse(w, reqContext)
		}
		if errDescription != nil && errDescription.statusCode != http.StatusOK {
			log.Errorf("action(%v) statusCode(%v) %v", listObjectsByBucketRouterName, errDescription.statusCode, reqContext.generateRequestDetail())
		} else {
			log.Infof("action(%v) statusCode(200) %v", listObjectsByBucketRouterName, reqContext.generateRequestDetail())
		}
	}()

	if gateway.metadata == nil {
		log.Errorw("failed to list objects by bucket name due to not config metadata")
		errDescription = NotExistComponentError
		return
	}

	req := &metatypes.MetadataServiceListObjectsByBucketNameRequest{
		BucketName: reqContext.bucketName,
	}

	ctx := log.Context(context.Background(), req)
	resp, err := gateway.metadata.ListObjectsByBucketName(ctx, req)
	if err != nil {
		log.Errorf("failed to list objects by bucket name", "error", err)
		return
	}

	m := jsonpb.Marshaler{EmitDefaults: true}
	if err = m.Marshal(&b, resp); err != nil {
		log.Errorf("failed to list objects by bucket name", "error", err)
		return
	}

	w.Header().Set(model.ContentTypeHeader, model.ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// TODO barry remove all sp internal handlers
// getBucketHandler
func (g *Gateway) getBucketByBucketNameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		isFullList bool
		b          bytes.Buffer
	)
	if vars["is_full_list"] == "true" {
		isFullList = true
	} else {
		isFullList = false
	}
	req := &metatypes.MetadataServiceGetBucketByBucketNameRequest{
		BucketName: vars["bucket_name"],
		IsFullList: isFullList,
	}

	ctx := log.Context(context.Background(), req)
	resp, err := g.metadata.GetBucketByBucketName(ctx, req)
	if err != nil {
		log.Errorf("failed to get bucket by bucket name", "error", err)
		return
	}
	m := jsonpb.Marshaler{EmitDefaults: true}
	if err = m.Marshal(&b, resp); err != nil {
		log.Errorf("failed to get bucket by bucket name", "error", err)
		return
	}

	w.Header().Set(model.ContentTypeHeader, model.ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// TODO barry remove all sp internal handlers
// getBucketHandler
func (g *Gateway) getBucketByBucketIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		isFullList bool
		b          bytes.Buffer
	)
	if vars["is_full_list"] == "true" {
		isFullList = true
	} else {
		isFullList = false
	}
	bucketId, err := strconv.ParseInt(vars["bucket_id"], 10, 64)
	if err != nil {
		log.Errorf("failed to parse bucket id", "error", err)
		return
	}
	req := &metatypes.MetadataServiceGetBucketByBucketIDRequest{
		BucketId:   bucketId,
		IsFullList: isFullList,
	}

	ctx := log.Context(context.Background(), req)
	resp, err := g.metadata.GetBucketByBucketID(ctx, req)
	if err != nil {
		log.Errorf("failed to get bucket by bucket id", "error", err)
		return
	}

	m := jsonpb.Marshaler{EmitDefaults: true}
	if err = m.Marshal(&b, resp); err != nil {
		log.Errorf("failed to get bucket by bucket id", "error", err)
		return
	}

	w.Header().Set(model.ContentTypeHeader, model.ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// TODO barry remove all sp internal handlers
// listDeletedObjectsByBlockNumberRangeHandler handle list deleted objects by block number range request
func (g *Gateway) listDeletedObjectsByBlockNumberRangeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var (
		isFullList bool
		b          bytes.Buffer
	)
	if vars["is_full_list"] == "true" {
		isFullList = true
	} else {
		isFullList = false
	}

	startBlockNumberStr := vars["start"]
	endBlockNumberStr := vars["end"]

	startBlockNumber, err := strconv.ParseInt(startBlockNumberStr, 10, 64)
	if err != nil {
		log.Errorf("failed to parse start block number", "error", err)
		return
	}

	endBlockNumber, err := strconv.ParseInt(endBlockNumberStr, 10, 64)
	if err != nil {
		log.Errorf("failed to parse end block number", "error", err)
		return
	}

	req := &metatypes.MetadataServiceListDeletedObjectsByBlockNumberRangeRequest{
		StartBlockNumber: startBlockNumber,
		EndBlockNumber:   endBlockNumber,
		IsFullList:       isFullList,
	}

	ctx := log.Context(context.Background(), req)
	resp, err := g.metadata.ListDeletedObjectsByBlockNumberRange(ctx, req)
	if err != nil {
		log.Errorf("failed to list deleted objects by block number range request", "error", err)
		return
	}

	m := jsonpb.Marshaler{EmitDefaults: true}
	if err = m.Marshal(&b, resp); err != nil {
		log.Errorf("failed to marshal list deleted objects by block number range response", "error", err)
		return
	}

	w.Header().Set(model.ContentTypeHeader, model.ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}

// TODO barry remove all sp internal handlers
// getUserBucketsCountHandler handle get object request
func (g *Gateway) getUserBucketsCountHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		b              bytes.Buffer
		errDescription *errorDescription
		reqContext     *requestContext
	)

	reqContext = newRequestContext(r)
	defer func() {
		if errDescription != nil {
			_ = errDescription.errorJSONResponse(w, reqContext)
		}
		if errDescription != nil && errDescription.statusCode != http.StatusOK {
			log.Errorf("action(%v) statusCode(%v) %v", getUserBucketsRouterName, errDescription.statusCode, reqContext.generateRequestDetail())
		} else {
			log.Infof("action(%v) statusCode(200) %v", getUserBucketsRouterName, reqContext.generateRequestDetail())
		}
	}()

	if g.metadata == nil {
		log.Errorw("failed to get user buckets count due to not config metadata")
		errDescription = NotExistComponentError
		return
	}

	req := &metatypes.MetadataServiceGetUserBucketsCountRequest{
		AccountId: reqContext.accountID,
	}

	ctx := log.Context(context.Background(), req)
	resp, err := g.metadata.GetUserBucketsCount(ctx, req)
	if err != nil {
		log.Errorf("failed to get user buckets count", "error", err)
		return
	}

	m := jsonpb.Marshaler{EmitDefaults: true}
	if err = m.Marshal(&b, resp); err != nil {
		log.Errorf("failed to get user buckets count", "error", err)
		return
	}

	w.Header().Set(model.ContentTypeHeader, model.ContentTypeJSONHeaderValue)
	w.Write(b.Bytes())
}
