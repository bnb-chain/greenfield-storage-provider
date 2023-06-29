package gater

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	localhttp "github.com/bnb-chain/greenfield-storage-provider/pkg/middleware/http"
)

const (
	approvalRouterName                             = "GetApproval"
	putObjectRouterName                            = "PutObject"
	resumablePutObjectRouterName                   = "ResumablePutObject"
	queryResumeOffsetName                          = "QueryResumeOffsetName"
	getObjectRouterName                            = "GetObject"
	getChallengeInfoRouterName                     = "GetChallengeInfo"
	replicateObjectPieceRouterName                 = "ReplicateObjectPiece"
	getUserBucketsRouterName                       = "GetUserBuckets"
	listObjectsByBucketRouterName                  = "ListObjectsByBucketName"
	verifyPermissionRouterName                     = "VerifyPermission"
	getBucketReadQuotaRouterName                   = "GetBucketReadQuota"
	listBucketReadRecordRouterName                 = "ListBucketReadRecord"
	requestNonceName                               = "RequestNonce"
	updateUserPublicKey                            = "UpdateUserPublicKey"
	queryUploadProgressRouterName                  = "QueryUploadProgress"
	downloadObjectByUniversalEndpointName          = "DownloadObjectByUniversalEndpoint"
	viewObjectByUniversalEndpointName              = "ViewObjectByUniversalEndpoint"
	getObjectMetaRouterName                        = "GetObjectMeta"
	getBucketMetaRouterName                        = "GetBucketMeta"
	getGroupListRouterName                         = "GetGroupList"
	listBucketsByBucketIDRouterName                = "ListBucketsByBucketID"
	listObjectsByObjectIDRouterName                = "ListObjectsByObjectID"
	recoveryPieceRouterName                        = "RecoveryObjectPiece"
	getPieceFromSecondaryRouterName                = "GetPieceFromSecondary"
	getPaymentByBucketIDRouterName                 = "GetPaymentByBucketID"
	getPaymentByBucketNameRouterName               = "GetPaymentByBucketName"
	getBucketByBucketNameRouterName                = "GetBucketByBucketName"
	getBucketByBucketIDRouterName                  = "GetBucketByBucketID"
	listDeletedObjectsByBlockNumberRangeRouterName = "ListDeletedObjectsByBlockNumberRange"
	getUserBucketsCountRouterName                  = "GetUserBucketsCount"
	listExpiredBucketsBySpRouterName               = "ListExpiredBucketsBySp"
)

const (
	createBucketApprovalAction  = "CreateBucket"
	createObjectApprovalAction  = "CreateObject"
	migrateBucketApprovalAction = "MigrateBucket"
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
	router.Path("/permission/{operator:.+}/{bucket:[^/]*}/{action-type:.+}").Name(verifyPermissionRouterName).Methods(http.MethodGet).HandlerFunc(g.verifyPermissionHandler)

	// admin router, path style
	// Get Approval
	router.Path(GetApprovalPath).Name(approvalRouterName).Methods(http.MethodGet).HandlerFunc(g.getApprovalHandler).Queries(
		ActionQuery, "{action}")

	// Challenge
	router.Path(GetChallengeInfoPath).Name(getChallengeInfoRouterName).Methods(http.MethodGet).HandlerFunc(g.getChallengeInfoHandler)

	// replicate piece to receiver
	router.Path(ReplicateObjectPiecePath).Name(replicateObjectPieceRouterName).Methods(http.MethodPut).HandlerFunc(g.replicateHandler)
	router.Path(RecoverObjectPiecePath).Name(recoveryPieceRouterName).Methods(http.MethodGet).HandlerFunc(g.recoverPrimaryHandler)
	// universal endpoint download
	router.Path("/download/{bucket:[^/]*}/{object:.+}").Name(downloadObjectByUniversalEndpointName).Methods(http.MethodGet).
		HandlerFunc(g.downloadObjectByUniversalEndpointHandler)
	// universal endpoint view
	router.Path("/view/{bucket:[^/]*}/{object:.+}").Name(viewObjectByUniversalEndpointName).Methods(http.MethodGet).
		HandlerFunc(g.viewObjectByUniversalEndpointHandler)

	var routers []*mux.Router
	routers = append(routers, router.Host("{bucket:.+}."+g.domain).Subrouter())
	routers = append(routers, router.PathPrefix("/{bucket}").Subrouter())
	for _, r := range routers {
		// Put Object By Offset
		r.NewRoute().Name(resumablePutObjectRouterName).Methods(http.MethodPost).Path("/{object:.+}").HandlerFunc(g.resumablePutObjectHandler).Queries(
			"offset", "{offset}",
			"complete", "{complete}")
		// Put Object
		r.NewRoute().Name(putObjectRouterName).Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(g.putObjectHandler)

		// QueryPutObjectOffset
		r.NewRoute().Name(queryResumeOffsetName).Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(g.queryResumeOffsetHandler).Queries(
			UploadContextQuery, "")

		// Query upload progress
		r.NewRoute().Name(queryUploadProgressRouterName).Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(g.queryUploadProgressHandler).Queries(
			UploadProgressQuery, "")

		r.NewRoute().Name(getPieceFromSecondaryRouterName).Methods(http.MethodGet).Path("/{object:.+}").Queries(GetSecondaryPieceData, "").HandlerFunc(g.getRecoveryPieceHandler)
		// Get Bucket Meta
		r.NewRoute().Name(getBucketMetaRouterName).Methods(http.MethodGet).Queries(GetBucketMetaQuery, "").HandlerFunc(g.getBucketMetaHandler)

		// Get Object Meta
		r.NewRoute().Name(getObjectMetaRouterName).Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(g.getObjectMetaHandler).Queries(
			GetObjectMetaQuery, "")

		// Get Object
		r.NewRoute().Name(getObjectRouterName).Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(g.getObjectHandler)

		// Get Bucket Read Quota
		r.NewRoute().Name(getBucketReadQuotaRouterName).Methods(http.MethodGet).HandlerFunc(g.getBucketReadQuotaHandler).Queries(
			GetBucketReadQuotaQuery, "",
			GetBucketReadQuotaMonthQuery, "{year_month}")

		if g.env != gfspapp.EnvMainnet {
			// Get Bucket By Bucket Name
			r.NewRoute().Name(getBucketByBucketNameRouterName).Methods(http.MethodGet).Queries(GetBucketByBucketNameQuery, "").HandlerFunc(g.getBucketByBucketNameHandler)

			// Get Payment By Bucket Name
			r.NewRoute().Name(getPaymentByBucketNameRouterName).Methods(http.MethodGet).Queries(GetPaymentByBucketNameQuery, "").HandlerFunc(g.getPaymentByBucketNameHandler)
		}

		// List Bucket Read Record
		r.NewRoute().Name(listBucketReadRecordRouterName).Methods(http.MethodGet).HandlerFunc(g.listBucketReadRecordHandler).Queries(
			ListBucketReadRecordQuery, "",
			ListBucketReadRecordMaxRecordsQuery, "{max_records}",
			StartTimestampUs, "{start_ts}",
			EndTimestampUs, "{end_ts}")

		// List Objects by bucket
		r.NewRoute().Name(listObjectsByBucketRouterName).Methods(http.MethodGet).Path("/").HandlerFunc(g.listObjectsByBucketNameHandler)

		// Not found
		r.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)
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
		Queries(ListObjectsByObjectIDQuery, "").
		HandlerFunc(g.listObjectsByObjectIDHandler)
	router.Path("/").
		Name(listBucketsByBucketIDRouterName).
		Methods(http.MethodPost).
		Queries(ListBucketsByBucketIDQuery, "").
		HandlerFunc(g.listBucketsByBucketIDHandler)

	if g.env != gfspapp.EnvMainnet {
		// Get Payment By Bucket ID
		router.Path("/").Name(getPaymentByBucketIDRouterName).Methods(http.MethodGet).Queries(GetPaymentByBucketIDQuery, "").HandlerFunc(g.getPaymentByBucketIDHandler)

		// Get Bucket By Bucket ID
		router.Path("/").Name(getBucketByBucketIDRouterName).Methods(http.MethodGet).Queries(GetBucketByBucketIDQuery, "").HandlerFunc(g.getBucketByBucketIDHandler)

		// List Deleted Objects
		router.Path("/").Name(listDeletedObjectsByBlockNumberRangeRouterName).Methods(http.MethodGet).Queries(ListDeletedObjectsQuery, "").HandlerFunc(g.listDeletedObjectsByBlockNumberRangeHandler)

		//Get User Buckets Count
		router.Path("/").Name(getUserBucketsCountRouterName).Methods(http.MethodGet).Queries(GetUserBucketsCountQuery, "").HandlerFunc(g.getUserBucketsCountHandler)

		//List Expired Buckets By Sp
		router.Path("/").Name(listExpiredBucketsBySpRouterName).Methods(http.MethodGet).Queries(ListExpiredBucketsBySpQuery, "").HandlerFunc(g.listExpiredBucketsBySpHandler)
	}

	router.Path("/").
		Name(getUserBucketsRouterName).
		Methods(http.MethodGet).
		HandlerFunc(g.getUserBucketsHandler)

	// bucket list router, path style
	router.Path("/").Name(getUserBucketsRouterName).Methods(http.MethodGet).HandlerFunc(g.getUserBucketsHandler)

	// redirect for universal endpoint
	http.Handle("/", router)

	router.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)
	router.Use(localhttp.Limit)
}
