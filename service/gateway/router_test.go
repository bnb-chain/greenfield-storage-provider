package gateway

import (
	"net/http"
	"strings"
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var (
	testDomain = "www.route-test.com"
	config     = &GatewayConfig{
		Domain: testDomain,
	}
	gw = &Gateway{
		config: config,
	}
	scheme     = "https://"
	bucketName = "test-bucket-name"
	objectName = "test-object-name"
)

type routerTestContext struct {
	title       string      // title of the test case
	router      *mux.Router // the router being tested
	method      string      // the request method
	url         string      // the request url, include path + query
	shouldMatch bool        // whether the request is expected to match the route at all
	routerName  string      // the matched router name
}

// newRequest is a helper function to create a new request with params context.
func newRequest(ctx *routerTestContext) *http.Request {
	req, err := http.NewRequest(ctx.method, ctx.url, strings.NewReader(""))
	if err != nil {
		panic(err)
	}
	return req
}

func checkRouter(t *testing.T, ctx *routerTestContext) {
	request := newRequest(ctx)
	router := ctx.router

	var match mux.RouteMatch
	ok := router.Match(request, &match)
	if ok != ctx.shouldMatch {
		t.Errorf("(%v) %v:\nRouter: %#v\nRequest: %#v\n", ctx.title, "should match", router, request)
	}
	assert.Equal(t, match.Route.GetName(), ctx.routerName)
}

func TestRouters(t *testing.T) {
	gwRouter := mux.NewRouter().SkipClean(true)
	gw.registerHandler(gwRouter)

	testCases := []routerTestContext{
		{
			title:       "Get create bucket approval router",
			router:      gwRouter,
			method:      http.MethodGet,
			url:         scheme + testDomain + model.GetApprovalPath + "?" + model.ActionQuery + "=" + createBucketApprovalAction,
			shouldMatch: true,
			routerName:  approvalRouterName,
		},
		{
			title:       "Get create object approval router",
			router:      gwRouter,
			method:      http.MethodGet,
			url:         scheme + testDomain + model.GetApprovalPath + "?" + model.ActionQuery + "=" + createObjectApprovalAction,
			shouldMatch: true,
			routerName:  approvalRouterName,
		},
		{
			title:       "Put object router",
			router:      gwRouter,
			method:      http.MethodPut,
			url:         scheme + bucketName + "." + testDomain + "/" + objectName,
			shouldMatch: true,
			routerName:  putObjectRouterName,
		},
		{
			title:       "Get object router",
			router:      gwRouter,
			method:      http.MethodGet,
			url:         scheme + bucketName + "." + testDomain + "/" + objectName,
			shouldMatch: true,
			routerName:  getObjectRouterName,
		},
		{
			title:       "Get bucket read quota router",
			router:      gwRouter,
			method:      http.MethodGet,
			url:         scheme + bucketName + "." + testDomain + "/?" + model.GetBucketReadQuotaQuery + "&" + model.GetBucketReadQuotaMonthQuery,
			shouldMatch: true,
			routerName:  getBucketReadQuotaRouterName,
		},
		{
			title:  "List bucket read records router",
			router: gwRouter,
			method: http.MethodGet,
			url: scheme + bucketName + "." + testDomain + "/?" + model.ListBucketReadRecordQuery +
				"&" + model.ListBucketReadRecordMaxRecordsQuery +
				"&" + model.StartTimestampUs + "&" + model.EndTimestampUs,
			shouldMatch: true,
			routerName:  listBucketReadRecordRouterName,
		},
		{
			title:       "List bucket objects router",
			router:      gwRouter,
			method:      http.MethodGet,
			url:         scheme + bucketName + "." + testDomain + "/",
			shouldMatch: true,
			routerName:  listObjectsByBucketRouterName,
		},
		{
			title:       "Get user buckets router",
			router:      gwRouter,
			method:      http.MethodGet,
			url:         scheme + testDomain + "/",
			shouldMatch: true,
			routerName:  getUserBucketsRouterName,
		},
		{
			title:       "Challenge router",
			router:      gwRouter,
			method:      http.MethodGet,
			url:         scheme + testDomain + model.ChallengePath,
			shouldMatch: true,
			routerName:  challengeRouterName,
		},
		{
			title:       "Sync router",
			router:      gwRouter,
			method:      http.MethodPut,
			url:         scheme + testDomain + model.SyncPath,
			shouldMatch: true,
			routerName:  syncPieceRouterName,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.title, func(t *testing.T) {
			checkRouter(t, &testCase)
		})
	}
}
