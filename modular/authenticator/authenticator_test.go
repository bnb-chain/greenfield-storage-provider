package authenticator

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	permissiontypes "github.com/bnb-chain/greenfield/x/permission/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var mockErr = errors.New("mock error")
var TestSpAddress = "TestSpAddress"
var TestUnsignedMsg = "I want to get approval from sp before creating the bucket."

func setup(t *testing.T) *AuthenticationModular {
	return &AuthenticationModular{baseApp: &gfspapp.GfSpBaseApp{}}
}

func TestAuthModular_Name(t *testing.T) {
	a := setup(t)
	result := a.Name()
	assert.Equal(t, module.AuthenticationModularName, result)
}

func TestAuthModular_StartSuccess(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m1 := rcmgr.NewMockResourceManager(ctrl)
	a.baseApp.SetResourceManager(m1)
	m2 := rcmgr.NewMockResourceScope(ctrl)
	m1.EXPECT().OpenService(gomock.Any()).DoAndReturn(func(svc string) (rcmgr.ResourceScope, error) {
		return m2, nil
	})
	err := a.Start(context.TODO())
	assert.Nil(t, err)
}

func TestAuthModular_Stop(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	a.scope = m
	m.EXPECT().Release().AnyTimes()
	err := a.Stop(context.TODO())
	assert.Nil(t, err)
}

func TestAuthModular_ReserveResourceSuccess(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	a.scope = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().BeginSpan().DoAndReturn(func() (rcmgr.ResourceScopeSpan, error) {
		return m1, nil
	})
	m1.EXPECT().ReserveResources(gomock.Any()).DoAndReturn(func(st *rcmgr.ScopeStat) error { return nil }).AnyTimes()
	result, err := a.ReserveResource(context.TODO(), &rcmgr.ScopeStat{Memory: 1})
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestAuthModular_ReserveResourceFailure1(t *testing.T) {
	t.Log("Failure case description: mock BeginSpan returns error")
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	a.scope = m
	m.EXPECT().BeginSpan().DoAndReturn(func() (rcmgr.ResourceScopeSpan, error) {
		return nil, mockErr
	})
	result, err := a.ReserveResource(context.TODO(), &rcmgr.ScopeStat{Memory: 1})
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestAuthModular_ReserveResourceFailure2(t *testing.T) {
	t.Log("Failure case description: mock ReserveResources returns error")
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScope(ctrl)
	a.scope = m
	m1 := rcmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().BeginSpan().DoAndReturn(func() (rcmgr.ResourceScopeSpan, error) {
		return m1, nil
	})
	m1.EXPECT().ReserveResources(gomock.Any()).DoAndReturn(func(st *rcmgr.ScopeStat) error { return mockErr }).AnyTimes()
	result, err := a.ReserveResource(context.TODO(), &rcmgr.ScopeStat{Memory: 1})
	assert.Equal(t, mockErr, err)
	assert.Nil(t, result)
}

func TestAuthModular_ReleaseResource(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := rcmgr.NewMockResourceScopeSpan(ctrl)
	m.EXPECT().Done().AnyTimes()
	a.ReleaseResource(context.TODO(), m)
}

func TestAuthModular_GetAuthNonce(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	m.EXPECT().GetAuthKey(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string) (*spdb.OffChainAuthKey, error) { return nil, nil },
	).Times(1)
	a.baseApp.SetGfSpDB(m)
	_, err := a.GetAuthNonce(context.Background(), "test_account", "https://domain.com")
	assert.Nil(t, err)
}

func TestAuthModular_GetAuthNonce_Failure(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	m.EXPECT().GetAuthKey(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string) (*spdb.OffChainAuthKey, error) { return nil, gorm.ErrInvalidDB },
	).Times(1)
	a.baseApp.SetGfSpDB(m)
	_, err := a.GetAuthNonce(context.Background(), "test_account", "https://domain.com")
	assert.NotNil(t, err)
}

func TestAuthModular_UpdateUserPublicKey(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	m.EXPECT().UpdateAuthKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	a.baseApp.SetGfSpDB(m)
	_, err := a.UpdateUserPublicKey(context.Background(), "test_account", "https://domain.com", 1, 2, "user_test_public_key", time.Now().UnixMilli())
	assert.Nil(t, err)
}

func TestAuthModular_UpdateUserPublicKey_Failure(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	m.EXPECT().UpdateAuthKey(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(gorm.ErrInvalidDB).Times(1)
	a.baseApp.SetGfSpDB(m)
	_, err := a.UpdateUserPublicKey(context.Background(), "test_account", "https://domain.com", 1, 2, "user_test_public_key", time.Now().UnixMilli())
	assert.NotNil(t, err)
}

func TestAuthModular_VerifyGNFD1EddsaSignature(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	privateKey, _ := crypto.GenerateKey()
	userEddsaSeed := "test_seed"
	// get the EDDSA private and public key
	userEddsaPrivateKey, _ := GenerateEddsaPrivateKey(userEddsaSeed)

	sig, _ := userEddsaPrivateKey.Sign([]byte(TestUnsignedMsg), mimc.NewMiMC())
	userEddsaPublicKeyStr := GetEddsaCompressedPublicKey(userEddsaSeed)

	userAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	userDomain := "testDomain"
	mockedData := &spdb.OffChainAuthKey{
		UserAddress:      userAddress,
		Domain:           userDomain,
		CurrentNonce:     1,
		CurrentPublicKey: userEddsaPublicKeyStr,
		NextNonce:        2,
		ExpiryDate:       time.Now().Add(7 * time.Hour),
		CreatedTime:      time.Now(),
		ModifiedTime:     time.Now(),
	}
	m.EXPECT().GetAuthKey(gomock.Any(), gomock.Any()).Return(mockedData, nil).Times(1)
	a.baseApp.SetGfSpDB(m)
	verifyResult, err := a.VerifyGNFD1EddsaSignature(context.Background(), userAddress, userDomain, hex.EncodeToString(sig), []byte(TestUnsignedMsg))
	assert.True(t, verifyResult)
	assert.Nil(t, err)
}

func TestAuthModular_VerifyGNFD1EddsaSignature_BadSig(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	privateKey, _ := crypto.GenerateKey()
	userEddsaSeed := "test_seed"
	// get the EDDSA private and public key
	userEddsaPrivateKey, _ := GenerateEddsaPrivateKey(userEddsaSeed)

	_, _ = userEddsaPrivateKey.Sign([]byte(TestUnsignedMsg), mimc.NewMiMC())
	userEddsaPublicKeyStr := GetEddsaCompressedPublicKey(userEddsaSeed)

	userAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	userDomain := "testDomain"
	mockedData := &spdb.OffChainAuthKey{
		UserAddress:      userAddress,
		Domain:           userDomain,
		CurrentNonce:     1,
		CurrentPublicKey: userEddsaPublicKeyStr,
		NextNonce:        2,
		ExpiryDate:       time.Now().Add(7 * time.Hour),
		CreatedTime:      time.Now(),
		ModifiedTime:     time.Now(),
	}
	m.EXPECT().GetAuthKey(gomock.Any(), gomock.Any()).Return(mockedData, nil).Times(0)
	a.baseApp.SetGfSpDB(m)
	verifyResult, err := a.VerifyGNFD1EddsaSignature(context.Background(), userAddress, userDomain, "badSignature", []byte(TestUnsignedMsg))
	assert.False(t, verifyResult)
	assert.NotNil(t, err)
}

func TestAuthModular_VerifyGNFD1EddsaSignature_GetNonceError(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	privateKey, _ := crypto.GenerateKey()
	userEddsaSeed := "test_seed"
	// get the EDDSA private and public key
	userEddsaPrivateKey, _ := GenerateEddsaPrivateKey(userEddsaSeed)

	sig, _ := userEddsaPrivateKey.Sign([]byte(TestUnsignedMsg), mimc.NewMiMC())

	userAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	userDomain := "testDomain"

	m.EXPECT().GetAuthKey(gomock.Any(), gomock.Any()).Return(nil, gorm.ErrInvalidDB).Times(1)
	a.baseApp.SetGfSpDB(m)
	verifyResult, err := a.VerifyGNFD1EddsaSignature(context.Background(), userAddress, userDomain, hex.EncodeToString(sig), []byte(TestUnsignedMsg))
	assert.False(t, verifyResult)
	assert.NotNil(t, err)
}

func TestAuthModular_VerifyGNFD1EddsaSignature_NonceExpired(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	privateKey, _ := crypto.GenerateKey()
	userEddsaSeed := "test_seed"
	// get the EDDSA private and public key
	userEddsaPrivateKey, _ := GenerateEddsaPrivateKey(userEddsaSeed)

	sig, _ := userEddsaPrivateKey.Sign([]byte(TestUnsignedMsg), mimc.NewMiMC())
	userEddsaPublicKeyStr := GetEddsaCompressedPublicKey(userEddsaSeed)

	userAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	userDomain := "testDomain"
	mockedData := &spdb.OffChainAuthKey{
		UserAddress:      userAddress,
		Domain:           userDomain,
		CurrentNonce:     1,
		CurrentPublicKey: userEddsaPublicKeyStr,
		NextNonce:        2,
		ExpiryDate:       time.Now(),
		CreatedTime:      time.Now(),
		ModifiedTime:     time.Now(),
	}
	m.EXPECT().GetAuthKey(gomock.Any(), gomock.Any()).Return(mockedData, nil).Times(1)
	a.baseApp.SetGfSpDB(m)
	verifyResult, err := a.VerifyGNFD1EddsaSignature(context.Background(), userAddress, userDomain, hex.EncodeToString(sig), []byte(TestUnsignedMsg))

	assert.False(t, verifyResult)
	assert.NotNil(t, err)
}

func TestAuthModular_VerifyGNFD1EddsaSignature_SigVerificationFailure(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	privateKey, _ := crypto.GenerateKey()
	userEddsaSeed := "test_seed"
	// get the EDDSA private and public key
	userEddsaPrivateKey, _ := GenerateEddsaPrivateKey(userEddsaSeed)

	sig, _ := userEddsaPrivateKey.Sign([]byte(TestUnsignedMsg), mimc.NewMiMC())
	userEddsaPublicKeyStr := GetEddsaCompressedPublicKey(userEddsaSeed)

	userAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	userDomain := "testDomain"
	mockedData := &spdb.OffChainAuthKey{
		UserAddress:      userAddress,
		Domain:           userDomain,
		CurrentNonce:     1,
		CurrentPublicKey: userEddsaPublicKeyStr,
		NextNonce:        2,
		ExpiryDate:       time.Now().Add(7 * time.Hour),
		CreatedTime:      time.Now(),
		ModifiedTime:     time.Now(),
	}
	m.EXPECT().GetAuthKey(gomock.Any(), gomock.Any()).Return(mockedData, nil).Times(1)
	a.baseApp.SetGfSpDB(m)
	verifyResult, err := a.VerifyGNFD1EddsaSignature(context.Background(), userAddress, userDomain, hex.EncodeToString(sig), []byte("wrong message"))

	assert.False(t, verifyResult)
	assert.NotNil(t, err)
}

func Test_VerifyAuth_AskCreateBucketApproval(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	a.baseApp.SetGfSpDB(m)

	_, err := a.VerifyAuthentication(context.Background(), coremodule.AuthOpAskCreateBucketApproval, "test_account", "test_bucket", "test_object")
	assert.Equal(t, ErrInvalidAddress, err)

	privateKey, _ := crypto.GenerateKey()
	userAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// bucket exists
	mockedConsensus := consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err = a.VerifyAuthentication(context.Background(), coremodule.AuthOpAskCreateBucketApproval, userAddress, "test_bucket", "test_object")
	assert.Equal(t, ErrRepeatedBucket, err)

	// bucket does not exist
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	result, _ := a.VerifyAuthentication(context.Background(), coremodule.AuthOpAskCreateBucketApproval, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, result)
}

func Test_VerifyAuth_MigrateBucketApproval(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	a.baseApp.SetGfSpDB(m)

	_, err := a.VerifyAuthentication(context.Background(), coremodule.AuthOpAskMigrateBucketApproval, "test_account", "test_bucket", "test_object")
	assert.Equal(t, ErrInvalidAddress, err)

	privateKey, _ := crypto.GenerateKey()
	userAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// bucket does not exist
	mockedConsensus := consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err = a.VerifyAuthentication(context.Background(), coremodule.AuthOpAskMigrateBucketApproval, userAddress, "test_bucket", "test_object")
	assert.Equal(t, ErrNoSuchBucket, err)

	// bucket exists
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	result, _ := a.VerifyAuthentication(context.Background(), coremodule.AuthOpAskMigrateBucketApproval, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, result)
}

func Test_VerifyAuth_AskCreateObjectApproval(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	a.baseApp.SetGfSpDB(m)

	privateKey, _ := crypto.GenerateKey()
	userAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// bucket does not exist
	mockedConsensus := consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err := a.VerifyAuthentication(context.Background(), coremodule.AuthOpAskCreateObjectApproval, userAddress, "test_bucket", "test_object")
	assert.Equal(t, ErrNoSuchBucket, err)

	// object exists
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{}, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err = a.VerifyAuthentication(context.Background(), coremodule.AuthOpAskCreateObjectApproval, userAddress, "test_bucket", "test_object")
	assert.Equal(t, ErrRepeatedObject, err)

	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, nil, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	result, _ := a.VerifyAuthentication(context.Background(), coremodule.AuthOpAskCreateObjectApproval, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, result)
}

func VerifyObjectAndBucketAndSPID(t *testing.T, authType coremodule.AuthOpType) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	a.baseApp.SetGfSpDB(m)

	privateKey, _ := crypto.GenerateKey()
	userAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// bucket does not exist
	mockedConsensus := consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, errors.New("error: No such bucket")).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err := a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, ErrNoSuchBucket, err)

	// object does not exist
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, errors.New("error: No such object")).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err = a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, ErrNoSuchObject, err)

	// other error when QueryBucketInfoAndObjectInfo
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, errors.New("other error")).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err = a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, strings.Contains(err.Error(), "failed to get bucket and object info from consensus"))

	a.baseApp.SetOperatorAddress(TestSpAddress)

	// object exists but get SP ID returns error
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{}, nil).Times(1)

	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err = a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, strings.Contains(err.Error(), "getSPID error"))

	// QueryVirtualGroupFamily returns error
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
		Id: 1,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err = a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, strings.Contains(err.Error(), "GetBucketPrimarySPID error"))

	// bucketSPID != spID
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
		Id: 1,
	}, nil).Times(0) // the SPID query result is cached already in above tests
	mockedConsensus.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtualgrouptypes.GlobalVirtualGroupFamily{
		PrimarySpId: 2,
	}, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err = a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, ErrMismatchSp, err)
}

func Test_VerifyAuth_PutObject(t *testing.T) {
	authType := coremodule.AuthOpTypePutObject
	VerifyObjectAndBucketAndSPID(t, authType)
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	a.baseApp.SetGfSpDB(m)

	privateKey, _ := crypto.GenerateKey()
	userAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	mockedConsensus := consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{
		ObjectName:   "test_object",
		ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
		Id: 1,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtualgrouptypes.GlobalVirtualGroupFamily{
		PrimarySpId: 1,
	}, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err := a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, ErrUnexpectedObjectStatusWithDetail("test_object", storagetypes.OBJECT_STATUS_CREATED, storagetypes.OBJECT_STATUS_SEALED), err)

	// VerifyPutObjectPermission get error
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{
		ObjectStatus: storagetypes.OBJECT_STATUS_CREATED,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
		Id: 1,
	}, nil).Times(0) // the SPID query result is cached already in above tests
	mockedConsensus.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtualgrouptypes.GlobalVirtualGroupFamily{
		PrimarySpId: 1,
	}, nil).Times(1)
	mockedConsensus.EXPECT().VerifyPutObjectPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, errors.New("error")).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	verifyResult, _ := a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, false, verifyResult)

	// VerifyPutObjectPermission doesn't have an error
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{
		ObjectStatus: storagetypes.OBJECT_STATUS_CREATED,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
		Id: 1,
	}, nil).Times(0) // the SPID query result is cached already in above tests
	mockedConsensus.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtualgrouptypes.GlobalVirtualGroupFamily{
		PrimarySpId: 1,
	}, nil).Times(1)
	mockedConsensus.EXPECT().VerifyPutObjectPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	verifyResult, _ = a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, verifyResult)

}
func Test_VerifyAuth_GetUploadingState(t *testing.T) {
	authType := coremodule.AuthOpTypeGetUploadingState
	VerifyObjectAndBucketAndSPID(t, authType)
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	a.baseApp.SetGfSpDB(m)

	privateKey, _ := crypto.GenerateKey()
	userAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// VerifyPutObjectPermission doesn't have an error
	mockedConsensus := consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{
		ObjectStatus: storagetypes.OBJECT_STATUS_CREATED,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
		Id: 1,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtualgrouptypes.GlobalVirtualGroupFamily{
		PrimarySpId: 1,
	}, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	verifyResult, _ := a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, verifyResult)
}

func Test_VerifyAuth_GetObject(t *testing.T) {
	VerifyObjectAndBucketAndSPID(t, coremodule.AuthOpTypeGetObject)

	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	a.baseApp.SetGfSpDB(m)

	privateKey, _ := crypto.GenerateKey()
	userAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// QueryPaymentStreamRecord get error
	mockedConsensus := consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{
		ObjectStatus: storagetypes.OBJECT_STATUS_CREATED,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
		Id: 1,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtualgrouptypes.GlobalVirtualGroupFamily{
		PrimarySpId: 1,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QueryPaymentStreamRecord(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err := a.VerifyAuthentication(context.Background(), coremodule.AuthOpTypeGetObject, userAddress, "test_bucket", "test_object")
	assert.Equal(t, errors.New("error"), err)

	// QueryPaymentStreamRecord get error
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{
		ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
		Id: 1,
	}, nil).Times(0) // the SPID query result is cached already in above tests
	mockedConsensus.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtualgrouptypes.GlobalVirtualGroupFamily{
		PrimarySpId: 1,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QueryPaymentStreamRecord(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	verifyResult, _ := a.VerifyAuthentication(context.Background(), coremodule.AuthOpTypeGetObject, userAddress, "test_bucket", "test_object")
	assert.Equal(t, false, verifyResult)

	// streamRecord.Status != paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{
		ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
		Id: 1,
	}, nil).Times(0) // the SPID query result is cached already in above tests
	mockedConsensus.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtualgrouptypes.GlobalVirtualGroupFamily{
		PrimarySpId: 1,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QueryPaymentStreamRecord(gomock.Any(), gomock.Any()).Return(&paymenttypes.StreamRecord{
		Status: paymenttypes.STREAM_ACCOUNT_STATUS_FROZEN,
	}, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err = a.VerifyAuthentication(context.Background(), coremodule.AuthOpTypeGetObject, userAddress, "test_bucket", "test_object")
	assert.Equal(t, ErrPaymentState, err)

	// VerifyPermission got an error
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{
		ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
		Id: 1,
	}, nil).Times(0) // the SPID query result is cached already in above tests
	mockedConsensus.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtualgrouptypes.GlobalVirtualGroupFamily{
		PrimarySpId: 1,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QueryPaymentStreamRecord(gomock.Any(), gomock.Any()).Return(&paymenttypes.StreamRecord{
		Status: paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE,
	}, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)

	mockedSpClient := gfspclient.NewMockGfSpClientAPI(ctrl)
	a.baseApp.SetGfSpClient(mockedSpClient)
	mockedSpClient.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(permissiontypes.ACTION_GET_OBJECT)).Return(nil, errors.New("error")).Times(1)

	verifyResult, _ = a.VerifyAuthentication(context.Background(), coremodule.AuthOpTypeGetObject, userAddress, "test_bucket", "test_object")
	assert.Equal(t, false, verifyResult)

	// Normal Case
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{
		ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
		Id: 1,
	}, nil).Times(0) // the SPID query result is cached already in above tests
	mockedConsensus.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtualgrouptypes.GlobalVirtualGroupFamily{
		PrimarySpId: 1,
	}, nil).Times(1)
	mockedConsensus.EXPECT().QueryPaymentStreamRecord(gomock.Any(), gomock.Any()).Return(&paymenttypes.StreamRecord{
		Status: paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE,
	}, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)

	mockedSpClient = gfspclient.NewMockGfSpClientAPI(ctrl)
	a.baseApp.SetGfSpClient(mockedSpClient)
	resp := &storagetypes.QueryVerifyPermissionResponse{Effect: permissiontypes.EFFECT_ALLOW}
	mockedSpClient.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(permissiontypes.ACTION_GET_OBJECT)).DoAndReturn(
		func(ctx context.Context, Operator string, bucketName string, objectName string,
			actionType permissiontypes.ActionType, opts ...grpc.DialOption) (*permissiontypes.Effect, error) {
			return &resp.Effect, nil
		},
	).Times(1)

	verifyResult, _ = a.VerifyAuthentication(context.Background(), coremodule.AuthOpTypeGetObject, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, verifyResult)

}

func Test_VerifyAuth_RecoveryPiece(t *testing.T) {
	authType := coremodule.AuthOpTypeGetRecoveryPiece
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	a.baseApp.SetGfSpDB(m)

	privateKey, _ := crypto.GenerateKey()
	userAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// bucket does not exist
	mockedConsensus := consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, errors.New("error: No such bucket")).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err := a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, ErrNoSuchBucket, err)

	// object does not exist
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, errors.New("error: No such object")).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err = a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, ErrNoSuchObject, err)

	// other error when QueryBucketInfoAndObjectInfo
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, errors.New("other error")).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err = a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, strings.Contains(err.Error(), "failed to get bucket and object info from consensus"))

	a.baseApp.SetOperatorAddress(TestSpAddress)

	// object exists but get SP ID returns error
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, &storagetypes.ObjectInfo{}, nil).Times(1)

	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err = a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, strings.Contains(err.Error(), "getSPID error"))

	// GetGlobalVirtualGroup returns error
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&storagetypes.BucketInfo{Id: sdkmath.NewUint(10)},
		&storagetypes.ObjectInfo{LocalVirtualGroupId: 1}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
		Id: 1,
	}, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)

	mockedSpClient := gfspclient.NewMockGfSpClientAPI(ctrl)
	a.baseApp.SetGfSpClient(mockedSpClient)
	mockedSpClient.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Eq(uint32(1))).Return(nil, errors.New("error")).Times(1)

	_, err = a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, strings.Contains(err.Error(), "failed to global virtual group info from metaData"))

	// mismatched SecondarySp
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&storagetypes.BucketInfo{Id: sdkmath.NewUint(10)},
		&storagetypes.ObjectInfo{LocalVirtualGroupId: 1}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{
		Id: 1,
	}, nil).Times(0) // the SPID query result is cached already in above tests
	a.baseApp.SetConsensus(mockedConsensus)

	mockedSpClient = gfspclient.NewMockGfSpClientAPI(ctrl)
	a.baseApp.SetGfSpClient(mockedSpClient)

	mockedGvgData := &virtualgrouptypes.GlobalVirtualGroup{
		SecondarySpIds: []uint32{},
	}
	mockedSpClient.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Eq(uint32(1))).Return(mockedGvgData, nil).Times(1)

	_, err = a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, strings.Contains(err.Error(), "mismatch"))

	// objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_SEALED
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&storagetypes.BucketInfo{Id: sdkmath.NewUint(10)},
		&storagetypes.ObjectInfo{LocalVirtualGroupId: 1}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(0) // the SPID query result is cached already in above tests
	a.baseApp.SetConsensus(mockedConsensus)

	mockedSpClient = gfspclient.NewMockGfSpClientAPI(ctrl)
	a.baseApp.SetGfSpClient(mockedSpClient)

	mockedGvgData = &virtualgrouptypes.GlobalVirtualGroup{SecondarySpIds: []uint32{1}}
	mockedSpClient.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Eq(uint32(1))).Return(mockedGvgData, nil).Times(1)

	_, err = a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, ErrNotSealedState, err)

	// VerifyPermission got an error
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&storagetypes.BucketInfo{Id: sdkmath.NewUint(10)},
		&storagetypes.ObjectInfo{LocalVirtualGroupId: 1, ObjectStatus: storagetypes.OBJECT_STATUS_SEALED}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(0) // the SPID query result is cached already in above tests
	a.baseApp.SetConsensus(mockedConsensus)

	mockedSpClient = gfspclient.NewMockGfSpClientAPI(ctrl)
	a.baseApp.SetGfSpClient(mockedSpClient)

	mockedGvgData = &virtualgrouptypes.GlobalVirtualGroup{SecondarySpIds: []uint32{1}}
	mockedSpClient.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Eq(uint32(1))).Return(mockedGvgData, nil).Times(1)
	mockedSpClient.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(permissiontypes.ACTION_GET_OBJECT)).Return(nil, errors.New("error")).Times(1)

	verifyResult, _ := a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, false, verifyResult)

	// Normal Case
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfoAndObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&storagetypes.BucketInfo{Id: sdkmath.NewUint(10)},
		&storagetypes.ObjectInfo{LocalVirtualGroupId: 1, ObjectStatus: storagetypes.OBJECT_STATUS_SEALED}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(0) // the SPID query result is cached already in above tests
	a.baseApp.SetConsensus(mockedConsensus)

	mockedSpClient = gfspclient.NewMockGfSpClientAPI(ctrl)
	a.baseApp.SetGfSpClient(mockedSpClient)

	mockedGvgData = &virtualgrouptypes.GlobalVirtualGroup{SecondarySpIds: []uint32{1}}
	mockedSpClient.EXPECT().GetGlobalVirtualGroup(gomock.Any(), gomock.Any(), gomock.Eq(uint32(1))).Return(mockedGvgData, nil).Times(1)
	resp := &storagetypes.QueryVerifyPermissionResponse{Effect: permissiontypes.EFFECT_ALLOW}
	mockedSpClient.EXPECT().VerifyPermission(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(permissiontypes.ACTION_GET_OBJECT)).Return(&resp.Effect, nil).Times(1)

	verifyResult, _ = a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, verifyResult)

}

func Test_VerifyAuth_GetBucketQuota(t *testing.T) {
	VerifyAuthBucket(t, coremodule.AuthOpTypeGetBucketQuota)
}
func Test_VerifyAuth_ListBucketReadRecord(t *testing.T) {
	VerifyAuthBucket(t, coremodule.AuthOpTypeListBucketReadRecord)

}
func VerifyAuthBucket(t *testing.T, authType coremodule.AuthOpType) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := spdb.NewMockSPDB(ctrl)
	a.baseApp.SetGfSpDB(m)

	privateKey, _ := crypto.GenerateKey()
	userAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// bucket does not exist
	mockedConsensus := consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err := a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, strings.Contains(err.Error(), "failed to get bucket info from consensus, error"))

	// get SP ID returns error
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err = a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, strings.Contains(err.Error(), "getSPID error"))

	// GetBucketPrimarySPID get errors
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(1)
	mockedConsensus.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(nil, errors.New("error")).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err = a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, strings.Contains(err.Error(), "GetBucketPrimarySPID error"))

	// ErrMismatchSp
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(0)
	mockedConsensus.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtualgrouptypes.GlobalVirtualGroupFamily{
		PrimarySpId: 2,
	}, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err = a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, ErrMismatchSp, err)

	// bucketInfo.GetOwner() != account
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{Owner: "another_account"}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(0)
	mockedConsensus.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtualgrouptypes.GlobalVirtualGroupFamily{
		PrimarySpId: 1,
	}, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	_, err = a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, ErrNoPermission, err)

	// Normal Case
	mockedConsensus = consensus.NewMockConsensus(ctrl)
	mockedConsensus.EXPECT().QueryBucketInfo(gomock.Any(), gomock.Any()).Return(&storagetypes.BucketInfo{Owner: userAddress}, nil).Times(1)
	mockedConsensus.EXPECT().QuerySP(gomock.Any(), gomock.Any()).Return(&sptypes.StorageProvider{Id: 1}, nil).Times(0)
	mockedConsensus.EXPECT().QueryVirtualGroupFamily(gomock.Any(), gomock.Any()).Return(&virtualgrouptypes.GlobalVirtualGroupFamily{
		PrimarySpId: 1,
	}, nil).Times(1)
	a.baseApp.SetConsensus(mockedConsensus)
	verifyResult, _ := a.VerifyAuthentication(context.Background(), authType, userAddress, "test_bucket", "test_object")
	assert.Equal(t, true, verifyResult)

}
