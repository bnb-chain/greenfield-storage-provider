package gater

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/modular/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

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

func TestGateModularCode_getObjectHandler(t *testing.T) {
	cases := []struct {
		name       string
		fn         func() *GateModular
		request    func() *http.Request
		wantedCode int
	}{
		{
			name: "failed to verify authentication no object",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				c1 := clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New(`rpc error: code = Unknown desc = {"code_space":"metadata","http_status_code":404,"inner_code":90008,"description":"the specified bucket does not exist"}`)).Times(1)
				c2 := clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).AnyTimes()
				gomock.InOrder(c1, c2)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(false, mockErr).AnyTimes()
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
			wantedCode: 404,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetObjectHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Equal(t, w.Code, tt.wantedCode)
		})
	}
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

	router.Path("/view/{bucket:[^/]*}/{object:.+}").Name(viewObjectByUniversalEndpointName).Methods(http.MethodGet).
		HandlerFunc(g.viewObjectByUniversalEndpointHandler)
	router.Path("/view").Name(viewObjectByUniversalEndpointName).Methods(http.MethodGet).
		Queries(UniversalEndpointSpecialSuffixQuery, "{bucket:[^/]*}/{object:.+}").HandlerFunc(g.viewObjectByUniversalEndpointHandler)
	return router
}

func TestGateModular_getObjectByUniversalEndpointHandler(t *testing.T) {
	cases := []struct {
		name                 string
		fn                   func() *GateModular
		request              func() *http.Request
		wantedResult         string
		wantedHttpRespCode   int
		wantedResponseHeader map[string]string
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
				req.Header.Set("User-Agent", "Chrome")
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
				req.Header.Set("User-Agent", "Chrome")
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
				req.Header.Set("User-Agent", "Chrome")
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
				req.Header.Set("User-Agent", "Chrome")
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
				req.Header.Set("User-Agent", "Chrome")
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
				req.Header.Set("User-Agent", "Chrome")
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
				clientMock.EXPECT().GetEndpointBySpID(gomock.Any(), gomock.Any()).Return("https://another.sp.com", nil).Times(1)
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
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult:       "<a href=\"https://another.sp.com/download/mock-bucket-name/mock-object-name\">Found</a>.",
			wantedHttpRespCode: 302,
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
				req.Header.Set("User-Agent", "Chrome")
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
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "",
		},

		{
			name: "download public object",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PUBLIC_READ}},
					nil).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(0) // public file, so no need to verify permission
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

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
				path := fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, mockBucketName, url.QueryEscape(mockObjectName))
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "",
			wantedResponseHeader: map[string]string{
				ContentDispositionHeader: ContentDispositionAttachmentValue + "; filename=\"" + "mock-object-name" + "\"",
			},
		},
		{
			name: "download special name object - ?limit=1",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PUBLIC_READ}},
					nil).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(0) // public file, so no need to verify permission
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

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
				path := fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, mockBucketName, url.QueryEscape("?limit=1"))
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "",
			wantedResponseHeader: map[string]string{
				ContentDispositionHeader: ContentDispositionAttachmentValue + "; filename=\"" + "%3Flimit%3D1" + "\"",
			},
		},
		{
			name: "download special name object - hello%20world!",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PUBLIC_READ}},
					nil).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(0) // public file, so no need to verify permission
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

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
				path := fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, mockBucketName, url.QueryEscape("hello%20world!"))
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "",
			wantedResponseHeader: map[string]string{
				ContentDispositionHeader: ContentDispositionAttachmentValue + "; filename=\"" + "hello%2520world%21" + "\"",
			},
		},
		{
			name: "download special name object - name&value=123.txt",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PUBLIC_READ}},
					nil).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(0) // public file, so no need to verify permission
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

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
				path := fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, mockBucketName, url.QueryEscape("name&value=123.txt"))
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "",
			wantedResponseHeader: map[string]string{
				ContentDispositionHeader: ContentDispositionAttachmentValue + "; filename=\"" + "name%26value%3D123.txt" + "\"",
			},
		},
		{
			name: "download special name object - file%20=name.txt",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PUBLIC_READ}},
					nil).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(0) // public file, so no need to verify permission
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

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
				path := fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, mockBucketName, url.QueryEscape("file%20=name.txt"))
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "",
			wantedResponseHeader: map[string]string{
				ContentDispositionHeader: ContentDispositionAttachmentValue + "; filename=\"" + "file%2520%3Dname.txt" + "\"",
			},
		},
		{
			name: "download non-escaped name object - 1~~~1",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PUBLIC_READ}},
					nil).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(0) // public file, so no need to verify permission
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

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
				path := fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, mockBucketName, url.QueryEscape("1~~~1"))
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "",
			wantedResponseHeader: map[string]string{
				ContentDispositionHeader: ContentDispositionAttachmentValue + "; filename=\"" + "1~~~1" + "\"",
			},
		},
		{
			name: "download non-escaped name object - object-name",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PUBLIC_READ}},
					nil).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(0) // public file, so no need to verify permission
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

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
				path := fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, mockBucketName, url.QueryEscape("object-name"))
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "",
			wantedResponseHeader: map[string]string{
				ContentDispositionHeader: ContentDispositionAttachmentValue + "; filename=\"" + "object-name" + "\"",
			},
		},
		{
			name: "view public object",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PUBLIC_READ}},
					nil).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(0) // public file, so no need to verify permission
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

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
				path := fmt.Sprintf("%s%s/view/%s/%s", scheme, testDomain, mockBucketName, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "",
			wantedResponseHeader: map[string]string{
				ContentDispositionHeader: ContentDispositionInlineValue,
			},
		},
		{
			name: "view private object",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PRIVATE}},
					nil).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(0) // public file, so no need to verify permission
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				g.baseApp.SetPieceOp(pieceOpMock)
				pieceOpMock.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").AnyTimes()

				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/view/%s/%s", scheme, testDomain, mockBucketName, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "<!doctype html>",
			wantedResponseHeader: map[string]string{
				ContentDispositionHeader: "",
			},
		},

		{
			name: "view private object with expired ExpiryDateParam",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PRIVATE}},
					nil).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(0) // can't reach here, so no need to verify permission
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				g.baseApp.SetPieceOp(pieceOpMock)
				pieceOpMock.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").AnyTimes()

				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/view/%s/%s", scheme, testDomain, mockBucketName, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set("User-Agent", "Chrome")
				validExpiryDateStr := time.Now().Add(-time.Hour * 60).Format(ExpiryDateFormat)
				queryParams := req.URL.Query()
				queryParams.Add(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				queryParams.Add("signature", "sample_signature")
				req.URL.RawQuery = queryParams.Encode()
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "INTERNAL_ERROR", // INTERNAL_ERROR as there is ErrInvalidExpiryDateParam
			wantedResponseHeader: map[string]string{
				ContentDispositionHeader: "",
			},
		},

		{
			name: "view private object with bad ExpiryDateParam",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PRIVATE}},
					nil).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(0) // can't reach here, so no need to verify permission
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				g.baseApp.SetPieceOp(pieceOpMock)
				pieceOpMock.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").AnyTimes()

				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/view/%s/%s", scheme, testDomain, mockBucketName, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set("User-Agent", "Chrome")
				validExpiryDateStr := time.Now().Add(-time.Hour * 60).Format(ExpiryDateFormat)
				queryParams := req.URL.Query()
				queryParams.Add(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr+"_bad_format")
				queryParams.Add("signature", "sample_signature")
				req.URL.RawQuery = queryParams.Encode()
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "INTERNAL_ERROR", // INTERNAL_ERROR as there is bad signature
			wantedResponseHeader: map[string]string{
				ContentDispositionHeader: "",
			},
		},

		{
			name: "view private object with bad signature",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PRIVATE}},
					nil).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(0) // can't reach here, so no need to verify permission
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				g.baseApp.SetPieceOp(pieceOpMock)
				pieceOpMock.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").AnyTimes()

				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/view/%s/%s", scheme, testDomain, mockBucketName, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set("User-Agent", "Chrome")
				validExpiryDateStr := time.Now().Add(+time.Hour * 60).Format(ExpiryDateFormat)
				queryParams := req.URL.Query()
				queryParams.Add(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				queryParams.Add("signature", "bad_signature")
				req.URL.RawQuery = queryParams.Encode()
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "INTERNAL_ERROR", // INTERNAL_ERROR as there is ErrInvalidExpiryDateParam
			wantedResponseHeader: map[string]string{
				ContentDispositionHeader: "",
			},
		},

		{
			name: "view private object with good signature but get error when verify permission",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{BucketName: mockBucketName}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName, ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PRIVATE}},
					nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(0) // can't reach here, so no need to verify permission
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				g.baseApp.SetPieceOp(pieceOpMock)
				pieceOpMock.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").AnyTimes()

				return g
			},
			request: func() *http.Request {

				path := fmt.Sprintf("%s%s/view/%s/%s", scheme, testDomain, mockBucketName, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set("User-Agent", "Chrome")
				validExpiryDateStr := time.Now().Add(+time.Hour * 60).Format(ExpiryDateFormat)
				queryParams := req.URL.Query()
				queryParams.Add(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)

				// prepare sig
				// create a test account information.
				privateKey, _ := crypto.GenerateKey()
				address := crypto.PubkeyToAddress(privateKey.PublicKey)
				log.Infof("address is: " + address.Hex())
				signedMsg := fmt.Sprintf(GnfdBuiltInDappSignedContentTemplate, "gnfd://"+mockBucketName+"/"+mockObjectName, validExpiryDateStr)
				unSignedContentHash := accounts.TextHash([]byte(signedMsg)) // personal sign.
				// sign data.
				sig, _ := crypto.Sign(unSignedContentHash, privateKey)

				queryParams.Add("signature", hexutil.Encode(sig))
				req.URL.RawQuery = queryParams.Encode()
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "NO_PERMISSION", // NO_PERMISSION as there is error occurs when verifying permission
			wantedResponseHeader: map[string]string{
				ContentDispositionHeader: "",
			},
		},

		{
			name: "view private object with good signature but no permission",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{BucketName: mockBucketName}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName, ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PRIVATE}},
					nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(0) // can't reach here, so no need to verify permission
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				g.baseApp.SetPieceOp(pieceOpMock)
				pieceOpMock.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").AnyTimes()

				return g
			},
			request: func() *http.Request {

				path := fmt.Sprintf("%s%s/view/%s/%s", scheme, testDomain, mockBucketName, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set("User-Agent", "Chrome")
				validExpiryDateStr := time.Now().Add(+time.Hour * 60).Format(ExpiryDateFormat)
				queryParams := req.URL.Query()
				queryParams.Add(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)

				// prepare sig
				// create a test account information.
				privateKey, _ := crypto.GenerateKey()
				address := crypto.PubkeyToAddress(privateKey.PublicKey)
				log.Infof("address is: " + address.Hex())
				signedMsg := fmt.Sprintf(GnfdBuiltInDappSignedContentTemplate, "gnfd://"+mockBucketName+"/"+mockObjectName, validExpiryDateStr)
				unSignedContentHash := accounts.TextHash([]byte(signedMsg)) // personal sign.
				// sign data.
				sig, _ := crypto.Sign(unSignedContentHash, privateKey)

				queryParams.Add("signature", hexutil.Encode(sig))
				req.URL.RawQuery = queryParams.Encode()
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "NO_PERMISSION", // NO_PERMISSION as there is error occurs when verifying permission
			wantedResponseHeader: map[string]string{
				ContentDispositionHeader: "",
			},
		},

		{
			name: "view private object with bad object name",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				var a = permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&a, nil).Times(0) // can't reach here, so no need to verify permission
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				g.baseApp.SetPieceOp(pieceOpMock)
				pieceOpMock.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").AnyTimes()

				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s/view/%s/%s", scheme, testDomain, mockBucketName, "bad/./object.name")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set("User-Agent", "Chrome")
				validExpiryDateStr := time.Now().Add(-time.Hour * 60).Format(ExpiryDateFormat)
				queryParams := req.URL.Query()
				queryParams.Add(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				queryParams.Add("signature", "sample_signature")
				req.URL.RawQuery = queryParams.Encode()
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "INTERNAL_ERROR", // INTERNAL_ERROR as there is ErrInvalidExpiryDateParam
			wantedResponseHeader: map[string]string{
				ContentDispositionHeader: "",
			},
		},
		{
			name: "redirect for pdf/xml suffix objects",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PRIVATE}},
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
				path := fmt.Sprintf("%s%s/download/%s/%s", scheme, testDomain, mockBucketName, "test.pdf")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				log.Info("req.RequestURI is " + req.RequestURI)
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult:       "<a href=\"/download?objectPath=mock-bucket-name/test.pdf\">Found</a>.",
			wantedHttpRespCode: 302,
		},

		{
			name: "download for pdf/xml suffix objects",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PRIVATE}},
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
				path := fmt.Sprintf("%s%s/view?objectPath=%s/%s", scheme, testDomain, mockBucketName, "test.xml")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				log.Info("req.RequestURI is " + req.RequestURI)
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult:       "<!doctype html>",
			wantedHttpRespCode: 200,
		},

		{
			name: "download from non-browser",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PRIVATE}},
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
				path := fmt.Sprintf("%s%s/view?objectPath=%s/%s", scheme, testDomain, mockBucketName, "test.xml")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				log.Info("req.RequestURI is " + req.RequestURI)
				return req
			},
			wantedResult:       "Forbidden to access",
			wantedHttpRespCode: 403,
		},

		{
			name: "view private object with good signature and correct permission",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{BucketName: mockBucketName}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName, ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PRIVATE}},
					nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil)
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

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

				path := fmt.Sprintf("%s%s/view/%s/%s", scheme, testDomain, mockBucketName, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set("User-Agent", "Chrome")
				validExpiryDateStr := time.Now().Add(+time.Hour * 60).Format(ExpiryDateFormat)
				queryParams := req.URL.Query()
				queryParams.Add(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)

				// prepare sig
				// create a test account information.
				privateKey, _ := crypto.GenerateKey()
				address := crypto.PubkeyToAddress(privateKey.PublicKey)
				log.Infof("address is: " + address.Hex())
				signedMsg := fmt.Sprintf(GnfdBuiltInDappSignedContentTemplate, "gnfd://"+mockBucketName+"/"+mockObjectName, validExpiryDateStr)
				unSignedContentHash := accounts.TextHash([]byte(signedMsg)) // personal sign.
				// sign data.
				sig, _ := crypto.Sign(unSignedContentHash, privateKey)

				queryParams.Add("signature", hexutil.Encode(sig))
				req.URL.RawQuery = queryParams.Encode()
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "",
			wantedResponseHeader: map[string]string{
				ContentDispositionHeader: ContentDispositionInlineValue,
			},
		},

		{
			name: "view private object with but no enough quota",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetBucketByBucketName(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Bucket{BucketInfo: &storagetypes.BucketInfo{BucketName: mockBucketName}}, nil).Times(1)
				clientMock.EXPECT().GetObjectMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&metadatatypes.Object{ObjectInfo: &storagetypes.ObjectInfo{ObjectName: mockObjectName, ObjectStatus: storagetypes.OBJECT_STATUS_SEALED, Visibility: storagetypes.VISIBILITY_TYPE_PRIVATE}},
					nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil)

				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return(nil, downloader.ErrExceedBucketQuota).AnyTimes()
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

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

				path := fmt.Sprintf("%s%s/view/%s/%s", scheme, testDomain, mockBucketName, mockObjectName)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				req.Header.Set("User-Agent", "Chrome")
				validExpiryDateStr := time.Now().Add(+time.Hour * 60).Format(ExpiryDateFormat)
				queryParams := req.URL.Query()
				queryParams.Add(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)

				// prepare sig
				// create a test account information.
				privateKey, _ := crypto.GenerateKey()
				address := crypto.PubkeyToAddress(privateKey.PublicKey)
				log.Infof("address is: " + address.Hex())
				signedMsg := fmt.Sprintf(GnfdBuiltInDappSignedContentTemplate, "gnfd://"+mockBucketName+"/"+mockObjectName, validExpiryDateStr)
				unSignedContentHash := accounts.TextHash([]byte(signedMsg)) // personal sign.
				// sign data.
				sig, _ := crypto.Sign(unSignedContentHash, privateKey)

				queryParams.Add("signature", hexutil.Encode(sig))
				req.URL.RawQuery = queryParams.Encode()
				req.Header.Set("User-Agent", "Chrome")
				return req
			},
			wantedResult: "NO_ENOUGH_QUOTA",
			wantedResponseHeader: map[string]string{
				ContentDispositionHeader: "",
				ContentLengthHeader:      "",
				ContentRangeHeader:       "",
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetObjectByUniversalEndpointHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Contains(t, w.Body.String(), tt.wantedResult)
			if tt.wantedResponseHeader != nil {
				log.Info("tt.wantedResponseHeader is not nil")
				for k, v := range tt.wantedResponseHeader {
					assert.Equal(t, v, w.Header().Get(k))
				}
			}
			if tt.wantedHttpRespCode != 0 {
				assert.Equal(t, tt.wantedHttpRespCode, w.Code)
			}
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
