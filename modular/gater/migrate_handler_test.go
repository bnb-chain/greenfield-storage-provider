package gater

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield/sdk/keys"
	permissiontypes "github.com/bnb-chain/greenfield/x/permission/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	commonhttp "github.com/bnb-chain/greenfield-common/go/http"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func mockNotifyMigrateSwapOutHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path(NotifyMigrateSwapOutTaskPath).Name(notifyMigrateSwapOutRouterName).Methods(http.MethodPost).
		HandlerFunc(g.notifyMigrateSwapOutHandler)
	return router
}

func TestGateModular_notifyMigrateSwapOutHandler(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *GateModular
		request      func() *http.Request
		wantedResult string
	}{
		{
			name: "failed to parse migrate swap out header",
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
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, NotifyMigrateSwapOutTaskPath)
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdMigrateSwapOutMsgHeader, "48656c6c6f20476f706865722")
				return req
			},
			wantedResult: "gnfd msg decoding error",
		},
		{
			name: "failed to unmarshal migrate swap out msg",
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
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, NotifyMigrateSwapOutTaskPath)
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdMigrateSwapOutMsgHeader, "48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "gnfd msg decoding error",
		},
		{
			name: "failed to notify migrate swap out",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().NotifyMigrateSwapOut(gomock.Any(), gomock.Any()).Return(mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, NotifyMigrateSwapOutTaskPath)
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdMigrateSwapOutMsgHeader, "7b2273746f726167655f70726f7669646572223a226d6f636b53746f7261676550726f7669646572227d")
				return req
			},
			wantedResult: "failed to notify migrate swap out",
		},
		{
			name: "success",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().NotifyMigrateSwapOut(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, NotifyMigrateSwapOutTaskPath)
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdMigrateSwapOutMsgHeader, "7b2273746f726167655f70726f7669646572223a226d6f636b53746f7261676550726f7669646572227d")
				return req
			},
			wantedResult: "",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockNotifyMigrateSwapOutHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Contains(t, w.Body.String(), tt.wantedResult)
		})
	}
}

func mockMigratePieceHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path(MigratePiecePath).Name(migratePieceRouterName).Methods(http.MethodGet).HandlerFunc(g.migratePieceHandler)
	return router
}

func makeMockMigrateGVGTaskHeader(t *testing.T, addValidExpireTime bool) string {
	mockTask := &gfsptask.GfSpMigrateGVGTask{}
	if addValidExpireTime {
		mockTask.ExpireTime = time.Now().Unix() + 5*60
	} else {
		mockTask.ExpireTime = time.Now().Unix() - 5*60
	}
	mockKM, err := keys.NewPrivateKeyManager(util.RandHexKey())
	assert.Nil(t, err)
	signature, err := mockKM.Sign(mockTask.GetSignBytes())
	assert.Nil(t, err)
	mockTask.SetSignature(signature)
	msg, err := json.Marshal(mockTask)
	assert.Nil(t, err)
	return hex.EncodeToString(msg)
}

func TestGateModular_migratePieceHandler(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *GateModular
		request      func() *http.Request
		wantedResult string
	}{
		{
			name: "failed to parse migrate piece header",
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
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, MigratePiecePath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdMigratePieceMsgHeader, "48656c6c6f20476f706865722")
				req.Header.Set(GnfdMigrateGVGMsgHeader, makeMockMigrateGVGTaskHeader(t, true))
				return req
			},
			wantedResult: "gnfd msg decoding error",
		},
		{
			name: "failed to unmarshal migrate piece msg",
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
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, MigratePiecePath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdMigratePieceMsgHeader, "48656c6c6f20476f7068657221")
				req.Header.Set(GnfdMigrateGVGMsgHeader, makeMockMigrateGVGTaskHeader(t, true))
				return req
			},
			wantedResult: "gnfd msg decoding error",
		},
		{
			name: "failed to get migrate piece object info due to gvg expire time",
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
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, MigratePiecePath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdMigratePieceMsgHeader, "7b227461736b223a7b7d2c2273746f726167655f706172616d73223a7b2276657273696f6e65645f706172616d73223a7b226d61785f7365676d656e745f73697a65223a31302c22726564756e64616e745f646174615f6368756e6b5f6e756d223a342c22726564756e64616e745f7061726974795f6368756e6b5f6e756d223a327d7d7d")
				req.Header.Set(GnfdMigrateGVGMsgHeader, makeMockMigrateGVGTaskHeader(t, false))
				return req
			},
			wantedResult: "no permission",
		},
		{
			name: "failed to get migrate piece object info due to query sp error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				chainMock := consensus.NewMockConsensus(ctrl)
				chainMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{}, fmt.Errorf("failed to query sp")).Times(1)
				g.spCachePool = NewSPCachePool(chainMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, MigratePiecePath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdMigratePieceMsgHeader, "7b227461736b223a7b7d2c2273746f726167655f706172616d73223a7b2276657273696f6e65645f706172616d73223a7b226d61785f7365676d656e745f73697a65223a31302c22726564756e64616e745f646174615f6368756e6b5f6e756d223a342c22726564756e64616e745f7061726974795f6368756e6b5f6e756d223a327d7d7d")
				req.Header.Set(GnfdMigrateGVGMsgHeader, makeMockMigrateGVGTaskHeader(t, true))
				return req
			},
			wantedResult: "failed to query sp",
		},
		{
			name: "failed to get migrate piece object info due to metadata api error",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().VerifyMigrateGVGPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("failed to query metadata")).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				chainMock := consensus.NewMockConsensus(ctrl)
				chainMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{}, nil).Times(1)
				g.spCachePool = NewSPCachePool(chainMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, MigratePiecePath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdMigratePieceMsgHeader, "7b227461736b223a7b7d2c2273746f726167655f706172616d73223a7b2276657273696f6e65645f706172616d73223a7b226d61785f7365676d656e745f73697a65223a31302c22726564756e64616e745f646174615f6368756e6b5f6e756d223a342c22726564756e64616e745f7061726974795f6368756e6b5f6e756d223a327d7d7d")
				req.Header.Set(GnfdMigrateGVGMsgHeader, makeMockMigrateGVGTaskHeader(t, true))
				return req
			},
			wantedResult: "failed to query metadata",
		},
		{
			name: "failed to get migrate piece object info due to metadata no permission",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				mockEffect := permissiontypes.EFFECT_DENY
				clientMock.EXPECT().VerifyMigrateGVGPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&mockEffect, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				chainMock := consensus.NewMockConsensus(ctrl)
				chainMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{}, nil).Times(1)
				g.spCachePool = NewSPCachePool(chainMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, MigratePiecePath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdMigratePieceMsgHeader, "7b227461736b223a7b7d2c2273746f726167655f706172616d73223a7b2276657273696f6e65645f706172616d73223a7b226d61785f7365676d656e745f73697a65223a31302c22726564756e64616e745f646174615f6368756e6b5f6e756d223a342c22726564756e64616e745f7061726974795f6368756e6b5f6e756d223a327d7d7d")
				req.Header.Set(GnfdMigrateGVGMsgHeader, makeMockMigrateGVGTaskHeader(t, true))
				return req
			},
			wantedResult: "no permission",
		},
		{
			name: "failed to get migrate piece object info due to has no object info",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				mockEffect := permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyMigrateGVGPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&mockEffect, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				chainMock := consensus.NewMockConsensus(ctrl)
				chainMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{}, nil).Times(1)
				g.spCachePool = NewSPCachePool(chainMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, MigratePiecePath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdMigratePieceMsgHeader, "7b227461736b223a7b7d2c2273746f726167655f706172616d73223a7b2276657273696f6e65645f706172616d73223a7b226d61785f7365676d656e745f73697a65223a31302c22726564756e64616e745f646174615f6368756e6b5f6e756d223a342c22726564756e64616e745f7061726974795f6368756e6b5f6e756d223a327d7d7d")
				req.Header.Set(GnfdMigrateGVGMsgHeader, makeMockMigrateGVGTaskHeader(t, true))
				return req
			},
			wantedResult: "invalid request header",
		},
		{
			name: "failed to get object on chain meta",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				mockEffect := permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyMigrateGVGPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&mockEffect, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil,
					mockErr).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				chainMock := consensus.NewMockConsensus(ctrl)
				chainMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{}, nil).Times(1)
				g.spCachePool = NewSPCachePool(chainMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, MigratePiecePath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdMigratePieceMsgHeader, "7b227461736b223a7b7d2c226f626a6563745f696e666f223a7b226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c226964223a2231227d2c2273746f726167655f706172616d73223a7b2276657273696f6e65645f706172616d73223a7b226d61785f7365676d656e745f73697a65223a31302c22726564756e64616e745f646174615f6368756e6b5f6e756d223a342c22726564756e64616e745f7061726974795f6368756e6b5f6e756d223a327d7d7d")
				req.Header.Set(GnfdMigrateGVGMsgHeader, makeMockMigrateGVGTaskHeader(t, true))
				return req
			},
			wantedResult: "invalid request header",
		},
		{
			name: "invalid redundancy index",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				mockEffect := permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyMigrateGVGPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&mockEffect, nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{
					ObjectName: mockObjectName, CreateAt: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketName: mockBucketName}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(&storagetypes.Params{
					MaxPayloadSize: DefaultMaxPayloadSize}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)
				chainMock := consensus.NewMockConsensus(ctrl)
				chainMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{}, nil).Times(1)
				g.spCachePool = NewSPCachePool(chainMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, MigratePiecePath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdMigratePieceMsgHeader, "7b227461736b223a7b7d2c226f626a6563745f696e666f223a7b226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c226964223a2231227d2c2273746f726167655f706172616d73223a7b2276657273696f6e65645f706172616d73223a7b226d61785f7365676d656e745f73697a65223a31302c22726564756e64616e745f646174615f6368756e6b5f6e756d223a342c22726564756e64616e745f7061726974795f6368756e6b5f6e756d223a327d7d2c22726564756e64616e63795f696478223a2d327d")
				req.Header.Set(GnfdMigrateGVGMsgHeader, makeMockMigrateGVGTaskHeader(t, true))
				return req
			},
			wantedResult: "invalid redundancy index",
		},
		{
			name: "redundancy index is -1 and failed to download piece",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				mockEffect := permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyMigrateGVGPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&mockEffect, nil).Times(1)
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{
					ObjectName: mockObjectName, CreateAt: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketName: mockBucketName}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(&storagetypes.Params{
					MaxPayloadSize: DefaultMaxPayloadSize}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				g.baseApp.SetPieceOp(pieceOpMock)
				pieceOpMock.EXPECT().SegmentPieceKey(gomock.Any(), gomock.Any()).Return("test").Times(1)
				pieceOpMock.EXPECT().SegmentPieceSize(gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)

				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{}, nil).Times(1)
				g.spCachePool = NewSPCachePool(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, MigratePiecePath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdMigratePieceMsgHeader, "7b227461736b223a7b7d2c226f626a6563745f696e666f223a7b226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c226964223a2231227d2c2273746f726167655f706172616d73223a7b2276657273696f6e65645f706172616d73223a7b226d61785f7365676d656e745f73697a65223a31302c22726564756e64616e745f646174615f6368756e6b5f6e756d223a342c22726564756e64616e745f7061726974795f6368756e6b5f6e756d223a327d7d2c22726564756e64616e63795f696478223a2d317d")
				req.Header.Set(GnfdMigrateGVGMsgHeader, makeMockMigrateGVGTaskHeader(t, true))
				return req
			},
			wantedResult: "mock error",
		},
		{
			name: "redundancy index is not  -1 and succeed to migrate one piece",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				mockEffect := permissiontypes.EFFECT_ALLOW
				clientMock.EXPECT().VerifyMigrateGVGPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&mockEffect, nil).Times(1)
				clientMock.EXPECT().GetPiece(gomock.Any(), gomock.Any()).Return([]byte("data"), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)

				consensusMock := consensus.NewMockConsensus(ctrl)
				consensusMock.EXPECT().QueryObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.ObjectInfo{
					ObjectName: mockObjectName, CreateAt: 1}, nil).Times(1)
				consensusMock.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{
					BucketName: mockBucketName}, nil).Times(1)
				consensusMock.EXPECT().QueryStorageParamsByTimestamp(gomock.Any(), gomock.Any()).Return(&storagetypes.Params{
					MaxPayloadSize: DefaultMaxPayloadSize}, nil).Times(1)
				g.baseApp.SetConsensus(consensusMock)

				pieceOpMock := piecestore.NewMockPieceOp(ctrl)
				g.baseApp.SetPieceOp(pieceOpMock)
				pieceOpMock.EXPECT().ECPieceKey(gomock.Any(), gomock.Any(), gomock.Any()).Return("test").Times(1)
				pieceOpMock.EXPECT().ECPieceSize(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1)).Times(1)

				consensusMock.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{}, nil).Times(1)
				g.spCachePool = NewSPCachePool(consensusMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, MigratePiecePath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdMigratePieceMsgHeader, "7b227461736b223a7b7d2c226f626a6563745f696e666f223a7b226f626a6563745f6e616d65223a226d6f636b2d6f626a6563742d6e616d65222c226964223a2231227d2c2273746f726167655f706172616d73223a7b2276657273696f6e65645f706172616d73223a7b226d61785f7365676d656e745f73697a65223a31302c22726564756e64616e745f646174615f6368756e6b5f6e756d223a342c22726564756e64616e745f7061726974795f6368756e6b5f6e756d223a327d7d2c22726564756e64616e63795f696478223a317d")
				req.Header.Set(GnfdMigrateGVGMsgHeader, makeMockMigrateGVGTaskHeader(t, true))
				return req
			},
			wantedResult: "",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockMigratePieceHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Contains(t, w.Body.String(), tt.wantedResult)
		})
	}
}

func mockGetSecondaryBlsMigrationBucketApprovalHandlerRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path(SecondarySPMigrationBucketApprovalPath).Name(migrationBucketApprovalName).Methods(http.MethodGet).
		HandlerFunc(g.getSecondaryBlsMigrationBucketApprovalHandler)
	return router
}

func TestGateModular_getSecondaryBlsMigrationBucketApprovalHandler(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *GateModular
		request      func() *http.Request
		wantedResult string
	}{
		{
			name: "failed to parse secondary migration bucket approval header",
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
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, SecondarySPMigrationBucketApprovalPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdSecondarySPMigrationBucketMsgHeader, "48656c6c6f20476f706865722")
				return req
			},
			wantedResult: "gnfd msg decoding error",
		},
		{
			name: "failed to unmarshal migration bucket approval msg",
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
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, SecondarySPMigrationBucketApprovalPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdSecondarySPMigrationBucketMsgHeader, "48656c6c6f20476f7068657221")
				return req
			},
			wantedResult: "gnfd msg decoding error",
		},
		{
			name: "failed to sign secondary sp migration bucket",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().SignSecondarySPMigrationBucket(gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, SecondarySPMigrationBucketApprovalPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdSecondarySPMigrationBucketMsgHeader, "7b22636861696e5f6964223a2231222c226473745f7072696d6172795f73705f6964223a312c227372635f676c6f62616c5f7669727475616c5f67726f75705f6964223a322c226473745f676c6f62616c5f7669727475616c5f67726f75705f6964223a332c226275636b65745f6964223a2231227d")
				return req
			},
			wantedResult: "failed to sign secondary sp migration bucket",
		},
		{
			name: "success",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().SignSecondarySPMigrationBucket(gomock.Any(), gomock.Any()).Return([]byte("mockSig"), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, SecondarySPMigrationBucketApprovalPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdSecondarySPMigrationBucketMsgHeader, "7b22636861696e5f6964223a2231222c226473745f7072696d6172795f73705f6964223a312c227372635f676c6f62616c5f7669727475616c5f67726f75705f6964223a322c226473745f676c6f62616c5f7669727475616c5f67726f75705f6964223a332c226275636b65745f6964223a2231227d")
				return req
			},
			wantedResult: "",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetSecondaryBlsMigrationBucketApprovalHandlerRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Contains(t, w.Body.String(), tt.wantedResult)
		})
	}
}

func mockGetSwapOutApprovalRoute(t *testing.T, g *GateModular) *mux.Router {
	t.Helper()
	router := mux.NewRouter().SkipClean(true)
	router.Path(SwapOutApprovalPath).Name(swapOutApprovalName).Methods(http.MethodGet).HandlerFunc(g.getSwapOutApproval)
	return router
}

func TestGateModular_getSwapOutApproval(t *testing.T) {
	cases := []struct {
		name         string
		fn           func() *GateModular
		request      func() *http.Request
		wantedResult string
	}{
		{
			name: "failed to parse swap out approval header",
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
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, SwapOutApprovalPath)
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
			name: "failed to unmarshal swap out approval msg",
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
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, SwapOutApprovalPath)
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
			name: "failed to basic check approval msg",
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
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, SwapOutApprovalPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2273746f726167655f70726f7669646572223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364636222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a322c22676c6f62616c5f7669727475616c5f67726f75705f696473223a5b5d2c22737563636573736f725f73705f6964223a302c22737563636573736f725f73705f617070726f76616c223a6e756c6c7d")
				return req
			},
			wantedResult: "gnfd msg validate error",
		},
		{
			name: "failed to sign swap out",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().SignSwapOut(gomock.Any(), gomock.Any()).Return(nil, mockErr)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, SwapOutApprovalPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2273746f726167655f70726f7669646572223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364636222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a322c22676c6f62616c5f7669727475616c5f67726f75705f696473223a5b5d2c22737563636573736f725f73705f6964223a312c22737563636573736f725f73705f617070726f76616c223a6e756c6c7d")
				return req
			},
			wantedResult: "failed to sign swap out",
		},
		{
			name: "success",
			fn: func() *GateModular {
				g := setup(t)
				ctrl := gomock.NewController(t)
				clientMock := gfspclient.NewMockGfSpClientAPI(ctrl)
				clientMock.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(false, nil).Times(1)
				clientMock.EXPECT().SignSwapOut(gomock.Any(), gomock.Any()).Return([]byte("mockSig"), nil).Times(1)
				g.baseApp.SetGfSpClient(clientMock)
				return g
			},
			request: func() *http.Request {
				path := fmt.Sprintf("%s%s%s", scheme, testDomain, SwapOutApprovalPath)
				req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
				validExpiryDateStr := time.Now().Add(time.Hour * 60).Format(ExpiryDateFormat)
				req.Header.Set(commonhttp.HTTPHeaderExpiryTimestamp, validExpiryDateStr)
				req.Header.Set(GnfdAuthorizationHeader, "GNFD1-EDDSA,Signature=48656c6c6f20476f7068657221")
				req.Header.Set(GnfdUnsignedApprovalMsgHeader, "7b2273746f726167655f70726f7669646572223a22307831433743384136363865323361454432393166373866433266336231383635416363383762364636222c22676c6f62616c5f7669727475616c5f67726f75705f66616d696c795f6964223a322c22676c6f62616c5f7669727475616c5f67726f75705f696473223a5b5d2c22737563636573736f725f73705f6964223a312c22737563636573736f725f73705f617070726f76616c223a6e756c6c7d")
				return req
			},
			wantedResult: "",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := mockGetSwapOutApprovalRoute(t, tt.fn())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request())
			assert.Contains(t, w.Body.String(), tt.wantedResult)
		})
	}
}
