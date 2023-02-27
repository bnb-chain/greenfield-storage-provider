package gateway

import (
	"context"
	"encoding/json"
	"net/http"

	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
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
		log.Warnf("failed to get user buckets", "error", err)
		return
	}
	ret, err := json.Marshal(resp)
	if err != nil {
		log.Warnf("failed to get user buckets", "error", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
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
		log.Warnf("failed to list objects by bucket name", "error", err)
		return
	}
	ret, err := json.Marshal(resp)
	if err != nil {
		log.Warnf("failed to list objects by bucket name", "error", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}
