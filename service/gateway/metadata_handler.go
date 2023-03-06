package gateway

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
	"github.com/gorilla/mux"
)

// getUserBucketsHandler handle get object request
func (g *Gateway) getUserBucketsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	req := &stypes.MetadataServiceGetUserBucketsRequest{
		AccountId: vars["account_id"],
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
	vars := mux.Vars(r)

	req := &stypes.MetadataServiceListObjectsByBucketNameRequest{
		BucketName: vars["bucket_name"],
		AccountId:  vars["account_id"],
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
