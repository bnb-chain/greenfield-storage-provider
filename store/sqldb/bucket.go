package sqldb

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/model/metadata"
)

// GetUserBuckets get buckets info by a user address
func (s *SpDBImpl) GetUserBuckets(accountID common.Address) ([]*metadata.Bucket, error) {
	var (
		buckets []*metadata.Bucket
		err     error
	)

	err = s.db.Find(&buckets, "owner = ?", accountID).Error
	return buckets, err
}

// GetBucketByName get buckets info by a bucket name
func (s *SpDBImpl) GetBucketByName(bucketName string, isFullList bool) (*metadata.Bucket, error) {
	var (
		bucket *metadata.Bucket
		err    error
	)

	if isFullList {
		err = s.db.Take(&bucket, "bucket_name = ?", bucketName).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return bucket, nil
	}

	err = s.db.Take(&bucket, "bucket_name = ? and is_public = ?", bucketName, true).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return bucket, err
}

// GetBucketByID get buckets info by by a bucket id
func (s *SpDBImpl) GetBucketByID(bucketID int64, isFullList bool) (*metadata.Bucket, error) {
	var (
		bucket *metadata.Bucket
		err    error
	)

	if isFullList {
		err = s.db.Take(&bucket, "bucket_id = ?", bucketID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return bucket, err
	}

	err = s.db.Take(&bucket, "bucket_id = ? and is_public = ?", bucketID, true).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return bucket, err
}

// GetUserBucketsCount get buckets count by a user address
func (s *SpDBImpl) GetUserBucketsCount(accountID common.Address) (int64, error) {
	var (
		count int64
		err   error
	)

	err = s.db.Table((&metadata.Bucket{}).TableName()).Select("count(1)").Take(&count, "owner = ?", accountID).Error
	return count, err
}
