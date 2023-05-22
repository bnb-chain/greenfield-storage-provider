package gater

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const (
	approvalRouterName                    = "GetApproval"
	putObjectRouterName                   = "PutObject"
	getObjectRouterName                   = "GetObject"
	challengeRouterName                   = "Challenge"
	replicateObjectPieceRouterName        = "ReplicateObjectPiece"
	getUserBucketsRouterName              = "GetUserBuckets"
	listObjectsByBucketRouterName         = "ListObjectsByBucketName"
	getBucketReadQuotaRouterName          = "GetBucketReadQuota"
	listBucketReadRecordRouterName        = "ListBucketReadRecord"
	requestNonceName                      = "RequestNonce"
	updateUserPublicKey                   = "UpdateUserPublicKey"
	queryUploadProgressRouterName         = "queryUploadProgress"
	downloadObjectByUniversalEndpointName = "DownloadObjectByUniversalEndpoint"
	viewObjectByUniversalEndpointName     = "ViewObjectByUniversalEndpoint"
	getObjectMetaRouterName               = "getObjectMeta"
	getBucketMetaRouterName               = "getBucketMeta"
)

const (
	createBucketApprovalAction = "CreateBucket"
	createObjectApprovalAction = "CreateObject"
)

// notFoundHandler log not found request info.
func (g *GateModular) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Errorw("failed to find the corresponding handler", "header", r.Header, "host", r.Host, "url", r.URL)
	if _, err := io.ReadAll(r.Body); err != nil {
		log.Errorw("failed to read the unknown request", "error", err)
	}
}

// RegisterHandler registers the handlers to the gateway router.
func (g *GateModular) RegisterHandler(router *mux.Router) {
	// bucket router, virtual-hosted style
	hostBucketRouter := router.Host("{bucket:.+}." + g.domain).Subrouter()
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
		Name(getBucketMetaRouterName).
		Methods(http.MethodGet).
		Queries(model.GetBucketMetaQuery, "").
		HandlerFunc(g.getBucketMetaHandler)
	hostBucketRouter.NewRoute().
		Name(getObjectMetaRouterName).
		Methods(http.MethodGet).
		Path("/{object:.+}").
		Queries(model.GetObjectMetaQuery, "").
		HandlerFunc(g.getObjectMetaHandler)
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

	// bucket list router, path style
	router.Path("/").
		Name(getUserBucketsRouterName).
		Methods(http.MethodGet).
		HandlerFunc(g.getUserBucketsHandler)

	// admin router, path style
	router.Path(model.GetApprovalPath).
		Name(approvalRouterName).
		Methods(http.MethodGet).
		Queries(model.ActionQuery, "{action}").
		HandlerFunc(g.getApprovalHandler)
	router.Path(model.ChallengePath).
		Name(challengeRouterName).
		Methods(http.MethodGet).
		HandlerFunc(g.challengeHandler)
	// replicate piece to receiver
	router.Path(model.ReplicateObjectPiecePath).
		Name(replicateObjectPieceRouterName).
		Methods(http.MethodPut).
		HandlerFunc(g.replicateHandler)
	//universal endpoint
	//router.Path(model.UniversalEndpointPath).
	//	Name(getObjectByUniversalEndpointName).
	//	Methods(http.MethodGet).
	//	HandlerFunc(g.getObjectByUniversalEndpointHandler)
	//redirect for universal endpoint
	http.Handle("/", router)

	// off-chain-auth router
	//r.Path(model.AuthRequestNoncePath).
	//	Name(requestNonceName).
	//	Methods(http.MethodGet).
	//	HandlerFunc(g.requestNonceHandler)
	//r.Path(model.AuthUpdateKeyPath).
	//	Name(updateUserPublicKey).
	//	Methods(http.MethodPost).
	//	HandlerFunc(g.updateUserPublicKeyHandler)

	// path style
	pathBucketRouter := router.PathPrefix("/{bucket}").Subrouter()
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
		Name(getBucketMetaRouterName).
		Methods(http.MethodGet).
		Queries(model.GetBucketMetaQuery, "").
		HandlerFunc(g.getBucketMetaHandler)
	pathBucketRouter.NewRoute().
		Name(getObjectMetaRouterName).
		Methods(http.MethodGet).
		Path("/{object:.+}").
		Queries(model.GetObjectMetaQuery, "").
		HandlerFunc(g.getObjectMetaHandler)
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

	router.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)
}
