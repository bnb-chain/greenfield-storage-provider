package gateway

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	metatypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
)

// getUserBucketsHandler handle get object request
func (g *Gateway) getUserBucketsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		errDescription *errorDescription
		reqContext     *requestContext
		statusCode     = http.StatusOK
	)

	reqContext = newRequestContext(r)
	defer func() {
		if errDescription != nil {
			statusCode = errDescription.statusCode
			_ = errDescription.errorJSONResponse(w, reqContext)
		}
		if statusCode == http.StatusOK {
			log.Infof("action(%v) statusCode(%v) %v", getUserBucketsRouterName, statusCode, reqContext.generateRequestDetail())
		} else {
			log.Errorf("action(%v) statusCode(%v) %v", getUserBucketsRouterName, statusCode, reqContext.generateRequestDetail())
		}
	}()

	if g.metadata == nil {
		log.Errorw("failed to get user buckets due to not config metadata")
		errDescription = NotExistComponentError
		return
	}

	req := &metatypes.MetadataServiceGetUserBucketsRequest{
		AccountId: reqContext.accountID,
	}
	ctx := log.Context(context.Background(), req)
	resp, err := g.metadata.GetUserBuckets(ctx, req)
	if err != nil {
		log.Errorf("failed to get user buckets", "error", err)
		return
	}
	ret, err := json.Marshal(resp)
	if err != nil {
		log.Errorf("failed to get user buckets", "error", err)
		return
	}
	w.Header().Set(model.ContentTypeHeader, model.ContentTypeJSONHeaderValue)
	w.Write(ret)
}

// listObjectsByBucketNameHandler handle list objects by bucket name request
func (g *Gateway) listObjectsByBucketNameHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err            error
		errDescription *errorDescription
		reqContext     *requestContext
		statusCode     = http.StatusOK
	)

	reqContext = newRequestContext(r)
	defer func() {
		if errDescription != nil {
			statusCode = errDescription.statusCode
			_ = errDescription.errorJSONResponse(w, reqContext)
		}
		if statusCode == http.StatusOK {
			log.Infof("action(%v) statusCode(%v) %v", listObjectsByBucketRouterName, statusCode, reqContext.generateRequestDetail())
		} else {
			log.Errorf("action(%v) statusCode(%v) %v", listObjectsByBucketRouterName, statusCode, reqContext.generateRequestDetail())
		}
	}()

	if g.metadata == nil {
		log.Errorw("failed to list objects by bucket name due to not config metadata")
		errDescription = NotExistComponentError
		return
	}

	req := &metatypes.MetadataServiceListObjectsByBucketNameRequest{
		BucketName: reqContext.bucketName,
		AccountId:  reqContext.accountID,
	}

	ctx := log.Context(context.Background(), req)
	resp, err := g.metadata.ListObjectsByBucketName(ctx, req)
	if err != nil {
		log.Errorf("failed to list objects by bucket name", "error", err)
		return
	}
	ret, err := json.Marshal(resp)
	if err != nil {
		log.Errorf("failed to list objects by bucket name", "error", err)
		return
	}
	w.Header().Set(model.ContentTypeHeader, model.ContentTypeJSONHeaderValue)
	w.Write(ret)
}
