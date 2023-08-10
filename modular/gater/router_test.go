package gater

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
)

var (
	testDomain = "www.route-test.com"
	gw         = &GateModular{
		env:    gfspapp.EnvLocal,
		domain: testDomain,
	}
	scheme     = "https://"
	bucketName = "test-bucket-name"
	objectName = "test-object-name"
)

func setupRouter(t *testing.T) *mux.Router {
	gwRouter := mux.NewRouter().SkipClean(true)
	gw.RegisterHandler(gwRouter)
	return gwRouter
}

func TestRouters(t *testing.T) {
	gwRouter := setupRouter(t)
	testCases := []struct {
		name             string
		router           *mux.Router // the router being tested
		method           string      // the request method
		url              string      // the request url, include path + query
		shouldMatch      bool        // whether the request is expected to match the route at all
		wantedRouterName string      // the matched router name
	}{
		{
			name:             "Get create bucket approval router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + GetApprovalPath + "?" + ActionQuery + "=" + createBucketApprovalAction,
			shouldMatch:      true,
			wantedRouterName: approvalRouterName,
		},
		{
			name:             "Get create object approval router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + GetApprovalPath + "?" + ActionQuery + "=" + createObjectApprovalAction,
			shouldMatch:      true,
			wantedRouterName: approvalRouterName,
		},
		{
			name:             "Notify migrate gvg task",
			router:           gwRouter,
			method:           http.MethodPost,
			url:              scheme + testDomain + NotifyMigrateSwapOutTaskPath,
			shouldMatch:      true,
			wantedRouterName: notifyMigrateSwapOutRouterName,
		},
		{
			name:             "Swap out approval",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + SwapOutApprovalPath,
			shouldMatch:      true,
			wantedRouterName: swapOutApprovalName,
		},
		{
			name:             "Migrate piece data",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + MigratePiecePath,
			shouldMatch:      true,
			wantedRouterName: migratePieceRouterName,
		},
		{
			name:             "Get secondary migrate bucket approval",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + SecondarySPMigrationBucketApprovalPath,
			shouldMatch:      true,
			wantedRouterName: migrationBucketApprovalName,
		},
		{
			name:             "Put object router，virtual host style",
			router:           gwRouter,
			method:           http.MethodPut,
			url:              scheme + bucketName + "." + testDomain + "/" + objectName,
			shouldMatch:      true,
			wantedRouterName: putObjectRouterName,
		},
		{
			name:             "Put object router，path style",
			router:           gwRouter,
			method:           http.MethodPut,
			url:              scheme + testDomain + "/" + bucketName + "/" + objectName,
			shouldMatch:      true,
			wantedRouterName: putObjectRouterName,
		},
		{
			name:             "PutObjectByOffset router, path style",
			router:           gwRouter,
			method:           http.MethodPost,
			url:              scheme + testDomain + "/" + bucketName + "/" + objectName + "?" + ResumableUploadComplete + "&" + ResumableUploadOffset,
			shouldMatch:      true,
			wantedRouterName: resumablePutObjectRouterName,
		},
		{
			name:             "PutObjectByOffset router, virtual host style",
			router:           gwRouter,
			method:           http.MethodPost,
			url:              scheme + bucketName + "." + testDomain + "/" + objectName + "?" + ResumableUploadComplete + "&" + ResumableUploadOffset,
			shouldMatch:      true,
			wantedRouterName: resumablePutObjectRouterName,
		},
		{
			name:             "QueryUploadOffset router, virtual host style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + bucketName + "." + testDomain + "/" + objectName + "?" + UploadContextQuery,
			shouldMatch:      true,
			wantedRouterName: queryResumeOffsetName,
		},
		{
			name:             "QueryUploadOffset router, path style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + bucketName + "." + testDomain + "/" + objectName + "?" + UploadContextQuery,
			shouldMatch:      true,
			wantedRouterName: queryResumeOffsetName,
		},
		{
			name:             "Get object upload progress router, virtual host style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + bucketName + "." + testDomain + "/" + objectName + "?" + UploadProgressQuery,
			shouldMatch:      true,
			wantedRouterName: queryUploadProgressRouterName,
		},
		{
			name:             "Get object upload progress router, path style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/" + bucketName + "/" + objectName + "?" + UploadProgressQuery,
			shouldMatch:      true,
			wantedRouterName: queryUploadProgressRouterName,
		},
		{
			name:             "Get object router, virtual host style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + bucketName + "." + testDomain + "/" + objectName,
			shouldMatch:      true,
			wantedRouterName: getObjectRouterName,
		},

		{
			name:             "Get object router, path style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/" + bucketName + "/" + objectName,
			shouldMatch:      true,
			wantedRouterName: getObjectRouterName,
		},
		{
			name:             "Get bucket read quota router, virtual host style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + bucketName + "." + testDomain + "/?" + GetBucketReadQuotaQuery + "&" + GetBucketReadQuotaMonthQuery,
			shouldMatch:      true,
			wantedRouterName: getBucketReadQuotaRouterName,
		},
		{
			name:             "Get bucket read quota router, path style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/" + bucketName + "?" + GetBucketReadQuotaQuery + "&" + GetBucketReadQuotaMonthQuery,
			shouldMatch:      true,
			wantedRouterName: getBucketReadQuotaRouterName,
		},
		{
			name:   "List bucket read records router, virtual host style",
			router: gwRouter,
			method: http.MethodGet,
			url: scheme + bucketName + "." + testDomain + "/?" + ListBucketReadRecordQuery +
				"&" + ListBucketReadRecordMaxRecordsQuery +
				"&" + StartTimestampUs + "&" + EndTimestampUs,
			shouldMatch:      true,
			wantedRouterName: listBucketReadRecordRouterName,
		},
		{
			name:   "List bucket read records router, path style",
			router: gwRouter,
			method: http.MethodGet,
			url: scheme + testDomain + "/" + bucketName + "?" + ListBucketReadRecordQuery +
				"&" + ListBucketReadRecordMaxRecordsQuery +
				"&" + StartTimestampUs + "&" + EndTimestampUs,
			shouldMatch:      true,
			wantedRouterName: listBucketReadRecordRouterName,
		},
		{
			name:             "List bucket objects router, virtual host style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + bucketName + "." + testDomain + "/",
			shouldMatch:      true,
			wantedRouterName: listObjectsByBucketRouterName,
		},
		{
			name:             "List bucket objects router, path style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/" + bucketName + "/",
			shouldMatch:      true,
			wantedRouterName: listObjectsByBucketRouterName,
		},
		{
			name:             "Get user buckets router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/",
			shouldMatch:      true,
			wantedRouterName: getUserBucketsRouterName,
		},
		{
			name:             "Get object metadata router, virtual host style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + bucketName + "." + testDomain + "/" + objectName + "?" + GetObjectMetaQuery,
			shouldMatch:      true,
			wantedRouterName: getObjectMetaRouterName,
		},
		{
			name:             "Get object metadata router, path style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/" + bucketName + "/" + objectName + "?" + GetObjectMetaQuery,
			shouldMatch:      true,
			wantedRouterName: getObjectMetaRouterName,
		},
		{
			name:             "Get bucket metadata router, virtual host style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + bucketName + "." + testDomain + "?" + GetBucketMetaQuery,
			shouldMatch:      true,
			wantedRouterName: getBucketMetaRouterName,
		},
		{
			name:             "Get bucket metadata router, path style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/" + bucketName + "?" + GetBucketMetaQuery,
			shouldMatch:      true,
			wantedRouterName: getBucketMetaRouterName,
		},
		{
			name:             "Challenge router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + GetChallengeInfoPath,
			shouldMatch:      true,
			wantedRouterName: getChallengeInfoRouterName,
		},
		{
			name:             "Replicate router",
			router:           gwRouter,
			method:           http.MethodPut,
			url:              scheme + testDomain + ReplicateObjectPiecePath,
			shouldMatch:      true,
			wantedRouterName: replicateObjectPieceRouterName,
		},
		{
			name:             "Recovery router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + RecoverObjectPiecePath,
			shouldMatch:      true,
			wantedRouterName: recoveryPieceRouterName,
		},
		{
			name:             "Get group list router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + GetGroupListGroupQuery + "&" + GetGroupListNameQuery + "&" + GetGroupListPrefixQuery + "&" + GetGroupListSourceTypeQuery + "&" + GetGroupListLimitQuery + "&" + GetGroupListOffsetQuery,
			shouldMatch:      true,
			wantedRouterName: getGroupListRouterName,
		},
		{
			name:             "List objects by object ids router",
			router:           gwRouter,
			method:           http.MethodPost,
			url:              scheme + testDomain + "/?" + ListObjectsByObjectIDQuery,
			shouldMatch:      true,
			wantedRouterName: listObjectsByObjectIDRouterName,
		},
		{
			name:             "List buckets by bucket ids router",
			router:           gwRouter,
			method:           http.MethodPost,
			url:              scheme + testDomain + "/?" + ListBucketsByBucketIDQuery,
			shouldMatch:      true,
			wantedRouterName: listBucketsByBucketIDRouterName,
		},
		{
			name:             "universal endpoint download router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/download/test_bucket_name/test_object_name",
			shouldMatch:      true,
			wantedRouterName: downloadObjectByUniversalEndpointName,
		},
		{
			name:             "universal endpoint view router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/view/test_bucket_name/test_object_name",
			shouldMatch:      true,
			wantedRouterName: viewObjectByUniversalEndpointName,
		},
		{
			name:             "universal endpoint download pdf/xml special router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/download" + objectSpecialSuffixUrlReplacement + "test_bucket_name/test_object_name",
			shouldMatch:      true,
			wantedRouterName: downloadObjectByUniversalEndpointName,
		},
		{
			name:             "universal endpoint view pdf/xml special router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/view" + objectSpecialSuffixUrlReplacement + "test_bucket_name/test_object_name",
			shouldMatch:      true,
			wantedRouterName: viewObjectByUniversalEndpointName,
		},
		{
			name:             "offchain-auth request nonce router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + AuthRequestNoncePath,
			shouldMatch:      true,
			wantedRouterName: requestNonceRouterName,
		},
		{
			name:             "offchain-auth update key router",
			router:           gwRouter,
			method:           http.MethodPost,
			url:              scheme + testDomain + AuthUpdateKeyPath,
			shouldMatch:      true,
			wantedRouterName: updateUserPublicKeyRouterName,
		},
		{
			name:             "Get payment by bucket id router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + GetPaymentByBucketIDQuery + "&" + BucketIDQuery,
			shouldMatch:      true,
			wantedRouterName: getPaymentByBucketIDRouterName,
		},
		{
			name:             "Get payment by bucket name router, virtual host style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + bucketName + "." + testDomain + "?" + GetPaymentByBucketNameQuery,
			shouldMatch:      true,
			wantedRouterName: getPaymentByBucketNameRouterName,
		},
		{
			name:             "Get payment by bucket name router, path style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/" + bucketName + "?" + GetPaymentByBucketNameQuery,
			shouldMatch:      true,
			wantedRouterName: getPaymentByBucketNameRouterName,
		},
		{
			name:             "Get bucket by bucket name router, virtual host style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + bucketName + "." + testDomain + "?" + GetBucketByBucketNameQuery,
			shouldMatch:      true,
			wantedRouterName: getBucketByBucketNameRouterName,
		},
		{
			name:             "Get bucket by bucket name router, path style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/" + bucketName + "?" + GetBucketByBucketNameQuery,
			shouldMatch:      true,
			wantedRouterName: getBucketByBucketNameRouterName,
		},
		{
			name:             "Get bucket by bucket id router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + GetBucketByBucketIDQuery + "&" + BucketIDQuery,
			shouldMatch:      true,
			wantedRouterName: getBucketByBucketIDRouterName,
		},
		{
			name:             "List deleted objects by block number range router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + ListDeletedObjectsQuery + "&" + SpOperatorAddressQuery + "&" + StartBlockNumberQuery + "&" + EndBlockNumberQuery,
			shouldMatch:      true,
			wantedRouterName: listDeletedObjectsByBlockNumberRangeRouterName,
		},
		{
			name:             "Get user buckets count router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + GetUserBucketsCountQuery,
			shouldMatch:      true,
			wantedRouterName: getUserBucketsCountRouterName,
		},
		{
			name:             "List expired buckets by sp router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + ListExpiredBucketsBySpQuery + "&" + LimitQuery + "&" + CreateAtQuery + "&" + PrimarySpIDQuery,
			shouldMatch:      true,
			wantedRouterName: listExpiredBucketsBySpRouterName,
		},
		{
			name:             "Verify permission by resource id router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + VerifyPermissionByIDQuery + "&" + ResourceIDQuery + "&" + ResourceTypeQuery + "&" + VerifyPermissionOperator + "&" + VerifyPermissionActionType,
			shouldMatch:      true,
			wantedRouterName: verifyPermissionByIDRouterName,
		},
		{
			name:             "List virtual group families by sp id router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + ListVirtualGroupFamiliesBySpIDQuery + "&" + SpIDQuery,
			shouldMatch:      true,
			wantedRouterName: listVirtualGroupFamiliesBySpIDRouterName,
		},
		{
			name:             "Get virtual group families by vgf id",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + GetVirtualGroupFamilyQuery + "&" + VgfIDQuery,
			shouldMatch:      true,
			wantedRouterName: getVirtualGroupFamilyRouterName,
		},
		{
			name:             "Get global virtual group by gvg id",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + GetGlobalVirtualGroupByGvgIDQuery + "&" + GvgIDQuery,
			shouldMatch:      true,
			wantedRouterName: getGlobalVirtualGroupByGvgIDRouterName,
		},
		{
			name:             "Get global virtual group by lvg id and bucket id",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + GetGlobalVirtualGroupQuery + "&" + LvgIDQuery + "&" + BucketIDQuery,
			shouldMatch:      true,
			wantedRouterName: getGlobalVirtualGroupRouterName,
		},
		{
			name:             "List global virtual group by secondary sp id",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + ListGlobalVirtualGroupsBySecondarySPQuery + "&" + SpIDQuery,
			shouldMatch:      true,
			wantedRouterName: listGlobalVirtualGroupsBySecondarySPRouterName,
		},
		{
			name:             "List global virtual group by bucket id",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + ListGlobalVirtualGroupsByBucketQuery + "&" + BucketIDQuery,
			shouldMatch:      true,
			wantedRouterName: listGlobalVirtualGroupsByBucketRouterName,
		},
		{
			name:             "List objects by gvg and bucket id",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + ListObjectsInGVGAndBucketQuery + "&" + GvgIDQuery + "&" + BucketIDQuery + "&" + ListObjectsStartAfterQuery + "&" + GetGroupListLimitQuery,
			shouldMatch:      true,
			wantedRouterName: listObjectsInGVGAndBucketRouterName,
		},
		{
			name:             "List objects by gvg and bucket id for gc",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + ListObjectsByGVGAndBucketForGCQuery + "&" + GvgIDQuery + "&" + BucketIDQuery + "&" + ListObjectsStartAfterQuery + "&" + GetGroupListLimitQuery,
			shouldMatch:      true,
			wantedRouterName: listObjectsByGVGAndBucketForGCRouterName,
		},
		{
			name:             "List objects by gvg id",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + ListObjectsInGVGQuery + "&" + GvgIDQuery + "&" + ListObjectsStartAfterQuery + "&" + GetGroupListLimitQuery,
			shouldMatch:      true,
			wantedRouterName: listObjectsInGVGRouterName,
		},
		{
			name:             "List migrate bucket events",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + ListMigrateBucketEventsQuery + "&" + SpIDQuery + "&" + BlockIDQuery,
			shouldMatch:      true,
			wantedRouterName: listMigrateBucketEventsRouterName,
		},
		{
			name:             "List swap out events",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + ListSwapOutEventsQuery + "&" + SpIDQuery + "&" + BlockIDQuery,
			shouldMatch:      true,
			wantedRouterName: listSwapOutEventsRouterName,
		},
		{
			name:             "List sp exit events",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + ListSpExitEventsQuery + "&" + SpIDQuery + "&" + BlockIDQuery,
			shouldMatch:      true,
			wantedRouterName: listSpExitEventsRouterName,
		},
		{
			name:             "Get sp info by operator address",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + GetSPInfoQuery + "&" + OperatorAddressQuery,
			shouldMatch:      true,
			wantedRouterName: getSPInfoRouterName,
		},
		{
			name:             "Get sp status",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/status",
			shouldMatch:      true,
			wantedRouterName: getStatusRouterName,
		},
		{
			name:             "Get groups info by a user address",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + GetUserGroupsQuery + "&" + ListObjectsStartAfterQuery + "&" + GetGroupListLimitQuery,
			shouldMatch:      true,
			wantedRouterName: getUserGroupsRouterName,
		},
		{
			name:             "Get group members by group id",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/?" + GetGroupMembersQuery + "&" + GroupIDQuery + "&" + ListObjectsStartAfterQuery + "&" + GetGroupListLimitQuery,
			shouldMatch:      true,
			wantedRouterName: getGroupMembersRouterName,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.url)
			request := httptest.NewRequest(testCase.method, testCase.url, strings.NewReader(""))
			router := testCase.router

			var match mux.RouteMatch
			ok := router.Match(request, &match)
			if ok != testCase.shouldMatch {
				t.Errorf("(%v) %v:\nRouter: %#v\nRequest: %#v\n", testCase.name, "should match", router, request)
			}
			assert.Equal(t, match.Route.GetName(), testCase.wantedRouterName)
		})
	}
}
