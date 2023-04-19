package bsdb

import (
	"errors"
	"strconv"

	"github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"
)

// GetUserBuckets get buckets info by a user address
func (b *BsDBImpl) GetUserBuckets(accountID common.Address) ([]*Bucket, error) {
	var (
		buckets []*Bucket
		err     error
	)

	err = b.db.Table((&Bucket{}).TableName()).
		Select("*").
		Where("owner = ?", accountID).
		Order("create_at desc").
		Find(&buckets).Error
	return buckets, err
}

// GetBucketByName get buckets info by a bucket name
func (b *BsDBImpl) GetBucketByName(bucketName string, isFullList bool) (*Bucket, error) {
	var (
		bucket *Bucket
		err    error
	)

	if isFullList {
		err = b.db.Take(&bucket, "bucket_name = ?", bucketName).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return bucket, err
	}

	err = b.db.Take(&bucket, "bucket_name = ? and visibility = ?", bucketName, types.VISIBILITY_TYPE_PUBLIC_READ.String()).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return bucket, err
}

// GetBucketByID get buckets info by a bucket id
func (b *BsDBImpl) GetBucketByID(bucketID int64, isFullList bool) (*Bucket, error) {
	var (
		bucket       *Bucket
		err          error
		bucketIDHash common.Hash
	)

	bucketIDHash = common.HexToHash(strconv.FormatInt(bucketID, 10))
	if isFullList {
		err = b.db.Take(&bucket, "bucket_id = ?", bucketIDHash).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return bucket, err
	}

	err = b.db.Take(&bucket, "bucket_id = ? and visibility = ?", bucketIDHash, types.VISIBILITY_TYPE_PUBLIC_READ.String()).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return bucket, err
}

// GetUserBucketsCount get buckets count by a user address
func (b *BsDBImpl) GetUserBucketsCount(accountID common.Address) (int64, error) {
	var (
		count int64
		err   error
	)

	err = b.db.Table((&Bucket{}).TableName()).Select("count(1)").Take(&count, "owner = ?", accountID).Error
	return count, err
}
