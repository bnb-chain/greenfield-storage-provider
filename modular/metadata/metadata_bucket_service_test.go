package metadata

import (
	"context"
	"math/big"
	"net/http"
	"testing"

	"github.com/forbole/juno/v4/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

func TestErrGfSpDBWithDetail(t *testing.T) {
	testDetail := "test detail error"
	expectedErr := &gfsperrors.GfSpError{
		CodeSpace:      coremodule.MetadataModularName,
		HttpStatusCode: int32(http.StatusInternalServerError),
		InnerCode:      95202,
		Description:    testDetail,
	}

	result := ErrGfSpDBWithDetail(testDetail)

	assert.NotNil(t, result)
	assert.Equal(t, expectedErr, result)
}

func TestMetadataModular_GfSpGetUserBuckets1(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	m.EXPECT().GetUserBuckets(gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Address, bool) (*bsdb.Bucket, error) { return nil, nil },
	).Times(1)
	m.EXPECT().ListVirtualGroupFamiliesByVgfIDs(gomock.Any()).DoAndReturn(
		func([]uint32) ([]*bsdb.GlobalVirtualGroupFamily, error) { return nil, nil },
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	_, err := a.GfSpGetUserBuckets(context.Background(), &types.GfSpGetUserBucketsRequest{
		AccountId:      "0x6a45de47a2cd53084b4793fca7c1e706b9f54ed1",
		IncludeRemoved: false,
	})
	assert.Nil(t, err)
}

func TestMetadataModular_GfSpGetUserBuckets2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	m.EXPECT().GetUserBuckets(gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Address, bool) ([]*bsdb.Bucket, error) {
			return []*bsdb.Bucket{&bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "44yei",
				Visibility:                 "VISIBILITY_TYPE_PRIVATE",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000529"),
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
			}}, nil
		},
	).Times(1)
	m.EXPECT().ListVirtualGroupFamiliesByVgfIDs(gomock.Any()).DoAndReturn(
		func([]uint32) ([]*bsdb.GlobalVirtualGroupFamily, error) {
			return []*bsdb.GlobalVirtualGroupFamily{&bsdb.GlobalVirtualGroupFamily{
				ID:                         4,
				GlobalVirtualGroupFamilyId: 4,
				PrimarySpId:                4,
				GlobalVirtualGroupIds:      nil,
				VirtualPaymentAddress:      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				CreateAt:                   0,
				CreateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				CreateTime:                 0,
				UpdateAt:                   0,
				UpdateTxHash:               common.HexToHash("0x0F508E101FF83B79DF357212029B05D1FCC585B50D479FB7E68D6E1A68E8BDD4"),
				UpdateTime:                 0,
				Removed:                    false,
			}}, nil
		},
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	buckets, err := a.GfSpGetUserBuckets(context.Background(), &types.GfSpGetUserBucketsRequest{
		AccountId:      "0x11E0A11A7A01E2E757447B52FBD7152004AC699D",
		IncludeRemoved: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, "44yei", buckets.Buckets[0].BucketInfo.BucketName)
}

func TestMetadataModular_GfSpGetUserBuckets_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	m.EXPECT().GetUserBuckets(gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Address, bool) ([]*bsdb.Bucket, error) {
			return []*bsdb.Bucket{&bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "44yei",
				Visibility:                 "VISIBILITY_TYPE_PRIVATE",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000529"),
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
			}}, nil
		},
	).Times(1)
	m.EXPECT().ListVirtualGroupFamiliesByVgfIDs(gomock.Any()).DoAndReturn(
		func([]uint32) ([]*bsdb.GlobalVirtualGroupFamily, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	buckets, err := a.GfSpGetUserBuckets(context.Background(), &types.GfSpGetUserBucketsRequest{
		AccountId:      "0x11E0A11A7A01E2E757447B52FBD7152004AC699D",
		IncludeRemoved: true,
	})
	assert.NotNil(t, err)
	assert.Nil(t, buckets, nil)
}

func TestMetadataModular_GfSpGetUserBuckets_Failed2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	m.EXPECT().GetUserBuckets(gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Address, bool) ([]*bsdb.Bucket, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	buckets, err := a.GfSpGetUserBuckets(context.Background(), &types.GfSpGetUserBucketsRequest{
		AccountId:      "0x11E0A11A7A01E2E757447B52FBD7152004AC699D",
		IncludeRemoved: true,
	})
	assert.NotNil(t, err)
	assert.Nil(t, buckets, nil)
}

func TestMetadataModular_GfSpGetBucketByBucketName_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) { return nil, nil },
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	_, err := a.GfSpGetBucketByBucketName(context.Background(), &types.GfSpGetBucketByBucketNameRequest{
		BucketName:     "11111",
		IncludePrivate: true,
	})
	assert.Nil(t, err)
}

func TestMetadataModular_GfSpGetBucketByBucketName_Success2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) {
			return &bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "44yei",
				Visibility:                 "VISIBILITY_TYPE_PRIVATE",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000529"),
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
	a.baseApp.SetGfBsDB(m)
	b, err := a.GfSpGetBucketByBucketName(context.Background(), &types.GfSpGetBucketByBucketNameRequest{
		BucketName:     "44yei",
		IncludePrivate: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, "44yei", b.Bucket.BucketInfo.BucketName)
}

func TestMetadataModular_GfSpGetBucketByBucketName_CheckValidBucketName_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	_, err := a.GfSpGetBucketByBucketName(context.Background(), &types.GfSpGetBucketByBucketNameRequest{
		BucketName:     "0",
		IncludePrivate: true,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpGetBucketByBucketName_GetBucketByName_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.Bucket, error) { return nil, ErrExceedRequest },
	).Times(1)
	_, err := a.GfSpGetBucketByBucketName(context.Background(), &types.GfSpGetBucketByBucketNameRequest{
		BucketName:     "hello",
		IncludePrivate: true,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpGetBucketByBucketID_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	m.EXPECT().GetBucketByID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(int64, bool) (*bsdb.Bucket, error) { return nil, nil },
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	_, err := a.GfSpGetBucketByBucketID(context.Background(), &types.GfSpGetBucketByBucketIDRequest{
		BucketId:       1,
		IncludePrivate: true,
	})
	assert.Nil(t, err)
}

func TestMetadataModular_GfSpGetBucketByBucketID_Success2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	m.EXPECT().GetBucketByID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(int64, bool) (*bsdb.Bucket, error) {
			return &bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "44yei",
				Visibility:                 "VISIBILITY_TYPE_PRIVATE",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000529"),
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
	a.baseApp.SetGfBsDB(m)
	b, err := a.GfSpGetBucketByBucketID(context.Background(), &types.GfSpGetBucketByBucketIDRequest{
		BucketId:       529,
		IncludePrivate: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, "44yei", b.Bucket.BucketInfo.BucketName)
}

func TestMetadataModular_GfSpGetBucketByBucketID_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketByID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(int64, bool) (*bsdb.Bucket, error) { return nil, ErrExceedRequest },
	).Times(1)
	_, err := a.GfSpGetBucketByBucketID(context.Background(), &types.GfSpGetBucketByBucketIDRequest{
		BucketId:       529,
		IncludePrivate: true,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpGetUserBucketsCount_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	m.EXPECT().GetUserBucketsCount(gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Address, bool) (int64, error) {
			return 1, nil
		},
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	count, err := a.GfSpGetUserBucketsCount(context.Background(), &types.GfSpGetUserBucketsCountRequest{
		AccountId:      "0x11E0A11A7A01E2E757447B52FBD7152004AC699D",
		IncludeRemoved: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), count.Count)
}

func TestMetadataModular_GfSpGetUserBucketsCount_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetUserBucketsCount(gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Address, bool) (int64, error) {
			return 0, ErrExceedRequest
		},
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	_, err := a.GfSpGetUserBucketsCount(context.Background(), &types.GfSpGetUserBucketsCountRequest{
		AccountId:      "0x11E0A11A7A01E2E757447B52FBD7152004AC699D",
		IncludeRemoved: true,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpListExpiredBucketsBySp_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListExpiredBucketsBySp(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(int64, uint32, int64) ([]*bsdb.Bucket, error) {
			return []*bsdb.Bucket{&bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "44yei",
				Visibility:                 "VISIBILITY_TYPE_PRIVATE",
				BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000529"),
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
			}}, nil
		},
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	buckets, err := a.GfSpListExpiredBucketsBySp(context.Background(), &types.GfSpListExpiredBucketsBySpRequest{
		CreateAt:    0,
		PrimarySpId: 0,
		Limit:       0,
	})
	assert.Nil(t, err)
	assert.Equal(t, "44yei", buckets.Buckets[0].BucketInfo.BucketName)
}

func TestMetadataModular_GfSpListExpiredBucketsBySp_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListExpiredBucketsBySp(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(int64, uint32, int64) ([]*bsdb.Bucket, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	_, err := a.GfSpListExpiredBucketsBySp(context.Background(), &types.GfSpListExpiredBucketsBySpRequest{
		CreateAt:    0,
		PrimarySpId: 0,
		Limit:       0,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpGetBucketMeta_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	big, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	m.EXPECT().GetBucketMetaByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.BucketFullMeta, error) {
			return &bsdb.BucketFullMeta{
				Bucket: bsdb.Bucket{
					ID:                         848,
					Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					BucketName:                 "44yei",
					Visibility:                 "VISIBILITY_TYPE_PRIVATE",
					BucketID:                   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000529"),
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
				},
				StreamRecord: bsdb.StreamRecord{
					ID:                0,
					Account:           common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					CrudTimestamp:     0,
					NetflowRate:       (*common.Big)(big),
					StaticBalance:     (*common.Big)(big),
					BufferBalance:     (*common.Big)(big),
					LockBalance:       (*common.Big)(big),
					Status:            "STREAM_ACCOUNT_STATUS_ACTIVE",
					SettleTimestamp:   0,
					OutFlowCount:      0,
					FrozenNetflowRate: (*common.Big)(big),
				},
			}, nil
		},
	).Times(1)
	m.EXPECT().ListVirtualGroupFamiliesByVgfIDs(gomock.Any()).DoAndReturn(
		func([]uint32) ([]*bsdb.GlobalVirtualGroupFamily, error) {
			return []*bsdb.GlobalVirtualGroupFamily{
				&bsdb.GlobalVirtualGroupFamily{
					ID:                         1,
					GlobalVirtualGroupFamilyId: 1,
					PrimarySpId:                0,
					GlobalVirtualGroupIds:      nil,
					VirtualPaymentAddress:      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					CreateAt:                   0,
					CreateTxHash:               common.HexToHash("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					CreateTime:                 0,
					UpdateAt:                   0,
					UpdateTxHash:               common.HexToHash("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					UpdateTime:                 0,
					Removed:                    false,
				},
			}, nil
		},
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	buckets, err := a.GfSpGetBucketMeta(context.Background(), &types.GfSpGetBucketMetaRequest{
		BucketName:     "44yei",
		IncludePrivate: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, "44yei", buckets.Bucket.BucketInfo.BucketName)
}

func TestMetadataModular_GfSpGetBucketMeta_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetBucketMetaByName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.BucketFullMeta, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	_, err := a.GfSpGetBucketMeta(context.Background(), &types.GfSpGetBucketMetaRequest{
		BucketName:     "44yei",
		IncludePrivate: true,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpListBucketsByIDs_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListBucketsByIDs(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Bucket, error) {
			return []*bsdb.Bucket{&bsdb.Bucket{
				ID:                         848,
				Owner:                      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				BucketName:                 "44yei",
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
			}}, nil
		},
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	buckets, err := a.GfSpListBucketsByIDs(context.Background(), &types.GfSpListBucketsByIDsRequest{
		BucketIds:      []uint64{1},
		IncludeRemoved: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, "44yei", buckets.Buckets[1].BucketInfo.BucketName)
}

func TestMetadataModular_GfSpListBucketsByIDs_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListBucketsByIDs(gomock.Any(), gomock.Any()).DoAndReturn(
		func([]common.Hash, bool) ([]*bsdb.Bucket, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	_, err := a.GfSpListBucketsByIDs(context.Background(), &types.GfSpListBucketsByIDsRequest{
		BucketIds:      []uint64{848},
		IncludeRemoved: true,
	})
	assert.NotNil(t, err)
}
