package gateway

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const (
	approvalRouterName               = "GetApproval"
	putObjectRouterName              = "PutObject"
	getObjectRouterName              = "GetObject"
	challengeRouterName              = "Challenge"
	replicateObjectPieceRouterName   = "ReplicateObjectPiece"
	getUserBucketsRouterName         = "GetUserBuckets"
	listObjectsByBucketRouterName    = "ListObjectsByBucketName"
	getBucketReadQuotaRouterName     = "GetBucketReadQuota"
	listBucketReadRecordRouterName   = "ListBucketReadRecord"
	getObjectByUniversalEndpointName = "GetObjectByUniversalEndpoint"
)

const (
	createBucketApprovalAction = "CreateBucket"
	createObjectApprovalAction = "CreateObject"
)

// notFoundHandler log not found request info.
func (g *Gateway) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Errorw("not found handler", "header", r.Header, "host", r.Host, "url", r.URL)
	if _, err := io.ReadAll(r.Body); err != nil {
		log.Errorw("failed to read the unknown request", "error", err)
	}
	if err := NoRouter.errorResponse(w, &requestContext{}); err != nil {
		log.Errorw("failed to response the unknown request", "error", err)
	}
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
		Name(getBucketReadQuotaRouterName).
		Methods(http.MethodGet).
		Queries(model.GetBucketReadQuotaQuery, "",
			model.GetBucketReadQuotaMonthQuery, "{year_month}").
		HandlerFunc(g.getBucketReadQuotaHandler)
	bucketRouter.NewRoute().
		Name(listBucketReadRecordRouterName).
		Methods(http.MethodGet).
		Queries(model.ListBucketReadRecordQuery, "",
			model.ListBucketReadRecordMaxRecordsQuery, "{max_records}",
			model.StartTimestampUs, "{start_ts}",
			model.EndTimestampUs, "{end_ts}").
		HandlerFunc(g.listBucketReadRecordHandler)
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
	// admin router, path style, new router will prefer use virtual-hosted style
	r.Path(model.GetApprovalPath).
		Name(approvalRouterName).
		Methods(http.MethodGet).
		Queries(model.ActionQuery, "{action}").
		HandlerFunc(g.getApprovalHandler)
	r.Path(model.ChallengePath).
		Name(challengeRouterName).
		Methods(http.MethodGet).
		HandlerFunc(g.challengeHandler)
	// replicate piece to receiver
	r.Path(model.ReplicateObjectPiecePath).
		Name(replicateObjectPieceRouterName).
		Methods(http.MethodPut).
		HandlerFunc(g.replicatePieceHandler)
	r.Path(model.UniversalENdpointPath).
		Name(getObjectByUniversalEndpointName).
		Methods(http.MethodGet).
		HandlerFunc(g.getObjectByUniversalEndpointHandler)
	//redirect for universal endpoint
	http.Handle("/", r)
	r.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)
}
