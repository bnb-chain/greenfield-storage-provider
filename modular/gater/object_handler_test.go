package gater

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	metadatatypes "github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	commonhttp "github.com/bnb-chain/greenfield-common/go/http"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	permissiontypes "github.com/bnb-chain/greenfield/x/permission/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func mockPutObjectHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	var routers []*mux.Router
	routers = append(routers, router.Host("{bucket:.+}."+g.domain).Subrouter())
	routers = append(routers, router.PathPrefix("/{bucket}").Subrouter())
	for _, r := range routers {
		r.NewRoute().Name(putObjectRouterName).Methods(http.MethodPut).Path("/{object:.+}").HandlerFunc(g.putObjectHandler)
	}
	return router
}

func TestGateModular_putObjectHandler(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *GateModular
		request      func() *http.Request
		wantedResult string
	}{
		{
			name: "new request context error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "put object failed to check sp and bucket status",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "failed to query sp by operator address",
		},
		{
			name: "failed to verify authentication",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "no permission to operate",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "no permission",
		},
		{
			name: "failed to get object info from consensus",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil,
					nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "failed to get object info from consensus",
		},
		{
			name: "failed to put object payload size is zero",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil, &storagetypes.ObjectInfo{}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "invalid payload",
		},
		{
			name: "failed to get storage params from consensus",
			fn: func() *GateModular {
				g := setup(t)
				g.maxPayloadSize = 100
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{PayloadSize: 10}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(nil,
					mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "failed to get storage params from consensus",
		},
		{
			name: "failed to upload payload data",
			fn: func() *GateModular {
				g := setup(t)
				g.maxPayloadSize = 100
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				clientMock.EXPECT().UploadObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{PayloadSize: 10}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(
					&storagetypes.Params{}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "success",
			fn: func() *GateModular {
				g := setup(t)
				g.maxPayloadSize = 100
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				clientMock.EXPECT().UploadObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{PayloadSize: 10}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(
					&storagetypes.Params{}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockPutObjectHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Contains(t, w.Body.String(), tt.wantedResult)
		})
	}
}

func Test_parseRange(t *testing.T) {
	cases := []struct {
		name          string
		rangeStr      string
		wantedResult1 bool
		wantedResult2 int64
		wantedResult3 int64
	}{
		{
			name:          "1",
			rangeStr:      "",
			wantedResult1: false,
			wantedResult2: -1,
			wantedResult3: -1,
		},
		{
			name:          "2",
			rangeStr:      "abc",
			wantedResult1: false,
			wantedResult2: -1,
			wantedResult3: -1,
		},
		{
			name:          "3",
			rangeStr:      "bytes=-1-",
			wantedResult1: false,
			wantedResult2: -1,
			wantedResult3: -1,
		},
		{
			name:          "4",
			rangeStr:      "bytes=1-",
			wantedResult1: true,
			wantedResult2: 1,
			wantedResult3: -1,
		},
		{
			name:          "5",
			rangeStr:      "bytes=a-2",
			wantedResult1: false,
			wantedResult2: -1,
			wantedResult3: -1,
		},
		{
			name:          "6",
			rangeStr:      "bytes=1-b",
			wantedResult1: false,
			wantedResult2: -1,
			wantedResult3: -1,
		},
		{
			name:          "7",
			rangeStr:      "bytes=1-2",
			wantedResult1: true,
			wantedResult2: 1,
			wantedResult3: 2,
		},
		{
			name:          "8",
			rangeStr:      "bytes=-1-2",
			wantedResult1: false,
			wantedResult2: -1,
			wantedResult3: -1,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result1, result2, result3 := parseRange(tt.rangeStr)
			assert.Equal(t, tt.wantedResult1, result1)
			assert.Equal(t, tt.wantedResult2, result2)
			assert.Equal(t, tt.wantedResult3, result3)
		})
	}
}

func mockResumablePutObjectHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	var routers []*mux.Router
	routers = append(routers, router.Host("{bucket:.+}."+g.domain).Subrouter())
	routers = append(routers, router.PathPrefix("/{bucket}").Subrouter())
	for _, r := range routers {
		r.NewRoute().Name(resumablePutObjectRouterName).Methods(http.MethodPost).Path("/{object:.+}").HandlerFunc(g.resumablePutObjectHandler).Queries(
			"offset", "{offset}", "complete", "{complete}")
	}
	return router
}

func TestGateModular_resumablePutObjectHandler(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *GateModular
		request      func() *http.Request
		wantedResult string
	}{
		{
			name: "new request context error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, mockObjectName,
					ResumableUploadComplete, ResumableUploadOffset)
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "resumable put object failed to check sp and bucket status",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, mockObjectName,
					ResumableUploadComplete, ResumableUploadOffset)
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "failed to query sp by operator address",
		},
		{
			name: "failed to verify authentication",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, mockObjectName,
					ResumableUploadComplete, ResumableUploadOffset)
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "no permission to operate",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, mockObjectName,
					ResumableUploadComplete, ResumableUploadOffset)
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "no permission",
		},
		{
			name: "failed to get object info from consensus",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil,
					nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, mockObjectName,
					ResumableUploadComplete, ResumableUploadOffset)
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "failed to get object info from consensus",
		},
		{
			name: "failed to get storage params from consensus",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil, &storagetypes.ObjectInfo{}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParams(gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, mockObjectName,
					ResumableUploadComplete, ResumableUploadOffset)
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "failed to get storage params from consensus",
		},
		{
			name: "failed to put object payload size is zero",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParams(gomock.Any()).Return(&storagetypes.Params{}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s&%s", scheme, mockBucketName, testDomain, mockObjectName,
					ResumableUploadComplete, ResumableUploadOffset)
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "invalid payload",
		},
		{
			name: "failed to parse complete from url",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{PayloadSize: 10}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParams(gomock.Any()).Return(&storagetypes.Params{
					MaxPayloadSize: 100}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s=%s&%s", scheme, mockBucketName, testDomain, mockObjectName,
					ResumableUploadComplete, "a", ResumableUploadOffset)
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "invalid complete",
		},
		{
			name: "complete query param is empty",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{PayloadSize: 10}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParams(gomock.Any()).Return(&storagetypes.Params{
					MaxPayloadSize: 100}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s=%s&%s", scheme, mockBucketName, testDomain, mockObjectName,
					ResumableUploadComplete, "", ResumableUploadOffset)
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "invalid complete",
		},
		{
			name: "failed to parse offset from url",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{PayloadSize: 10}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParams(gomock.Any()).Return(&storagetypes.Params{
					MaxPayloadSize: 100}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s=%s&%s=%s", scheme, mockBucketName, testDomain, mockObjectName,
					ResumableUploadComplete, "false", ResumableUploadOffset, "a")
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "invalid offset",
		},
		{
			name: "offset query param is empty",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{PayloadSize: 10}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParams(gomock.Any()).Return(&storagetypes.Params{
					MaxPayloadSize: 100}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s=%s&%s=%s", scheme, mockBucketName, testDomain, mockObjectName,
					ResumableUploadComplete, "false", ResumableUploadOffset, "")
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "invalid offset",
		},
		{
			name: "failed to resumable upload payload data",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				clientMock.EXPECT().ResumableUploadObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{PayloadSize: 10}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParams(gomock.Any()).Return(&storagetypes.Params{
					MaxPayloadSize: 100}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s=%s&%s=%s", scheme, mockBucketName, testDomain, mockObjectName,
					ResumableUploadComplete, "true", ResumableUploadOffset, "1")
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
		{
			name: "success",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				clientMock.EXPECT().ResumableUploadObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{PayloadSize: 10}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParams(gomock.Any()).Return(&storagetypes.Params{
					MaxPayloadSize: 100}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s=%s&%s=%s", scheme, mockBucketName, testDomain, mockObjectName,
					ResumableUploadComplete, "true", ResumableUploadOffset, "1")
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockResumablePutObjectHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Contains(t, w.Body.String(), tt.wantedResult)
		})
	}
}

func mockQueryResumeOffsetHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	var routers []*mux.Router
	routers = append(routers, router.Host("{bucket:.+}."+g.domain).Subrouter())
	routers = append(routers, router.PathPrefix("/{bucket}").Subrouter())
	for _, r := range routers {
		r.NewRoute().Name(queryResumeOffsetName).Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(g.queryResumeOffsetHandler).
			Queries(UploadContextQuery, "")
	}
	return router
}

func TestGateModular_queryResumeOffsetHandlerHandler(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *GateModular
		request      func() *http.Request
		wantedResult string
	}{
		{
			name: "new request context error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, UploadContextQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "failed to verify authentication",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, UploadContextQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "no permission to operate",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, UploadContextQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "no permission",
		},
		{
			name: "failed to get object info from consensus",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).
					Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, UploadContextQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "failed to get object info from consensus",
		},
		{
			name: "failed to get storage params from consensus",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{},
					nil).Times(1)
				consensusMock.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, UploadContextQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "failed to get storage params from consensus",
		},
		{
			name: "failed to get uploading object segment",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				clientMock.EXPECT().GetUploadObjectSegment(gomock.Any(), gomock.Any()).Return(uint32(0), mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{
					Id: sdkmath.NewUint(1)}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(
					&storagetypes.Params{}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, UploadContextQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "success",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				clientMock.EXPECT().GetUploadObjectSegment(gomock.Any(), gomock.Any()).Return(uint32(0), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{
					Id: sdkmath.NewUint(1)}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(
					&storagetypes.Params{VersionedParams: storagetypes.VersionedParams{MaxSegmentSize: 10}}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, UploadContextQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockQueryResumeOffsetHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Contains(t, w.Body.String(), tt.wantedResult)
		})
	}
}

func mockGetObjectHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	var routers []*mux.Router
	routers = append(routers, router.Host("{bucket:.+}."+g.domain).Subrouter())
	routers = append(routers, router.PathPrefix("/{bucket}").Subrouter())
	for _, r := range routers {
		r.NewRoute().Name(getObjectRouterName).Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(g.getObjectHandler)
	}
	return router
}

func TestGateModular_getObjectHandler(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *GateModular
		request      func() *http.Request
		wantedResult string
	}{
		{
			name: "failed to verify authentication for getting public object",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "<Error><Code>999999</Code><Message>mock error</Message></Error>",
		},
		{
			name: "new request returns error and unsupported sign type",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				var a = permissiontypes.EFFECT_DENY
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?Authorization=a", scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "unsupported sign type",
		},
		{
			name: "no permission to operate, object is not public and preSignedURLErr is not nil",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				c1 := clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				var a = permissiontypes.EFFECT_DENY
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(1)
				c2 := clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				gomock.InOrder(c1, c2)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?Authorization=GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221&X-Gnfd-User-Address=user&X-Gnfd-App-Domain=app&X-Gnfd-Expiry-Timestamp=time",
					scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "failed to verify authentication",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				c1 := clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				var a = permissiontypes.EFFECT_DENY
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(1)
				c2 := clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				gomock.InOrder(c1, c2)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(false, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?Authorization=GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221&X-Gnfd-User-Address=user&X-Gnfd-App-Domain=app&X-Gnfd-Expiry-Timestamp=time&view=true",
					scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "no permission to operate",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				c1 := clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				var a = permissiontypes.EFFECT_DENY
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(1)
				c2 := clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				gomock.InOrder(c1, c2)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?Authorization=GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221&X-Gnfd-User-Address=user&X-Gnfd-App-Domain=app&X-Gnfd-Expiry-Timestamp=time&view=true",
					scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "no permission",
		},
		{
			name: "failed to get object info from consensus",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "failed to get object info from consensus",
		},
		{
			name: "failed to get bucket info from consensus",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "failed to get bucket info from consensus",
		},
		{
			name: "failed to get storage params from consensus",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					Id: sdkmath.NewUint(2)}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "failed to get storage params from consensus",
		},
		{
			name: "request params invalid",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.ObjectInfo{Id: sdkmath.NewUint(1)}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					Id: sdkmath.NewUint(2)}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(
					&storagetypes.Params{MaxPayloadSize: 10}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(RangeHeader, "bytes=-1")
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "request params invalid",
		},
		{
			name: "failed to download piece",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(1)
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return(nil, mockErr).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.ObjectInfo{
						Id:          sdkmath.NewUint(1),
						PayloadSize: 10,
					}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					Id: sdkmath.NewUint(2)}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(
					&storagetypes.Params{MaxPayloadSize: 10}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				g.baseApp.SetPieceOp(pieceOpMock)
				pieceOpMock.EXPECT().SegmentPieceCount(gomock.Any(), gomock.Any()).Return(uint32(1)).Times(1)
				pieceOpMock.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").AnyTimes()
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(RangeHeader, "bytes=-1")
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "success",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(1)
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.ObjectInfo{
						Id:          sdkmath.NewUint(1),
						PayloadSize: 10,
					}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					Id: sdkmath.NewUint(2)}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(
					&storagetypes.Params{MaxPayloadSize: 10}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				g.baseApp.SetPieceOp(pieceOpMock)
				pieceOpMock.EXPECT().SegmentPieceCount(gomock.Any(), gomock.Any()).Return(uint32(1)).Times(1)
				pieceOpMock.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").AnyTimes()
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, mockBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(RangeHeader, "bytes=-1")
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
		{
			name: "failed to check bucket name",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, invalidBucketName, testDomain, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "<Error><Code>50010</Code><Message>invalid request params for query</Message></Error>",
		},
		{
			name: "failed to check object name",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s", scheme, mockBucketName, testDomain, invalidObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "<Error><Code>50010</Code><Message>invalid request params for query</Message></Error>",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetObjectHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Contains(t, w.Body.String(), tt.wantedResult)
		})
	}
}

func mockQueryUploadProgressHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	var routers []*mux.Router
	routers = append(routers, router.Host("{bucket:.+}."+g.domain).Subrouter())
	routers = append(routers, router.PathPrefix("/{bucket}").Subrouter())
	for _, r := range routers {
		r.NewRoute().Name(queryUploadProgressRouterName).Methods(http.MethodGet).Path("/{object:.+}").HandlerFunc(g.queryUploadProgressHandler).
			Queries(UploadProgressQuery, "")
	}
	return router
}

func TestGateModular_queryUploadProgressHandler(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *GateModular
		request      func() *http.Request
		wantedResult string
	}{
		{
			name: "new request context error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, UploadProgressQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "failed to verify authentication",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, UploadProgressQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "no permission to operate",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, UploadProgressQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "no permission",
		},
		{
			name: "failed to get object info from consensus",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, UploadProgressQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "failed to get object info from consensus",
		},
		{
			name: "failed to get uploading job state and OBJECT_STATUS_CREATED",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil)
				clientMock.EXPECT().GetUploadObjectState(gomock.Any(), gomock.Any()).Return(int32(0), "", mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_CREATED, Id: sdkmath.NewUint(1)}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, UploadProgressQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
		{
			name: "get uploading job state no error and OBJECT_STATUS_CREATED",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil)
				clientMock.EXPECT().GetUploadObjectState(gomock.Any(), gomock.Any()).Return(int32(0), "", nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_CREATED, Id: sdkmath.NewUint(1)}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, UploadProgressQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
		{
			name: "OBJECT_STATUS_SEALED and success",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Id: sdkmath.NewUint(1)}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, UploadProgressQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
		{
			name: "OBJECT_STATUS_DISCONTINUED and success",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_DISCONTINUED, Id: sdkmath.NewUint(1)}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s.%s/%s?%s", scheme, mockBucketName, testDomain, mockObjectName, UploadProgressQuery)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockQueryUploadProgressHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Contains(t, w.Body.String(), tt.wantedResult)
		})
	}
}

func mockGetObjectByUniversalEndpointHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path("/download/{bucket:[^/]*}/{object:.+}").Name(downloadObjectByUniversalEndpointName).Methods(http.MethodGet).
		HandlerFunc(g.downloadObjectByUniversalEndpointHandler)
	router.Path("/download").Name(downloadObjectByUniversalEndpointName).Methods(http.MethodGet).
		Queries(UniversalEndpointSpecialSuffixQuery, "{bucket:[^/]*}/{object:.+}").HandlerFunc(g.downloadObjectByUniversalEndpointHandler)
	return router
}

func TestGateModular_getObjectByUniversalEndpointHandler(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *GateModular
		request      func() *http.Request
		wantedResult string
	}{
		{
			name: "failed to check bucket name",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, "mo", mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set("User-Agent", "Chrome")
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
		{
			name: "failed to check object name",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, mockBucketName, invalidObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set("User-Agent", "Chrome")
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
		{
			name: "failed to check bucket info",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, mockBucketName, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set("User-Agent", "Chrome")
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
		{
			name: "getSPID error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, mockBucketName, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set("User-Agent", "Chrome")
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
		{
			name: "GetBucketPrimarySPID error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, mockBucketName, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set("User-Agent", "Chrome")
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
		{
			name: "failed to get endpoint by id",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetEndpointBySpID(gomock.Any(), gomock.Any()).Return("", mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 2}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, mockBucketName, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set("User-Agent", "Chrome")
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
		{
			name: "redirect",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetEndpointBySpID(gomock.Any(), gomock.Any()).Return("a", nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 2}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, mockBucketName, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set("User-Agent", "Chrome")
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
		{
			name: "failed to check object meta",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, mockBucketName, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set("User-Agent", "Chrome")
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
		{
			name: "object is not sealed",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_CREATED}},
					nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, mockBucketName, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set("User-Agent", "Chrome")
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetObjectByUniversalEndpointHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Contains(t, w.Body.String(), tt.wantedResult)
		})
	}
}

func Test_isPrivateObject(t *testing.T) {
	bucket := &storagetypes.BucketInfo{Visibility: 2}
	object := &storagetypes.ObjectInfo{Visibility: 1}
	result := isPrivateObject(bucket, object)
	assert.Equal(t, false, result)
}

func Test_checkIfRequestFromBrowser(t *testing.T) {
	cases := []struct {
		name         string
		userAgent    string
		wantedResult bool
	}{
		{
			name:         "1",
			userAgent:    "Chrome",
			wantedResult: true,
		},
		{
			name:         "2",
			userAgent:    "unknown",
			wantedResult: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := checkIfRequestFromBrowser(tt.userAgent)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func Test_downloadObjectByUniversalEndpointHandler(t *testing.T) {
	g := setup(t)
	r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, mockBucketName, mockObjectName),
		nil)
	w := httptest.NewRecorder()
	g.downloadObjectByUniversalEndpointHandler(w, r)
}

func Test_viewObjectByUniversalEndpointHandler(t *testing.T) {
	g := setup(t)
	r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s%s/view/%s/%s", scheme, testDomain, mockBucketName, mockObjectName),
		nil)
	w := httptest.NewRecorder()
	g.viewObjectByUniversalEndpointHandler(w, r)
}
