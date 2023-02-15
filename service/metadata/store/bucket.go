package store

import (
	"context"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
	"github.com/gofrs/uuid"
)

func (s *Store) GetUserBuckets(ctx context.Context) (ret []*model.Bucket, err error) {
	dummyUUID1, _ := uuid.FromString("46765cbc-d30c-4f4a-a814-b68181fcab7a")
	bucket1 := model.Bucket{
		UserID: dummyUUID1,
	}
	ret = append(ret, &bucket1)
	return ret, nil
	//return ret, s.getCTXUserDB(ctx).Find(&ret).Error
}
