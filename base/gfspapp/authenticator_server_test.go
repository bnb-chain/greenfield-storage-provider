package gfspapp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
)

func TestGfSpBaseApp_GfSpVerifyAuthenticationSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(true, nil).Times(1)
	req := &gfspserver.GfSpAuthenticationRequest{
		AuthType:    1,
		UserAccount: "mockUserAccount",
		BucketName:  "mockBucketName",
		ObjectName:  "mockObjectName",
	}
	result, err := g.GfSpVerifyAuthentication(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, true, result.GetAllowed())
}

func TestGfSpBaseApp_GfSpVerifyAuthenticationFailure(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().VerifyAuthentication(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(false, mockErr).Times(1)
	req := &gfspserver.GfSpAuthenticationRequest{
		AuthType:    1,
		UserAccount: "mockUserAccount",
		BucketName:  "mockBucketName",
		ObjectName:  "mockObjectName",
	}
	result, err := g.GfSpVerifyAuthentication(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
	assert.Equal(t, false, result.GetAllowed())
}

func TestGfSpBaseApp_GetAuthNonceSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return(&corespdb.OffChainAuthKey{CurrentNonce: 1},
		nil).Times(1)
	req := &gfspserver.GetAuthNonceRequest{
		AccountId: "mockAccountID",
		Domain:    "mockDomain",
	}
	result, err := g.GetAuthNonce(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, int32(1), result.GetCurrentNonce())
}

func TestGfSpBaseApp_GetAuthNonceFailure(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().GetAuthNonce(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, mockErr).Times(1)
	req := &gfspserver.GetAuthNonceRequest{
		AccountId: "mockAccountID",
		Domain:    "mockDomain",
	}
	result, err := g.GetAuthNonce(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, mockErr.Error(), result.GetErr().GetDescription())
}

func TestGfSpBaseApp_UpdateUserPublicKeySuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().UpdateUserPublicKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any()).Return(true, nil).Times(1)
	req := &gfspserver.UpdateUserPublicKeyRequest{
		AccountId: "mockAccountID",
		Domain:    "mockDomain",
	}
	result, err := g.UpdateUserPublicKey(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, true, result.GetResult())
}

func TestGfSpBaseApp_VerifyGNFD1EddsaSignatureSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().VerifyGNFD1EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		true, nil).Times(1)
	req := &gfspserver.VerifyGNFD1EddsaSignatureRequest{
		AccountId: "mockAccountID",
		Domain:    "mockDomain",
	}
	result, err := g.VerifyGNFD1EddsaSignature(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, true, result.GetResult())
}
