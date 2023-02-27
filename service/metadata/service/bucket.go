package service

import (
	"context"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

func (metadata *Metadata) GetUserBuckets(ctx context.Context, req *stypes.MetadataServiceBucketNameRequest) (resp *stypes.MetadataServiceGetUserBucketsResponse, err error) {
	ctx = log.Context(ctx, req)
	defer func() {
		if err != nil {
			resp.ErrMessage = merrors.MakeErrMsgResponse(err)
			log.CtxErrorw(ctx, "failed to get user buckets", "err", err)
		} else {
			log.CtxInfow(ctx, "succeed to get user buckets")
		}
	}()
	//buckets, err := metadata.store.GetUserBuckets(ctx, req.AccountID)
	var buckets []*model.Bucket
	bucket1 := model.Bucket{
		Owner:            "46765cbc-d30c-4f4a-a814-b68181fcab12",
		BucketName:       "BBC News",
		IsPublic:         true,
		Id:               "1",
		SourceType:       1,
		CreateAt:         1676530547,
		PaymentAddress:   "0x000000006b4BD0274e8f943201A922295D13fc28",
		PrimarySpAddress: "0x000000006b4BD0274e8f943201A922295D13fc28",
		ReadQuota:        2,
		PaymentPriceTime: 0,
	}
	buckets = append(buckets, &bucket1)
	res := make([]*stypes.Bucket, 0)

	for _, bucket := range buckets {
		res = append(res, &stypes.Bucket{
			Owner:            bucket.Owner,
			BucketName:       bucket.BucketName,
			IsPublic:         false,
			Id:               bucket.Id,
			SourceType:       stypes.SourceType(bucket.SourceType),
			CreateAt:         bucket.CreateAt,
			PaymentAddress:   bucket.PaymentAddress,
			PrimarySpAddress: bucket.PrimarySpAddress,
			ReadQuota:        stypes.ReadQuota(bucket.ReadQuota),
			PaymentPriceTime: bucket.PaymentPriceTime,
			PaymentOutFlows:  nil,
		})
	}
	resp = &stypes.MetadataServiceGetUserBucketsResponse{Buckets: res}
	return resp, nil
}
