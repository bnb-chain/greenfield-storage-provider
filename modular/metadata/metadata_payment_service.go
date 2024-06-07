package metadata

import (
	"context"

	"cosmossdk.io/math"
	"github.com/forbole/juno/v4/common"

	paymenttypes "github.com/evmos/evmos/v12/x/payment/types"
	storagetypes "github.com/evmos/evmos/v12/x/storage/types"
	"github.com/zkMeLabs/mechain-storage-provider/modular/metadata/types"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/log"
	model "github.com/zkMeLabs/mechain-storage-provider/store/bsdb"
)

// GfSpGetPaymentByBucketName get bucket payment info by a bucket name
func (r *MetadataModular) GfSpGetPaymentByBucketName(ctx context.Context, req *types.GfSpGetPaymentByBucketNameRequest) (resp *types.GfSpGetPaymentByBucketNameResponse, err error) {
	var (
		streamRecord *model.StreamRecord
		res          *paymenttypes.StreamRecord
	)

	ctx = log.Context(ctx, req)

	streamRecord, err = r.baseApp.GfBsDB().GetPaymentByBucketName(req.BucketName, req.IncludePrivate)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get payment by bucket name", "error", err)
		return
	}

	if streamRecord != nil {
		res = &paymenttypes.StreamRecord{
			Account:           streamRecord.Account.String(),
			CrudTimestamp:     streamRecord.CrudTimestamp,
			NetflowRate:       math.NewIntFromBigInt(streamRecord.NetflowRate.Raw()),
			StaticBalance:     math.NewIntFromBigInt(streamRecord.StaticBalance.Raw()),
			BufferBalance:     math.NewIntFromBigInt(streamRecord.BufferBalance.Raw()),
			LockBalance:       math.NewIntFromBigInt(streamRecord.LockBalance.Raw()),
			Status:            paymenttypes.StreamAccountStatus(paymenttypes.StreamAccountStatus_value[streamRecord.Status]),
			SettleTimestamp:   streamRecord.SettleTimestamp,
			OutFlowCount:      streamRecord.OutFlowCount,
			FrozenNetflowRate: math.NewIntFromBigInt(streamRecord.FrozenNetflowRate.Raw()),
		}
	}

	resp = &types.GfSpGetPaymentByBucketNameResponse{StreamRecord: res}
	log.CtxInfow(ctx, "succeed to get payment by bucket name")
	return resp, nil
}

// GfSpGetPaymentByBucketID get bucket payment info by a bucket id
func (r *MetadataModular) GfSpGetPaymentByBucketID(ctx context.Context, req *types.GfSpGetPaymentByBucketIDRequest) (resp *types.GfSpGetPaymentByBucketIDResponse, err error) {
	var (
		streamRecord *model.StreamRecord
		res          *paymenttypes.StreamRecord
	)

	ctx = log.Context(ctx, req)

	streamRecord, err = r.baseApp.GfBsDB().GetPaymentByBucketID(req.BucketId, req.IncludePrivate)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get payment by bucket id", "error", err)
		return
	}

	if streamRecord != nil {
		res = &paymenttypes.StreamRecord{
			Account:           streamRecord.Account.String(),
			CrudTimestamp:     streamRecord.CrudTimestamp,
			NetflowRate:       math.NewIntFromBigInt(streamRecord.NetflowRate.Raw()),
			StaticBalance:     math.NewIntFromBigInt(streamRecord.StaticBalance.Raw()),
			BufferBalance:     math.NewIntFromBigInt(streamRecord.BufferBalance.Raw()),
			LockBalance:       math.NewIntFromBigInt(streamRecord.LockBalance.Raw()),
			Status:            paymenttypes.StreamAccountStatus(paymenttypes.StreamAccountStatus_value[streamRecord.Status]),
			SettleTimestamp:   streamRecord.SettleTimestamp,
			OutFlowCount:      streamRecord.OutFlowCount,
			FrozenNetflowRate: math.NewIntFromBigInt(streamRecord.FrozenNetflowRate.Raw()),
		}
	}

	resp = &types.GfSpGetPaymentByBucketIDResponse{StreamRecord: res}
	log.CtxInfow(ctx, "succeed to get payment by bucket id")
	return resp, nil
}

// GfSpListPaymentAccountStreams list payment account streams
func (r *MetadataModular) GfSpListPaymentAccountStreams(ctx context.Context, req *types.GfSpListPaymentAccountStreamsRequest) (resp *types.GfSpListPaymentAccountStreamsResponse, err error) {
	var (
		buckets []*model.Bucket
		res     []*types.Bucket
	)

	ctx = log.Context(ctx, req)

	buckets, err = r.baseApp.GfBsDB().ListPaymentAccountStreams(common.HexToAddress(req.PaymentAccount))
	if err != nil {
		log.CtxErrorw(ctx, "failed to list payment account streams", "error", err)
		return
	}

	res = make([]*types.Bucket, len(buckets))
	for i, bucket := range buckets {
		res[i] = &types.Bucket{
			BucketInfo: &storagetypes.BucketInfo{
				Owner:                      bucket.Owner.String(),
				BucketName:                 bucket.BucketName,
				Visibility:                 storagetypes.VisibilityType(storagetypes.VisibilityType_value[bucket.Visibility]),
				Id:                         math.NewUintFromBigInt(bucket.BucketID.Big()),
				SourceType:                 storagetypes.SourceType(storagetypes.SourceType_value[bucket.SourceType]),
				CreateAt:                   bucket.CreateTime,
				PaymentAddress:             bucket.PaymentAddress.String(),
				GlobalVirtualGroupFamilyId: bucket.GlobalVirtualGroupFamilyID,
				ChargedReadQuota:           bucket.ChargedReadQuota,
				BucketStatus:               storagetypes.BucketStatus(storagetypes.BucketStatus_value[bucket.Status]),
				Tags:                       bucket.GetResourceTags(),
			},
			Removed:        bucket.Removed,
			DeleteAt:       bucket.DeleteAt,
			DeleteReason:   bucket.DeleteReason,
			Operator:       bucket.Operator.String(),
			CreateTxHash:   bucket.CreateTxHash.String(),
			UpdateTxHash:   bucket.UpdateTxHash.String(),
			UpdateAt:       bucket.UpdateAt,
			UpdateTime:     bucket.UpdateTime,
			OffChainStatus: int32(bucket.OffChainStatus),
		}
	}

	resp = &types.GfSpListPaymentAccountStreamsResponse{Buckets: res}
	log.CtxInfow(ctx, "succeed to list payment account streams")
	return resp, nil
}

// GfSpListUserPaymentAccounts list payment accounts by owner address
func (r *MetadataModular) GfSpListUserPaymentAccounts(ctx context.Context, req *types.GfSpListUserPaymentAccountsRequest) (resp *types.GfSpListUserPaymentAccountsResponse, err error) {
	var (
		accounts []*model.StreamRecordPaymentAccount
		res      []*types.PaymentAccountMeta
	)

	ctx = log.Context(ctx, req)

	accounts, err = r.baseApp.GfBsDB().ListUserPaymentAccounts(common.HexToAddress(req.AccountId))
	if err != nil {
		log.CtxErrorw(ctx, "failed to list payment accounts by owner address", "error", err)
		return
	}

	res = make([]*types.PaymentAccountMeta, len(accounts))
	for i, account := range accounts {
		res[i] = &types.PaymentAccountMeta{
			StreamRecord: &paymenttypes.StreamRecord{
				Account:           account.Account.String(),
				CrudTimestamp:     account.CrudTimestamp,
				NetflowRate:       math.NewIntFromBigInt(account.NetflowRate.Raw()),
				StaticBalance:     math.NewIntFromBigInt(account.StaticBalance.Raw()),
				BufferBalance:     math.NewIntFromBigInt(account.BufferBalance.Raw()),
				LockBalance:       math.NewIntFromBigInt(account.LockBalance.Raw()),
				Status:            paymenttypes.StreamAccountStatus(paymenttypes.StreamAccountStatus_value[account.Status]),
				SettleTimestamp:   account.SettleTimestamp,
				OutFlowCount:      account.OutFlowCount,
				FrozenNetflowRate: math.NewIntFromBigInt(account.FrozenNetflowRate.Raw()),
			},
			PaymentAccount: &types.PaymentAccount{
				Address:    account.Addr.String(),
				Owner:      account.Owner.String(),
				Refundable: account.Refundable,
				UpdateAt:   account.UpdateAt,
				UpdateTime: account.UpdateTime,
			},
		}
	}

	resp = &types.GfSpListUserPaymentAccountsResponse{PaymentAccounts: res}
	log.CtxInfow(ctx, "succeed to list payment accounts by owner address")
	return resp, nil
}
