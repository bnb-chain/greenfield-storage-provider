package metadata

import (
	"context"
	"math/big"
	"testing"

	"github.com/forbole/juno/v4/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

func TestMetadataModular_GfSpGetPaymentByBucketName_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	big, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	m.EXPECT().GetPaymentByBucketName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.StreamRecord, error) {
			return &bsdb.StreamRecord{
				ID:                0,
				Account:           common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				CrudTimestamp:     1000,
				NetflowRate:       (*common.Big)(big),
				StaticBalance:     (*common.Big)(big),
				BufferBalance:     (*common.Big)(big),
				LockBalance:       (*common.Big)(big),
				Status:            "STREAM_ACCOUNT_STATUS_ACTIVE",
				SettleTimestamp:   0,
				OutFlowCount:      0,
				FrozenNetflowRate: (*common.Big)(big),
			}, nil
		},
	).Times(1)
	record, err := a.GfSpGetPaymentByBucketName(context.Background(), &types.GfSpGetPaymentByBucketNameRequest{
		BucketName:     "barry",
		IncludePrivate: false,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(1000), record.StreamRecord.CrudTimestamp)
}

func TestMetadataModular_GfSpGetPaymentByBucketName_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetPaymentByBucketName(gomock.Any(), gomock.Any()).DoAndReturn(
		func(string, bool) (*bsdb.StreamRecord, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpGetPaymentByBucketName(context.Background(), &types.GfSpGetPaymentByBucketNameRequest{
		BucketName:     "barry",
		IncludePrivate: false,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpGetPaymentByBucketID_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	big, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	m.EXPECT().GetPaymentByBucketID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(int64, bool) (*bsdb.StreamRecord, error) {
			return &bsdb.StreamRecord{
				ID:                0,
				Account:           common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
				CrudTimestamp:     1000,
				NetflowRate:       (*common.Big)(big),
				StaticBalance:     (*common.Big)(big),
				BufferBalance:     (*common.Big)(big),
				LockBalance:       (*common.Big)(big),
				Status:            "STREAM_ACCOUNT_STATUS_ACTIVE",
				SettleTimestamp:   0,
				OutFlowCount:      0,
				FrozenNetflowRate: (*common.Big)(big),
			}, nil
		},
	).Times(1)
	record, err := a.GfSpGetPaymentByBucketID(context.Background(), &types.GfSpGetPaymentByBucketIDRequest{
		BucketId:       1,
		IncludePrivate: false,
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(1000), record.StreamRecord.CrudTimestamp)
}

func TestMetadataModular_GfSpGetPaymentByBucketID_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().GetPaymentByBucketID(gomock.Any(), gomock.Any()).DoAndReturn(
		func(int64, bool) (*bsdb.StreamRecord, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpGetPaymentByBucketID(context.Background(), &types.GfSpGetPaymentByBucketIDRequest{
		BucketId:       1,
		IncludePrivate: false,
	})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpListPaymentAccountStreams_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListPaymentAccountStreams(gomock.Any()).DoAndReturn(
		func(common.Address) ([]*bsdb.Bucket, error) {
			return []*bsdb.Bucket{
				&bsdb.Bucket{
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
			}, nil
		},
	).Times(1)
	buckets, err := a.GfSpListPaymentAccountStreams(context.Background(), &types.GfSpListPaymentAccountStreamsRequest{PaymentAccount: "0x11E0A11A7A01E2E757447B52FBD7152004AC699D"})
	assert.Nil(t, err)
	assert.Equal(t, "44yei", buckets.Buckets[0].BucketInfo.BucketName)
}

func TestMetadataModular_GfSpListPaymentAccountStreams_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListPaymentAccountStreams(gomock.Any()).DoAndReturn(
		func(common.Address) ([]*bsdb.Bucket, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListPaymentAccountStreams(context.Background(), &types.GfSpListPaymentAccountStreamsRequest{PaymentAccount: "0x11E0A11A7A01E2E757447B52FBD7152004AC699D"})
	assert.NotNil(t, err)
}

func TestMetadataModular_GfSpListUserPaymentAccounts_Success(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	big, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	m.EXPECT().ListUserPaymentAccounts(gomock.Any()).DoAndReturn(
		func(common.Address) ([]*bsdb.StreamRecordPaymentAccount, error) {
			return []*bsdb.StreamRecordPaymentAccount{
				&bsdb.StreamRecordPaymentAccount{
					PaymentAccount: bsdb.PaymentAccount{
						ID:         0,
						Addr:       common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
						Owner:      common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
						Refundable: false,
						UpdateAt:   0,
						UpdateTime: 0,
					},
					StreamRecord: bsdb.StreamRecord{
						ID:                0,
						Account:           common.HexToAddress("0x11E0A11A7A01E2E757447B52FBD7152004AC699D"),
						CrudTimestamp:     1000,
						NetflowRate:       (*common.Big)(big),
						StaticBalance:     (*common.Big)(big),
						BufferBalance:     (*common.Big)(big),
						LockBalance:       (*common.Big)(big),
						Status:            "STREAM_ACCOUNT_STATUS_ACTIVE",
						SettleTimestamp:   0,
						OutFlowCount:      0,
						FrozenNetflowRate: (*common.Big)(big),
					},
				},
			}, nil
		},
	).Times(1)
	records, err := a.GfSpListUserPaymentAccounts(context.Background(), &types.GfSpListUserPaymentAccountsRequest{
		AccountId: "0x11E0A11A7A01E2E757447B52FBD7152004AC699D",
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(1000), records.PaymentAccounts[0].StreamRecord.CrudTimestamp)
}

func TestMetadataModular_GfSpListUserPaymentAccounts_Failed(t *testing.T) {
	a := setup(t)
	ctrl := gomock.NewController(t)
	m := bsdb.NewMockBSDB(ctrl)
	a.baseApp.SetGfBsDB(m)
	m.EXPECT().ListUserPaymentAccounts(gomock.Any()).DoAndReturn(
		func(common.Address) ([]*bsdb.StreamRecord, error) {
			return nil, ErrExceedRequest
		},
	).Times(1)
	_, err := a.GfSpListUserPaymentAccounts(context.Background(), &types.GfSpListUserPaymentAccountsRequest{
		AccountId: "0x11E0A11A7A01E2E757447B52FBD7152004AC699D",
	})
	assert.NotNil(t, err)
}
