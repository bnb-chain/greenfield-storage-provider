package gater

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	localhttp "github.com/bnb-chain/greenfield-storage-provider/pkg/middleware/http"
)

const (
	approvalRouterName                    = "GetApproval"
	putObjectRouterName                   = "PutObject"
	getObjectRouterName                   = "GetObject"
	getChallengeInfoRouterName            = "GetChallengeInfo"
	replicateObjectPieceRouterName        = "ReplicateObjectPiece"
	getUserBucketsRouterName              = "GetUserBuckets"
	listObjectsByBucketRouterName         = "ListObjectsByBucketName"
	verifyPermissionRouterName            = "VerifyPermission"
	getBucketReadQuotaRouterName          = "GetBucketReadQuota"
	listBucketReadRecordRouterName        = "ListBucketReadRecord"
	requestNonceName                      = "RequestNonce"
	updateUserPublicKey                   = "UpdateUserPublicKey"
	queryUploadProgressRouterName         = "QueryUploadProgress"
	downloadObjectByUniversalEndpointName = "DownloadObjectByUniversalEndpoint"
	viewObjectByUniversalEndpointName     = "ViewObjectByUniversalEndpoint"
	getObjectMetaRouterName               = "GetObjectMeta"
	getBucketMetaRouterName               = "GetBucketMeta"
	getGroupListRouterName                = "GetGroupList"
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
		Queries(UploadProgressQuery, "").
		HandlerFunc(g.queryUploadProgressHandler)
	hostBucketRouter.NewRoute().
		Name(getBucketMetaRouterName).
		Methods(http.MethodGet).
		Queries(GetBucketMetaQuery, "").
		HandlerFunc(g.getBucketMetaHandler)
	hostBucketRouter.NewRoute().
		Name(getObjectMetaRouterName).
		Methods(http.MethodGet).
		Path("/{object:.+}").
		Queries(GetObjectMetaQuery, "").
		HandlerFunc(g.getObjectMetaHandler)
	hostBucketRouter.NewRoute().
		Name(getObjectRouterName).
		Methods(http.MethodGet).
		Path("/{object:.+}").
		HandlerFunc(g.getObjectHandler)
	hostBucketRouter.NewRoute().
		Name(getBucketReadQuotaRouterName).
		Methods(http.MethodGet).
		Queries(GetBucketReadQuotaQuery, "",
			GetBucketReadQuotaMonthQuery, "{year_month}").
		HandlerFunc(g.getBucketReadQuotaHandler)
	hostBucketRouter.NewRoute().
		Name(listBucketReadRecordRouterName).
		Methods(http.MethodGet).
		Queries(ListBucketReadRecordQuery, "",
			ListBucketReadRecordMaxRecordsQuery, "{max_records}",
			StartTimestampUs, "{start_ts}",
			EndTimestampUs, "{end_ts}").
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
	router.Path(GetApprovalPath).
		Name(approvalRouterName).
		Methods(http.MethodGet).
		Queries(ActionQuery, "{action}").
		HandlerFunc(g.getApprovalHandler)
	router.Path(GetChallengeInfoPath).
		Name(getChallengeInfoRouterName).
		Methods(http.MethodGet).
		HandlerFunc(g.getChallengeInfoHandler)
	// replicate piece to receiver
	router.Path(ReplicateObjectPiecePath).
		Name(replicateObjectPieceRouterName).
		Methods(http.MethodPut).
		HandlerFunc(g.replicateHandler)
	// universal endpoint download
	router.Path("/download/{bucket:[^/]*}/{object:.+}").
		Name(downloadObjectByUniversalEndpointName).
		Methods(http.MethodGet).
		HandlerFunc(g.downloadObjectByUniversalEndpointHandler)
	// universal endpoint view
	router.Path("/view/{bucket:[^/]*}/{object:.+}").
		Name(viewObjectByUniversalEndpointName).
		Methods(http.MethodGet).
		HandlerFunc(g.viewObjectByUniversalEndpointHandler)
	//redirect for universal endpoint
	http.Handle("/", router)

	// off-chain-auth router
	router.Path(AuthRequestNoncePath).
		Name(requestNonceName).
		Methods(http.MethodGet).
		HandlerFunc(g.requestNonceHandler)
	router.Path(AuthUpdateKeyPath).
		Name(updateUserPublicKey).
		Methods(http.MethodPost).
		HandlerFunc(g.updateUserPublicKeyHandler)

	// verify permission router
	router.Path("/permission/{operator:.+}/{bucket:[^/]*}/{action-type:.+}").
		Name(verifyPermissionRouterName).
		Methods(http.MethodGet).
		HandlerFunc(g.verifyPermissionHandler)

	// group router
	router.Path(GroupListPath).
		Name(getGroupListRouterName).
		Methods(http.MethodGet).
		HandlerFunc(g.getGroupListHandler)

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
		Queries(UploadProgressQuery, "").
		HandlerFunc(g.queryUploadProgressHandler)
	pathBucketRouter.NewRoute().
		Name(getBucketMetaRouterName).
		Methods(http.MethodGet).
		Queries(GetBucketMetaQuery, "").
		HandlerFunc(g.getBucketMetaHandler)
	pathBucketRouter.NewRoute().
		Name(getObjectMetaRouterName).
		Methods(http.MethodGet).
		Path("/{object:.+}").
		Queries(GetObjectMetaQuery, "").
		HandlerFunc(g.getObjectMetaHandler)
	pathBucketRouter.NewRoute().
		Name(getObjectRouterName).
		Methods(http.MethodGet).
		Path("/{object:.+}").
		HandlerFunc(g.getObjectHandler)
	pathBucketRouter.NewRoute().
		Name(getBucketReadQuotaRouterName).
		Methods(http.MethodGet).
		Queries(GetBucketReadQuotaQuery, "",
			GetBucketReadQuotaMonthQuery, "{year_month}").
		HandlerFunc(g.getBucketReadQuotaHandler)
	pathBucketRouter.NewRoute().
		Name(listBucketReadRecordRouterName).
		Methods(http.MethodGet).
		Queries(ListBucketReadRecordQuery, "",
			ListBucketReadRecordMaxRecordsQuery, "{max_records}",
			StartTimestampUs, "{start_ts}",
			EndTimestampUs, "{end_ts}").
		HandlerFunc(g.listBucketReadRecordHandler)
	pathBucketRouter.NewRoute().
		Name(listObjectsByBucketRouterName).
		Methods(http.MethodGet).
		Path("/").
		HandlerFunc(g.listObjectsByBucketNameHandler)
	pathBucketRouter.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)

	router.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)
	router.Use(localhttp.Limit)
}
