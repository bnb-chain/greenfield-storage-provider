package gateway

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const (
	putObjectRouterName      = "PutObject"
	getObjectRouterName      = "GetObject"
	approvalRouterName       = "GetApproval"
	challengeRouterName      = "Challenge"
	syncPieceRouterName      = "SyncPiece"
	getUserBucketsRouterName = "GetUserBuckets"
	//TODO delete redundent name
	getUserBucketsCountRouterName                  = "GetUserBucketsCount"
	getBucketByBucketIDRouterName                  = "GetBucketByBucketID"
	getBucketByBucketNameRouterName                = "GetBucketByBucketName"
	listObjectsByBucketRouterName                  = "ListObjectsByBucketName"
	listDeletedObjectsByBlockNumberRangeRouterName = "ListDeletedObjectsByBlockNumberRange"
)

const (
	createBucketApprovalAction = "CreateBucket"
	createObjectApprovalAction = "CreateObject"
)

// notFoundHandler log not found request info.
func (g *Gateway) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Errorw("not found handler", "header", r.Header, "host", r.Host, "url", r.URL)
	s, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorw("failed to read the unknown request", "error", err)
	}
	w.WriteHeader(http.StatusNotFound)
	w.Write(s)
}

// registerHandler is used to register mux handlers.
func (g *Gateway) registerHandler(r *mux.Router) {
	// bucket router, virtual-hosted style
	bucketRouter := r.Host("{bucket:.+}." + g.config.Domain).Subrouter()
	bucketRouter.NewRoute().
		Name(putObjectRouterName).
		Methods(http.MethodPut).
		Path("/{object:.+}").
		HandlerFunc(g.putObjectHandler)
	bucketRouter.NewRoute().
		Name(getObjectRouterName).
		Methods(http.MethodGet).
		Path("/{object:.+}").
		HandlerFunc(g.getObjectHandler)
	bucketRouter.NewRoute().
		Name(listObjectsByBucketRouterName).
		Methods(http.MethodGet).
		Path("/").
		HandlerFunc(g.listObjectsByBucketNameHandler)
	bucketRouter.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)

	// bucket list router, virtual-hosted style
	bucketListRouter := r.Host(g.config.Domain).Subrouter()
	bucketListRouter.NewRoute().
		Name(getUserBucketsRouterName).
		Methods(http.MethodGet).
		Path("/").
		HandlerFunc(g.getUserBucketsHandler)
	bucketListRouter.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)

	// admin router, path style
	r.Path(model.GetApprovalPath).
		Name(approvalRouterName).
		Methods(http.MethodGet).
		Queries(model.ActionQuery, "{action}").
		HandlerFunc(g.getApprovalHandler)
	r.Path(model.ChallengePath).
		Name(challengeRouterName).
		Methods(http.MethodGet).
		HandlerFunc(g.challengeHandler)
	// sync piece to receiver
	r.Path(model.SyncPath).
		Name(syncPieceRouterName).
		Methods(http.MethodPut).
		HandlerFunc(g.syncPieceHandler)
	//metadata router
	//r.Name(getUserBucketsRouterName).
	//	Methods(http.MethodGet).
	//	Path("/accounts/{account_id:.+}/buckets").
	//	HandlerFunc(g.getUserBucketsHandler)
	r.Name(getUserBucketsCountRouterName).
		Methods(http.MethodGet).
		Path("/accounts/{account_id:.+}/buckets/count").
		HandlerFunc(g.getUserBucketsCountHandler)
	r.Name(getBucketByBucketNameRouterName).
		Methods(http.MethodGet).
		Path("/buckets/name/{bucket_name:.+}/{is_full_list:.+}").
		HandlerFunc(g.getBucketByBucketNameHandler)
	//TODO barry remove all sp internal handlers
	r.Name(getBucketByBucketIDRouterName).
		Methods(http.MethodGet).
		Path("/buckets/id/{bucket_id:.+}/{is_full_list:.+}").
		HandlerFunc(g.getBucketByBucketIDHandler)
	//r.Name(listObjectsByBucketRouterName).
	//	Methods(http.MethodGet).
	//	Path("/accounts/{account_id:.+}/buckets/{bucket:.+}/objects").
	//	HandlerFunc(g.listObjectsByBucketNameHandler)
	r.Name(listDeletedObjectsByBlockNumberRangeRouterName).
		Methods(http.MethodGet).
		Path("/delete/{start:.+}/{end:.+}/{is_full_list:.+}").
		HandlerFunc(g.listDeletedObjectsByBlockNumberRangeHandler)

	r.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)
}
