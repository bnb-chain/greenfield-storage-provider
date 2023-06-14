package gater

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var (
	testDomain = "www.route-test.com"
	gw         = &GateModular{
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
			url:              scheme + testDomain + "/" + bucketName + "/" + objectName + "?offset=0&complete=false&context=",
			shouldMatch:      true,
			wantedRouterName: resumablePutObjectRouterName,
		},
		{
			name:             "PutObjectByOffset router, virtual host style",
			router:           gwRouter,
			method:           http.MethodPost,
			url:              scheme + bucketName + "." + testDomain + "/" + objectName + "?offset=0&complete=false&context=",
			shouldMatch:      true,
			wantedRouterName: resumablePutObjectRouterName,
		},
		{
			name:             "QueryUploadOffset router, virtual host style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + bucketName + "." + testDomain + "/" + objectName + "?upload-context=",
			shouldMatch:      true,
			wantedRouterName: queryResumeOffsetName,
		},
		{
			name:             "QueryUploadOffset router, path style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + bucketName + "." + testDomain + "/" + objectName + "?upload-context=",
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
			url:              scheme + testDomain + "/?" + ListObjectsByObjectID,
			shouldMatch:      true,
			wantedRouterName: listObjectsByObjectIDRouterName,
		}, {
			name:             "List buckets by bucket ids router",
			router:           gwRouter,
			method:           http.MethodPost,
			url:              scheme + testDomain + "/?" + ListBucketsByBucketID,
			shouldMatch:      true,
			wantedRouterName: listBucketsByBucketIDRouterName,
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
