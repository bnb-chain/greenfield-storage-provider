package gater

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield/sdk/keys"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

func makeMockSPOperatorAuthHeader(t *testing.T, payload []byte, expired bool) string {
	t.Helper()
	bindingHash := binary.BigEndian.Uint64(crypto.Keccak256(payload)[:8])
	authMsg := &gfsptask.GfSpBucketMigrationInfo{
		BucketId: bindingHash,
	}
	if expired {
		authMsg.ExpireTime = time.Now().Unix() - 300
	} else {
		authMsg.ExpireTime = time.Now().Unix() + 3600
	}
	mockKM, err := keys.NewPrivateKeyManager(util.RandHexKey())
	assert.Nil(t, err)
	signature, err := mockKM.Sign(authMsg.GetSignBytes())
	assert.Nil(t, err)
	authMsg.SetSignature(signature)
	msg, err := json.Marshal(authMsg)
	assert.Nil(t, err)
	return hex.EncodeToString(msg)
}

func setupGateWithSPCachePool(t *testing.T) *GateModular {
	t.Helper()
	g := setup(t)
	ctrl := gomock.NewController(t)
	mockConsensus := consensus.NewMockConsensus(ctrl)
	mockConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
		Id:              1,
		OperatorAddress: "mock-operator-addr",
	}, nil).AnyTimes()
	g.spCachePool = NewSPCachePool(mockConsensus)
	return g
}

func TestPayloadBindingHash(t *testing.T) {
	payload1 := []byte("test-payload-1")
	payload2 := []byte("test-payload-2")

	hash1 := payloadBindingHash(payload1)
	hash1Again := payloadBindingHash(payload1)
	hash2 := payloadBindingHash(payload2)

	assert.Equal(t, hash1, hash1Again)
	assert.NotEqual(t, hash1, hash2)
}

func TestParseAndVerifySPAuth_MissingHeader(t *testing.T) {
	g := setupGateWithSPCachePool(t)
	path := fmt.Sprintf("%s%s/test", scheme, testDomain)
	req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
	reqCtx, _ := NewRequestContext(req, g)

	_, err := g.ParseAndVerifySPAuth(reqCtx, req, []byte("payload"))
	assert.NotNil(t, err)
}

func TestParseAndVerifySPAuth_InvalidHex(t *testing.T) {
	g := setupGateWithSPCachePool(t)
	path := fmt.Sprintf("%s%s/test", scheme, testDomain)
	req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
	req.Header.Set(GnfdSPOperatorAuthHeader, "zzzz-not-hex")
	reqCtx, _ := NewRequestContext(req, g)

	_, err := g.ParseAndVerifySPAuth(reqCtx, req, []byte("payload"))
	assert.NotNil(t, err)
}

func TestParseAndVerifySPAuth_Expired(t *testing.T) {
	payload := []byte("test-payload")
	g := setupGateWithSPCachePool(t)
	path := fmt.Sprintf("%s%s/test", scheme, testDomain)
	req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
	req.Header.Set(GnfdSPOperatorAuthHeader, makeMockSPOperatorAuthHeader(t, payload, true))
	reqCtx, _ := NewRequestContext(req, g)

	_, err := g.ParseAndVerifySPAuth(reqCtx, req, payload)
	assert.NotNil(t, err)
}

func TestParseAndVerifySPAuth_PayloadBindingMismatch(t *testing.T) {
	payload := []byte("test-payload")
	differentPayload := []byte("different-payload")
	g := setupGateWithSPCachePool(t)
	path := fmt.Sprintf("%s%s/test", scheme, testDomain)
	req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
	req.Header.Set(GnfdSPOperatorAuthHeader, makeMockSPOperatorAuthHeader(t, payload, false))
	reqCtx, _ := NewRequestContext(req, g)

	_, err := g.ParseAndVerifySPAuth(reqCtx, req, differentPayload)
	assert.NotNil(t, err)
}

func TestParseAndVerifySPAuth_UnregisteredSP(t *testing.T) {
	payload := []byte("test-payload")
	g := setup(t)
	ctrl := gomock.NewController(t)
	mockConsensus := consensus.NewMockConsensus(ctrl)
	mockConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).
		Return(nil, fmt.Errorf("storage provider not found")).AnyTimes()
	g.spCachePool = NewSPCachePool(mockConsensus)

	path := fmt.Sprintf("%s%s/test", scheme, testDomain)
	req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
	req.Header.Set(GnfdSPOperatorAuthHeader, makeMockSPOperatorAuthHeader(t, payload, false))
	reqCtx, _ := NewRequestContext(req, g)

	sp, err := g.ParseAndVerifySPAuth(reqCtx, req, payload)
	assert.NotNil(t, err)
	assert.Nil(t, sp)
	assert.Contains(t, err.Error(), "storage provider not found")
}

func TestParseAndVerifySPAuth_TamperedSignature(t *testing.T) {
	payload := []byte("test-payload")
	g := setupGateWithSPCachePool(t)

	bindingHash := binary.BigEndian.Uint64(crypto.Keccak256(payload)[:8])
	authMsg := &gfsptask.GfSpBucketMigrationInfo{
		BucketId:   bindingHash,
		ExpireTime: time.Now().Unix() + 3600,
	}
	authMsg.SetSignature([]byte("too-short-invalid-signature"))
	msg, err := json.Marshal(authMsg)
	assert.Nil(t, err)

	path := fmt.Sprintf("%s%s/test", scheme, testDomain)
	req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
	req.Header.Set(GnfdSPOperatorAuthHeader, hex.EncodeToString(msg))
	reqCtx, _ := NewRequestContext(req, g)

	sp, err := g.ParseAndVerifySPAuth(reqCtx, req, payload)
	assert.NotNil(t, err)
	assert.Nil(t, sp)
}

func TestParseAndVerifySPAuth_Success(t *testing.T) {
	payload := []byte("test-payload")
	g := setupGateWithSPCachePool(t)
	path := fmt.Sprintf("%s%s/test", scheme, testDomain)
	req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
	req.Header.Set(GnfdSPOperatorAuthHeader, makeMockSPOperatorAuthHeader(t, payload, false))
	reqCtx, _ := NewRequestContext(req, g)

	sp, err := g.ParseAndVerifySPAuth(reqCtx, req, payload)
	assert.Nil(t, err)
	assert.NotNil(t, sp)
}
