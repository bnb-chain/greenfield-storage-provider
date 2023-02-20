package store

import (
	"context"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
)

func (s *Store) GetUserBuckets(ctx context.Context) (ret []*model.Bucket, err error) {
	//dummyUUID1, _ := uuid.FromString("46765cbc-d30c-4f4a-a814-b68181fcab7a")
	bucket1 := model.Bucket{
		Owner:            "46765cbc-d30c-4f4a-a814-b68181fcab12",
		BucketName:       "BBC News",
		IsPublic:         true,
		Id:               "1",
		SourceType:       "SOURCE_TYPE_BSC_CROSS_CHAIN",
		CreateAt:         1676530547,
		PaymentAddress:   "0x000000006b4BD0274e8f943201A922295D13fc28",
		PrimarySpAddress: "0x000000006b4BD0274e8f943201A922295D13fc28",
		ReadQuota:        "1000",
		PaymentPriceTime: 0,
		PaymentOutFlows:  nil,
	}
	ret = append(ret, &bucket1)
	return ret, nil
}
