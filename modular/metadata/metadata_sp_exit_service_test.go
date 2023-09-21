package metadata

import (
	"context"
	"math/big"
	"testing"

	"github.com/forbole/juno/v4/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

func TestMetadataModularGfSpListVirtualGroupFamiliesBySpID_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListVirtualGroupFamiliesBySpID(gomock.Any()).DoAndReturn(
		func(uint32) ([]*bsdb.GlobalVirtualGroupFamily, error) {
			return []*bsdb.GlobalVirtualGroupFamily{
				&bsdb.GlobalVirtualGroupFamily{
					ID:                         1,
					GlobalVirtualGroupFamilyId: 1,
					PrimarySpId:                1,
					GlobalVirtualGroupIds:      []uint32{1},
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
	vgf, err := a.GfSpListVirtualGroupFamiliesBySpID(context.Background(), &types.GfSpListVirtualGroupFamiliesBySpIDRequest{
		SpId: 1,
	})
	assert.Nil(t, err)
	assert.Equal(t, []uint32{1}, vgf.GlobalVirtualGroupFamilies[0].GlobalVirtualGroupIds)
}

func TestMetadataModularGfSpListVirtualGroupFamiliesBySpID_Fail(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListVirtualGroupFamiliesBySpID(gomock.Any()).DoAndReturn(
		func(uint32) ([]*bsdb.GlobalVirtualGroupFamily, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListVirtualGroupFamiliesBySpID(context.Background(), &types.GfSpListVirtualGroupFamiliesBySpIDRequest{
		SpId: 1,
	})
	assert.NotNil(t, err)
}

func TestMetadataModularGfSpGetGlobalVirtualGroupByGvgID_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetGlobalVirtualGroupByGvgID(gomock.Any()).DoAndReturn(
		func(uint32) (*bsdb.GlobalVirtualGroup, error) {
			return &bsdb.GlobalVirtualGroup{
				ID:                    1,
				GlobalVirtualGroupId:  1,
				FamilyId:              1,
				PrimarySpId:           1,
				SecondarySpIds:        []uint32{1},
				StoredSize:            1,
				VirtualPaymentAddress: common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				TotalDeposit:          nil,
				CreateAt:              0,
				CreateTxHash:          common.HexToHash("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				CreateTime:            0,
				UpdateAt:              0,
				UpdateTxHash:          common.HexToHash("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				UpdateTime:            0,
				Removed:               false,
			}, nil
		},
	).Times(1)
	gvg, err := a.GfSpGetGlobalVirtualGroupByGvgID(context.Background(), &types.GfSpGetGlobalVirtualGroupByGvgIDRequest{
		GvgId: 1,
	})
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), gvg.GlobalVirtualGroup.FamilyId)
}

func TestMetadataModularGfSpGetGlobalVirtualGroupByGvgID_Fail(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetGlobalVirtualGroupByGvgID(gomock.Any()).DoAndReturn(
		func(uint32) (*bsdb.GlobalVirtualGroup, error) {
			return nil, ErrDanglingPointer
		},
	).Times(1)
	_, err := a.GfSpGetGlobalVirtualGroupByGvgID(context.Background(), &types.GfSpGetGlobalVirtualGroupByGvgIDRequest{
		GvgId: 1,
	})
	assert.NotNil(t, err)
}

func TestMetadataModularGfSpGetVirtualGroupFamily_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetVirtualGroupFamiliesByVgfID(gomock.Any()).DoAndReturn(
		func(uint32) (*bsdb.GlobalVirtualGroupFamily, error) {
			return &bsdb.GlobalVirtualGroupFamily{
				ID:                         1,
				GlobalVirtualGroupFamilyId: 1,
				PrimarySpId:                1,
				GlobalVirtualGroupIds:      []uint32{1},
				VirtualPaymentAddress:      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				CreateAt:                   0,
				CreateTxHash:               common.HexToHash("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				CreateTime:                 0,
				UpdateAt:                   0,
				UpdateTxHash:               common.HexToHash("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				UpdateTime:                 0,
				Removed:                    false,
			}, nil
		},
	).Times(1)
	vgf, err := a.GfSpGetVirtualGroupFamily(context.Background(), &types.GfSpGetVirtualGroupFamilyRequest{
		VgfId: 1,
	})
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), vgf.Vgf.Id)
}

func TestMetadataModularGfSpGetVirtualGroupFamily_Fail(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetVirtualGroupFamiliesByVgfID(gomock.Any()).DoAndReturn(
		func(uint32) (*bsdb.GlobalVirtualGroupFamily, error) {
			return nil, ErrDanglingPointer
		},
	).Times(1)
	_, err := a.GfSpGetVirtualGroupFamily(context.Background(), &types.GfSpGetVirtualGroupFamilyRequest{
		VgfId: 1,
	})
	assert.NotNil(t, err)
}

func TestMetadataModularGfSpGetGlobalVirtualGroup_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetGvgByBucketAndLvgID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32) (*bsdb.GlobalVirtualGroup, error) {
			return &bsdb.GlobalVirtualGroup{
				ID:                    1,
				GlobalVirtualGroupId:  1,
				FamilyId:              1,
				PrimarySpId:           1,
				SecondarySpIds:        []uint32{1},
				StoredSize:            1,
				VirtualPaymentAddress: common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				TotalDeposit:          nil,
				CreateAt:              0,
				CreateTxHash:          common.HexToHash("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				CreateTime:            0,
				UpdateAt:              0,
				UpdateTxHash:          common.HexToHash("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				UpdateTime:            0,
				Removed:               false,
			}, nil
		},
	).Times(1)
	gvg, err := a.GfSpGetGlobalVirtualGroup(context.Background(), &types.GfSpGetGlobalVirtualGroupRequest{
		BucketId: 1,
		LvgId:    1,
	})
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), gvg.Gvg.Id)
}

func TestMetadataModularGfSpGetGlobalVirtualGroup_Fail(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetGvgByBucketAndLvgID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(common.Hash, uint32) (*bsdb.GlobalVirtualGroup, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpGetGlobalVirtualGroup(context.Background(), &types.GfSpGetGlobalVirtualGroupRequest{
		BucketId: 1,
		LvgId:    1,
	})
	assert.NotNil(t, err)
}

func TestMetadataModularGfSpListMigrateBucketEvents_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 100000, nil
		},
	).Times(1)
	m.EXPECT().ListMigrateBucketEvents(gomock.Any(), gomock.Any()).DoAndReturn(
		func(spID uint32, filters ...func(*gorm.DB) *gorm.DB) ([]*bsdb.EventMigrationBucket, []*bsdb.EventCompleteMigrationBucket, []*bsdb.EventCancelMigrationBucket, error) {
			return []*bsdb.EventMigrationBucket{
					&bsdb.EventMigrationBucket{
						ID:             1,
						BucketID:       common.HexToHash("1"),
						Operator:       common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
						BucketName:     "",
						DstPrimarySpId: 1,
						CreateAt:       0,
						CreateTxHash:   common.HexToHash("1"),
						CreateTime:     0,
					},
				},
				[]*bsdb.EventCompleteMigrationBucket{
					&bsdb.EventCompleteMigrationBucket{
						ID:                         0,
						BucketID:                   common.HexToHash("1"),
						Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
						BucketName:                 "",
						GlobalVirtualGroupFamilyId: 0,
						SrcPrimarySpId:             1,
						CreateAt:                   2,
						CreateTxHash:               common.HexToHash("1"),
						CreateTime:                 2,
					},
				},
				[]*bsdb.EventCancelMigrationBucket{
					&bsdb.EventCancelMigrationBucket{
						ID:           0,
						BucketID:     common.HexToHash("1"),
						Operator:     common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
						BucketName:   "",
						CreateAt:     1,
						CreateTxHash: common.HexToHash("1"),
						CreateTime:   1,
					},
				}, nil
		},
	).Times(1)
	res, err := a.GfSpListMigrateBucketEvents(context.Background(), &types.GfSpListMigrateBucketEventsRequest{
		BlockId: 999,
		SpId:    1,
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(res.Events))
}

func TestMetadataModularGfSpListMigrateBucketEvents_Success2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 100000, nil
		},
	).Times(1)
	m.EXPECT().ListMigrateBucketEvents(gomock.Any(), gomock.Any()).DoAndReturn(
		func(spID uint32, filters ...func(*gorm.DB) *gorm.DB) ([]*bsdb.EventMigrationBucket, []*bsdb.EventCompleteMigrationBucket, []*bsdb.EventCancelMigrationBucket, error) {
			return []*bsdb.EventMigrationBucket{
					&bsdb.EventMigrationBucket{
						ID:             1,
						BucketID:       common.HexToHash("1"),
						Operator:       common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
						BucketName:     "",
						DstPrimarySpId: 1,
						CreateAt:       0,
						CreateTxHash:   common.HexToHash("1"),
						CreateTime:     0,
					},
				}, nil,
				[]*bsdb.EventCancelMigrationBucket{
					&bsdb.EventCancelMigrationBucket{
						ID:           0,
						BucketID:     common.HexToHash("1"),
						Operator:     common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
						BucketName:   "",
						CreateAt:     1,
						CreateTxHash: common.HexToHash("1"),
						CreateTime:   1,
					},
				}, nil
		},
	).Times(1)
	res, err := a.GfSpListMigrateBucketEvents(context.Background(), &types.GfSpListMigrateBucketEventsRequest{
		BlockId: 999,
		SpId:    1,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(res.Events))
}

func TestMetadataModularGfSpListMigrateBucketEvents_Fail(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 100000, nil
		},
	).Times(1)
	m.EXPECT().ListMigrateBucketEvents(gomock.Any(), gomock.Any()).DoAndReturn(
		func(spID uint32, filters ...func(*gorm.DB) *gorm.DB) ([]*bsdb.EventMigrationBucket, []*bsdb.EventCompleteMigrationBucket, []*bsdb.EventCancelMigrationBucket, error) {
			return nil, nil, nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListMigrateBucketEvents(context.Background(), &types.GfSpListMigrateBucketEventsRequest{
		BlockId: 999,
		SpId:    1,
	})
	assert.NotNil(t, err)
}

func TestMetadataModularGfSpListMigrateBucketEvents_Fail2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 100000, nil
		},
	).Times(1)
	_, err := a.GfSpListMigrateBucketEvents(context.Background(), &types.GfSpListMigrateBucketEventsRequest{
		BlockId: 1000001,
		SpId:    1,
	})
	assert.NotNil(t, err)
}

func TestMetadataModularGfSpListMigrateBucketEvents_Fail3(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 0, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListMigrateBucketEvents(context.Background(), &types.GfSpListMigrateBucketEventsRequest{
		BlockId: 1000001,
		SpId:    1,
	})
	assert.NotNil(t, err)
}

func TestMetadataModularGfSpGetSPMigratingBucketNumber_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListMigrateBucketEvents(gomock.Any(), gomock.Any()).DoAndReturn(
		func(spID uint32, filters ...func(*gorm.DB) *gorm.DB) ([]*bsdb.EventMigrationBucket, []*bsdb.EventCompleteMigrationBucket, []*bsdb.EventCancelMigrationBucket, error) {
			return []*bsdb.EventMigrationBucket{
					&bsdb.EventMigrationBucket{
						ID:             1,
						BucketID:       common.HexToHash("1"),
						Operator:       common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
						BucketName:     "",
						DstPrimarySpId: 1,
						CreateAt:       0,
						CreateTxHash:   common.HexToHash("1"),
						CreateTime:     0,
					},
				},
				[]*bsdb.EventCompleteMigrationBucket{
					&bsdb.EventCompleteMigrationBucket{
						ID:                         0,
						BucketID:                   common.HexToHash("1"),
						Operator:                   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
						BucketName:                 "",
						GlobalVirtualGroupFamilyId: 0,
						SrcPrimarySpId:             1,
						CreateAt:                   2,
						CreateTxHash:               common.HexToHash("1"),
						CreateTime:                 2,
					},
				},
				[]*bsdb.EventCancelMigrationBucket{
					&bsdb.EventCancelMigrationBucket{
						ID:           0,
						BucketID:     common.HexToHash("1"),
						Operator:     common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
						BucketName:   "",
						CreateAt:     1,
						CreateTxHash: common.HexToHash("1"),
						CreateTime:   1,
					},
				}, nil
		},
	).Times(1)
	count, err := a.GfSpGetSPMigratingBucketNumber(context.Background(), &types.GfSpGetSPMigratingBucketNumberRequest{
		SpId: 1,
	})
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), count.Count)
}

func TestMetadataModularGfSpGetSPMigratingBucketNumber_Success2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListMigrateBucketEvents(gomock.Any(), gomock.Any()).DoAndReturn(
		func(spID uint32, filters ...func(*gorm.DB) *gorm.DB) ([]*bsdb.EventMigrationBucket, []*bsdb.EventCompleteMigrationBucket, []*bsdb.EventCancelMigrationBucket, error) {
			return []*bsdb.EventMigrationBucket{
					&bsdb.EventMigrationBucket{
						ID:             1,
						BucketID:       common.HexToHash("1"),
						Operator:       common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
						BucketName:     "",
						DstPrimarySpId: 1,
						CreateAt:       0,
						CreateTxHash:   common.HexToHash("1"),
						CreateTime:     0,
					},
				}, nil,
				[]*bsdb.EventCancelMigrationBucket{
					&bsdb.EventCancelMigrationBucket{
						ID:           0,
						BucketID:     common.HexToHash("1"),
						Operator:     common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
						BucketName:   "",
						CreateAt:     1,
						CreateTxHash: common.HexToHash("1"),
						CreateTime:   1,
					},
				}, nil
		},
	).Times(1)
	count, err := a.GfSpGetSPMigratingBucketNumber(context.Background(), &types.GfSpGetSPMigratingBucketNumberRequest{
		SpId: 1,
	})
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), count.Count)
}

func TestMetadataModularGfSpGetSPMigratingBucketNumber_Success3(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListMigrateBucketEvents(gomock.Any(), gomock.Any()).DoAndReturn(
		func(spID uint32, filters ...func(*gorm.DB) *gorm.DB) ([]*bsdb.EventMigrationBucket, []*bsdb.EventCompleteMigrationBucket, []*bsdb.EventCancelMigrationBucket, error) {
			return []*bsdb.EventMigrationBucket{
				&bsdb.EventMigrationBucket{
					ID:             1,
					BucketID:       common.HexToHash("1"),
					Operator:       common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					BucketName:     "",
					DstPrimarySpId: 1,
					CreateAt:       0,
					CreateTxHash:   common.HexToHash("1"),
					CreateTime:     0,
				},
			}, nil, nil, nil
		},
	).Times(1)
	count, err := a.GfSpGetSPMigratingBucketNumber(context.Background(), &types.GfSpGetSPMigratingBucketNumberRequest{
		SpId: 1,
	})
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), count.Count)
}

func TestMetadataModularGfSpGetSPMigratingBucketNumber_Fail(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListMigrateBucketEvents(gomock.Any(), gomock.Any()).DoAndReturn(
		func(spID uint32, filters ...func(*gorm.DB) *gorm.DB) ([]*bsdb.EventMigrationBucket, []*bsdb.EventCompleteMigrationBucket, []*bsdb.EventCancelMigrationBucket, error) {
			return nil, nil, nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpGetSPMigratingBucketNumber(context.Background(), &types.GfSpGetSPMigratingBucketNumberRequest{
		SpId: 1,
	})
	assert.NotNil(t, err)
}

func TestMetadataModularGfSpListSwapOutEvents_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 100000, nil
		},
	).Times(1)
	m.EXPECT().ListSwapOutEvents(gomock.Any(), gomock.Any()).DoAndReturn(
		func(blockID uint64, spID uint32) ([]*bsdb.EventSwapOut, []*bsdb.EventCompleteSwapOut, []*bsdb.EventCancelSwapOut, error) {
			return []*bsdb.EventSwapOut{
					&bsdb.EventSwapOut{
						ID:                         1,
						StorageProviderId:          1,
						GlobalVirtualGroupFamilyId: 1,
						GlobalVirtualGroupIds:      []uint32{1},
						SuccessorSpId:              1,
						CreateAt:                   1,
						CreateTxHash:               common.HexToHash("1"),
						CreateTime:                 0,
					},
				},
				[]*bsdb.EventCompleteSwapOut{
					&bsdb.EventCompleteSwapOut{
						ID:                         1,
						StorageProviderId:          1,
						SrcStorageProviderId:       1,
						GlobalVirtualGroupFamilyId: 1,
						GlobalVirtualGroupIds:      []uint32{1},
						CreateAt:                   3,
						CreateTxHash:               common.HexToHash("1"),
						CreateTime:                 3,
					},
				},
				[]*bsdb.EventCancelSwapOut{
					&bsdb.EventCancelSwapOut{
						ID:                         1,
						StorageProviderId:          1,
						SuccessorSpId:              1,
						GlobalVirtualGroupFamilyId: 1,
						GlobalVirtualGroupIds:      []uint32{1},
						CreateAt:                   2,
						CreateTxHash:               common.HexToHash("1"),
						CreateTime:                 2,
					},
				}, nil
		},
	).Times(1)
	res, err := a.GfSpListSwapOutEvents(context.Background(), &types.GfSpListSwapOutEventsRequest{
		BlockId: 999,
		SpId:    1,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(res.Events))
}

func TestMetadataModularGfSpGfSpListSwapOutEvents_Success2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 100000, nil
		},
	).Times(1)
	m.EXPECT().ListSwapOutEvents(gomock.Any(), gomock.Any()).DoAndReturn(
		func(blockID uint64, spID uint32) ([]*bsdb.EventSwapOut, []*bsdb.EventCompleteSwapOut, []*bsdb.EventCancelSwapOut, error) {
			return []*bsdb.EventSwapOut{
					&bsdb.EventSwapOut{
						ID:                         1,
						StorageProviderId:          1,
						GlobalVirtualGroupFamilyId: 1,
						GlobalVirtualGroupIds:      []uint32{1},
						SuccessorSpId:              1,
						CreateAt:                   1,
						CreateTxHash:               common.HexToHash("1"),
						CreateTime:                 0,
					},
				}, nil,
				[]*bsdb.EventCancelSwapOut{
					&bsdb.EventCancelSwapOut{
						ID:                         1,
						StorageProviderId:          1,
						SuccessorSpId:              1,
						GlobalVirtualGroupFamilyId: 1,
						GlobalVirtualGroupIds:      []uint32{1},
						CreateAt:                   2,
						CreateTxHash:               common.HexToHash("1"),
						CreateTime:                 2,
					},
				}, nil
		},
	).Times(1)
	res, err := a.GfSpListSwapOutEvents(context.Background(), &types.GfSpListSwapOutEventsRequest{
		BlockId: 999,
		SpId:    1,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(res.Events))
}

func TestMetadataModularGfSpGfSpListSwapOutEvents_Fail(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 100000, nil
		},
	).Times(1)
	m.EXPECT().ListSwapOutEvents(gomock.Any(), gomock.Any()).DoAndReturn(
		func(blockID uint64, spID uint32) ([]*bsdb.EventSwapOut, []*bsdb.EventCompleteSwapOut, []*bsdb.EventCancelSwapOut, error) {
			return nil, nil, nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListSwapOutEvents(context.Background(), &types.GfSpListSwapOutEventsRequest{
		BlockId: 999,
		SpId:    1,
	})

	assert.NotNil(t, err)
}

func TestMetadataModularGfSpGfSpListSwapOutEvents_Fail2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 100000, nil
		},
	).Times(1)
	_, err := a.GfSpListSwapOutEvents(context.Background(), &types.GfSpListSwapOutEventsRequest{
		BlockId: 1000001,
		SpId:    1,
	})
	assert.NotNil(t, err)
}

func TestMetadataModularGfSpListSwapOutEvents_Fail3(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 0, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListSwapOutEvents(context.Background(), &types.GfSpListSwapOutEventsRequest{
		BlockId: 1000001,
		SpId:    1,
	})
	assert.NotNil(t, err)
}

func TestMetadataModularGfSpListSpExitEvents_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 100000, nil
		},
	).Times(1)
	big, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	m.EXPECT().ListSpExitEvents(gomock.Any(), gomock.Any()).DoAndReturn(
		func(blockID uint64, spID uint32) (*bsdb.EventStorageProviderExit, *bsdb.EventCompleteStorageProviderExit, error) {
			return &bsdb.EventStorageProviderExit{
					ID:                1,
					StorageProviderId: 1,
					OperatorAddress:   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					CreateAt:          0,
					CreateTxHash:      common.HexToHash("1"),
					CreateTime:        0,
				}, &bsdb.EventCompleteStorageProviderExit{
					ID:                1,
					StorageProviderId: 1,
					TotalDeposit:      (*common.Big)(big),
					OperatorAddress:   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
					CreateAt:          1,
					CreateTxHash:      common.HexToHash("1"),
					CreateTime:        1,
				}, nil
		},
	).Times(1)
	res, err := a.GfSpListSpExitEvents(context.Background(), &types.GfSpListSpExitEventsRequest{
		BlockId: 999,
		SpId:    1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, res.Events.Event)
}

func TestMetadataModularGfSpGfSpGfSpListSpExitEvents_Success2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 100000, nil
		},
	).Times(1)
	m.EXPECT().ListSpExitEvents(gomock.Any(), gomock.Any()).DoAndReturn(
		func(blockID uint64, spID uint32) (*bsdb.EventStorageProviderExit, *bsdb.EventCompleteStorageProviderExit, error) {
			return &bsdb.EventStorageProviderExit{
				ID:                1,
				StorageProviderId: 1,
				OperatorAddress:   common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				CreateAt:          0,
				CreateTxHash:      common.HexToHash("1"),
				CreateTime:        0,
			}, nil, nil
		},
	).Times(1)
	res, err := a.GfSpListSpExitEvents(context.Background(), &types.GfSpListSpExitEventsRequest{
		BlockId: 999,
		SpId:    1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, res.Events.Event)
}

func TestMetadataModularGfSpListSpExitEvents_Fail(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 100000, nil
		},
	).Times(1)
	m.EXPECT().ListSpExitEvents(gomock.Any(), gomock.Any()).DoAndReturn(
		func(blockID uint64, spID uint32) (*bsdb.EventStorageProviderExit, *bsdb.EventCompleteStorageProviderExit, error) {
			return nil, nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListSpExitEvents(context.Background(), &types.GfSpListSpExitEventsRequest{
		BlockId: 999,
		SpId:    1,
	})

	assert.NotNil(t, err)
}

func TestMetadataModularGfSpListSpExitEvents_Fail2(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 100000, nil
		},
	).Times(1)
	_, err := a.GfSpListSpExitEvents(context.Background(), &types.GfSpListSpExitEventsRequest{
		BlockId: 1000001,
		SpId:    1,
	})
	assert.NotNil(t, err)
}

func TestMetadataModularGfSpGfSpListSpExitEvents_Fail3(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetLatestBlockNumber().DoAndReturn(
		func() (int64, error) {
			return 0, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListSpExitEvents(context.Background(), &types.GfSpListSpExitEventsRequest{
		BlockId: 1000001,
		SpId:    1,
	})
	assert.NotNil(t, err)
}
