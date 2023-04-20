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
	requestNonceName                 = "RequestNonce"
	updateUserPublicKey              = "UpdateUserPublicKey"
	queryUploadProgressRouterName    = "queryUploadProgress"
	getObjectByUniversalEndpointName = "GetObjectByUniversalEndpoint"
)

const (
	createBucketApprovalAction = "CreateBucket"
	createObjectApprovalAction = "CreateObject"
)

// notFoundHandler log not found request info.
func (g *Gateway) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Errorw("failed to find the corresponding handler", "header", r.Header, "host", r.Host, "url", r.URL)
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
	hostBucketRouter := r.Host("{bucket:.+}." + g.config.Domain).Subrouter()
	hostBucketRouter.NewRoute().
		Name(putObjectRouterName).
		Methods(http.MethodPut).
		Path("/{object:.+}").
		HandlerFunc(g.putObjectHandler)
	hostBucketRouter.NewRoute().
		Name(queryUploadProgressRouterName).
		Methods(http.MethodGet).
		Path("/{object:.+}").
		Queries(model.UploadProgressQuery, "").
		HandlerFunc(g.queryUploadProgressHandler)
	hostBucketRouter.NewRoute().
		Name(getObjectRouterName).
		Methods(http.MethodGet).
		Path("/{object:.+}").
		HandlerFunc(g.getObjectHandler)
	hostBucketRouter.NewRoute().
		Name(getBucketReadQuotaRouterName).
		Methods(http.MethodGet).
		Queries(model.GetBucketReadQuotaQuery, "",
			model.GetBucketReadQuotaMonthQuery, "{year_month}").
		HandlerFunc(g.getBucketReadQuotaHandler)
	hostBucketRouter.NewRoute().
		Name(listBucketReadRecordRouterName).
		Methods(http.MethodGet).
		Queries(model.ListBucketReadRecordQuery, "",
			model.ListBucketReadRecordMaxRecordsQuery, "{max_records}",
			model.StartTimestampUs, "{start_ts}",
			model.EndTimestampUs, "{end_ts}").
		HandlerFunc(g.listBucketReadRecordHandler)
	hostBucketRouter.NewRoute().
		Name(listObjectsByBucketRouterName).
		Methods(http.MethodGet).
		Path("/").
		HandlerFunc(g.listObjectsByBucketNameHandler)
	hostBucketRouter.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)

	// bucket list router, virtual-hosted style
	bucketListRouter := r.Host(g.config.Domain).Subrouter()
	bucketListRouter.NewRoute().
		Name(getUserBucketsRouterName).
		Methods(http.MethodGet).
		Path("/").
		HandlerFunc(g.getUserBucketsHandler)

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
	// replicate piece to receiver
	r.Path(model.ReplicateObjectPiecePath).
		Name(replicateObjectPieceRouterName).
		Methods(http.MethodPut).
		HandlerFunc(g.replicatePieceHandler)
	//universal endpoint
	r.Path(model.UniversalEndpointPath).
		Name(getObjectByUniversalEndpointName).
		Methods(http.MethodGet).
		HandlerFunc(g.getObjectByUniversalEndpointHandler)
	//redirect for universal endpoint
	http.Handle("/", r)

	// off-chain-auth router
	r.Path(model.AuthRequestNoncePath).
		Name(requestNonceName).
		Methods(http.MethodGet).
		HandlerFunc(g.requestNonceHandler)
	r.Path(model.AuthUpdateKeyPath).
		Name(updateUserPublicKey).
		Methods(http.MethodPost).
		HandlerFunc(g.updateUserPublicKeyHandler)

	// path style
	pathBucketRouter := r.PathPrefix("/{bucket}").Subrouter()
	pathBucketRouter.NewRoute().
		Name(putObjectRouterName).
		Methods(http.MethodPut).
		Path("/{object:.+}").
		HandlerFunc(g.putObjectHandler)
	pathBucketRouter.NewRoute().
		Name(queryUploadProgressRouterName).
		Methods(http.MethodGet).
		Path("/{object:.+}").
		Queries(model.UploadProgressQuery, "").
		HandlerFunc(g.queryUploadProgressHandler)
	pathBucketRouter.NewRoute().
		Name(getObjectRouterName).
		Methods(http.MethodGet).
		Path("/{object:.+}").
		HandlerFunc(g.getObjectHandler)
	pathBucketRouter.NewRoute().
		Name(getBucketReadQuotaRouterName).
		Methods(http.MethodGet).
		Queries(model.GetBucketReadQuotaQuery, "",
			model.GetBucketReadQuotaMonthQuery, "{year_month}").
		HandlerFunc(g.getBucketReadQuotaHandler)
	pathBucketRouter.NewRoute().
		Name(listBucketReadRecordRouterName).
		Methods(http.MethodGet).
		Queries(model.ListBucketReadRecordQuery, "",
			model.ListBucketReadRecordMaxRecordsQuery, "{max_records}",
			model.StartTimestampUs, "{start_ts}",
			model.EndTimestampUs, "{end_ts}").
		HandlerFunc(g.listBucketReadRecordHandler)
	pathBucketRouter.NewRoute().
		Name(listObjectsByBucketRouterName).
		Methods(http.MethodGet).
		Path("/").
		HandlerFunc(g.listObjectsByBucketNameHandler)
	pathBucketRouter.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)

	r.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)
	//r.Use(localhttp.Limit)
}
