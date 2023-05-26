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
	putObjectByOffsetRouterName           = "PutObjectByOffset"
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
	listBucketsByBucketIDRouterName       = "ListBucketsByBucketID"
	listObjectsByObjectIDRouterName       = "ListObjectsByObjectID"
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
	var routers []*mux.Router

	routers = append(routers, router.Host("{bucket:.+}."+g.domain).Subrouter())
	routers = append(routers, router.PathPrefix("/{bucket}").Subrouter())

	for _, router := range routers {
		// Put Object
		router.NewRoute().Name(putObjectRouterName).Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(g.putObjectHandler)

		// Put Object By Offset
		router.NewRoute().Name(putObjectByOffsetRouterName).Methods(http.MethodPost).Path("/{object:.+}").HandlerFunc(g.putObjectByOffsetHandler).Queries(
			"offset", "{offset}",
			"complete", "{complete}",
			"context", "{context}")

		// Get Object
		router.NewRoute().Name(getObjectRouterName).Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(g.getObjectHandler)

		// Get Bucket Read Quota
		router.NewRoute().Name(getBucketReadQuotaRouterName).Methods(http.MethodGet).HandlerFunc(g.getBucketReadQuotaHandler).Queries(
			model.GetBucketReadQuotaQuery, "",
			model.GetBucketReadQuotaMonthQuery, "{year_month}")

		// List Bucket Read Record
		router.NewRoute().Name(listBucketReadRecordRouterName).Methods(http.MethodGet).HandlerFunc(g.listBucketReadRecordHandler).Queries(
			model.ListBucketReadRecordQuery, "",
			model.ListBucketReadRecordMaxRecordsQuery, "{max_records}",
			model.StartTimestampUs, "{start_ts}",
			model.EndTimestampUs, "{end_ts}")

		// List Objects by bucket
		router.NewRoute().Name(listObjectsByBucketRouterName).Methods(http.MethodGet).Path("/").HandlerFunc(g.listObjectsByBucketNameHandler)

		// Get Bucket Meta
		router.NewRoute().Name(getBucketMetaRouterName).Methods(http.MethodGet).Queries(model.GetBucketMetaQuery, "").HandlerFunc(g.getBucketMetaHandler)

		// Get Object Meta
		router.NewRoute().Name(getObjectMetaRouterName).Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(g.getObjectMetaHandler).Queries(
			model.GetObjectMetaQuery, "")

		// Query upload progress
		router.NewRoute().Name(queryUploadProgressRouterName).Methods(http.MethodGet).Path("/{object:.+}.").HandlerFunc(g.queryUploadProgressHandler).Queries(
			model.UploadProgressQuery, "")

		// Not found
		router.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)
	}

	// group router
	router.Path("/").
		Name(getGroupListRouterName).
		Methods(http.MethodGet).
		Queries(GetGroupListGroupQuery, "").
		HandlerFunc(g.getGroupListHandler)
	router.Path("/").
		Name(listObjectsByObjectIDRouterName).
		Methods(http.MethodPost).
		Queries(ListObjectsByObjectID, "").
		HandlerFunc(g.listObjectsByObjectIDHandler)
	router.Path("/").
		Name(listBucketsByBucketIDRouterName).
		Methods(http.MethodPost).
		Queries(ListBucketsByBucketID, "").
		HandlerFunc(g.listBucketsByBucketIDHandler)
	router.Path("/").
		Name(getUserBucketsRouterName).
		Methods(http.MethodGet).
		HandlerFunc(g.getUserBucketsHandler)

	// bucket list router, path style
	router.Path("/").Name(getUserBucketsRouterName).Methods(http.MethodGet).HandlerFunc(g.getUserBucketsHandler)

	// verify permission router
	router.Path("/permission/{operator:.+}/{bucket:[^/]*}/{action-type:.+}").Name(verifyPermissionRouterName).Methods(http.MethodGet).HandlerFunc(g.verifyPermissionHandler)

	// admin router, path style
	// Get Approval
	router.Path(model.GetApprovalPath).Name(approvalRouterName).Methods(http.MethodGet).HandlerFunc(g.getApprovalHandler).
		Queries(model.ActionQuery, "{action}")

	// Challenge
	router.Path(model.ChallengePath).Name(challengeRouterName).Methods(http.MethodGet).HandlerFunc(g.challengeHandler)

	// replicate piece to receiver
	router.Path(model.ReplicateObjectPiecePath).Name(replicateObjectPieceRouterName).Methods(http.MethodPut).HandlerFunc(g.replicateHandler)

	// universal endpoint download
	router.Path("/download/{bucket:[^/]*}/{object:.+}").Name(downloadObjectByUniversalEndpointName).Methods(http.MethodGet).
		HandlerFunc(g.downloadObjectByUniversalEndpointHandler)
	// universal endpoint view
	router.Path("/view/{bucket:[^/]*}/{object:.+}").Name(viewObjectByUniversalEndpointName).Methods(http.MethodGet).
		HandlerFunc(g.viewObjectByUniversalEndpointHandler)

	// redirect for universal endpoint
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

	router.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)
	router.Use(localhttp.Limit)
}
