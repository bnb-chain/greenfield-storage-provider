package bsdb

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/model/metadata"
)

// GetUserBuckets get buckets info by a user address
func (s *Store) GetUserBuckets(ctx context.Context, accountID string) (ret []*metadata.Bucket, err error) {
	//TODO:: cancel mock after impl db
	bucket1 := metadata.Bucket{
		Owner:            "46765cbc-d30c-4f4a-a814-b68181fcab12",
		BucketName:       "BBC News",
		IsPublic:         true,
		ID:               "1",
		SourceType:       1,
		CreateAt:         1676530547,
		PaymentAddress:   "0x000000006b4BD0274e8f943201A922295D13fc28",
		PrimarySpAddress: "0x000000006b4BD0274e8f943201A922295D13fc28",
		ReadQuota:        2,
		PaymentPriceTime: 0,
	}
	ret = append(ret, &bucket1)
	return ret, nil
}
