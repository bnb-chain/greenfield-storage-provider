package gateway

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

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
			url:              scheme + testDomain + model.GetApprovalPath + "?" + model.ActionQuery + "=" + createBucketApprovalAction,
			shouldMatch:      true,
			wantedRouterName: approvalRouterName,
		},
		{
			name:             "Get create object approval router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + model.GetApprovalPath + "?" + model.ActionQuery + "=" + createObjectApprovalAction,
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
			name:             "Get object upload progress router, virtual host style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + bucketName + "." + testDomain + "/" + objectName + "?" + model.UploadProgressQuery,
			shouldMatch:      true,
			wantedRouterName: queryUploadProgressRouterName,
		},
		{
			name:             "Get object upload progress router, path style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/" + bucketName + "/" + objectName + "?" + model.UploadProgressQuery,
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
			url:              scheme + bucketName + "." + testDomain + "/?" + model.GetBucketReadQuotaQuery + "&" + model.GetBucketReadQuotaMonthQuery,
			shouldMatch:      true,
			wantedRouterName: getBucketReadQuotaRouterName,
		},
		{
			name:             "Get bucket read quota router, path style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/" + bucketName + "?" + model.GetBucketReadQuotaQuery + "&" + model.GetBucketReadQuotaMonthQuery,
			shouldMatch:      true,
			wantedRouterName: getBucketReadQuotaRouterName,
		},
		{
			name:   "List bucket read records router, virtual host style",
			router: gwRouter,
			method: http.MethodGet,
			url: scheme + bucketName + "." + testDomain + "/?" + model.ListBucketReadRecordQuery +
				"&" + model.ListBucketReadRecordMaxRecordsQuery +
				"&" + model.StartTimestampUs + "&" + model.EndTimestampUs,
			shouldMatch:      true,
			wantedRouterName: listBucketReadRecordRouterName,
		},
		{
			name:   "List bucket read records router, path style",
			router: gwRouter,
			method: http.MethodGet,
			url: scheme + testDomain + "/" + bucketName + "?" + model.ListBucketReadRecordQuery +
				"&" + model.ListBucketReadRecordMaxRecordsQuery +
				"&" + model.StartTimestampUs + "&" + model.EndTimestampUs,
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
			url:              scheme + bucketName + "." + testDomain + "/" + objectName + "?" + model.GetObjectMetaQuery,
			shouldMatch:      true,
			wantedRouterName: getObjectMetaRouterName,
		},
		{
			name:             "Get object metadata router, path style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/" + bucketName + "/" + objectName + "?" + model.GetObjectMetaQuery,
			shouldMatch:      true,
			wantedRouterName: getObjectMetaRouterName,
		},
		{
			name:             "Get bucket metadata router, virtual host style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + bucketName + "." + testDomain + "?" + model.GetBucketMetaQuery,
			shouldMatch:      true,
			wantedRouterName: getBucketMetaRouterName,
		},
		{
			name:             "Get bucket metadata router, path style",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + "/" + bucketName + "?" + model.GetBucketMetaQuery,
			shouldMatch:      true,
			wantedRouterName: getBucketMetaRouterName,
		},
		{
			name:             "Challenge router",
			router:           gwRouter,
			method:           http.MethodGet,
			url:              scheme + testDomain + model.ChallengePath,
			shouldMatch:      true,
			wantedRouterName: challengeRouterName,
		},
		{
			name:             "Replicate router",
			router:           gwRouter,
			method:           http.MethodPut,
			url:              scheme + testDomain + model.ReplicateObjectPiecePath,
			shouldMatch:      true,
			wantedRouterName: replicateObjectPieceRouterName,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
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
