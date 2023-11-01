package metadata

import (
	"context"
	"math/big"
	"testing"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	"github.com/forbole/juno/v4/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGfSpPrimarySpIncomeDetails_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	big, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	m.EXPECT().GetPrimarySPStreamRecordBySpID(gomock.Any()).DoAndReturn(
		func(spID any) ([]*bsdb.PrimarySpIncomeMeta, error) {
			return []*bsdb.PrimarySpIncomeMeta{&bsdb.PrimarySpIncomeMeta{
				GlobalVirtualGroupFamilyId: 48,
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
			}}, nil
		},
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	resp, err := a.GfSpPrimarySpIncomeDetails(context.Background(), &types.GfSpPrimarySpIncomeDetailsRequest{
		SpId: 1,
	})
	assert.Nil(t, err)
	assert.Equal(t, uint32(48), resp.PrimarySpIncomeDetails[0].VgfId)
	assert.Equal(t, math.NewIntFromBigInt((*common.Big)(big).Raw()), resp.PrimarySpIncomeDetails[0].StreamRecord.NetflowRate)
}

func TestGfSpPrimarySpIncomeDetails_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetPrimarySPStreamRecordBySpID(gomock.Any()).DoAndReturn(
		func(spID any) ([]*bsdb.PrimarySpIncomeMeta, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	_, err := a.GfSpPrimarySpIncomeDetails(context.Background(), &types.GfSpPrimarySpIncomeDetailsRequest{
		SpId: 1,
	})
	assert.NotNil(t, err)
}

func TestGfSpSecondarySpIncomeDetails_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	big, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	m.EXPECT().GetSecondarySPStreamRecordBySpID(gomock.Any()).DoAndReturn(
		func(spID any) ([]*bsdb.SecondarySpIncomeMeta, error) {
			return []*bsdb.SecondarySpIncomeMeta{&bsdb.SecondarySpIncomeMeta{
				GlobalVirtualGroupId: 4,
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
			}}, nil
		},
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	resp, err := a.GfSpSecondarySpIncomeDetails(context.Background(), &types.GfSpSecondarySpIncomeDetailsRequest{
		SpId: 1,
	})
	assert.Nil(t, err)
	assert.Equal(t, uint32(4), resp.SecondarySpIncomeDetails[0].GvgId)
	assert.Equal(t, math.NewIntFromBigInt((*common.Big)(big).Raw()), resp.SecondarySpIncomeDetails[0].StreamRecord.NetflowRate)
}

func TestGfSpSecondarySpIncomeDetails_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetSecondarySPStreamRecordBySpID(gomock.Any()).DoAndReturn(
		func(spID any) ([]*bsdb.SecondarySpIncomeMeta, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	a.baseApp.SetGfBsDB(m)
	_, err := a.GfSpSecondarySpIncomeDetails(context.Background(), &types.GfSpSecondarySpIncomeDetailsRequest{
		SpId: 1,
	})
	assert.NotNil(t, err)
}
