package service

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/types/resource"
	gnfdresource "github.com/bnb-chain/greenfield/types/resource"
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/juno/v4/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

func TestVerifyBucketPermission(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	//nolint:all
	mockDB := bsdb.NewMockBSDB(ctrl)
	m := &Metadata{
		name: "mockMetadata",
		bsDB: mockDB,
	}

	mockDB.EXPECT().GetPermissionByResourceAndPrincipal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).MaxTimes(100)
	operator, _ := sdk.AccAddressFromHexUnsafe("0x260CA9838382D1D71897423ED796C3443A4DE3FC")
	testCases := []struct {
		name           string
		m              *Metadata
		bucketInfo     *bsdb.Bucket
		operator       sdk.AccAddress
		action         permtypes.ActionType
		options        *permtypes.VerifyOptions
		isErr          bool
		expectedEffect permtypes.Effect
	}{
		{
			"operator is same as bucket owner",
			m,
			&bsdb.Bucket{
				ID:         1,
				Owner:      common.HexToAddress("0x260CA9838382D1D71897423ED796C3443A4DE3FC"),
				BucketName: "qquh6wekw1",
				Visibility: "VISIBILITY_TYPE_PRIVATE",
				BucketID:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
			},
			operator,
			6,
			nil,
			false,
			permtypes.EFFECT_ALLOW,
		},
		{
			"public read bucket allowed actions",
			m,
			&bsdb.Bucket{
				ID:         1,
				Owner:      common.HexToAddress("0x1517425A984D1A139DDE31337174B960B33EE58B"),
				BucketName: "u31ji4db",
				Visibility: "VISIBILITY_TYPE_PUBLIC_READ",
				BucketID:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
			},
			operator,
			6,
			nil,
			false,
			permtypes.EFFECT_ALLOW,
		},
		{
			"private visibility with public read bucket allowed actions",
			m,
			&bsdb.Bucket{
				ID:         1,
				Owner:      common.HexToAddress("0x1517425A984D1A139DDE31337174B960B33EE58B"),
				BucketName: "u31ji4db",
				Visibility: "VISIBILITY_TYPE_PRIVATE",
				BucketID:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
			},
			operator,
			6,
			nil,
			false,
			permtypes.EFFECT_DENY,
		},
		{
			"public visibility without public read bucket allowed actions",
			m,
			&bsdb.Bucket{
				ID:         1,
				Owner:      common.HexToAddress("0x1517425A984D1A139DDE31337174B960B33EE58B"),
				BucketName: "u31ji4db",
				Visibility: "VISIBILITY_TYPE_PUBLIC_READ",
				BucketID:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
			},
			operator,
			1,
			nil,
			false,
			permtypes.EFFECT_DENY,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			effect, err := testCase.m.VerifyBucketPermission(ctx, testCase.bucketInfo, testCase.operator,
				testCase.action, testCase.options)
			if !testCase.isErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				return
			}
			assert.Equal(t, testCase.expectedEffect, effect)
		})
	}
}

func TestVerifyPolicy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	//nolint:all
	mockDB := bsdb.NewMockBSDB(ctrl)
	m := &Metadata{
		name: "mockMetadata",
		bsDB: mockDB,
	}
	operator, _ := sdk.AccAddressFromHexUnsafe("0x260CA9838382D1D71897423ED796C3443A4DE3FC")
	gomock.InOrder(
		mockDB.EXPECT().GetPermissionByResourceAndPrincipal(
			gnfdresource.RESOURCE_TYPE_BUCKET.String(),
			"1",
			permtypes.PRINCIPAL_TYPE_GNFD_ACCOUNT.String(),
			operator.String()).Return(&bsdb.Permission{
			ID:              1,
			PrincipalType:   1,
			PrincipalValue:  operator.String(),
			ResourceType:    "RESOURCE_TYPE_BUCKET",
			ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
			PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
			CreateTimestamp: 1681198000,
			UpdateTimestamp: 1681198003,
			ExpirationTime:  1682864111,
			Removed:         false,
		}, nil).MaxTimes(100),
		mockDB.EXPECT().GetStatementsByPolicyID(gomock.Any()).Return([]*bsdb.Statement{
			{
				ID:             1,
				PolicyID:       common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				Effect:         "EFFECT_ALLOW",
				ActionValue:    8,
				Resources:      nil,
				ExpirationTime: 1682864111,
				LimitSize:      0,
				Removed:        false,
			},
		}, nil).MaxTimes(100),
		mockDB.EXPECT().GetPermissionsByResourceAndPrincipleType(
			gnfdresource.RESOURCE_TYPE_BUCKET.String(),
			"1",
			permtypes.PRINCIPAL_TYPE_GNFD_GROUP.String()).Return([]*bsdb.Permission{
			{
				ID:              7,
				PrincipalType:   2,
				PrincipalValue:  "0x0000000000000000000000000000000000000000000000000000000000000001",
				ResourceType:    "RESOURCE_TYPE_BUCKET",
				ResourceID:      common.HexToHash("0x000000000000000000000000000000000000000000000000000000000000000B"),
				PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000006"),
				CreateTimestamp: 1681198000,
				UpdateTimestamp: 1681198003,
				ExpirationTime:  1682864111,
				Removed:         false,
			},
			{
				ID:              5,
				PrincipalType:   2,
				PrincipalValue:  "0x0000000000000000000000000000000000000000000000000000000000000002",
				ResourceType:    "RESOURCE_TYPE_BUCKET",
				ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000009"),
				PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000004"),
				CreateTimestamp: 1681191000,
				UpdateTimestamp: 1681191234,
				ExpirationTime:  1682833333,
				Removed:         false,
			},
		}, nil).MaxTimes(100),
		mockDB.EXPECT().GetGroupsByGroupIDAndAccount([]common.Hash{
			common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
			common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
		}, common.HexToHash(operator.String())).Return([]*bsdb.Group{
			{
				ID:              1,
				Owner:           common.HexToAddress("0x5ca6d69Ac76B42f6035C6FfE719c0d1a460a1045"),
				GroupID:         common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				GroupName:       "x9uahlb0",
				AccountID:       common.Hash{},
				OperatorAddress: common.HexToAddress(operator.String()),
				Removed:         false,
			},
		}, nil).MaxTimes(100),
		mockDB.EXPECT().GetStatementsByPolicyID([]common.Hash{
			common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000006"),
		}).Return([]*bsdb.Statement{{
			ID:             1,
			PolicyID:       common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000006"),
			Effect:         "EFFECT_ALLOW",
			ActionValue:    8,
			Resources:      nil,
			ExpirationTime: 1682833333,
			LimitSize:      0,
			Removed:        false,
		}}, nil).MaxTimes(100),
	)

	gomock.InOrder(
		mockDB.EXPECT().GetPermissionByResourceAndPrincipal(
			gnfdresource.RESOURCE_TYPE_OBJECT.String(),
			"2",
			permtypes.PRINCIPAL_TYPE_GNFD_ACCOUNT.String(),
			operator.String()).Return(&bsdb.Permission{
			ID:              1,
			PrincipalType:   1,
			PrincipalValue:  operator.String(),
			ResourceType:    "RESOURCE_TYPE_OBJECT",
			ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
			PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
			CreateTimestamp: 1681198000,
			UpdateTimestamp: 1681198003,
			ExpirationTime:  1682864111,
			Removed:         false,
		}, nil).MaxTimes(100),
		mockDB.EXPECT().GetStatementsByPolicyID(gomock.Any()).Return([]*bsdb.Statement{
			{
				ID:             2,
				PolicyID:       common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
				Effect:         "EFFECT_ALLOW",
				ActionValue:    8,
				Resources:      nil,
				ExpirationTime: 1682864111,
				LimitSize:      0,
				Removed:        false,
			},
		}, nil).MaxTimes(100),
		mockDB.EXPECT().GetPermissionsByResourceAndPrincipleType(
			gnfdresource.RESOURCE_TYPE_OBJECT.String(),
			"1",
			permtypes.PRINCIPAL_TYPE_GNFD_GROUP.String()).Return([]*bsdb.Permission{
			{
				ID:              7,
				PrincipalType:   2,
				PrincipalValue:  "0x0000000000000000000000000000000000000000000000000000000000000002",
				ResourceType:    "RESOURCE_TYPE_OBJECT",
				ResourceID:      common.HexToHash("0x000000000000000000000000000000000000000000000000000000000000000B"),
				PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000006"),
				CreateTimestamp: 1681198000,
				UpdateTimestamp: 1681198003,
				ExpirationTime:  1682864111,
				Removed:         false,
			},
			{
				ID:              5,
				PrincipalType:   2,
				PrincipalValue:  "0x0000000000000000000000000000000000000000000000000000000000000002",
				ResourceType:    "RESOURCE_TYPE_OBJECT",
				ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000009"),
				PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000004"),
				CreateTimestamp: 1681191000,
				UpdateTimestamp: 1681191234,
				ExpirationTime:  1682833333,
				Removed:         false,
			},
		}, nil).MaxTimes(100),
		mockDB.EXPECT().GetGroupsByGroupIDAndAccount([]common.Hash{
			common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
			common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
		}, common.HexToHash(operator.String())).Return([]*bsdb.Group{
			{
				ID:              1,
				Owner:           common.HexToAddress("0x5ca6d69Ac76B42f6035C6FfE719c0d1a460a1045"),
				GroupID:         common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				GroupName:       "x9uahlb0",
				AccountID:       common.Hash{},
				OperatorAddress: common.HexToAddress(operator.String()),
				Removed:         false,
			},
		}, nil).MaxTimes(100),
		mockDB.EXPECT().GetStatementsByPolicyID([]common.Hash{
			common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000006"),
		}).Return([]*bsdb.Statement{{
			ID:             1,
			PolicyID:       common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000006"),
			Effect:         "EFFECT_DENY",
			ActionValue:    8,
			Resources:      nil,
			ExpirationTime: 1682833333,
			LimitSize:      0,
			Removed:        false,
		}}, nil).MaxTimes(100),
	)

	testCases := []struct {
		name           string
		m              *Metadata
		resourceID     math.Uint
		resourceType   resource.ResourceType
		operator       sdk.AccAddress
		action         permtypes.ActionType
		opts           *permtypes.VerifyOptions
		isErr          bool
		expectedEffect permtypes.Effect
	}{
		{
			"verify bucket permission",
			m,
			math.NewUintFromBigInt(common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001").Big()),
			gnfdresource.RESOURCE_TYPE_BUCKET,
			operator,
			3,
			nil,
			false,
			permtypes.EFFECT_ALLOW,
		},
		{
			"verify object permission",
			m,
			math.NewUintFromBigInt(common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002").Big()),
			gnfdresource.RESOURCE_TYPE_OBJECT,
			operator,
			3,
			nil,
			false,
			permtypes.EFFECT_ALLOW,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			effect, err := testCase.m.VerifyPolicy(ctx, testCase.resourceID, testCase.resourceType,
				testCase.operator, testCase.action, testCase.opts)
			if !testCase.isErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				return
			}
			assert.Equal(t, testCase.expectedEffect, effect)
		})
	}
}
