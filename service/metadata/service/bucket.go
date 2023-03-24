package service

import (
	"context"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/types/s3util"
	"github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/ethereum/go-ethereum/common"

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
				IsPublic:         bucket.IsPublic,
				Id:               math.NewUint(uint64(bucket.BucketID)),
				SourceType:       types.SourceType(types.SourceType_value[bucket.SourceType]),
				CreateAt:         bucket.CreateAt,
				PaymentAddress:   bucket.PaymentAddress.String(),
				PrimarySpAddress: bucket.PrimarySpAddress.String(),
				ReadQuota:        bucket.ReadQuota,
				BillingInfo: types.BillingInfo{
					PriceTime:              0,
					TotalChargeSize:        0,
					SecondarySpObjectsSize: nil,
				},
			},
			Removed: bucket.Removed,
		})
	}
	resp = &metatypes.GetUserBucketsResponse{Buckets: res}
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
				IsPublic:         bucket.IsPublic,
				Id:               math.NewUint(uint64(bucket.BucketID)),
				SourceType:       types.SourceType(types.SourceType_value[bucket.SourceType]),
				CreateAt:         bucket.CreateAt,
				PaymentAddress:   bucket.PaymentAddress.String(),
				PrimarySpAddress: bucket.PrimarySpAddress.String(),
				ReadQuota:        bucket.ReadQuota,
			},
			Removed: bucket.Removed,
		}
	}
	resp = &metatypes.GetBucketByBucketNameResponse{Bucket: res}
	log.CtxInfo(ctx, "success to get bucket by bucket name")
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
				IsPublic:         bucket.IsPublic,
				Id:               math.NewUint(uint64(bucket.BucketID)),
				SourceType:       types.SourceType(types.SourceType_value[bucket.SourceType]),
				CreateAt:         bucket.CreateAt,
				PaymentAddress:   bucket.PaymentAddress.String(),
				PrimarySpAddress: bucket.PrimarySpAddress.String(),
				ReadQuota:        bucket.ReadQuota,
			},
			Removed: bucket.Removed,
		}
	}
	resp = &metatypes.GetBucketByBucketIDResponse{Bucket: res}
	log.CtxInfow(ctx, "success to get bucket by bucket id")
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
	log.CtxInfow(ctx, "success to get buckets count by a user address")
	return resp, nil
}
