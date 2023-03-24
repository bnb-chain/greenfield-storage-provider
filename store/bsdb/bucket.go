package bsdb

import (
	"errors"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
)

// GetUserBuckets get buckets info by a user address
func (b *BsDBImpl) GetUserBuckets(accountID common.Address) ([]*Bucket, error) {
	var (
		buckets []*Bucket
		err     error
	)

	err = b.db.Find(&buckets, "owner_address = ?", accountID).Error
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
		return bucket, nil
	}

	err = b.db.Take(&bucket, "bucket_name = ? and is_public = ?", bucketName, true).Error
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

	err = b.db.Take(&bucket, "bucket_id = ? and is_public = ?", bucketIDHash, true).Error
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

	err = b.db.Table((&Bucket{}).TableName()).Select("count(1)").Take(&count, "owner_address = ?", accountID).Error
	return count, err
}
