package service

import (
	"context"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/ethereum/go-ethereum/common"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	model "github.com/bnb-chain/greenfield-storage-provider/model/metadata"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	metatypes "github.com/bnb-chain/greenfield-storage-provider/service/metadata/types"
)

// GetUserBuckets get buckets info by a user address
func (metadata *Metadata) GetUserBuckets(ctx context.Context, req *metatypes.MetadataServiceGetUserBucketsRequest) (resp *metatypes.MetadataServiceGetUserBucketsResponse, err error) {
	ctx = log.Context(ctx, req)
	buckets, err := metadata.spDB.GetUserBuckets(common.HexToAddress(req.AccountId))
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
				IsPublic:         false,
				Id:               math.NewUint(uint64(bucket.BucketId)),
				SourceType:       types.SourceType(types.SourceType_value[bucket.SourceType]),
				CreateAt:         bucket.CreateAt,
				PaymentAddress:   bucket.PaymentAddress.String(),
				PrimarySpAddress: bucket.PrimarySpAddress.String(),
				ReadQuota:        0,
				BillingInfo: types.BillingInfo{
					PriceTime:              0,
					TotalChargeSize:        0,
					SecondarySpObjectsSize: nil,
				},
			},
		})
	}
	resp = &metatypes.MetadataServiceGetUserBucketsResponse{Buckets: res}
	return resp, nil
}

// GetBucketByBucketName get buckets info by a bucket name
func (metadata *Metadata) GetBucketByBucketName(ctx context.Context, req *metatypes.MetadataServiceGetBucketByBucketNameRequest) (resp *metatypes.MetadataServiceGetBucketByBucketNameResponse, err error) {
	var (
		bucket *model.Bucket
		res    *metatypes.Bucket
	)

	ctx = log.Context(ctx, req)
	if req.BucketName == "" {
		return nil, merrors.ErrInvalidBucketName
	}

	bucket, err = metadata.spDB.GetBucketByName(req.BucketName, req.IsFullList)
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
				Id:               math.NewUint(uint64(bucket.BucketId)),
				SourceType:       types.SourceType(types.SourceType_value[bucket.SourceType]),
				CreateAt:         bucket.CreateAt,
				PaymentAddress:   bucket.PaymentAddress.String(),
				PrimarySpAddress: bucket.PrimarySpAddress.String(),
				ReadQuota:        0,
			},
			Removed: bucket.Removed,
		}
	}
	resp = &metatypes.MetadataServiceGetBucketByBucketNameResponse{Bucket: res}
	log.CtxInfow(ctx, "success to get bucket by bucket name")
	return resp, nil
}

// GetBucketByBucketID get buckets info by by a bucket id
func (metadata *Metadata) GetBucketByBucketID(ctx context.Context, req *metatypes.MetadataServiceGetBucketByBucketIDRequest) (resp *metatypes.MetadataServiceGetBucketByBucketIDResponse, err error) {
	var (
		bucket *model.Bucket
		res    *metatypes.Bucket
	)

	ctx = log.Context(ctx, req)
	bucket, err = metadata.spDB.GetBucketByID(req.BucketId, req.IsFullList)
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
				Id:               math.NewUint(uint64(bucket.BucketId)),
				SourceType:       types.SourceType(types.SourceType_value[bucket.SourceType]),
				CreateAt:         bucket.CreateAt,
				PaymentAddress:   bucket.PaymentAddress.String(),
				PrimarySpAddress: bucket.PrimarySpAddress.String(),
				ReadQuota:        0,
			},
			Removed: bucket.Removed,
		}
	}
	resp = &metatypes.MetadataServiceGetBucketByBucketIDResponse{Bucket: res}
	log.CtxInfow(ctx, "success to get bucket by bucket id")
	return resp, nil
}

// GetUserBucketsCount get buckets count by a user address
func (metadata *Metadata) GetUserBucketsCount(ctx context.Context, req *metatypes.MetadataServiceGetUserBucketsCountRequest) (resp *metatypes.MetadataServiceGetUserBucketsCountResponse, err error) {
	ctx = log.Context(ctx, req)

	if req.AccountId == "" {
		return nil, merrors.ErrInvalidAccountID
	}

	count, err := metadata.spDB.GetUserBucketsCount(common.HexToAddress(req.AccountId))
	if err != nil {
		log.CtxErrorw(ctx, "failed to get user buckets count", "error", err)
		return
	}

	resp = &metatypes.MetadataServiceGetUserBucketsCountResponse{Count: count}
	log.CtxInfow(ctx, "success to get buckets count by a user address")
	return resp, nil
}
