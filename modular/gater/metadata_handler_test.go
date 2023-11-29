package gater

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
)

func mockListObjectPoliciesRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	var routers []*mux.Router
	routers = append(routers, router.Host("{bucket:.+}."+g.domain).Subrouter())
	routers = append(routers, router.PathPrefix("/{bucket}").Subrouter())
	for _, r := range routers {
		r.NewRoute().Name(listObjectPoliciesRouterName).Methods(http.MethodGet).Path("/{object:.+}").Queries(ListObjectPoliciesQuery, "").
			HandlerFunc(g.listObjectPoliciesHandler)
	}
	return router
}

func TestGateModular_listObjectPoliciesHandler(t *testing.T) {
	cases := []struct {
		name       string
		fn         func() *GateModular
		request    func() *http.Request
		wantedCode int
	}{
		{
			name: "bukcet name error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, invalidBucketName, testDomain, invalidObjectName, ListObjectPoliciesQuery,
					"limit=10&action-type=6")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedCode: 400,
		},
		{
			name: "object name error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, invalidObjectName, ListObjectPoliciesQuery,
					"limit=10&action-type=6")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedCode: 400,
		},
		{
			name: "limit value error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, mockObjectName, ListObjectPoliciesQuery,
					"limit=a&action-type=6")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedCode: 400,
		},
		{
			name: "start-after value error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, mockObjectName, ListObjectPoliciesQuery,
					"limit=10&action-type=6&start-after=a")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedCode: 400,
		},
		{
			name: "action type value error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, mockObjectName, ListObjectPoliciesQuery,
					"limit=10&action-type=100")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedCode: 400,
		},
		{
			name: "action type value error 2",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, mockObjectName, ListObjectPoliciesQuery,
					"limit=10&action-type=a")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedCode: 400,
		},
		{
			name: "rpc error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().ListObjectPolicies(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(nil, errors.New(`rpc error: code = Unknown desc = {"code_space":"metadata","http_status_code":404,"inner_code":90008,"description":"the specified bucket does not exist"}`)).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, mockObjectName, ListObjectPoliciesQuery,
					"limit=10&action-type=6")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				return req
			},
			wantedCode: 404,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockListObjectPoliciesRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Equal(t, w.Code, tt.wantedCode)
		})
	}
}
