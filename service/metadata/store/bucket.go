package store

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
)

//func extractUserID(ctx *gin.Context) (string, error) {
//	userID := ctx.Param("user_id")
//	return userID, nil
//}

func (s *Store) GetUserBuckets(ctx context.Context, accountID string) (ret []*model.Bucket, err error) {
	//err = s.userDB.WithContext(ctx).Find(&ret, "owner = ?", accountID).Error
	//return
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
	ret = append(ret, &bucket1)
	return ret, nil
}
