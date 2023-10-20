package gater

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	commonhttp "github.com/bnb-chain/greenfield-common/go/http"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func mockGetApprovalRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path(GetApprovalPath).Name(approvalRouterName).Methods(http.MethodGet).HandlerFunc(g.getApprovalHandler).
		Queries(ActionQuery, "{action}")
	return router
}

func TestGateModular_getApprovalHandlerCreateBucketApproval(t *testing.T) {
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
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "48656c6c6f20476f706865722")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "failed to parse approval header",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "48656c6c6f20476f706865722")
				return req
			},
			wantedResult: "gnfd msg decoding error",
		},
		{
			name: "failed to query sp by operator address",
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
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "sp is not in service status",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_JAILED}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "sp is not in service status",
		},
		{
			name: "failed to unmarshal approval",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "gnfd msg decoding error",
		},
		{
			name: "failed to basic check bucket approval msg",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364636222c226275636b65745f6e616d65223a226d6f222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c227061796d656e745f61646472657373223a22222c227072696d6172795f73705f61646472657373223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364636222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c22636861726765645f726561645f71756f7461223a2230227d")
				return req
			},
			wantedResult: "gnfd msg validate error",
		},
		{
			name: "failed to verify authentication",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c227061796d656e745f61646472657373223a22222c227072696d6172795f73705f61646472657373223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c22636861726765645f726561645f71756f7461223a2230227d")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "no permission to operate",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c227061796d656e745f61646472657373223a22222c227072696d6172795f73705f61646472657373223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c22636861726765645f726561645f71756f7461223a2230227d")
				return req
			},
			wantedResult: "no permission",
		},
		{
			name: "failed to ask create bucket approval",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				clientMock.EXPECT().AskCreateBucketApproval(gomock.Any(), gomock.Any()).Return(false, nil,
					mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c227061796d656e745f61646472657373223a22222c227072696d6172795f73705f61646472657373223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c22636861726765645f726561645f71756f7461223a2230227d")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "refuse the ask create bucket approval",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				clientMock.EXPECT().AskCreateBucketApproval(gomock.Any(), gomock.Any()).Return(false, nil,
					nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c227061796d656e745f61646472657373223a22222c227072696d6172795f73705f61646472657373223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c22636861726765645f726561645f71756f7461223a2230227d")
				return req
			},
			wantedResult: "approval request is refused",
		},
		{
			name: "success",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				approvalTask := &gfsptask.GfSpCreateBucketApprovalTask{
					Task:             &gfsptask.GfSpTask{},
					CreateBucketInfo: &storagetypes.MsgCreateBucket{BucketName: mockBucketName},
					Fingerprint:      []byte("mockSig"),
				}
				clientMock.EXPECT().AskCreateBucketApproval(gomock.Any(), gomock.Any()).Return(true,
					approvalTask, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c227061796d656e745f61646472657373223a22222c227072696d6172795f73705f61646472657373223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c22636861726765645f726561645f71756f7461223a2230227d")
				return req
			},
			wantedResult: "",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetApprovalRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Contains(t, w.Body.String(), tt.wantedResult)
		})
	}
}

func TestGateModular_getApprovalHandlerMigrateBucketApproval(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *GateModular
		request      func() *http.Request
		wantedResult string
	}{
		{
			name: "failed to unmarshal approval",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, migrateBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "gnfd msg decoding error",
		},
		{
			name: "failed to basic check migrate bucket approval msg",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, migrateBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b226f70657261746f72223a226d6f636b4f70657261746f72222c226275636b65745f6e616d65223a226d6f222c226473745f7072696d6172795f73705f6964223a312c226473745f7072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d7d")
				return req
			},
			wantedResult: "gnfd msg validate error",
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
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, migrateBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b226f70657261746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226473745f7072696d6172795f73705f6964223a312c226473745f7072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d7d")
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
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, migrateBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b226f70657261746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226473745f7072696d6172795f73705f6964223a312c226473745f7072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d7d")
				return req
			},
			wantedResult: "no permission",
		},
		{
			name: "failed to ask migrate bucket approval",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(true, nil).Times(1)
				clientMock.EXPECT().AskMigrateBucketApproval(gomock.Any(), gomock.Any()).Return(false, nil,
					mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, migrateBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b226f70657261746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226473745f7072696d6172795f73705f6964223a312c226473745f7072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d7d")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "refuse the ask create bucket approval",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				clientMock.EXPECT().AskMigrateBucketApproval(gomock.Any(), gomock.Any()).Return(false, nil,
					nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, migrateBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b226f70657261746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226473745f7072696d6172795f73705f6964223a312c226473745f7072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d7d")
				return req
			},
			wantedResult: "approval request is refused",
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
				approvalTask := &gfsptask.GfSpMigrateBucketApprovalTask{
					Task:              &gfsptask.GfSpTask{},
					MigrateBucketInfo: &storagetypes.MsgMigrateBucket{BucketName: mockBucketName},
				}
				clientMock.EXPECT().AskMigrateBucketApproval(gomock.Any(), gomock.Any()).Return(true, approvalTask,
					nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, migrateBucketApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b226f70657261746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226473745f7072696d6172795f73705f6964223a312c226473745f7072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d7d")
				return req
			},
			wantedResult: "",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetApprovalRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Contains(t, w.Body.String(), tt.wantedResult)
		})
	}
}

func TestGateModular_getApprovalHandlerCreateObjectApproval(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *GateModular
		request      func() *http.Request
		wantedResult string
	}{
		{
			name: "failed to unmarshal approval",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createObjectApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "gnfd msg decoding error",
		},
		{
			name: "failed to basic check migrate bucket approval msg",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createObjectApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f222c226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c227061796c6f61645f73697a65223a2230222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c22636f6e74656e745f74797065223a226170706c69636174696f6e2f6a736f6e222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c226578706563745f636865636b73756d73223a5b5d2c22726564756e64616e63795f74797065223a22524544554e44414e43595f45435f54595045227d")
				return req
			},
			wantedResult: "gnfd msg validate error",
		},
		{
			name: "failed to check sp and bucket status",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)

				m := consensus.NewMockConsensus(ctrl)
				g.baseApp.SetConsensus(m)
				m.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				m.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_DISCONTINUED}, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createObjectApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c227061796c6f61645f73697a65223a2230222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c22636f6e74656e745f74797065223a226170706c69636174696f6e2f6a736f6e222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c226578706563745f636865636b73756d73223a5b5d2c22726564756e64616e63795f74797065223a22524544554e44414e43595f45435f54595045227d")
				return req
			},
			wantedResult: "bucket is not in service status",
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

				m := consensus.NewMockConsensus(ctrl)
				g.baseApp.SetConsensus(m)
				m.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				m.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createObjectApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c227061796c6f61645f73697a65223a2230222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c22636f6e74656e745f74797065223a226170706c69636174696f6e2f6a736f6e222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c226578706563745f636865636b73756d73223a5b5d2c22726564756e64616e63795f74797065223a22524544554e44414e43595f45435f54595045227d")
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

				m := consensus.NewMockConsensus(ctrl)
				g.baseApp.SetConsensus(m)
				m.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				m.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createObjectApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c227061796c6f61645f73697a65223a2230222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c22636f6e74656e745f74797065223a226170706c69636174696f6e2f6a736f6e222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c226578706563745f636865636b73756d73223a5b5d2c22726564756e64616e63795f74797065223a22524544554e44414e43595f45435f54595045227d")
				return req
			},
			wantedResult: "no permission",
		},
		{
			name: "failed to ask migrate bucket approval",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				clientMock.EXPECT().AskCreateObjectApproval(gomock.Any(), gomock.Any()).Return(false, nil,
					mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				m := consensus.NewMockConsensus(ctrl)
				g.baseApp.SetConsensus(m)
				m.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				m.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createObjectApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c227061796c6f61645f73697a65223a2230222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c22636f6e74656e745f74797065223a226170706c69636174696f6e2f6a736f6e222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c226578706563745f636865636b73756d73223a5b5d2c22726564756e64616e63795f74797065223a22524544554e44414e43595f45435f54595045227d")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "refuse the ask create bucket approval",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil).Times(1)
				clientMock.EXPECT().AskCreateObjectApproval(gomock.Any(), gomock.Any()).Return(false, nil,
					nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				m := consensus.NewMockConsensus(ctrl)
				g.baseApp.SetConsensus(m)
				m.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				m.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createObjectApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c227061796c6f61645f73697a65223a2230222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c22636f6e74656e745f74797065223a226170706c69636174696f6e2f6a736f6e222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c226578706563745f636865636b73756d73223a5b5d2c22726564756e64616e63795f74797065223a22524544554e44414e43595f45435f54595045227d")
				return req
			},
			wantedResult: "approval request is refused",
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
				approvalTask := &gfsptask.GfSpCreateObjectApprovalTask{
					Task:             &gfsptask.GfSpTask{},
					CreateObjectInfo: &storagetypes.MsgCreateObject{BucketName: mockBucketName},
				}
				clientMock.EXPECT().AskCreateObjectApproval(gomock.Any(), gomock.Any()).Return(true, approvalTask,
					nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				m := consensus.NewMockConsensus(ctrl)
				g.baseApp.SetConsensus(m)
				m.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				m.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketStatus: storagetypes.BUCKET_STATUS_CREATED}, nil).Times(1)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createObjectApprovalAction)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c227061796c6f61645f73697a65223a2230222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c22636f6e74656e745f74797065223a226170706c69636174696f6e2f6a736f6e222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c226578706563745f636865636b73756d73223a5b5d2c22726564756e64616e63795f74797065223a22524544554e44414e43595f45435f54595045227d")
				return req
			},
			wantedResult: "",
		},
		{
			name: "unsupported request type",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, "unknown")
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "unsupported request type",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetApprovalRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Contains(t, w.Body.String(), tt.wantedResult)
		})
	}
}

func mockGetChallengeInfoRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path(GetChallengeInfoPath).Name(getChallengeInfoRouterName).Methods(http.MethodGet).HandlerFunc(g.getChallengeInfoHandler)
	return router
}

func TestGateModular_getChallengeInfoHandler(t *testing.T) {
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
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, GetChallengeInfoPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "failed to parse object id",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, GetChallengeInfoPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdObjectIDHeader, "test")
				return req
			},
			wantedResult: "invalid request header",
		},
		{
			name: "No such object",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfoByID(gomock.Any(), gomock.Any()).Return(nil,
					errors.New("No such object")).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, GetChallengeInfoPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdObjectIDHeader, "1")
				return req
			},
			wantedResult: "no such object",
		},
		{
			name: "failed to get object info from consensus",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfoByID(gomock.Any(), gomock.Any()).Return(nil,
					mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, GetChallengeInfoPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdObjectIDHeader, "1")
				return req
			},
			wantedResult: "failed to get object info from consensus",
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

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfoByID(gomock.Any(), gomock.Any()).Return(
					&storagetypes.ObjectInfo{ObjectName: mockObjectName}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, GetChallengeInfoPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdObjectIDHeader, "1")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "no permission",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfoByID(gomock.Any(), gomock.Any()).Return(
					&storagetypes.ObjectInfo{ObjectName: mockObjectName}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, GetChallengeInfoPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdObjectIDHeader, "1")
				return req
			},
			wantedResult: "no permission",
		},
		{
			name: "failed to get bucket info from consensus",
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
				consensusMock.EXPECT().QueryObjectInfoByID(gomock.Any(), gomock.Any()).Return(
					&storagetypes.ObjectInfo{ObjectName: mockObjectName}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, GetChallengeInfoPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdObjectIDHeader, "1")
				return req
			},
			wantedResult: "failed to get bucket info from consensus",
		},
		{
			name: "failed to parse segment index",
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
				consensusMock.EXPECT().QueryObjectInfoByID(gomock.Any(), gomock.Any()).Return(
					&storagetypes.ObjectInfo{ObjectName: mockObjectName}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketName: mockBucketName}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, GetChallengeInfoPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdObjectIDHeader, "1")
				req.Header.Set(GnfdRedundancyIndexHeader, "test")
				return req
			},
			wantedResult: "invalid request header",
		},
		{
			name: "failed to parse segment index",
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
				consensusMock.EXPECT().QueryObjectInfoByID(gomock.Any(), gomock.Any()).Return(
					&storagetypes.ObjectInfo{ObjectName: mockObjectName}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketName: mockBucketName}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, GetChallengeInfoPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdObjectIDHeader, "1")
				req.Header.Set(GnfdRedundancyIndexHeader, "0")
				req.Header.Set(GnfdPieceIndexHeader, "test")
				return req
			},
			wantedResult: "invalid request header",
		},
		{
			name: "failed to get storage params",
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
				consensusMock.EXPECT().QueryObjectInfoByID(gomock.Any(), gomock.Any()).Return(
					&storagetypes.ObjectInfo{ObjectName: mockObjectName}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketName: mockBucketName}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(nil, mockErr)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, GetChallengeInfoPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdObjectIDHeader, "1")
				req.Header.Set(GnfdRedundancyIndexHeader, "0")
				req.Header.Set(GnfdPieceIndexHeader, "2")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "redundancy index is less than 0 and failed to get challenge info",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil)
				clientMock.EXPECT().GetChallengeInfo(gomock.Any(), gomock.Any()).Return(nil, nil, nil, mockErr).
					Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfoByID(gomock.Any(), gomock.Any()).Return(
					&storagetypes.ObjectInfo{ObjectName: mockObjectName}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketName: mockBucketName}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(
					&storagetypes.Params{MaxPayloadSize: 10}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				g.baseApp.SetPieceOp(pieceOpMock)
				pieceOpMock.EXPECT().SegmentPieceSize(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, GetChallengeInfoPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdObjectIDHeader, "1")
				req.Header.Set(GnfdRedundancyIndexHeader, "-1")
				req.Header.Set(GnfdPieceIndexHeader, "2")
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "redundancy index is greater than 0 and success",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(true, nil)
				clientMock.EXPECT().GetChallengeInfo(gomock.Any(), gomock.Any()).Return([]byte("mockIntegrityHash"),
					[][]byte{[]byte("mockChecksums")}, []byte("mockData"), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfoByID(gomock.Any(), gomock.Any()).Return(
					&storagetypes.ObjectInfo{ObjectName: mockObjectName}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketName: mockBucketName}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(
					&storagetypes.Params{MaxPayloadSize: 10}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				g.baseApp.SetPieceOp(pieceOpMock)
				pieceOpMock.EXPECT().ECPieceSize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, GetChallengeInfoPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdObjectIDHeader, "1")
				req.Header.Set(GnfdRedundancyIndexHeader, "1")
				req.Header.Set(GnfdPieceIndexHeader, "2")
				return req
			},
			wantedResult: "",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetChallengeInfoRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Contains(t, w.Body.String(), tt.wantedResult)
		})
	}
}

func mockGetReplicateHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path(ReplicateObjectPiecePath).Name(replicateObjectPieceRouterName).Methods(http.MethodPut).HandlerFunc(g.replicateHandler)
	return router
}

func TestGateModular_getReplicateHandlerHandler(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *GateModular
		request      func() *http.Request
		wantedResult string
	}{
		{
			name: "failed to parse receive header",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, ReplicateObjectPiecePath)
				req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(""))
				// no expire header was set between the http request of SPs, the expiry info is judged by the task update time info of the receive task
				req.Header.Set(GnfdReceiveMsgHeader, "48656c6c6f20476f706865722")
				return req
			},
			wantedResult: "gnfd msg decoding error",
		},
		{
			name: "failed to unmarshal receive header",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, ReplicateObjectPiecePath)
				req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(""))
				// no expire header was set between the http request of SPs, the expiry info is judged by the task update time info of the receive task
				req.Header.Set(GnfdRecoveryMsgHeader, "48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "gnfd msg decoding error",
		},
		{
			name: "failed to verify receive task",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, ReplicateObjectPiecePath)
				req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(""))
				// no expire header was set between the http request of SPs, the expiry info is judged by the task update time info of the receive task
				req.Header.Set(GnfdReceiveMsgHeader, "7b227461736b223a7b7d2c226f626a6563745f696e666f223a7b226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c226964223a2230227d2c2273746f726167655f706172616d73223a7b2276657273696f6e65645f706172616d73223a7b7d2c226d61785f7061796c6f61645f73697a65223a31307d2c227365676d656e745f696478223a312c22726564756e64616e63795f696478223a327d")
				return req
			},
			wantedResult: "signature is invalid",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetReplicateHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Contains(t, w.Body.String(), tt.wantedResult)
		})
	}
}

func mockGetRecoverDataHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path(RecoverObjectPiecePath).Name(recoveryPieceRouterName).Methods(http.MethodGet).HandlerFunc(g.getRecoverDataHandler)
	return router
}

func TestGateModular_getRecoverDataHandler(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *GateModular
		request      func() *http.Request
		wantedResult string
	}{
		{
			name: "failed to parse recovery header",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, RecoverObjectPiecePath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				// no expire header was set between the http request of SPs, the expiry info is judged by the task update time info of the receive task
				req.Header.Set(GnfdRecoveryMsgHeader, "48656c6c6f20476f706865722")
				return req
			},
			wantedResult: "gnfd msg decoding error",
		},
		{
			name: "failed to unmarshal recovery msg header",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, RecoverObjectPiecePath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				// no expire header was set between the http request of SPs, the expiry info is judged by the task update time info of the receive task
				req.Header.Set(GnfdRecoveryMsgHeader, "48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "gnfd msg decoding error",
		},
		{
			name: "failed to get recover task address",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, RecoverObjectPiecePath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				// no expire header was set between the http request of SPs, the expiry info is judged by the task update time info of the receive task
				req.Header.Set(GnfdRecoveryMsgHeader, "7b227461736b223a7b7d2c226f626a6563745f696e666f223a7b226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c226964223a2230227d2c2273746f726167655f706172616d73223a7b2276657273696f6e65645f706172616d73223a7b7d2c226d61785f7061796c6f61645f73697a65223a31307d2c227365676d656e745f696478223a312c22726564756e64616e63795f696478223a327d")
				return req
			},
			wantedResult: "signature is invalid",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetRecoverDataHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Contains(t, w.Body.String(), tt.wantedResult)
		})
	}
}

func TestGateModular_getRecoverPiece(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *GateModular
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name: "GetBucketPrimarySPID returns error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			wantedIsErr:  true,
			wantedErrStr: mockErr.Error(),
		},
		{
			name: "QuerySPByID returns error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				consensusMock.EXPECT().QuerySPByID(gomock.Any(), gomock.Any()).Return(
					nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			wantedIsErr:  true,
			wantedErrStr: mockErr.Error(),
		},
		{
			name: "GetGlobalVirtualGroup returns error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				consensusMock.EXPECT().QuerySPByID(gomock.Any(), gomock.Any()).Return(
					&sptypes.StorageProvider{
						Id:              1,
						OperatorAddress: "0x1C7C8A668e23aED291f78fC2f3b1865Acc87b6F6",
					}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			wantedIsErr:  true,
			wantedErrStr: mockErr.Error(),
		},
		{
			name: "recovery request not come from the correct secondary sp",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				consensusMock.EXPECT().QuerySPByID(gomock.Any(), gomock.Any()).Return(
					&sptypes.StorageProvider{
						Id:              1,
						OperatorAddress: "test",
					}, nil).Times(2)
				g.baseApp.SetConsensus(consensusMock)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroup{
						Id:             1,
						SecondarySpIds: []uint32{0},
					}, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			wantedIsErr:  true,
			wantedErrStr: ErrRecoverySP.Error(),
		},
		{
			name: "not one of secondary sp",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				consensusMock.EXPECT().QuerySPByID(gomock.Any(), gomock.Any()).Return(
					&sptypes.StorageProvider{
						Id:              1,
						OperatorAddress: "0x1C7C8A668e23aED291f78fC2f3b1865Acc87b6F6",
					}, nil).Times(2)
				g.baseApp.SetConsensus(consensusMock)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroup{
						Id:             1,
						SecondarySpIds: []uint32{0},
					}, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			wantedIsErr:  true,
			wantedErrStr: ErrRecoverySP.Error(),
		},
		{
			name: "failed to download piece",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				consensusMock.EXPECT().QuerySPByID(gomock.Any(), gomock.Any()).Return(
					&sptypes.StorageProvider{
						Id:              1,
						OperatorAddress: "0x1C7C8A668e23aED291f78fC2f3b1865Acc87b6F6",
					}, nil).Times(2)
				g.baseApp.SetConsensus(consensusMock)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroup{
						Id:             1,
						SecondarySpIds: []uint32{0},
					}, nil).Times(1)
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				g.baseApp.SetOperatorAddress("0x1C7C8A668e23aED291f78fC2f3b1865Acc87b6F6")

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				pieceOpMock.EXPECT().ECPieceSize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
				pieceOpMock.EXPECT().ECPieceKey(gomock.Any(), gomock.Any(), gomock.Any()).Return("test").Times(1)
				g.baseApp.SetPieceOp(pieceOpMock)
				return g
			},
			wantedIsErr:  true,
			wantedErrStr: mockErr.Error(),
		},
		{
			name: "success",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				consensusMock.EXPECT().QuerySPByID(gomock.Any(), gomock.Any()).Return(
					&sptypes.StorageProvider{
						Id:              1,
						OperatorAddress: "0x1C7C8A668e23aED291f78fC2f3b1865Acc87b6F6",
					}, nil).Times(2)
				g.baseApp.SetConsensus(consensusMock)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroup{
						Id:             1,
						SecondarySpIds: []uint32{0},
					}, nil).Times(1)
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("a"), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				g.baseApp.SetOperatorAddress("0x1C7C8A668e23aED291f78fC2f3b1865Acc87b6F6")

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				pieceOpMock.EXPECT().ECPieceSize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
				pieceOpMock.EXPECT().ECPieceKey(gomock.Any(), gomock.Any(), gomock.Any()).Return("test").Times(1)
				g.baseApp.SetPieceOp(pieceOpMock)
				return g
			},
			wantedIsErr: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			objectInfo := &storagetypes.ObjectInfo{
				ObjectName: mockObjectName,
				Id:         sdkmath.NewUint(1),
			}
			bucketInfo := &storagetypes.BucketInfo{
				BucketName: mockBucketName,
				Id:         sdkmath.NewUint(2),
			}
			params := &storagetypes.Params{VersionedParams: storagetypes.VersionedParams{
				MaxSegmentSize:          10,
				RedundantDataChunkNum:   4,
				RedundantParityChunkNum: 2,
			}}
			recoveryTask := gfsptask.GfSpRecoverPieceTask{
				Task:          &gfsptask.GfSpTask{},
				ObjectInfo:    objectInfo,
				StorageParams: params,
			}
			addr := sdktypes.MustAccAddressFromHex("0x1C7C8A668e23aED291f78fC2f3b1865Acc87b6F6")
			result, err := tt.fn().getRecoverPiece(context.TODO(), objectInfo, bucketInfo, recoveryTask, params, addr)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErrStr)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestGateModular_getRecoverSegment(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *GateModular
		dataChunkNum uint32
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name: "getSPID returns error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(
					nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				pieceOpMock.EXPECT().ECPieceSize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
				g.baseApp.SetPieceOp(pieceOpMock)
				return g
			},
			dataChunkNum: 4,
			wantedIsErr:  true,
			wantedErrStr: mockErr.Error(),
		},
		{
			name: "GetBucketPrimarySPID returns error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(
					&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				pieceOpMock.EXPECT().ECPieceSize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
				g.baseApp.SetPieceOp(pieceOpMock)
				return g
			},
			dataChunkNum: 4,
			wantedIsErr:  true,
			wantedErrStr: mockErr.Error(),
		},
		{
			name: "not the right the primary SP",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(
					&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 2}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				pieceOpMock.EXPECT().ECPieceSize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
				g.baseApp.SetPieceOp(pieceOpMock)
				return g
			},
			dataChunkNum: 4,
			wantedIsErr:  true,
			wantedErrStr: ErrRecoverySP.Error(),
		},
		{
			name: "failed to get global virtual group",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(
					&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				pieceOpMock.EXPECT().ECPieceSize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
				g.baseApp.SetPieceOp(pieceOpMock)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			dataChunkNum: 4,
			wantedIsErr:  true,
			wantedErrStr: ErrRecoverySP.Error(),
		},
		{
			name: "QuerySPByID error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(
					&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				consensusMock.EXPECT().QuerySPByID(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				pieceOpMock.EXPECT().ECPieceSize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
				g.baseApp.SetPieceOp(pieceOpMock)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroup{
						Id:             1,
						SecondarySpIds: []uint32{0},
					}, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			dataChunkNum: 4,
			wantedIsErr:  true,
			wantedErrStr: mockErr.Error(),
		},
		{
			name: "isOneOfSecondary is false",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(
					&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				consensusMock.EXPECT().QuerySPByID(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Id: 1, OperatorAddress: "a"}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				pieceOpMock.EXPECT().ECPieceSize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
				g.baseApp.SetPieceOp(pieceOpMock)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroup{
						Id:             1,
						SecondarySpIds: []uint32{0},
					}, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			dataChunkNum: 4,
			wantedIsErr:  true,
			wantedErrStr: ErrRecoverySP.Error(),
		},
		{
			name: "redundancyIdx is less than data chunk num and get piece returns error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(
					&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				consensusMock.EXPECT().QuerySPByID(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Id:              1,
					OperatorAddress: "0x1C7C8A668e23aED291f78fC2f3b1865Acc87b6F6",
				}, nil).Times(2)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				pieceOpMock.EXPECT().ECPieceSize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
				pieceOpMock.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").Times(1)
				g.baseApp.SetPieceOp(pieceOpMock)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroup{
						Id:             1,
						SecondarySpIds: []uint32{0},
					}, nil).Times(1)
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			dataChunkNum: 4,
			wantedIsErr:  true,
			wantedErrStr: mockErr.Error(),
		},
		{
			name: "redundancyIdx is greater than data chunk num and get piece returns error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(
					&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				consensusMock.EXPECT().QuerySPByID(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Id:              1,
					OperatorAddress: "0x1C7C8A668e23aED291f78fC2f3b1865Acc87b6F6",
				}, nil).Times(2)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				pieceOpMock.EXPECT().ECPieceSize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
				pieceOpMock.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").Times(1)
				pieceOpMock.EXPECT().SegmentPieceSize(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
				g.baseApp.SetPieceOp(pieceOpMock)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroup{
						Id:             1,
						SecondarySpIds: []uint32{0},
					}, nil).Times(1)
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			dataChunkNum: 0,
			wantedIsErr:  true,
			wantedErrStr: mockErr.Error(),
		},
		{
			name: "failed to ec encode data when recovering secondary SP",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(
					&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				consensusMock.EXPECT().QuerySPByID(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Id:              1,
					OperatorAddress: "0x1C7C8A668e23aED291f78fC2f3b1865Acc87b6F6",
				}, nil).Times(2)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				pieceOpMock.EXPECT().ECPieceSize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
				pieceOpMock.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").Times(1)
				pieceOpMock.EXPECT().SegmentPieceSize(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
				g.baseApp.SetPieceOp(pieceOpMock)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroup{
						Id:             1,
						SecondarySpIds: []uint32{0},
					}, nil).Times(1)
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte("test"), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			dataChunkNum: 0,
			wantedIsErr:  true,
			wantedErrStr: "cannot create Encoder with less than one data shard or less than zero parity shards",
		},
		{
			name: "success",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(
					&sptypes.StorageProvider{Id: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroupFamily{PrimarySpId: 1}, nil).Times(1)
				consensusMock.EXPECT().QuerySPByID(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Id:              1,
					OperatorAddress: "0x1C7C8A668e23aED291f78fC2f3b1865Acc87b6F6",
				}, nil).Times(2)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				pieceOpMock.EXPECT().ECPieceSize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)
				pieceOpMock.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").Times(1)
				g.baseApp.SetPieceOp(pieceOpMock)

				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&virtualgrouptypes.GlobalVirtualGroup{
						Id:             1,
						SecondarySpIds: []uint32{0},
					}, nil).Times(1)
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte("test"), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			dataChunkNum: 4,
			wantedIsErr:  false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			objectInfo := &storagetypes.ObjectInfo{
				ObjectName: mockObjectName,
				Id:         sdkmath.NewUint(1),
			}
			bucketInfo := &storagetypes.BucketInfo{
				BucketName: mockBucketName,
				Id:         sdkmath.NewUint(2),
			}
			params := &storagetypes.Params{VersionedParams: storagetypes.VersionedParams{
				MaxSegmentSize:          10,
				RedundantDataChunkNum:   tt.dataChunkNum,
				RedundantParityChunkNum: 2,
			}}
			recoveryTask := gfsptask.GfSpRecoverPieceTask{
				Task:          &gfsptask.GfSpTask{},
				ObjectInfo:    objectInfo,
				StorageParams: params,
			}
			addr := sdktypes.MustAccAddressFromHex("0x1C7C8A668e23aED291f78fC2f3b1865Acc87b6F6")
			result, err := tt.fn().getRecoverSegment(context.TODO(), objectInfo, bucketInfo, recoveryTask, params, addr)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErrStr)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
