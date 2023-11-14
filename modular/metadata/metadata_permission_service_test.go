package metadata

import (
	"context"
	"testing"
	"time"

	permission_types "github.com/bnb-chain/greenfield/x/permission/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/forbole/juno/v4/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

func TestMetadataModular_GfSpVerifyPermission_VerifyBucketPermission_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) {
			return &bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "barry",
				Visibility:                 "VISIBILITY_TYPE_PRIVATE",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				SourceType:                 "SOURCE_TYPE_ORIGIN",
				CreateAt:                   0,
				CreateTime:                 0,
				CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				GlobalVirtualGroupFamilyID: 4,
				ChargedReadQuota:           0,
				PaymentPriceTime:           0,
				Removed:                    false,
				Status:                     "",
				DeleteAt:                   0,
				DeleteReason:               "",
				UpdateAt:                   0,
				UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				UpdateTime:                 0,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionByResourceAndPrincipal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, common.Hash) (*bsdb.Permission, error) {
			return &bsdb.Permission{
				ID:              2,
				PrincipalType:   2,
				PrincipalValue:  "3",
				ResourceType:    "RESOURCE_TYPE_BUCKET",
				ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
				CreateTimestamp: time.Now().Unix(),
				UpdateTimestamp: time.Now().Unix(),
				ExpirationTime:  time.Now().Unix() + 100000,
				Removed:         false,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return []*bsdb.Statement{
				&bsdb.Statement{
					ID:             1,
					PolicyID:       common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
					Effect:         "EFFECT_ALLOW",
					ActionValue:    64,
					Resources:      nil,
					ExpirationTime: 0,
					LimitSize:      0,
					Removed:        false,
				},
			}, nil
		},
	).Times(1)
	effect, err := a.GfSpVerifyPermission(context.Background(), &storagetypes.QueryVerifyPermissionRequest{
		Operator:   "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		BucketName: "barry",
		ObjectName: "",
		ActionType: 6,
	})
	assert.Nil(t, err)
	assert.Equal(t, "EFFECT_ALLOW", effect.Effect.String())
}

func TestMetadataModular_GfSpVerifyPermission_VerifyBucketPermission_Success2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) {
			return &bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "barry",
				Visibility:                 "VISIBILITY_TYPE_PRIVATE",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				SourceType:                 "SOURCE_TYPE_ORIGIN",
				CreateAt:                   0,
				CreateTime:                 0,
				CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				GlobalVirtualGroupFamilyID: 4,
				ChargedReadQuota:           0,
				PaymentPriceTime:           0,
				Removed:                    false,
				Status:                     "",
				DeleteAt:                   0,
				DeleteReason:               "",
				UpdateAt:                   0,
				UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				UpdateTime:                 0,
			}, nil
		},
	).Times(1)
	effect, err := a.GfSpVerifyPermission(context.Background(), &storagetypes.QueryVerifyPermissionRequest{
		Operator:   "0x11E0A11A7A01E2E757447B52FBD7152004AC699D",
		BucketName: "barry",
		ObjectName: "",
		ActionType: 6,
	})
	assert.Nil(t, err)
	assert.Equal(t, "EFFECT_ALLOW", effect.Effect.String())
}

func TestMetadataModular_GfSpVerifyPermission_VerifyBucketPermission_Success3(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) {
			return &bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "barry",
				Visibility:                 "VISIBILITY_TYPE_PRIVATE",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				SourceType:                 "SOURCE_TYPE_ORIGIN",
				CreateAt:                   0,
				CreateTime:                 0,
				CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				GlobalVirtualGroupFamilyID: 4,
				ChargedReadQuota:           0,
				PaymentPriceTime:           0,
				Removed:                    false,
				Status:                     "",
				DeleteAt:                   0,
				DeleteReason:               "",
				UpdateAt:                   0,
				UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				UpdateTime:                 0,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionByResourceAndPrincipal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, common.Hash) (*bsdb.Permission, error) {
			return &bsdb.Permission{
				ID:              2,
				PrincipalType:   2,
				PrincipalValue:  "3",
				ResourceType:    "RESOURCE_TYPE_BUCKET",
				ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
				CreateTimestamp: time.Now().Unix(),
				UpdateTimestamp: time.Now().Unix(),
				ExpirationTime:  time.Now().Unix() + 100000,
				Removed:         false,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return nil, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionsByResourceAndPrincipleType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, common.Hash, bool) ([]*bsdb.Permission, error) {
			return []*bsdb.Permission{
				&bsdb.Permission{
					ID:              2,
					PrincipalType:   3,
					PrincipalValue:  "3",
					ResourceType:    "RESOURCE_TYPE_BUCKET",
					ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
					CreateTimestamp: time.Now().Unix(),
					UpdateTimestamp: time.Now().Unix(),
					ExpirationTime:  time.Now().Unix() + 100000,
					Removed:         false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetGroupsByGroupIDAndAccount(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, common.Address, bool) ([]*bsdb.Group, error) {
			return []*bsdb.Group{
				&bsdb.Group{
					ID:             1,
					Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
					GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000003"),
					GroupName:      "test",
					SourceType:     "SOURCE_TYPE_ORIGIN",
					Extra:          "",
					AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
					Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
					ExpirationTime: 0,
					CreateAt:       0,
					CreateTime:     0,
					UpdateAt:       0,
					UpdateTime:     0,
					Removed:        false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return []*bsdb.Statement{
				&bsdb.Statement{
					ID:             1,
					PolicyID:       common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
					Effect:         "EFFECT_ALLOW",
					ActionValue:    64,
					Resources:      nil,
					ExpirationTime: 0,
					LimitSize:      0,
					Removed:        false,
				},
			}, nil
		},
	).Times(1)
	effect, err := a.GfSpVerifyPermission(context.Background(), &storagetypes.QueryVerifyPermissionRequest{
		Operator:   "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		BucketName: "barry",
		ObjectName: "",
		ActionType: 6,
	})
	assert.Nil(t, err)
	assert.Equal(t, "EFFECT_ALLOW", effect.Effect.String())
}

func TestMetadataModular_GfSpVerifyPermission_VerifyBucketPermission_Success4(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) {
			return &bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("ABC"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "barry",
				Visibility:                 "VISIBILITY_TYPE_PUBLIC_READ",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				SourceType:                 "SOURCE_TYPE_ORIGIN",
				CreateAt:                   0,
				CreateTime:                 0,
				CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				GlobalVirtualGroupFamilyID: 4,
				ChargedReadQuota:           0,
				PaymentPriceTime:           0,
				Removed:                    false,
				Status:                     "",
				DeleteAt:                   0,
				DeleteReason:               "",
				UpdateAt:                   0,
				UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				UpdateTime:                 0,
			}, nil
		},
	).Times(1)
	effect, err := a.GfSpVerifyPermission(context.Background(), &storagetypes.QueryVerifyPermissionRequest{
		Operator:   "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		BucketName: "barry",
		ObjectName: "",
		ActionType: 6,
	})
	assert.Nil(t, err)
	assert.Equal(t, "EFFECT_ALLOW", effect.Effect.String())
}

func TestMetadataModular_GfSpVerifyPermission_VerifyObjectPermission_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) {
			return &bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "barry",
				Visibility:                 "VISIBILITY_TYPE_PUBLIC_READ",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				SourceType:                 "SOURCE_TYPE_ORIGIN",
				CreateAt:                   0,
				CreateTime:                 0,
				CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				GlobalVirtualGroupFamilyID: 4,
				ChargedReadQuota:           0,
				PaymentPriceTime:           0,
				Removed:                    false,
				Status:                     "",
				DeleteAt:                   0,
				DeleteReason:               "",
				UpdateAt:                   0,
				UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				UpdateTime:                 0,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetObjectByName(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, bool) (*bsdb.Object, error) {
			return &bsdb.Object{
				ID:                  1,
				Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				LocalVirtualGroupId: 0,
				BucketName:          "barry",
				ObjectName:          "test",
				ObjectID:            common.HexToHash("1"),
				BucketID:            common.HexToHash("1"),
				PayloadSize:         0,
				Visibility:          "VISIBILITY_TYPE_INHERIT",
				ContentType:         "",
				CreateAt:            0,
				CreateTime:          0,
				ObjectStatus:        "",
				RedundancyType:      "",
				SourceType:          "",
				Checksums:           nil,
				LockedBalance:       common.HexToHash("1"),
				Removed:             false,
				UpdateTime:          0,
				UpdateAt:            0,
				DeleteAt:            0,
				DeleteReason:        "",
				CreateTxHash:        common.HexToHash("1"),
				UpdateTxHash:        common.HexToHash("1"),
				SealTxHash:          common.HexToHash("1"),
			}, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionByResourceAndPrincipal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, common.Hash) (*bsdb.Permission, error) {
			return &bsdb.Permission{
				ID:              2,
				PrincipalType:   2,
				PrincipalValue:  "3",
				ResourceType:    "RESOURCE_TYPE_BUCKET",
				ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
				CreateTimestamp: time.Now().Unix(),
				UpdateTimestamp: time.Now().Unix(),
				ExpirationTime:  time.Now().Unix() + 100000,
				Removed:         false,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return nil, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionsByResourceAndPrincipleType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, common.Hash, bool) ([]*bsdb.Permission, error) {
			return []*bsdb.Permission{
				&bsdb.Permission{
					ID:              2,
					PrincipalType:   3,
					PrincipalValue:  "3",
					ResourceType:    "RESOURCE_TYPE_BUCKET",
					ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
					CreateTimestamp: time.Now().Unix(),
					UpdateTimestamp: time.Now().Unix(),
					ExpirationTime:  time.Now().Unix() + 100000,
					Removed:         false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetGroupsByGroupIDAndAccount(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, common.Address, bool) ([]*bsdb.Group, error) {
			return []*bsdb.Group{
				&bsdb.Group{
					ID:             1,
					Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
					GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000003"),
					GroupName:      "test",
					SourceType:     "SOURCE_TYPE_ORIGIN",
					Extra:          "",
					AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
					Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
					ExpirationTime: 0,
					CreateAt:       0,
					CreateTime:     0,
					UpdateAt:       0,
					UpdateTime:     0,
					Removed:        false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return []*bsdb.Statement{
				&bsdb.Statement{
					ID:             1,
					PolicyID:       common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
					Effect:         "EFFECT_ALLOW",
					ActionValue:    512,
					Resources:      []string{"grn"},
					ExpirationTime: 0,
					LimitSize:      0,
					Removed:        false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionByResourceAndPrincipal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, common.Hash) (*bsdb.Permission, error) {
			return &bsdb.Permission{
				ID:              2,
				PrincipalType:   2,
				PrincipalValue:  "3",
				ResourceType:    "RESOURCE_TYPE_BUCKET",
				ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
				CreateTimestamp: time.Now().Unix(),
				UpdateTimestamp: time.Now().Unix(),
				ExpirationTime:  time.Now().Unix() + 100000,
				Removed:         false,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return nil, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionsByResourceAndPrincipleType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, common.Hash, bool) ([]*bsdb.Permission, error) {
			return []*bsdb.Permission{
				&bsdb.Permission{
					ID:              2,
					PrincipalType:   3,
					PrincipalValue:  "3",
					ResourceType:    "RESOURCE_TYPE_BUCKET",
					ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
					CreateTimestamp: time.Now().Unix(),
					UpdateTimestamp: time.Now().Unix(),
					ExpirationTime:  time.Now().Unix() + 100000,
					Removed:         false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetGroupsByGroupIDAndAccount(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, common.Address, bool) ([]*bsdb.Group, error) {
			return []*bsdb.Group{
				&bsdb.Group{
					ID:             1,
					Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
					GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000003"),
					GroupName:      "test",
					SourceType:     "SOURCE_TYPE_ORIGIN",
					Extra:          "",
					AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
					Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
					ExpirationTime: 0,
					CreateAt:       0,
					CreateTime:     0,
					UpdateAt:       0,
					UpdateTime:     0,
					Removed:        false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return []*bsdb.Statement{
				&bsdb.Statement{
					ID:             1,
					PolicyID:       common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
					Effect:         "EFFECT_ALLOW",
					ActionValue:    512,
					Resources:      []string{"grn"},
					ExpirationTime: 0,
					LimitSize:      0,
					Removed:        false,
				},
			}, nil
		},
	).Times(1)
	effect, err := a.GfSpVerifyPermission(context.Background(), &storagetypes.QueryVerifyPermissionRequest{
		Operator:   "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		BucketName: "barry",
		ObjectName: "test",
		ActionType: 9,
	})
	assert.Nil(t, err)
	assert.Equal(t, "EFFECT_ALLOW", effect.Effect.String())
}

func TestMetadataModular_GfSpVerifyPermission_VerifyBucketPermission_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) {
			return &bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "barry",
				Visibility:                 "VISIBILITY_TYPE_PRIVATE",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				SourceType:                 "SOURCE_TYPE_ORIGIN",
				CreateAt:                   0,
				CreateTime:                 0,
				CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				GlobalVirtualGroupFamilyID: 4,
				ChargedReadQuota:           0,
				PaymentPriceTime:           0,
				Removed:                    false,
				Status:                     "",
				DeleteAt:                   0,
				DeleteReason:               "",
				UpdateAt:                   0,
				UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				UpdateTime:                 0,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionByResourceAndPrincipal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, common.Hash) (*bsdb.Permission, error) {
			return &bsdb.Permission{
				ID:              2,
				PrincipalType:   2,
				PrincipalValue:  "3",
				ResourceType:    "RESOURCE_TYPE_BUCKET",
				ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
				CreateTimestamp: time.Now().Unix(),
				UpdateTimestamp: time.Now().Unix(),
				ExpirationTime:  time.Now().Unix() + 100000,
				Removed:         false,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return nil, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionsByResourceAndPrincipleType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, common.Hash, bool) ([]*bsdb.Permission, error) {
			return []*bsdb.Permission{
				&bsdb.Permission{
					ID:              2,
					PrincipalType:   3,
					PrincipalValue:  "3",
					ResourceType:    "RESOURCE_TYPE_BUCKET",
					ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
					CreateTimestamp: time.Now().Unix(),
					UpdateTimestamp: time.Now().Unix(),
					ExpirationTime:  time.Now().Unix() + 100000,
					Removed:         false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetGroupsByGroupIDAndAccount(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, common.Address, bool) ([]*bsdb.Group, error) {
			return []*bsdb.Group{
				&bsdb.Group{
					ID:             1,
					Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
					GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000003"),
					GroupName:      "test",
					SourceType:     "SOURCE_TYPE_ORIGIN",
					Extra:          "",
					AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
					Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
					ExpirationTime: 0,
					CreateAt:       0,
					CreateTime:     0,
					UpdateAt:       0,
					UpdateTime:     0,
					Removed:        false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpVerifyPermission(context.Background(), &storagetypes.QueryVerifyPermissionRequest{
		Operator:   "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		BucketName: "barry",
		ObjectName: "",
		ActionType: 6,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpVerifyPermission_VerifyBucketPermission_Failed2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) {
			return &bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "barry",
				Visibility:                 "VISIBILITY_TYPE_PRIVATE",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				SourceType:                 "SOURCE_TYPE_ORIGIN",
				CreateAt:                   0,
				CreateTime:                 0,
				CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				GlobalVirtualGroupFamilyID: 4,
				ChargedReadQuota:           0,
				PaymentPriceTime:           0,
				Removed:                    false,
				Status:                     "",
				DeleteAt:                   0,
				DeleteReason:               "",
				UpdateAt:                   0,
				UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				UpdateTime:                 0,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionByResourceAndPrincipal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, common.Hash) (*bsdb.Permission, error) {
			return &bsdb.Permission{
				ID:              2,
				PrincipalType:   2,
				PrincipalValue:  "3",
				ResourceType:    "RESOURCE_TYPE_BUCKET",
				ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
				CreateTimestamp: time.Now().Unix(),
				UpdateTimestamp: time.Now().Unix(),
				ExpirationTime:  time.Now().Unix() + 100000,
				Removed:         false,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return nil, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionsByResourceAndPrincipleType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, common.Hash, bool) ([]*bsdb.Permission, error) {
			return []*bsdb.Permission{
				&bsdb.Permission{
					ID:              2,
					PrincipalType:   3,
					PrincipalValue:  "3",
					ResourceType:    "RESOURCE_TYPE_BUCKET",
					ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
					CreateTimestamp: time.Now().Unix(),
					UpdateTimestamp: time.Now().Unix(),
					ExpirationTime:  time.Now().Unix() + 100000,
					Removed:         false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetGroupsByGroupIDAndAccount(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, common.Address, bool) ([]*bsdb.Group, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpVerifyPermission(context.Background(), &storagetypes.QueryVerifyPermissionRequest{
		Operator:   "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		BucketName: "barry",
		ObjectName: "",
		ActionType: 6,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpVerifyPermission_VerifyBucketPermission_Failed3(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) {
			return &bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "barry",
				Visibility:                 "VISIBILITY_TYPE_PRIVATE",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				SourceType:                 "SOURCE_TYPE_ORIGIN",
				CreateAt:                   0,
				CreateTime:                 0,
				CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				GlobalVirtualGroupFamilyID: 4,
				ChargedReadQuota:           0,
				PaymentPriceTime:           0,
				Removed:                    false,
				Status:                     "",
				DeleteAt:                   0,
				DeleteReason:               "",
				UpdateAt:                   0,
				UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				UpdateTime:                 0,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionByResourceAndPrincipal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, common.Hash) (*bsdb.Permission, error) {
			return &bsdb.Permission{
				ID:              2,
				PrincipalType:   2,
				PrincipalValue:  "3",
				ResourceType:    "RESOURCE_TYPE_BUCKET",
				ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
				CreateTimestamp: time.Now().Unix(),
				UpdateTimestamp: time.Now().Unix(),
				ExpirationTime:  time.Now().Unix() + 100000,
				Removed:         false,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return nil, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionsByResourceAndPrincipleType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, common.Hash, bool) ([]*bsdb.Permission, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpVerifyPermission(context.Background(), &storagetypes.QueryVerifyPermissionRequest{
		Operator:   "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		BucketName: "barry",
		ObjectName: "",
		ActionType: 6,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpVerifyPermission_VerifyBucketPermission_Failed4(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) {
			return &bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "barry",
				Visibility:                 "VISIBILITY_TYPE_PRIVATE",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				SourceType:                 "SOURCE_TYPE_ORIGIN",
				CreateAt:                   0,
				CreateTime:                 0,
				CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				GlobalVirtualGroupFamilyID: 4,
				ChargedReadQuota:           0,
				PaymentPriceTime:           0,
				Removed:                    false,
				Status:                     "",
				DeleteAt:                   0,
				DeleteReason:               "",
				UpdateAt:                   0,
				UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				UpdateTime:                 0,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionByResourceAndPrincipal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, common.Hash) (*bsdb.Permission, error) {
			return &bsdb.Permission{
				ID:              2,
				PrincipalType:   2,
				PrincipalValue:  "3",
				ResourceType:    "RESOURCE_TYPE_BUCKET",
				ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
				CreateTimestamp: time.Now().Unix(),
				UpdateTimestamp: time.Now().Unix(),
				ExpirationTime:  time.Now().Unix() + 100000,
				Removed:         false,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpVerifyPermission(context.Background(), &storagetypes.QueryVerifyPermissionRequest{
		Operator:   "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		BucketName: "barry",
		ObjectName: "",
		ActionType: 6,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpVerifyPermission_VerifyBucketPermission_Failed5(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) {
			return &bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "barry",
				Visibility:                 "VISIBILITY_TYPE_PRIVATE",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				SourceType:                 "SOURCE_TYPE_ORIGIN",
				CreateAt:                   0,
				CreateTime:                 0,
				CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				GlobalVirtualGroupFamilyID: 4,
				ChargedReadQuota:           0,
				PaymentPriceTime:           0,
				Removed:                    false,
				Status:                     "",
				DeleteAt:                   0,
				DeleteReason:               "",
				UpdateAt:                   0,
				UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				UpdateTime:                 0,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionByResourceAndPrincipal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, common.Hash) (*bsdb.Permission, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpVerifyPermission(context.Background(), &storagetypes.QueryVerifyPermissionRequest{
		Operator:   "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		BucketName: "barry",
		ObjectName: "",
		ActionType: 6,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpVerifyPermission_VerifyBucketPermission_Failed6(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) {
			return &bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "barry",
				Visibility:                 "VISIBILITY_TYPE_PRIVATE",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				SourceType:                 "SOURCE_TYPE_ORIGIN",
				CreateAt:                   0,
				CreateTime:                 0,
				CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				GlobalVirtualGroupFamilyID: 4,
				ChargedReadQuota:           0,
				PaymentPriceTime:           0,
				Removed:                    false,
				Status:                     "",
				DeleteAt:                   0,
				DeleteReason:               "",
				UpdateAt:                   0,
				UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				UpdateTime:                 0,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionByResourceAndPrincipal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, common.Hash) (*bsdb.Permission, error) {
			return nil, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionsByResourceAndPrincipleType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, common.Hash, bool) ([]*bsdb.Permission, error) {
			return nil, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return nil, nil
		},
	).Times(1)
	effect, err := a.GfSpVerifyPermission(context.Background(), &storagetypes.QueryVerifyPermissionRequest{
		Operator:   "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		BucketName: "barry",
		ObjectName: "",
		ActionType: 6,
	})
	assert.Nil(t, err)
	assert.Equal(t, "EFFECT_DENY", effect.Effect.String())
}

func TestMetadataModular_GfSpVerifyPermission_VerifyObjectPermission_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) {
			return &bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "barry",
				Visibility:                 "VISIBILITY_TYPE_PUBLIC_READ",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				SourceType:                 "SOURCE_TYPE_ORIGIN",
				CreateAt:                   0,
				CreateTime:                 0,
				CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				GlobalVirtualGroupFamilyID: 4,
				ChargedReadQuota:           0,
				PaymentPriceTime:           0,
				Removed:                    false,
				Status:                     "",
				DeleteAt:                   0,
				DeleteReason:               "",
				UpdateAt:                   0,
				UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				UpdateTime:                 0,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetObjectByName(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, bool) (*bsdb.Object, error) {
			return nil, gorm.ErrRecordNotFound
		},
	).Times(1)
	_, err := a.GfSpVerifyPermission(context.Background(), &storagetypes.QueryVerifyPermissionRequest{
		Operator:   "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		BucketName: "barry",
		ObjectName: "test",
		ActionType: 9,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpVerifyPermission_VerifyObjectPermission_Failed2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) {
			return nil, nil
		},
	).Times(1)
	_, err := a.GfSpVerifyPermission(context.Background(), &storagetypes.QueryVerifyPermissionRequest{
		Operator:   "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		BucketName: "barry",
		ObjectName: "test",
		ActionType: 9,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpVerifyPermission_VerifyObjectPermission_Failed3(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpVerifyPermission(context.Background(), &storagetypes.QueryVerifyPermissionRequest{
		Operator:   "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		BucketName: "barry",
		ObjectName: "test",
		ActionType: 9,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpVerifyPermission_VerifyObjectPermission_Failed4(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpVerifyPermission(context.Background(), &storagetypes.QueryVerifyPermissionRequest{
		Operator:   "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		BucketName: "barry",
		ObjectName: "test",
		ActionType: 9,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpVerifyPermission_VerifyObjectPermission_Failed5(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	_, err := a.GfSpVerifyPermission(context.Background(), nil)
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpVerifyPermission_VerifyObjectPermission_Failed6(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	_, err := a.GfSpVerifyPermission(context.Background(), &storagetypes.QueryVerifyPermissionRequest{
		Operator:   "111",
		BucketName: "barry",
		ObjectName: "test",
		ActionType: 9,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpVerifyPermission_VerifyObjectPermission_Failed7(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	_, err := a.GfSpVerifyPermission(context.Background(), &storagetypes.QueryVerifyPermissionRequest{
		Operator:   "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		BucketName: "",
		ObjectName: "test",
		ActionType: 9,
	})
	assert.NotNil(t, err)
}

func TestMetadataModularGfSpGfSpVerifyPermissionByID_VerifyObjectPermission_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetObjectByID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(int64, bool) (*bsdb.Object, error) {
			return &bsdb.Object{
				ID:                  1,
				Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				LocalVirtualGroupId: 0,
				BucketName:          "barry",
				ObjectName:          "test",
				ObjectID:            common.HexToHash("1"),
				BucketID:            common.HexToHash("1"),
				PayloadSize:         0,
				Visibility:          "VISIBILITY_TYPE_INHERIT",
				ContentType:         "",
				CreateAt:            0,
				CreateTime:          0,
				ObjectStatus:        "",
				RedundancyType:      "",
				SourceType:          "",
				Checksums:           nil,
				LockedBalance:       common.HexToHash("1"),
				Removed:             false,
				UpdateTime:          0,
				UpdateAt:            0,
				DeleteAt:            0,
				DeleteReason:        "",
				CreateTxHash:        common.HexToHash("1"),
				UpdateTxHash:        common.HexToHash("1"),
				SealTxHash:          common.HexToHash("1"),
			}, nil
		},
	).Times(1)
	m.EXPECT().GetBucketByID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(int64, bool) (*bsdb.Bucket, error) {
			return &bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "barry",
				Visibility:                 "VISIBILITY_TYPE_PRIVATE",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				SourceType:                 "SOURCE_TYPE_ORIGIN",
				CreateAt:                   0,
				CreateTime:                 0,
				CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				GlobalVirtualGroupFamilyID: 4,
				ChargedReadQuota:           0,
				PaymentPriceTime:           0,
				Removed:                    false,
				Status:                     "",
				DeleteAt:                   0,
				DeleteReason:               "",
				UpdateAt:                   0,
				UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				UpdateTime:                 0,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionByResourceAndPrincipal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, common.Hash) (*bsdb.Permission, error) {
			return &bsdb.Permission{
				ID:              2,
				PrincipalType:   2,
				PrincipalValue:  "3",
				ResourceType:    "RESOURCE_TYPE_BUCKET",
				ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
				CreateTimestamp: time.Now().Unix(),
				UpdateTimestamp: time.Now().Unix(),
				ExpirationTime:  time.Now().Unix() + 100000,
				Removed:         false,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return nil, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionsByResourceAndPrincipleType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, common.Hash, bool) ([]*bsdb.Permission, error) {
			return []*bsdb.Permission{
				&bsdb.Permission{
					ID:              2,
					PrincipalType:   3,
					PrincipalValue:  "3",
					ResourceType:    "RESOURCE_TYPE_BUCKET",
					ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
					CreateTimestamp: time.Now().Unix(),
					UpdateTimestamp: time.Now().Unix(),
					ExpirationTime:  time.Now().Unix() + 100000,
					Removed:         false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetGroupsByGroupIDAndAccount(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, common.Address, bool) ([]*bsdb.Group, error) {
			return []*bsdb.Group{
				&bsdb.Group{
					ID:             1,
					Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
					GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000003"),
					GroupName:      "test",
					SourceType:     "SOURCE_TYPE_ORIGIN",
					Extra:          "",
					AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
					Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
					ExpirationTime: 0,
					CreateAt:       0,
					CreateTime:     0,
					UpdateAt:       0,
					UpdateTime:     0,
					Removed:        false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return []*bsdb.Statement{
				&bsdb.Statement{
					ID:             1,
					PolicyID:       common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
					Effect:         "EFFECT_ALLOW",
					ActionValue:    512,
					Resources:      []string{"grn"},
					ExpirationTime: 0,
					LimitSize:      0,
					Removed:        false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionByResourceAndPrincipal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, common.Hash) (*bsdb.Permission, error) {
			return &bsdb.Permission{
				ID:              2,
				PrincipalType:   2,
				PrincipalValue:  "3",
				ResourceType:    "RESOURCE_TYPE_BUCKET",
				ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
				CreateTimestamp: time.Now().Unix(),
				UpdateTimestamp: time.Now().Unix(),
				ExpirationTime:  time.Now().Unix() + 100000,
				Removed:         false,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return nil, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionsByResourceAndPrincipleType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, common.Hash, bool) ([]*bsdb.Permission, error) {
			return []*bsdb.Permission{
				&bsdb.Permission{
					ID:              2,
					PrincipalType:   3,
					PrincipalValue:  "3",
					ResourceType:    "RESOURCE_TYPE_BUCKET",
					ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
					CreateTimestamp: time.Now().Unix(),
					UpdateTimestamp: time.Now().Unix(),
					ExpirationTime:  time.Now().Unix() + 100000,
					Removed:         false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetGroupsByGroupIDAndAccount(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, common.Address, bool) ([]*bsdb.Group, error) {
			return []*bsdb.Group{
				&bsdb.Group{
					ID:             1,
					Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
					GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000003"),
					GroupName:      "test",
					SourceType:     "SOURCE_TYPE_ORIGIN",
					Extra:          "",
					AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
					Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
					ExpirationTime: 0,
					CreateAt:       0,
					CreateTime:     0,
					UpdateAt:       0,
					UpdateTime:     0,
					Removed:        false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return []*bsdb.Statement{
				&bsdb.Statement{
					ID:             1,
					PolicyID:       common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
					Effect:         "EFFECT_ALLOW",
					ActionValue:    512,
					Resources:      []string{"grn"},
					ExpirationTime: 0,
					LimitSize:      0,
					Removed:        false,
				},
			}, nil
		},
	).Times(1)
	effect, err := a.GfSpVerifyPermissionByID(context.Background(), &types.GfSpVerifyPermissionByIDRequest{
		Operator:     "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		ResourceType: 2,
		ResourceId:   1,
		ActionType:   9,
	})
	assert.Nil(t, err)
	assert.Equal(t, "EFFECT_ALLOW", effect.Effect.String())
}

func TestMetadataModularGfSpGfSpVerifyPermissionByID_VerifyBucket_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(int64, bool) (*bsdb.Bucket, error) {
			return &bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "barry",
				Visibility:                 "VISIBILITY_TYPE_PRIVATE",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				SourceType:                 "SOURCE_TYPE_ORIGIN",
				CreateAt:                   0,
				CreateTime:                 0,
				CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				PaymentAddress:             common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				GlobalVirtualGroupFamilyID: 4,
				ChargedReadQuota:           0,
				PaymentPriceTime:           0,
				Removed:                    false,
				Status:                     "",
				DeleteAt:                   0,
				DeleteReason:               "",
				UpdateAt:                   0,
				UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				UpdateTime:                 0,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionByResourceAndPrincipal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, common.Hash) (*bsdb.Permission, error) {
			return &bsdb.Permission{
				ID:              2,
				PrincipalType:   2,
				PrincipalValue:  "3",
				ResourceType:    "RESOURCE_TYPE_BUCKET",
				ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
				CreateTimestamp: time.Now().Unix(),
				UpdateTimestamp: time.Now().Unix(),
				ExpirationTime:  time.Now().Unix() + 100000,
				Removed:         false,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return nil, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionsByResourceAndPrincipleType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, common.Hash, bool) ([]*bsdb.Permission, error) {
			return []*bsdb.Permission{
				&bsdb.Permission{
					ID:              2,
					PrincipalType:   3,
					PrincipalValue:  "3",
					ResourceType:    "RESOURCE_TYPE_BUCKET",
					ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
					CreateTimestamp: time.Now().Unix(),
					UpdateTimestamp: time.Now().Unix(),
					ExpirationTime:  time.Now().Unix() + 100000,
					Removed:         false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetGroupsByGroupIDAndAccount(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, common.Address, bool) ([]*bsdb.Group, error) {
			return []*bsdb.Group{
				&bsdb.Group{
					ID:             1,
					Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
					GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000003"),
					GroupName:      "test",
					SourceType:     "SOURCE_TYPE_ORIGIN",
					Extra:          "",
					AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
					Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
					ExpirationTime: 0,
					CreateAt:       0,
					CreateTime:     0,
					UpdateAt:       0,
					UpdateTime:     0,
					Removed:        false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return []*bsdb.Statement{
				&bsdb.Statement{
					ID:             1,
					PolicyID:       common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
					Effect:         "EFFECT_ALLOW",
					ActionValue:    512,
					Resources:      []string{"grn"},
					ExpirationTime: 0,
					LimitSize:      0,
					Removed:        false,
				},
			}, nil
		},
	).Times(1)
	effect, err := a.GfSpVerifyPermissionByID(context.Background(), &types.GfSpVerifyPermissionByIDRequest{
		Operator:     "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		ResourceType: 1,
		ResourceId:   1,
		ActionType:   9,
	})
	assert.Nil(t, err)
	assert.Equal(t, "EFFECT_ALLOW", effect.Effect.String())
}

func TestMetadataModularGfSpGfSpVerifyPermissionByID_VerifyGroup_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetGroupByID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(int64, bool) (*bsdb.Group, error) {
			return &bsdb.Group{
				ID:             1,
				Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
				GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				GroupName:      "test",
				SourceType:     "SOURCE_TYPE_ORIGIN",
				Extra:          "",
				AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
				Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
				ExpirationTime: 0,
				CreateAt:       0,
				CreateTime:     0,
				UpdateAt:       0,
				UpdateTime:     0,
				Removed:        false,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionByResourceAndPrincipal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, string, common.Hash) (*bsdb.Permission, error) {
			return &bsdb.Permission{
				ID:              2,
				PrincipalType:   2,
				PrincipalValue:  "3",
				ResourceType:    "RESOURCE_TYPE_BUCKET",
				ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
				CreateTimestamp: time.Now().Unix(),
				UpdateTimestamp: time.Now().Unix(),
				ExpirationTime:  time.Now().Unix() + 100000,
				Removed:         false,
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return nil, nil
		},
	).Times(1)
	m.EXPECT().GetPermissionsByResourceAndPrincipleType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, common.Hash, bool) ([]*bsdb.Permission, error) {
			return []*bsdb.Permission{
				&bsdb.Permission{
					ID:              2,
					PrincipalType:   3,
					PrincipalValue:  "3",
					ResourceType:    "RESOURCE_TYPE_BUCKET",
					ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
					CreateTimestamp: time.Now().Unix(),
					UpdateTimestamp: time.Now().Unix(),
					ExpirationTime:  time.Now().Unix() + 100000,
					Removed:         false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetGroupsByGroupIDAndAccount(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, common.Address, bool) ([]*bsdb.Group, error) {
			return []*bsdb.Group{
				&bsdb.Group{
					ID:             1,
					Owner:          common.HexToAddress("0x84A0D38D64498414B14CD979159D57557345CD8B"),
					GroupID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000003"),
					GroupName:      "test",
					SourceType:     "SOURCE_TYPE_ORIGIN",
					Extra:          "",
					AccountID:      common.HexToAddress("0x00000000000000000000000000000000000000000x0000000000000000000000000000000000000000"),
					Operator:       common.HexToAddress("0x0000000000000000000000000000000000000000"),
					ExpirationTime: 0,
					CreateAt:       0,
					CreateTime:     0,
					UpdateAt:       0,
					UpdateTime:     0,
					Removed:        false,
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().GetStatementsByPolicyID(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Statement, error) {
			return []*bsdb.Statement{
				&bsdb.Statement{
					ID:             1,
					PolicyID:       common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
					Effect:         "EFFECT_ALLOW",
					ActionValue:    512,
					Resources:      []string{"grn"},
					ExpirationTime: 0,
					LimitSize:      0,
					Removed:        false,
				},
			}, nil
		},
	).Times(1)
	effect, err := a.GfSpVerifyPermissionByID(context.Background(), &types.GfSpVerifyPermissionByIDRequest{
		Operator:     "0x11E0A11A7A01E2E757447B52FBD7152004AC699e",
		ResourceType: 3,
		ResourceId:   1,
		ActionType:   9,
	})
	assert.Nil(t, err)
	assert.Equal(t, "EFFECT_ALLOW", effect.Effect.String())
}

func TestMetadataModularGfSpListObjectPolicies_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetObjectByName(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, bool) (*bsdb.Object, error) {
			return &bsdb.Object{
				ID:                  1,
				Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				LocalVirtualGroupId: 0,
				BucketName:          "barry",
				ObjectName:          "1111",
				ObjectID:            common.HexToHash("1"),
				BucketID:            common.HexToHash("1"),
				PayloadSize:         0,
				Visibility:          "",
				ContentType:         "",
				CreateAt:            0,
				CreateTime:          0,
				ObjectStatus:        "",
				RedundancyType:      "",
				SourceType:          "",
				Checksums:           nil,
				LockedBalance:       common.HexToHash("1"),
				Removed:             false,
				UpdateTime:          0,
				UpdateAt:            0,
				DeleteAt:            0,
				DeleteReason:        "",
				CreateTxHash:        common.HexToHash("1"),
				UpdateTxHash:        common.HexToHash("1"),
				SealTxHash:          common.HexToHash("1"),
			}, nil
		},
	).Times(1)
	m.EXPECT().ListObjectPolicies(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, permission_types.ActionType, common.Hash, int) ([]*bsdb.PermissionWithStatement, error) {
			return []*bsdb.PermissionWithStatement{
				&bsdb.PermissionWithStatement{
					Permission: bsdb.Permission{
						ID:              2,
						PrincipalType:   2,
						PrincipalValue:  "3",
						ResourceType:    "RESOURCE_TYPE_OBJECT",
						ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
						PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
						CreateTimestamp: time.Now().Unix(),
						UpdateTimestamp: time.Now().Unix(),
						ExpirationTime:  time.Now().Unix() + 100000,
						Removed:         false,
					},
					Statement: bsdb.Statement{
						ID:             2,
						PolicyID:       common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
						Effect:         "EFFECT_ALLOW",
						ActionValue:    64,
						Resources:      nil,
						ExpirationTime: time.Now().Unix() + 100000,
						LimitSize:      1,
						Removed:        false,
					},
				},
			}, nil
		},
	).Times(1)
	policies, err := a.GfSpListObjectPolicies(context.Background(), &types.GfSpListObjectPoliciesRequest{
		ObjectName: "1111",
		BucketName: "barry",
		ActionType: 6,
		Limit:      1111,
		StartAfter: 0,
	})
	assert.Nil(t, err)
	assert.Equal(t, "RESOURCE_TYPE_OBJECT", policies.Policies[0].ResourceType.String())
}

func TestMetadataModularGfSpListObjectPolicies_Success2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetObjectByName(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, bool) (*bsdb.Object, error) {
			return &bsdb.Object{
				ID:                  1,
				Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				LocalVirtualGroupId: 0,
				BucketName:          "barry",
				ObjectName:          "1111",
				ObjectID:            common.HexToHash("1"),
				BucketID:            common.HexToHash("1"),
				PayloadSize:         0,
				Visibility:          "",
				ContentType:         "",
				CreateAt:            0,
				CreateTime:          0,
				ObjectStatus:        "",
				RedundancyType:      "",
				SourceType:          "",
				Checksums:           nil,
				LockedBalance:       common.HexToHash("1"),
				Removed:             false,
				UpdateTime:          0,
				UpdateAt:            0,
				DeleteAt:            0,
				DeleteReason:        "",
				CreateTxHash:        common.HexToHash("1"),
				UpdateTxHash:        common.HexToHash("1"),
				SealTxHash:          common.HexToHash("1"),
			}, nil
		},
	).Times(1)
	m.EXPECT().ListObjectPolicies(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, permission_types.ActionType, common.Hash, int) ([]*bsdb.PermissionWithStatement, error) {
			return []*bsdb.PermissionWithStatement{
				&bsdb.PermissionWithStatement{
					Permission: bsdb.Permission{
						ID:              2,
						PrincipalType:   2,
						PrincipalValue:  "3",
						ResourceType:    "RESOURCE_TYPE_OBJECT",
						ResourceID:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
						PolicyID:        common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
						CreateTimestamp: time.Now().Unix(),
						UpdateTimestamp: time.Now().Unix(),
						ExpirationTime:  time.Now().Unix() + 100000,
						Removed:         false,
					},
					Statement: bsdb.Statement{
						ID:             2,
						PolicyID:       common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
						Effect:         "EFFECT_ALLOW",
						ActionValue:    64,
						Resources:      nil,
						ExpirationTime: time.Now().Unix() + 100000,
						LimitSize:      1,
						Removed:        false,
					},
				},
			}, nil
		},
	).Times(1)
	policies, err := a.GfSpListObjectPolicies(context.Background(), &types.GfSpListObjectPoliciesRequest{
		ObjectName: "1111",
		BucketName: "barry",
		ActionType: 6,
		Limit:      0,
		StartAfter: 0,
	})
	assert.Nil(t, err)
	assert.Equal(t, "RESOURCE_TYPE_OBJECT", policies.Policies[0].ResourceType.String())
}

func TestMetadataModular_GfSpListObjectPolicies_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetObjectByName(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, bool) (*bsdb.Object, error) {
			return &bsdb.Object{
				ID:                  1,
				Creator:             common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				Operator:            common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				Owner:               common.HexToAddress("0xe978A9160BC061f602fa083e9C68539C549A421D"),
				LocalVirtualGroupId: 0,
				BucketName:          "barry",
				ObjectName:          "1111",
				ObjectID:            common.HexToHash("1"),
				BucketID:            common.HexToHash("1"),
				PayloadSize:         0,
				Visibility:          "",
				ContentType:         "",
				CreateAt:            0,
				CreateTime:          0,
				ObjectStatus:        "",
				RedundancyType:      "",
				SourceType:          "",
				Checksums:           nil,
				LockedBalance:       common.HexToHash("1"),
				Removed:             false,
				UpdateTime:          0,
				UpdateAt:            0,
				DeleteAt:            0,
				DeleteReason:        "",
				CreateTxHash:        common.HexToHash("1"),
				UpdateTxHash:        common.HexToHash("1"),
				SealTxHash:          common.HexToHash("1"),
			}, nil
		},
	).Times(1)
	m.EXPECT().ListObjectPolicies(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, permission_types.ActionType, common.Hash, int) ([]*bsdb.Permission, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListObjectPolicies(context.Background(), &types.GfSpListObjectPoliciesRequest{
		ObjectName: "1111",
		BucketName: "barry",
		ActionType: 6,
		Limit:      0,
		StartAfter: 0,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpListObjectPolicies_Failed2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetObjectByName(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, bool) (*bsdb.Object, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListObjectPolicies(context.Background(), &types.GfSpListObjectPoliciesRequest{
		ObjectName: "1111",
		BucketName: "barry",
		ActionType: 6,
		Limit:      0,
		StartAfter: 0,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpListObjectPolicies_Failed3(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetObjectByName(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, string, bool) (*bsdb.Object, error) {
			return nil, gorm.ErrRecordNotFound
		},
	).Times(1)
	_, err := a.GfSpListObjectPolicies(context.Background(), &types.GfSpListObjectPoliciesRequest{
		ObjectName: "1111",
		BucketName: "barry",
		ActionType: 6,
		Limit:      0,
		StartAfter: 0,
	})
	assert.NotNil(t, err)
}

func TestMetadataModularGfSpVerifyMigrateGVGPermission_verifyBucketMigration_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetEventMigrationBucketByBucketID(gomock.Any()).DoAndReturn(
		func(common.Hash) (*bsdb.EventMigrationBucket, error) {
			return &bsdb.EventMigrationBucket{
				ID:             1,
				BucketID:       common.HexToHash("1"),
				Operator:       common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:     "barry",
				DstPrimarySpId: 0,
				CreateAt:       0,
				CreateTxHash:   common.HexToHash("1"),
				CreateTime:     0,
			}, nil
		},
	).Times(1)
	effect, err := a.GfSpVerifyMigrateGVGPermission(context.Background(), &types.GfSpVerifyMigrateGVGPermissionRequest{
		BucketId: 1,
		GvgId:    0,
		DstSpId:  0,
	})
	assert.Nil(t, err)
	assert.Equal(t, "EFFECT_ALLOW", effect.Effect.String())
}

func TestMetadataModularGfSpVerifyMigrateGVGPermission_verifySwapOut_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetEventSwapOutByGvgID(gomock.Any()).DoAndReturn(
		func(uint32) (*bsdb.EventSwapOut, error) {
			return &bsdb.EventSwapOut{
				ID:                         1,
				StorageProviderId:          1,
				GlobalVirtualGroupFamilyId: 1,
				GlobalVirtualGroupIds:      []uint32{1},
				SuccessorSpId:              1,
				CreateAt:                   0,
				CreateTxHash:               common.HexToHash("1"),
				CreateTime:                 0,
			}, nil
		},
	).Times(1)
	effect, err := a.GfSpVerifyMigrateGVGPermission(context.Background(), &types.GfSpVerifyMigrateGVGPermissionRequest{
		BucketId: 0,
		GvgId:    1,
		DstSpId:  1,
	})
	assert.Nil(t, err)
	assert.Equal(t, "EFFECT_ALLOW", effect.Effect.String())
}

func TestMetadataModularGfSpVerifyMigrateGVGPermission_verifyBucketMigration_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetEventMigrationBucketByBucketID(gomock.Any()).DoAndReturn(
		func(common.Hash) (*bsdb.EventMigrationBucket, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpVerifyMigrateGVGPermission(context.Background(), &types.GfSpVerifyMigrateGVGPermissionRequest{
		BucketId: 1,
		GvgId:    0,
		DstSpId:  0,
	})
	assert.NotNil(t, err)
}

func TestMetadataModularGfSpVerifyMigrateGVGPermission_verifySwapOut_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetEventSwapOutByGvgID(gomock.Any()).DoAndReturn(
		func(uint32) (*bsdb.EventSwapOut, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpVerifyMigrateGVGPermission(context.Background(), &types.GfSpVerifyMigrateGVGPermissionRequest{
		BucketId: 0,
		GvgId:    1,
		DstSpId:  1,
	})
	assert.NotNil(t, err)
}

func TestMetadataModularGfSpVerifyMigrateGVGPermission_verifyBucketMigration_DENY(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetEventMigrationBucketByBucketID(gomock.Any()).DoAndReturn(
		func(common.Hash) (*bsdb.EventMigrationBucket, error) {
			return &bsdb.EventMigrationBucket{
				ID:             1,
				BucketID:       common.HexToHash("1"),
				Operator:       common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:     "barry",
				DstPrimarySpId: 0,
				CreateAt:       0,
				CreateTxHash:   common.HexToHash("1"),
				CreateTime:     0,
			}, nil
		},
	).Times(1)
	effect, err := a.GfSpVerifyMigrateGVGPermission(context.Background(), &types.GfSpVerifyMigrateGVGPermissionRequest{
		BucketId: 1,
		GvgId:    0,
		DstSpId:  1,
	})
	assert.Nil(t, err)
	assert.Equal(t, "EFFECT_DENY", effect.Effect.String())
}

func TestMetadataModularGfSpVerifyMigrateGVGPermission_verifySwapOut_DENY(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetEventSwapOutByGvgID(gomock.Any()).DoAndReturn(
		func(uint32) (*bsdb.EventSwapOut, error) {
			return &bsdb.EventSwapOut{
				ID:                         1,
				StorageProviderId:          1,
				GlobalVirtualGroupFamilyId: 1,
				GlobalVirtualGroupIds:      []uint32{1},
				SuccessorSpId:              1,
				CreateAt:                   0,
				CreateTxHash:               common.HexToHash("1"),
				CreateTime:                 0,
			}, nil
		},
	).Times(1)
	effect, err := a.GfSpVerifyMigrateGVGPermission(context.Background(), &types.GfSpVerifyMigrateGVGPermissionRequest{
		BucketId: 0,
		GvgId:    1,
		DstSpId:  11,
	})
	assert.Nil(t, err)
	assert.Equal(t, "EFFECT_DENY", effect.Effect.String())
}
