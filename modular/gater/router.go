package gater

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	mwhttp "github.com/bnb-chain/greenfield-storage-provider/pkg/middleware/http"
)

const (
	approvalRouterName                             = "GetApproval"
	putObjectRouterName                            = "PutObject"
	resumablePutObjectRouterName                   = "ResumablePutObject"
	delegatePutObjectRouterName                    = "DelegatePutObject"
	delegateResumablePutObjectRouterName           = "DelegateResumablePutObject"
	delegateCreateFolderRouterName                 = "DelegateCreateFolder"
	queryResumeOffsetName                          = "QueryResumeOffsetName"
	getObjectRouterName                            = "GetObject"
	getChallengeInfoRouterName                     = "GetChallengeInfo"
	getChallengeInfoV2RouterName                   = "GetChallengeInfoV2"
	replicateObjectPieceRouterName                 = "ReplicateObjectPiece"
	getUserBucketsRouterName                       = "GetUserBuckets"
	listObjectsByBucketRouterName                  = "ListObjectsByBucketName"
	verifyPermissionRouterName                     = "VerifyPermission"
	getBucketReadQuotaRouterName                   = "GetBucketReadQuota"
	listBucketReadRecordRouterName                 = "ListBucketReadRecord"
	listBucketReadQuotaRouterName                  = "ListBucketReadQuota"
	getBucketReadQuotaCountRouterName              = "GetBucketReadQuotaCount"
	requestNonceRouterName                         = "RequestNonce"
	updateUserPublicKeyRouterName                  = "UpdateUserPublicKey"
	updateUserPublicKeyV2RouterName                = "UpdateUserPublicKeyV2"
	listUserPublicKeyV2RouterName                  = "ListUserPublicKeyV2"
	deleteUserPublicKeyV2RouterName                = "DeleteUserPublicKeyV2"
	queryUploadProgressRouterName                  = "QueryUploadProgress"
	downloadObjectByUniversalEndpointName          = "DownloadObjectByUniversalEndpoint"
	viewObjectByUniversalEndpointName              = "ViewObjectByUniversalEndpoint"
	getObjectMetaRouterName                        = "GetObjectMeta"
	getBucketMetaRouterName                        = "GetBucketMeta"
	queryMigrationProgressRouterName               = "QueryMigrationProgress"
	getGroupListRouterName                         = "GetGroupList"
	listBucketsByIDsRouterName                     = "ListBucketsByIDs"
	listObjectsByIDsRouterName                     = "ListObjectsByIDs"
	recoveryPieceRouterName                        = "RecoveryObjectPiece"
	getPieceFromSecondaryRouterName                = "GetPieceFromSecondary"
	getPaymentByBucketIDRouterName                 = "GetPaymentByBucketID"
	getPaymentByBucketNameRouterName               = "GetPaymentByBucketName"
	getBucketByBucketNameRouterName                = "GetBucketByBucketName"
	getBucketByBucketIDRouterName                  = "GetBucketByBucketID"
	listDeletedObjectsByBlockNumberRangeRouterName = "ListDeletedObjectsByBlockNumberRange"
	getUserBucketsCountRouterName                  = "GetUserBucketsCount"
	listExpiredBucketsBySpRouterName               = "ListExpiredBucketsBySp"
	notifyMigrateSwapOutRouterName                 = "NotifyMigrateSwapOut"
	queryBucketQuotaRouterName                     = "QueryBucketQuota"
	queryBucketEnoughQuotaRouterName               = "QueryBucketEnoughQuota"
	PreMigrateBucketRouterName                     = "PreMigrateBucket"
	PostMigrateBucketRouterName                    = "PostMigrateBucket"
	migratePieceRouterName                         = "MigratePiece"
	migrationBucketApprovalName                    = "MigrationBucketApproval"
	swapOutApprovalName                            = "SwapOutApproval"
	listVirtualGroupFamiliesBySpIDRouterName       = "ListVirtualGroupFamiliesBySpID"
	getVirtualGroupFamilyRouterName                = "GetVirtualGroupFamily"
	getGlobalVirtualGroupByGvgIDRouterName         = "GetGlobalVirtualGroupByGvgID"
	getGlobalVirtualGroupRouterName                = "GetGlobalVirtualGroup"
	listGlobalVirtualGroupsBySecondarySPRouterName = "ListGlobalVirtualGroupsBySecondarySP"
	listGlobalVirtualGroupsByBucketRouterName      = "ListGlobalVirtualGroupsByBucket"
	listObjectsInGVGAndBucketRouterName            = "ListObjectsInGVGAndBucket"
	listObjectsByGVGAndBucketForGCRouterName       = "ListObjectsByGVGAndBucketForGC"
	listObjectsInGVGRouterName                     = "ListObjectsInGVG"
	listMigrateBucketEventsRouterName              = "ListMigrateBucketEvents"
	listSwapOutEventsRouterName                    = "ListSwapOutEvents"
	listSpExitEventsRouterName                     = "ListSpExitEvents"
	verifyPermissionByIDRouterName                 = "VerifyPermissionByID"
	getSPInfoRouterName                            = "GetSPInfo"
	getStatusRouterName                            = "GetStatus"
	getUserGroupsRouterName                        = "GetUserGroups"
	getGroupMembersRouterName                      = "GetGroupMembers"
	getUserOwnedGroupsRouterName                   = "GetUserOwnedGroups"
	listObjectPoliciesRouterName                   = "ListObjectPolicies"
	listUserPaymentAccountsRouterName              = "ListUserPaymentAccounts"
	listPaymentAccountStreamsRouterName            = "ListPaymentAccountStreams"
	listGroupsByIDsRouterName                      = "ListGroupsByIDs"
	getSPMigratingBucketNumberRouterName           = "GetSPMigratingBucketNumber"
	verifyMigrateGVGPermissionRouterName           = "VerifyMigrateGVGPermission"
	getBucketSizeRouterName                        = "GetBucketSize"
	getRecommendedVGFRouterName                    = "GetRecommendedVGF"
)

const (
	createBucketApprovalAction  = "CreateBucket"
	createObjectApprovalAction  = "CreateObject"
	migrateBucketApprovalAction = "MigrateBucket"
)

// notFoundHandler log not found request info.
func (g *GateModular) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Errorw("failed to find the corresponding handler", "method", r.Method, "host", r.Host, "url", r.URL)
	if _, err := io.ReadAll(r.Body); err != nil {
		log.Errorw("failed to read the unknown request", "error", err)
	}
}

// RegisterHandler registers the handlers to the gateway router.
func (g *GateModular) RegisterHandler(router *mux.Router) {
	// off-chain-auth router
	router.Path(AuthRequestNoncePath).Name(requestNonceRouterName).Methods(http.MethodGet).HandlerFunc(g.requestNonceHandler)
	router.Path(AuthUpdateKeyPath).Name(updateUserPublicKeyRouterName).Methods(http.MethodPost).HandlerFunc(g.updateUserPublicKeyHandler)
	router.Path(AuthUpdateKeyV2Path).Name(updateUserPublicKeyV2RouterName).Methods(http.MethodPost).HandlerFunc(g.updateUserPublicKeyV2Handler)
	router.Path(ListAuthKeyV2Path).Name(listUserPublicKeyV2RouterName).Methods(http.MethodGet).HandlerFunc(g.listUserPublicKeyV2Handler)
	router.Path(AuthDeleteKeysV2Path).Name(deleteUserPublicKeyV2RouterName).Methods(http.MethodPost).HandlerFunc(g.deleteUserPublicKeyV2Handler)

	// verify permission router
	router.Path("/permission/{operator:.+}/{bucket:[^/]*}/{action-type:.+}").Name(verifyPermissionRouterName).
		Methods(http.MethodGet).HandlerFunc(g.verifyPermissionHandler)

	// admin router, path style
	// Get Approval
	router.Path(GetApprovalPath).Name(approvalRouterName).Methods(http.MethodGet).HandlerFunc(g.getApprovalHandler).
		Queries(ActionQuery, "{action}")

	// Query recommended virtual group family for creating bucket
	router.Path(GetRecommendedVirtualGroupFamilyPath).Name(getRecommendedVGFRouterName).Methods(http.MethodGet).HandlerFunc(g.getRecommendedVGFIDHandler)

	// get challenge info
	router.Path(GetChallengeInfoPath).Name(getChallengeInfoRouterName).Methods(http.MethodGet).HandlerFunc(g.getChallengeInfoHandler)
	router.Path(GetChallengeInfoV2Path).Name(getChallengeInfoV2RouterName).Methods(http.MethodGet).HandlerFunc(g.getChallengeInfoV2Handler)

	// dest sp query latest bucket quota from src sp
	router.Path(MigrateQueryBucketQuotaPath).Name(queryBucketQuotaRouterName).Methods(http.MethodGet).HandlerFunc(g.getLatestBucketQuotaHandler)
	// dest sp query whether src sp has enough quota for bucket migration
	router.Path(MigrateQueryBucketQuotaHasEnoughQuotaPath).Name(queryBucketEnoughQuotaRouterName).Methods(http.MethodGet).HandlerFunc(g.sufficientQuotaForBucketMigrationHandler)
	// dest sp notify src sp to pre migrate bucket
	router.Path(PreMigrateBucketPath).Name(PreMigrateBucketRouterName).Methods(http.MethodGet).HandlerFunc(g.preMigrateBucketHandler)
	// dest sp notify src sp to post migrate bucket
	router.Path(PostMigrateBucketPath).Name(PostMigrateBucketRouterName).Methods(http.MethodGet).HandlerFunc(g.postMigrateBucketHandler)

	// replicate piece to receiver
	router.Path(ReplicateObjectPiecePath).Name(replicateObjectPieceRouterName).Methods(http.MethodPut).HandlerFunc(g.replicateHandler)
	// data recovery
	router.Path(RecoverObjectPiecePath).Name(recoveryPieceRouterName).Methods(http.MethodGet).HandlerFunc(g.getRecoverDataHandler)

	// dest sp receive swap out notify from src sp.
	router.Path(NotifyMigrateSwapOutTaskPath).Name(notifyMigrateSwapOutRouterName).Methods(http.MethodPost).HandlerFunc(g.notifyMigrateSwapOutHandler)
	// dest sp pull piece data from src sp, for sp exit and bucket migrate.
	router.Path(MigratePiecePath).Name(migratePieceRouterName).Methods(http.MethodGet).HandlerFunc(g.migratePieceHandler)
	// migration bucket approval for secondary sp bls signature.
	router.Path(SecondarySPMigrationBucketApprovalPath).Name(migrationBucketApprovalName).Methods(http.MethodGet).HandlerFunc(g.getSecondaryBlsMigrationBucketApprovalHandler)
	// swap out approval for sp exiting.
	router.Path(SwapOutApprovalPath).Name(swapOutApprovalName).Methods(http.MethodGet).HandlerFunc(g.getSwapOutApproval)

	// universal endpoint download
	router.Path("/download/{bucket:[^/]*}/{object:.+}").Name(downloadObjectByUniversalEndpointName).Methods(http.MethodGet).
		HandlerFunc(g.downloadObjectByUniversalEndpointHandler)
	router.Path("/download").Name(downloadObjectByUniversalEndpointName).Methods(http.MethodGet).
		Queries(UniversalEndpointSpecialSuffixQuery, "{bucket:[^/]*}/{object:.+}").HandlerFunc(g.downloadObjectByUniversalEndpointHandler)
	// universal endpoint view
	router.Path("/view/{bucket:[^/]*}/{object:.+}").Name(viewObjectByUniversalEndpointName).Methods(http.MethodGet).
		HandlerFunc(g.viewObjectByUniversalEndpointHandler)
	router.Path("/view").Name(viewObjectByUniversalEndpointName).Methods(http.MethodGet).
		Queries(UniversalEndpointSpecialSuffixQuery, "{bucket:[^/]*}/{object:.+}").HandlerFunc(g.viewObjectByUniversalEndpointHandler)

	router.Path(StatusPath).Name(getStatusRouterName).Methods(http.MethodGet).HandlerFunc(g.getStatusHandler)

	var routers []*mux.Router
	routers = append(routers, router.Host("{bucket:.+}."+g.domain).Subrouter())
	routers = append(routers, router.PathPrefix("/{bucket}").Subrouter())
	for _, r := range routers {
		// Delegate Put Object By Offset
		r.NewRoute().Name(delegateResumablePutObjectRouterName).Methods(http.MethodPost).Path("/{object:.+}").Queries(Delegate, "").HandlerFunc(g.delegateResumablePutObjectHandler)
		// Put Object By Offset
		r.NewRoute().Name(resumablePutObjectRouterName).Methods(http.MethodPost).Path("/{object:.+}").HandlerFunc(g.resumablePutObjectHandler).
			Queries("offset", "{offset}", "complete", "{complete}")
		// Delegate Put Object
		r.NewRoute().Name(delegatePutObjectRouterName).Methods(http.MethodPut).Path("/{object:.+}").Queries(Delegate, "").HandlerFunc(g.delegatePutObjectHandler)
		// Put Object
		r.NewRoute().Name(putObjectRouterName).Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(g.putObjectHandler)
		// create folder
		r.NewRoute().Name(delegateCreateFolderRouterName).Methods(http.MethodPost).Path("/{object:.+}").HandlerFunc(g.delegateCreateFolderHandler).Queries(CreateFolderQuery, "")

		// QueryPutObjectOffset
		r.NewRoute().Name(queryResumeOffsetName).Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(g.queryResumeOffsetHandler).
			Queries(UploadContextQuery, "")

		// Query upload progress
		r.NewRoute().Name(queryUploadProgressRouterName).Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(g.queryUploadProgressHandler).
			Queries(UploadProgressQuery, "")

		// Get Bucket Meta
		r.NewRoute().Name(getBucketMetaRouterName).Methods(http.MethodGet).Queries(GetBucketMetaQuery, "").HandlerFunc(g.getBucketMetaHandler)

		// Query migration progress
		r.NewRoute().Name(queryMigrationProgressRouterName).Methods(http.MethodGet).HandlerFunc(g.queryBucketMigrationProgressHandler).
			Queries(GetBucketMigrationProgressQuery, "")

		// Get Object Meta
		r.NewRoute().Name(getObjectMetaRouterName).Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(g.getObjectMetaHandler).
			Queries(GetObjectMetaQuery, "")

		// List Object Policies
		r.NewRoute().Name(listObjectPoliciesRouterName).Methods(http.MethodGet).Path("/{object:.+}").Queries(ListObjectPoliciesQuery, "").
			HandlerFunc(g.listObjectPoliciesHandler)

		// Get Object
		r.NewRoute().Name(getObjectRouterName).Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(g.getObjectHandler)

		// Get Bucket Read Quota
		r.NewRoute().Name(getBucketReadQuotaRouterName).Methods(http.MethodGet).HandlerFunc(g.getBucketReadQuotaHandler).Queries(
			GetBucketReadQuotaQuery, "", GetBucketReadQuotaMonthQuery, "{year_month}")

		if g.env != gfspapp.EnvMainnet {
			// Get Bucket By Bucket Name
			r.NewRoute().Name(getBucketByBucketNameRouterName).Methods(http.MethodGet).Queries(GetBucketByBucketNameQuery, "").HandlerFunc(g.getBucketByBucketNameHandler)

			// Get Payment By Bucket Name
			r.NewRoute().Name(getPaymentByBucketNameRouterName).Methods(http.MethodGet).Queries(GetPaymentByBucketNameQuery, "").HandlerFunc(g.getPaymentByBucketNameHandler)
		}

		// List Bucket Read Record
		r.NewRoute().Name(listBucketReadRecordRouterName).Methods(http.MethodGet).HandlerFunc(g.listBucketReadRecordHandler).Queries(
			ListBucketReadRecordQuery, "", ListBucketReadRecordMaxRecordsQuery, "{max_records}",
			StartTimestampUs, "{start_ts}", EndTimestampUs, "{end_ts}")

		// List Objects by bucket
		r.NewRoute().Name(listObjectsByBucketRouterName).Methods(http.MethodGet).Path("/").HandlerFunc(g.listObjectsByBucketNameHandler)

		// Not found
		r.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)
	}

	// Get Group List
	router.Path("/").Name(getGroupListRouterName).Methods(http.MethodGet).Queries(GetGroupListGroupQuery, "").HandlerFunc(g.getGroupListHandler)

	// List Objects By IDs
	router.Path("/").Name(listObjectsByIDsRouterName).Methods(http.MethodGet).Queries(ListObjectsByIDsQuery, "").HandlerFunc(g.listObjectsByIDsHandler)

	// List Buckets By IDs
	router.Path("/").Name(listBucketsByIDsRouterName).Methods(http.MethodGet).Queries(ListBucketsByIDsQuery, "").HandlerFunc(g.listBucketsByIDsHandler)

	// Verify Permission By ID
	router.Path("/").Name(verifyPermissionByIDRouterName).Methods(http.MethodGet).Queries(VerifyPermissionByIDQuery, "").HandlerFunc(g.verifyPermissionByIDHandler)

	// Get User Groups
	router.Path("/").Name(getUserGroupsRouterName).Methods(http.MethodGet).Queries(GetUserGroupsQuery, "").HandlerFunc(g.getUserGroupsHandler)

	// Get Group Members
	router.Path("/").Name(getGroupMembersRouterName).Methods(http.MethodGet).Queries(GetGroupMembersQuery, "").HandlerFunc(g.getGroupMembersHandler)

	// Get User Owned Groups
	router.Path("/").Name(getUserOwnedGroupsRouterName).Methods(http.MethodGet).Queries(GetUserOwnedGroupsQuery, "").HandlerFunc(g.getUserOwnedGroupsHandler)

	// List Payment Account Streams
	router.Path("/").Name(listPaymentAccountStreamsRouterName).Methods(http.MethodGet).Queries(ListPaymentAccountStreamsQuery, "").HandlerFunc(g.listPaymentAccountStreamsHandler)

	// List User Payment Accounts
	router.Path("/").Name(listUserPaymentAccountsRouterName).Methods(http.MethodGet).Queries(ListUserPaymentAccountsQuery, "").HandlerFunc(g.listUserPaymentAccountsHandler)

	// List Groups By IDs
	router.Path("/").Name(listGroupsByIDsRouterName).Methods(http.MethodGet).Queries(ListGroupsByIDsQuery, "").HandlerFunc(g.listGroupsByIDsHandler)

	// List Bucket Read Quota by time
	router.Path("/").Name(listBucketReadQuotaRouterName).Methods(http.MethodGet).Queries(ListBucketReadRecordQuery, "").HandlerFunc(g.listBucketReadQuotaHandler)

	// List Bucket Read Quota Count
	router.Path("/").Name(getBucketReadQuotaCountRouterName).Methods(http.MethodGet).Queries(ListBucketReadCountQuery, "").HandlerFunc(g.getBucketReadQuotaCountHandler)

	if g.env != gfspapp.EnvMainnet {
		// Get Payment By Bucket ID
		router.Path("/").Name(getPaymentByBucketIDRouterName).Methods(http.MethodGet).Queries(GetPaymentByBucketIDQuery, "").HandlerFunc(g.getPaymentByBucketIDHandler)

		// Get Bucket By Bucket ID
		router.Path("/").Name(getBucketByBucketIDRouterName).Methods(http.MethodGet).Queries(GetBucketByBucketIDQuery, "").HandlerFunc(g.getBucketByBucketIDHandler)

		// List Deleted Objects
		router.Path("/").Name(listDeletedObjectsByBlockNumberRangeRouterName).Methods(http.MethodGet).Queries(ListDeletedObjectsQuery, "").HandlerFunc(g.listDeletedObjectsByBlockNumberRangeHandler)

		// Get User Buckets Count
		router.Path("/").Name(getUserBucketsCountRouterName).Methods(http.MethodGet).Queries(GetUserBucketsCountQuery, "").HandlerFunc(g.getUserBucketsCountHandler)

		// List Expired Buckets By Sp
		router.Path("/").Name(listExpiredBucketsBySpRouterName).Methods(http.MethodGet).Queries(ListExpiredBucketsBySpQuery, "").HandlerFunc(g.listExpiredBucketsBySpHandler)

		// List Virtual Group Families By Sp ID
		router.Path("/").Name(listVirtualGroupFamiliesBySpIDRouterName).Methods(http.MethodGet).Queries(ListVirtualGroupFamiliesBySpIDQuery, "").HandlerFunc(g.listVirtualGroupFamiliesBySpIDHandler)

		// Get Virtual Group Families By Vgf ID
		router.Path("/").Name(getVirtualGroupFamilyRouterName).Methods(http.MethodGet).Queries(GetVirtualGroupFamilyQuery, "").HandlerFunc(g.getVirtualGroupFamilyHandler)

		// Get Global Virtual Group By Gvg ID
		router.Path("/").Name(getGlobalVirtualGroupByGvgIDRouterName).Methods(http.MethodGet).Queries(GetGlobalVirtualGroupByGvgIDQuery, "").HandlerFunc(g.getGlobalVirtualGroupByGvgIDHandler)

		// Get Global Virtual Group By Lvg ID And Bucket ID
		router.Path("/").Name(getGlobalVirtualGroupRouterName).Methods(http.MethodGet).Queries(GetGlobalVirtualGroupQuery, "").HandlerFunc(g.getGlobalVirtualGroupHandler)

		// List Global Virtual Group By Secondary Sp ID
		router.Path("/").Name(listGlobalVirtualGroupsBySecondarySPRouterName).Methods(http.MethodGet).Queries(ListGlobalVirtualGroupsBySecondarySPQuery, "").HandlerFunc(g.listGlobalVirtualGroupsBySecondarySPHandler)

		// List Global Virtual Group By Bucket ID
		router.Path("/").Name(listGlobalVirtualGroupsByBucketRouterName).Methods(http.MethodGet).Queries(ListGlobalVirtualGroupsByBucketQuery, "").HandlerFunc(g.listGlobalVirtualGroupsByBucketHandler)

		// List Objects By Gvg And Bucket ID
		router.Path("/").Name(listObjectsInGVGAndBucketRouterName).Methods(http.MethodGet).Queries(ListObjectsInGVGAndBucketQuery, "").HandlerFunc(g.listObjectsInGVGAndBucketHandler)

		// List Objects By Gvg And Bucket ID
		router.Path("/").Name(listObjectsByGVGAndBucketForGCRouterName).Methods(http.MethodGet).Queries(ListObjectsByGVGAndBucketForGCQuery, "").HandlerFunc(g.listObjectsByGVGAndBucketForGCHandler)

		// List Objects By Gvg And Bucket ID
		router.Path("/").Name(listObjectsInGVGRouterName).Methods(http.MethodGet).Queries(ListObjectsInGVGQuery, "").HandlerFunc(g.listObjectsInGVGHandler)

		// List Migrate Bucket Events
		router.Path("/").Name(listMigrateBucketEventsRouterName).Methods(http.MethodGet).Queries(ListMigrateBucketEventsQuery, "").HandlerFunc(g.listMigrateBucketEventsHandler)

		// List Swap Out Events
		router.Path("/").Name(listSwapOutEventsRouterName).Methods(http.MethodGet).Queries(ListSwapOutEventsQuery, "").HandlerFunc(g.listSwapOutEventsHandler)

		// List Sp Exit Events
		router.Path("/").Name(listSpExitEventsRouterName).Methods(http.MethodGet).Queries(ListSpExitEventsQuery, "").HandlerFunc(g.listSpExitEventsHandler)

		// Get Sp info by operator address
		router.Path("/").Name(getSPInfoRouterName).Methods(http.MethodGet).Queries(GetSPInfoQuery, "").HandlerFunc(g.getSPInfoHandler)

		// Get SP Migrate Bucket Number
		router.Path("/").Name(getSPMigratingBucketNumberRouterName).Methods(http.MethodGet).Queries(GetSPMigratingBucketNumberQuery, "").HandlerFunc(g.getSPMigratingBucketNumberHandler)

		// Verify the destination sp id of bucket migration & swap out
		router.Path("/").Name(verifyMigrateGVGPermissionRouterName).Methods(http.MethodGet).Queries(VerifyMigrateGVGPermissionQuery, "").HandlerFunc(g.verifyMigrateGVGPermissionHandler)

		// Verify the destination sp id of bucket migration & swap out
		router.Path("/").Name(getBucketSizeRouterName).Methods(http.MethodGet).Queries(GetBucketSizeQuery, "").HandlerFunc(g.getBucketSizeHandler)
	}

	// bucket list router, path style
	router.Path("/").Name(getUserBucketsRouterName).Methods(http.MethodGet).HandlerFunc(g.getUserBucketsHandler)

	// redirect for universal endpoint
	http.Handle("/", router)

	router.NotFoundHandler = http.HandlerFunc(g.notFoundHandler)
	router.Use(mwhttp.Limit(g.domain))
}
