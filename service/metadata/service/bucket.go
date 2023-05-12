package service

import (
	"context"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/types/s3util"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/forbole/juno/v4/common"
	jsoniter "github.com/json-iterator/go"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	metatypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
	model "github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

// GetUserBuckets get buckets info by a user address
func (metadata *Metadata) GetUserBuckets(ctx context.Context, req *metatypes.GetUserBucketsRequest) (resp *metatypes.GetUserBucketsResponse, err error) {
	ctx = log.Context(ctx, req)
	buckets, err := metadata.bsDB.GetUserBuckets(common.HexToAddress(req.AccountId))
	if err != nil {
		log.CtxErrorw(ctx, "failed to get user buckets", "error", err)
		return
	}

	res := make([]*metatypes.Bucket, 0)
	for _, bucket := range buckets {
		res = append(res, &metatypes.Bucket{
			BucketInfo: &types.BucketInfo{
				Owner:            bucket.Owner.String(),
				BucketName:       bucket.BucketName,
				Id:               math.NewUintFromBigInt(bucket.BucketID.Big()),
				SourceType:       types.SourceType(types.SourceType_value[bucket.SourceType]),
				CreateAt:         bucket.CreateTime,
				PaymentAddress:   bucket.PaymentAddress.String(),
				PrimarySpAddress: bucket.PrimarySpAddress.String(),
				ChargedReadQuota: bucket.ChargedReadQuota,
				Visibility:       types.VisibilityType(types.VisibilityType_value[bucket.Visibility]),
				BillingInfo: types.BillingInfo{
					PriceTime:              0,
					TotalChargeSize:        0,
					SecondarySpObjectsSize: nil,
				},
				BucketStatus: types.BucketStatus(types.BucketStatus_value[bucket.Status]),
			},
			Removed:      bucket.Removed,
			DeleteAt:     bucket.DeleteAt,
			DeleteReason: bucket.DeleteReason,
			Operator:     bucket.Operator.String(),
			CreateTxHash: bucket.CreateTxHash.String(),
			UpdateTxHash: bucket.UpdateTxHash.String(),
			UpdateAt:     bucket.UpdateAt,
			UpdateTime:   bucket.UpdateTime,
		})
	}
	resp = &metatypes.GetUserBucketsResponse{Buckets: res}
	log.CtxInfow(ctx, "succeed to get user buckets")
	return resp, nil
}

// GetBucketByBucketName get buckets info by a bucket name
func (metadata *Metadata) GetBucketByBucketName(ctx context.Context, req *metatypes.GetBucketByBucketNameRequest) (resp *metatypes.GetBucketByBucketNameResponse, err error) {
	var (
		bucket *model.Bucket
		res    *metatypes.Bucket
	)

	ctx = log.Context(ctx, req)
	if err = s3util.CheckValidBucketName(req.BucketName); err != nil {
		log.Errorw("failed to check bucket name", "bucket_name", req.BucketName, "error", err)
		return nil, err
	}

	bucket, err = metadata.bsDB.GetBucketByName(req.BucketName, req.IsFullList)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket by bucket name", "error", err)
		return nil, err
	}

	if bucket != nil {
		res = &metatypes.Bucket{
			BucketInfo: &types.BucketInfo{
				Owner:            bucket.Owner.String(),
				BucketName:       bucket.BucketName,
				Id:               math.NewUintFromBigInt(bucket.BucketID.Big()),
				SourceType:       types.SourceType(types.SourceType_value[bucket.SourceType]),
				CreateAt:         bucket.CreateTime,
				PaymentAddress:   bucket.PaymentAddress.String(),
				PrimarySpAddress: bucket.PrimarySpAddress.String(),
				ChargedReadQuota: bucket.ChargedReadQuota,
				Visibility:       types.VisibilityType(types.VisibilityType_value[bucket.Visibility]),
				BillingInfo: types.BillingInfo{
					PriceTime:              0,
					TotalChargeSize:        0,
					SecondarySpObjectsSize: nil,
				},
				BucketStatus: types.BucketStatus(types.BucketStatus_value[bucket.Status]),
			},
			Removed:      bucket.Removed,
			DeleteAt:     bucket.DeleteAt,
			DeleteReason: bucket.DeleteReason,
			Operator:     bucket.Operator.String(),
			CreateTxHash: bucket.CreateTxHash.String(),
			UpdateTxHash: bucket.UpdateTxHash.String(),
			UpdateAt:     bucket.UpdateAt,
			UpdateTime:   bucket.UpdateTime,
		}
	}
	resp = &metatypes.GetBucketByBucketNameResponse{Bucket: res}
	log.CtxInfo(ctx, "succeed to get bucket by bucket name")
	return resp, nil
}

// GetBucketByBucketID get buckets info by by a bucket id
func (metadata *Metadata) GetBucketByBucketID(ctx context.Context, req *metatypes.GetBucketByBucketIDRequest) (resp *metatypes.GetBucketByBucketIDResponse, err error) {
	var (
		bucket *model.Bucket
		res    *metatypes.Bucket
	)

	ctx = log.Context(ctx, req)
	bucket, err = metadata.bsDB.GetBucketByID(req.BucketId, req.IsFullList)
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket by bucket id", "error", err)
		return nil, err
	}

	if bucket != nil {
		res = &metatypes.Bucket{
			BucketInfo: &types.BucketInfo{
				Owner:            bucket.Owner.String(),
				BucketName:       bucket.BucketName,
				Id:               math.NewUintFromBigInt(bucket.BucketID.Big()),
				SourceType:       types.SourceType(types.SourceType_value[bucket.SourceType]),
				CreateAt:         bucket.CreateTime,
				PaymentAddress:   bucket.PaymentAddress.String(),
				PrimarySpAddress: bucket.PrimarySpAddress.String(),
				ChargedReadQuota: bucket.ChargedReadQuota,
				Visibility:       types.VisibilityType(types.VisibilityType_value[bucket.Visibility]),
				BillingInfo: types.BillingInfo{
					PriceTime:              0,
					TotalChargeSize:        0,
					SecondarySpObjectsSize: nil,
				},
				BucketStatus: types.BucketStatus(types.BucketStatus_value[bucket.Status]),
			},
			Removed:      bucket.Removed,
			DeleteAt:     bucket.DeleteAt,
			DeleteReason: bucket.DeleteReason,
			Operator:     bucket.Operator.String(),
			CreateTxHash: bucket.CreateTxHash.String(),
			UpdateTxHash: bucket.UpdateTxHash.String(),
			UpdateAt:     bucket.UpdateAt,
			UpdateTime:   bucket.UpdateTime,
		}
	}
	resp = &metatypes.GetBucketByBucketIDResponse{Bucket: res}
	log.CtxInfow(ctx, "succeed to get bucket by bucket id")
	return resp, nil
}

// GetUserBucketsCount get buckets count by a user address
func (metadata *Metadata) GetUserBucketsCount(ctx context.Context, req *metatypes.GetUserBucketsCountRequest) (resp *metatypes.GetUserBucketsCountResponse, err error) {
	ctx = log.Context(ctx, req)

	count, err := metadata.bsDB.GetUserBucketsCount(common.HexToAddress(req.AccountId))
	if err != nil {
		log.CtxErrorw(ctx, "failed to get user buckets count", "error", err)
		return
	}

	resp = &metatypes.GetUserBucketsCountResponse{Count: count}
	log.CtxInfow(ctx, "succeed to get buckets count by a user address")
	return resp, nil
}

// ListExpiredBucketsBySp list expired bucket by sp
func (metadata *Metadata) ListExpiredBucketsBySp(ctx context.Context, req *metatypes.ListExpiredBucketsBySpRequest) (resp *metatypes.ListExpiredBucketsBySpResponse, err error) {
	ctx = log.Context(ctx, req)
	buckets, err := metadata.bsDB.ListExpiredBucketsBySp(req.GetCreateAt(), req.GetPrimarySpAddress(), req.GetLimit())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get user buckets", "error", err)
		return
	}

	res := make([]*metatypes.Bucket, 0)
	for _, bucket := range buckets {
		res = append(res, &metatypes.Bucket{
			BucketInfo: &types.BucketInfo{
				Owner:            bucket.Owner.String(),
				BucketName:       bucket.BucketName,
				Id:               math.NewUintFromBigInt(bucket.BucketID.Big()),
				SourceType:       types.SourceType(types.SourceType_value[bucket.SourceType]),
				CreateAt:         bucket.CreateTime,
				PaymentAddress:   bucket.PaymentAddress.String(),
				PrimarySpAddress: bucket.PrimarySpAddress.String(),
				ChargedReadQuota: bucket.ChargedReadQuota,
				Visibility:       types.VisibilityType(types.VisibilityType_value[bucket.Visibility]),
				BillingInfo: types.BillingInfo{
					PriceTime:              0,
					TotalChargeSize:        0,
					SecondarySpObjectsSize: nil,
				},
				BucketStatus: types.BucketStatus(types.BucketStatus_value[bucket.Status]),
			},
			Removed:      bucket.Removed,
			DeleteAt:     bucket.DeleteAt,
			DeleteReason: bucket.DeleteReason,
		})
	}
	resp = &metatypes.ListExpiredBucketsBySpResponse{Buckets: res}
	log.CtxInfow(ctx, "succeed to get user buckets")
	return resp, nil
}

// GetBucketMeta get bucket metadata
func (metadata *Metadata) GetBucketMeta(ctx context.Context, req *metatypes.GetBucketMetaRequest) (resp *metatypes.GetBucketMetaResponse, err error) {
	var (
		bucket          *model.Bucket
		bucketRes       *metatypes.Bucket
		streamRecord    *model.StreamRecord
		streamRecordRes *paymenttypes.StreamRecord
		outflows        []paymenttypes.OutFlow
	)

	ctx = log.Context(ctx, req)
	bucketFullMeta, err := metadata.bsDB.GetBucketMetaByName(req.GetBucketName(), req.GetIsFullList())
	if err != nil {
		log.CtxErrorw(ctx, "failed to get bucket meta by name", "error", err)
		return
	}
	bucket = &bucketFullMeta.Bucket
	streamRecord = &bucketFullMeta.StreamRecord

	if bucket != nil {
		bucketRes = &metatypes.Bucket{
			BucketInfo: &types.BucketInfo{
				Owner:            bucket.Owner.String(),
				BucketName:       bucket.BucketName,
				Id:               math.NewUintFromBigInt(bucket.BucketID.Big()),
				SourceType:       types.SourceType(types.SourceType_value[bucket.SourceType]),
				CreateAt:         bucket.CreateTime,
				PaymentAddress:   bucket.PaymentAddress.String(),
				PrimarySpAddress: bucket.PrimarySpAddress.String(),
				ChargedReadQuota: bucket.ChargedReadQuota,
				Visibility:       types.VisibilityType(types.VisibilityType_value[bucket.Visibility]),
				BillingInfo: types.BillingInfo{
					PriceTime:              0,
					TotalChargeSize:        0,
					SecondarySpObjectsSize: nil,
				},
				BucketStatus: types.BucketStatus(types.BucketStatus_value[bucket.Status]),
			},
			Removed:      bucket.Removed,
			DeleteAt:     bucket.DeleteAt,
			DeleteReason: bucket.DeleteReason,
			Operator:     bucket.Operator.String(),
			CreateTxHash: bucket.CreateTxHash.String(),
			UpdateTxHash: bucket.UpdateTxHash.String(),
			UpdateAt:     bucket.UpdateAt,
			UpdateTime:   bucket.UpdateTime,
		}
	}

	if streamRecord != nil {
		err = jsoniter.Unmarshal(streamRecord.OutFlows, &outflows)
		if err != nil {
			log.CtxErrorw(ctx, "failed to unmarshal out flows", "error", err)
			return
		}
		streamRecordRes = &paymenttypes.StreamRecord{
			Account:         streamRecord.Account.String(),
			CrudTimestamp:   streamRecord.CrudTimestamp,
			NetflowRate:     math.NewIntFromBigInt(streamRecord.NetflowRate.Raw()),
			StaticBalance:   math.NewIntFromBigInt(streamRecord.StaticBalance.Raw()),
			BufferBalance:   math.NewIntFromBigInt(streamRecord.BufferBalance.Raw()),
			LockBalance:     math.NewIntFromBigInt(streamRecord.LockBalance.Raw()),
			Status:          paymenttypes.StreamAccountStatus(paymenttypes.StreamAccountStatus_value[streamRecord.Status]),
			SettleTimestamp: streamRecord.SettleTimestamp,
			OutFlows:        outflows,
		}
	}

	resp = &metatypes.GetBucketMetaResponse{Bucket: bucketRes, StreamRecord: streamRecordRes}
	log.CtxInfow(ctx, "succeed to get bucket meta by name")
	return resp, nil
}
