package bsdb

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
)

func (s *Store) ListObjectsByBucketName(ctx context.Context, bucketName string) (ret []*model.Object, err error) {
	//err = s.userDB.WithContext(ctx).Table((&model.Object{}).TableName()).Select("*").Where("bucket_name = ?", bucketName).Find(&ret).Error
	//return
	object1 := &model.Object{
		Owner:                "46765cbc-d30c-4f4a-a814-b68181fcab12",
		BucketName:           bucketName,
		ObjectName:           "test-object",
		Id:                   "1000",
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
	return ret, nil
}
