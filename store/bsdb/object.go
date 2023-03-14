package bsdb

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/model/metadata"
)

// ListObjectsByBucketName list objects info by a bucket name
func (s *Store) ListObjectsByBucketName(ctx context.Context, bucketName string) ([]*metadata.Object, error) {
	var (
		ret []*metadata.Object
		err error
	)

	//TODO:: cancel mock after impl db
	object1 := &metadata.Object{
		Owner:                "46765cbc-d30c-4f4a-a814-b68181fcab12",
		BucketName:           bucketName,
		ObjectName:           "test-object",
		ID:                   "1000",
		PayloadSize:          100,
		IsPublic:             false,
		ContentType:          "video",
		CreateAt:             0,
		ObjectStatus:         1,
		RedundancyType:       1,
		SourceType:           1,
		SecondarySpAddresses: nil,
		LockedBalance:        "1000",
	}
	ret = append(ret, object1)
	return ret, err
}
