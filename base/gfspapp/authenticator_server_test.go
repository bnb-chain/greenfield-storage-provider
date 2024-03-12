package gfspapp

import (
	"context"
	"testing"
	"time"

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

func TestGfSpBaseApp_GetAuthKeyV2Success(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().GetAuthKeyV2(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&corespdb.OffChainAuthKeyV2{PublicKey: "test"}, nil).Times(1)
	req := &gfspserver.GetAuthKeyV2Request{
		AccountId:     "mockAccountID",
		Domain:        "mockDomain",
		UserPublicKey: "test",
	}
	result, err := g.GetAuthKeyV2(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, "test", result.PublicKey)
}

func TestGfSpBaseApp_GetAuthKeyV2Failure1(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().GetAuthKeyV2(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&corespdb.OffChainAuthKeyV2{PublicKey: "test"}, nil).Times(0)

	result, err := g.GetAuthKeyV2(context.TODO(), nil)
	assert.Nil(t, err)
	assert.Equal(t, ErrAuthenticatorTaskDangling, result.Err)
}

func TestGfSpBaseApp_GetAuthKeyV2Failure2(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().GetAuthKeyV2(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		nil, mockErr).Times(1)
	req := &gfspserver.GetAuthKeyV2Request{
		AccountId:     "mockAccountID",
		Domain:        "mockDomain",
		UserPublicKey: "test",
	}
	result, err := g.GetAuthKeyV2(context.TODO(), req)
	assert.Nil(t, err)
	assert.NotNil(t, result.GetErr())
}

func TestGfSpBaseApp_GetAuthKeyV2Failure3(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().GetAuthKeyV2(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		nil, mockErr).Times(1)
	req := &gfspserver.GetAuthKeyV2Request{
		AccountId:     "mockAccountID",
		Domain:        "mockDomain",
		UserPublicKey: "test",
	}
	result, err := g.GetAuthKeyV2(context.TODO(), req)
	assert.Nil(t, err)
	assert.NotNil(t, result.GetErr())
}

func TestGfSpBaseApp_UpdateUserPublicKeyV2Success(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().UpdateUserPublicKeyV2(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		true, nil).Times(1)
	req := &gfspserver.UpdateUserPublicKeyV2Request{
		AccountId:     "mockAccountID",
		Domain:        "mockDomain",
		UserPublicKey: "test",
		ExpiryDate:    time.Now().Add(7 * time.Hour).UnixMilli(),
	}
	result, err := g.UpdateUserPublicKeyV2(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, true, result.Result)
}

func TestGfSpBaseApp_UpdateUserPublicKeyV2Failure(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().UpdateUserPublicKeyV2(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		true, nil).Times(0)

	result, err := g.UpdateUserPublicKeyV2(context.TODO(), nil)
	assert.Nil(t, err)
	assert.Equal(t, ErrAuthenticatorTaskDangling, result.Err)
}

func TestGfSpBaseApp_UpdateUserPublicKeyV2Failure2(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().UpdateUserPublicKeyV2(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		false, mockErr).Times(1)
	req := &gfspserver.UpdateUserPublicKeyV2Request{
		AccountId:     "mockAccountID",
		Domain:        "mockDomain",
		UserPublicKey: "test",
		ExpiryDate:    time.Now().Add(7 * time.Hour).UnixMilli(),
	}
	result, err := g.UpdateUserPublicKeyV2(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, false, result.Result)
	assert.Equal(t, "mock error", result.Err.Description)
}

func TestGfSpBaseApp_VerifyGNFD2EddsaSignatureSuccess(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().VerifyGNFD2EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		true, nil).Times(1)
	req := &gfspserver.VerifyGNFD2EddsaSignatureRequest{
		AccountId:     "mockAccountID",
		Domain:        "mockDomain",
		UserPublicKey: "test",
	}
	result, err := g.VerifyGNFD2EddsaSignature(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, true, result.Result)
}

func TestGfSpBaseApp_VerifyGNFD2EddsaSignatureFailure(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().VerifyGNFD2EddsaSignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		true, nil).Times(0)

	result, err := g.VerifyGNFD2EddsaSignature(context.TODO(), nil)
	assert.Nil(t, err)
	assert.Equal(t, ErrAuthenticatorTaskDangling, result.Err)
}

func TestGfSpBaseApp_ListAuthKeysV2Success(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().ListAuthKeysV2(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		[]string{"key1", "key2"}, nil).Times(1)
	req := &gfspserver.ListAuthKeysV2Request{
		AccountId: "mockAccountID",
		Domain:    "mockDomain",
	}
	result, err := g.ListAuthKeysV2(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, []string{"key1", "key2"}, result.PublicKeys)
}

func TestGfSpBaseApp_ListAuthKeysV2Failure1(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().ListAuthKeysV2(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		[]string{"key1", "key2"}, nil).Times(0)

	result, err := g.ListAuthKeysV2(context.TODO(), nil)
	assert.Nil(t, err)
	assert.Equal(t, ErrAuthenticatorTaskDangling, result.Err)
}

func TestGfSpBaseApp_ListAuthKeysV2Failure2(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().ListAuthKeysV2(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		nil, mockErr).Times(1)
	req := &gfspserver.ListAuthKeysV2Request{
		AccountId: "mockAccountID",
		Domain:    "mockDomain",
	}
	result, err := g.ListAuthKeysV2(context.TODO(), req)
	assert.Nil(t, err)
	assert.NotNil(t, result.GetErr())
}

func TestGfSpBaseApp_ListAuthKeysV2Failure3(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().ListAuthKeysV2(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		nil, mockErr).Times(1)
	req := &gfspserver.ListAuthKeysV2Request{
		AccountId: "mockAccountID",
		Domain:    "mockDomain",
	}
	result, err := g.ListAuthKeysV2(context.TODO(), req)
	assert.Nil(t, err)
	assert.NotNil(t, result.GetErr())
}

func TestGfSpBaseApp_DeleteAuthKeysV2Success(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().DeleteAuthKeysV2(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		true, nil).Times(1)
	req := &gfspserver.DeleteAuthKeysV2Request{
		AccountId:  "mockAccountID",
		Domain:     "mockDomain",
		PublicKeys: []string{"key1", "key2"},
	}
	result, err := g.DeleteAuthKeysV2(context.TODO(), req)
	assert.Nil(t, err)
	assert.Equal(t, true, result.Result)
}

func TestGfSpBaseApp_DeleteAuthKeysV2Failure(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().DeleteAuthKeysV2(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		true, nil).Times(0)
	result, err := g.DeleteAuthKeysV2(context.TODO(), nil)
	assert.Nil(t, err)
	assert.Equal(t, false, result.Result)
}

func TestGfSpBaseApp_DeleteAuthKeysV2Failure2(t *testing.T) {
	g := setup(t)
	ctrl := gomock.NewController(t)
	m := module.NewMockAuthenticator(ctrl)
	g.authenticator = m
	m.EXPECT().DeleteAuthKeysV2(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
		false, mockErr).Times(1)
	req := &gfspserver.DeleteAuthKeysV2Request{
		AccountId:  "mockAccountID",
		Domain:     "mockDomain",
		PublicKeys: []string{"key1", "key2"},
	}
	result, err := g.DeleteAuthKeysV2(context.TODO(), req)
	assert.Nil(t, err)
	assert.NotNil(t, result.Err)
}
