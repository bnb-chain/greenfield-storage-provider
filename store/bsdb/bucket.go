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
		Limit(GetUserBucketsLimitSize).
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

// ListExpiredBucketsBySp lists expired buckets
func (b *BsDBImpl) ListExpiredBucketsBySp(createAt int64, primarySpAddress string, limit int64) ([]*Bucket, error) {
	var (
		buckets []*Bucket
		err     error
	)

	if limit < 1 || limit > ExpiredBucketsDefaultSize {
		limit = ExpiredBucketsDefaultSize
	}

	err = b.db.Table((&Bucket{}).TableName()).
		Select("*").
		Where("primary_sp_address = ? and status = 'BUCKET_STATUS_CREATED' and create_time < ? and removed = false", common.HexToAddress(primarySpAddress), createAt).
		Limit(int(limit)).
		Order("create_at").
		Find(&buckets).Error

	return buckets, err
}

func (b *BsDBImpl) GetBucketMetaByName(bucketName string, isFullList bool) (*BucketFullMeta, error) {
	var (
		bucketFullMeta *BucketFullMeta
		err            error
	)

	if isFullList {
		err = b.db.Table((&Bucket{}).TableName()).
			Select("*").
			Joins("left join stream_records on buckets.payment_address = stream_records.account").
			Where("buckets.bucket_name = ?", bucketName).
			Take(&bucketFullMeta).Error
	} else {
		err = b.db.Table((&Bucket{}).TableName()).
			Select("*").
			Joins("left join stream_records on buckets.payment_address = stream_records.account").
			Where("buckets.bucket_name = ? and "+
				"buckets.visibility='VISIBILITY_TYPE_PUBLIC_READ'", bucketName).
			Take(&bucketFullMeta).Error
	}

	return bucketFullMeta, err
}
