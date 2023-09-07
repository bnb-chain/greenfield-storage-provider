package gater

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	commonhttp "github.com/bnb-chain/greenfield-common/go/http"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
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
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "48656c6c6f20476f706865722")
				return request
			},
			wantedResult: "mock error",
		},
		{
			name: "failed to parse approval header",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "48656c6c6f20476f706865722")
				return request
			},
			wantedResult: "gnfd msg encoding error",
		},
		{
			name: "failed to query sp by operator address",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "48656c6c6f20476f7068657221")
				return request
			},
			wantedResult: "mock error",
		},
		{
			name: "sp is not in service status",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_JAILED}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "48656c6c6f20476f7068657221")
				return request
			},
			wantedResult: "sp is not in service status",
		},
		{
			name: "failed to unmarshal approval",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "48656c6c6f20476f7068657221")
				return request
			},
			wantedResult: "gnfd msg encoding error",
		},
		{
			name: "failed to basic check bucket approval msg",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
					Status: sptypes.STATUS_IN_SERVICE}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364636222c226275636b65745f6e616d65223a226d6f222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c227061796d656e745f61646472657373223a22222c227072696d6172795f73705f61646472657373223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364636222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c22636861726765645f726561645f71756f7461223a2230227d")
				return request
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
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c227061796d656e745f61646472657373223a22222c227072696d6172795f73705f61646472657373223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c22636861726765645f726561645f71756f7461223a2230227d")
				return request
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
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c227061796d656e745f61646472657373223a22222c227072696d6172795f73705f61646472657373223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c22636861726765645f726561645f71756f7461223a2230227d")
				return request
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
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					true, nil).Times(1)
				clientMock.EXPECT().AskCreateBucketApproval(gomock.Any(), gomock.Any()).Return(false, nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c227061796d656e745f61646472657373223a22222c227072696d6172795f73705f61646472657373223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c22636861726765645f726561645f71756f7461223a2230227d")
				return request
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
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					true, nil).Times(1)
				clientMock.EXPECT().AskCreateBucketApproval(gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c227061796d656e745f61646472657373223a22222c227072696d6172795f73705f61646472657373223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c22636861726765645f726561645f71756f7461223a2230227d")
				return request
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
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					true, nil).Times(1)
				approvalTask := &gfsptask.GfSpCreateBucketApprovalTask{
					Task:             &gfsptask.GfSpTask{},
					CreateBucketInfo: &storagetypes.MsgCreateBucket{BucketName: mockBucketName},
					Fingerprint:      []byte("mockSig"),
				}
				clientMock.EXPECT().AskCreateBucketApproval(gomock.Any(), gomock.Any()).Return(true, approvalTask, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c227061796d656e745f61646472657373223a22222c227072696d6172795f73705f61646472657373223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c22636861726765645f726561645f71756f7461223a2230227d")
				return request
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
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, migrateBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "48656c6c6f20476f7068657221")
				return request
			},
			wantedResult: "gnfd msg encoding error",
		},
		{
			name: "failed to basic check migrate bucket approval msg",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, migrateBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b226f70657261746f72223a226d6f636b4f70657261746f72222c226275636b65745f6e616d65223a226d6f222c226473745f7072696d6172795f73705f6964223a312c226473745f7072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d7d")
				return request
			},
			wantedResult: "gnfd msg validate error",
		},
		{
			name: "failed to verify authentication",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, migrateBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b226f70657261746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226473745f7072696d6172795f73705f6964223a312c226473745f7072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d7d")
				return request
			},
			wantedResult: "mock error",
		},
		{
			name: "no permission to operate",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, migrateBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b226f70657261746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226473745f7072696d6172795f73705f6964223a312c226473745f7072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d7d")
				return request
			},
			wantedResult: "no permission",
		},
		{
			name: "failed to ask migrate bucket approval",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					true, nil).Times(1)
				clientMock.EXPECT().AskMigrateBucketApproval(gomock.Any(), gomock.Any()).Return(false, nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, migrateBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b226f70657261746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226473745f7072696d6172795f73705f6964223a312c226473745f7072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d7d")
				return request
			},
			wantedResult: "mock error",
		},
		{
			name: "refuse the ask create bucket approval",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					true, nil).Times(1)
				clientMock.EXPECT().AskMigrateBucketApproval(gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, migrateBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b226f70657261746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226473745f7072696d6172795f73705f6964223a312c226473745f7072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d7d")
				return request
			},
			wantedResult: "approval request is refused",
		},
		{
			name: "success",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					true, nil).Times(1)
				approvalTask := &gfsptask.GfSpMigrateBucketApprovalTask{
					Task:              &gfsptask.GfSpTask{},
					MigrateBucketInfo: &storagetypes.MsgMigrateBucket{BucketName: mockBucketName},
				}
				clientMock.EXPECT().AskMigrateBucketApproval(gomock.Any(), gomock.Any()).Return(true, approvalTask, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, migrateBucketApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b226f70657261746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226473745f7072696d6172795f73705f6964223a312c226473745f7072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d7d")
				return request
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
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createObjectApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "48656c6c6f20476f7068657221")
				return request
			},
			wantedResult: "gnfd msg encoding error",
		},
		{
			name: "failed to basic check migrate bucket approval msg",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, createObjectApprovalAction)
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f222c226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c227061796c6f61645f73697a65223a2230222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c22636f6e74656e745f74797065223a226170706c69636174696f6e2f6a736f6e222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c226578706563745f636865636b73756d73223a5b5d2c22726564756e64616e63795f74797065223a22524544554e44414e43595f45435f54595045227d")
				return request
			},
			wantedResult: "gnfd msg validate error",
		},
		{
			name: "failed to check sp and bucket status",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)

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
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c227061796c6f61645f73697a65223a2230222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c22636f6e74656e745f74797065223a226170706c69636174696f6e2f6a736f6e222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c226578706563745f636865636b73756d73223a5b5d2c22726564756e64616e63795f74797065223a22524544554e44414e43595f45435f54595045227d")
				return request
			},
			wantedResult: "bucket is not in service status",
		},
		{
			name: "failed to verify authentication",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, mockErr).Times(1)
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
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c227061796c6f61645f73697a65223a2230222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c22636f6e74656e745f74797065223a226170706c69636174696f6e2f6a736f6e222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c226578706563745f636865636b73756d73223a5b5d2c22726564756e64616e63795f74797065223a22524544554e44414e43595f45435f54595045227d")
				return request
			},
			wantedResult: "mock error",
		},
		{
			name: "no permission to operate",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
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
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c227061796c6f61645f73697a65223a2230222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c22636f6e74656e745f74797065223a226170706c69636174696f6e2f6a736f6e222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c226578706563745f636865636b73756d73223a5b5d2c22726564756e64616e63795f74797065223a22524544554e44414e43595f45435f54595045227d")
				return request
			},
			wantedResult: "no permission",
		},
		{
			name: "failed to ask migrate bucket approval",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					true, nil).Times(1)
				clientMock.EXPECT().AskCreateObjectApproval(gomock.Any(), gomock.Any()).Return(false, nil, mockErr).Times(1)
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
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c227061796c6f61645f73697a65223a2230222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c22636f6e74656e745f74797065223a226170706c69636174696f6e2f6a736f6e222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c226578706563745f636865636b73756d73223a5b5d2c22726564756e64616e63795f74797065223a22524544554e44414e43595f45435f54595045227d")
				return request
			},
			wantedResult: "mock error",
		},
		{
			name: "refuse the ask create bucket approval",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					true, nil).Times(1)
				clientMock.EXPECT().AskCreateObjectApproval(gomock.Any(), gomock.Any()).Return(false, nil, nil).Times(1)
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
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c227061796c6f61645f73697a65223a2230222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c22636f6e74656e745f74797065223a226170706c69636174696f6e2f6a736f6e222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c226578706563745f636865636b73756d73223a5b5d2c22726564756e64616e63795f74797065223a22524544554e44414e43595f45435f54595045227d")
				return request
			},
			wantedResult: "approval request is refused",
		},
		{
			name: "success",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				clientMock.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					true, nil).Times(1)
				approvalTask := &gfsptask.GfSpCreateObjectApprovalTask{
					Task:             &gfsptask.GfSpTask{},
					CreateObjectInfo: &storagetypes.MsgCreateObject{BucketName: mockBucketName},
				}
				clientMock.EXPECT().AskCreateObjectApproval(gomock.Any(), gomock.Any()).Return(true, approvalTask, nil).Times(1)
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
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				request.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2263726561746f72223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364635222c226275636b65745f6e616d65223a226d6f636b2d6275636b65742d6e616d65222c226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c227061796c6f61645f73697a65223a2230222c227669736962696c697479223a225649534942494c4954595f545950455f494e4845524954222c22636f6e74656e745f74797065223a226170706c69636174696f6e2f6a736f6e222c227072696d6172795f73705f617070726f76616c223a7b22657870697265645f686569676874223a223130222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a302c22736967223a6e756c6c7d2c226578706563745f636865636b73756d73223a5b5d2c22726564756e64616e63795f74797065223a22524544554e44414e43595f45435f54595045227d")
				return request
			},
			wantedResult: "",
		},
		{
			name: "unsupported request type",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s?%s=%s", scheme, testDomain, GetApprovalPath, ActionQuery, "unknown")
				request := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				request.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				request.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				return request
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
