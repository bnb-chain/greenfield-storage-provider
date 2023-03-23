package service

import (
	"context"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/x/payment/types"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	metatypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
	model "github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

// GetPaymentByBucketName get bucket payment info by a bucket name
func (metadata *Metadata) GetPaymentByBucketName(ctx context.Context, req *metatypes.GetPaymentByBucketNameRequest) (resp *metatypes.GetPaymentByBucketNameResponse, err error) {
	var (
		streamRecord *model.StreamRecord
		res          *types.StreamRecord
	)

	ctx = log.Context(ctx, req)

	streamRecord, err = metadata.bsDB.GetPaymentByBucketName(req.BucketName, req.IsFullList)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get payment by bucket name", "error", err)
		return
	}

	if streamRecord != nil {
		res = &types.StreamRecord{
			Account:         streamRecord.Account.String(),
			CrudTimestamp:   streamRecord.UpdateTime,
			NetflowRate:     math.NewIntFromBigInt(streamRecord.NetflowRate.BigInt()),
			StaticBalance:   math.NewIntFromBigInt(streamRecord.StaticBalance.BigInt()),
			BufferBalance:   math.NewIntFromBigInt(streamRecord.BufferBalance.BigInt()),
			LockBalance:     math.NewIntFromBigInt(streamRecord.LockBalance.BigInt()),
			Status:          types.StreamAccountStatus(types.StreamAccountStatus_value[streamRecord.Status]),
			SettleTimestamp: streamRecord.SettleTimestamp,
			OutFlows:        streamRecord.OutFlows,
		}
	}

	resp = &metatypes.GetPaymentByBucketNameResponse{StreamRecord: res}
	log.CtxInfow(ctx, "succeed to get payment by bucket name")
	return resp, nil
}

// GetPaymentByBucketID get bucket payment info by a bucket id
func (metadata *Metadata) GetPaymentByBucketID(ctx context.Context, req *metatypes.GetPaymentByBucketIDRequest) (resp *metatypes.GetPaymentByBucketIDResponse, err error) {
	var (
		streamRecord *model.StreamRecord
		res          *types.StreamRecord
	)

	ctx = log.Context(ctx, req)

	streamRecord, err = metadata.bsDB.GetPaymentByBucketID(req.BucketId, req.IsFullList)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get payment by bucket id", "error", err)
		return
	}

	if streamRecord != nil {
		res = &types.StreamRecord{
			Account:         streamRecord.Account.String(),
			CrudTimestamp:   streamRecord.UpdateTime,
			NetflowRate:     math.NewIntFromBigInt(streamRecord.NetflowRate.BigInt()),
			StaticBalance:   math.NewIntFromBigInt(streamRecord.StaticBalance.BigInt()),
			BufferBalance:   math.NewIntFromBigInt(streamRecord.BufferBalance.BigInt()),
			LockBalance:     math.NewIntFromBigInt(streamRecord.LockBalance.BigInt()),
			Status:          types.StreamAccountStatus(types.StreamAccountStatus_value[streamRecord.Status]),
			SettleTimestamp: streamRecord.SettleTimestamp,
			OutFlows:        streamRecord.OutFlows,
		}
	}

	resp = &metatypes.GetPaymentByBucketIDResponse{StreamRecord: res}
	log.CtxInfow(ctx, "succeed to get payment by bucket id")
	return resp, nil
}
