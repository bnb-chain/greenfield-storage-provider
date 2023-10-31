package bsdb

import (
	"errors"
	"time"

	"cosmossdk.io/math"
	"github.com/forbole/juno/v4/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

// GetUserBuckets get buckets info by a user address
func (b *BsDBImpl) GetUserBuckets(accountID common.Address, includeRemoved bool) ([]*Bucket, error) {
	var (
		buckets []*Bucket
		err     error
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	if includeRemoved {
		err = b.db.Table((&Bucket{}).TableName()).
			Select("*").
			Where("owner = ?", accountID).
			Order("create_at desc").
			Limit(GetUserBucketsLimitSize).
			Find(&buckets).Error
	} else {
		err = b.db.Table((&Bucket{}).TableName()).
			Select("*").
			Where("owner = ? and removed = false", accountID).
			Order("create_at desc").
			Limit(GetUserBucketsLimitSize).
			Find(&buckets).Error
	}
	return buckets, err
}

// GetBucketByName get buckets info by a bucket name
func (b *BsDBImpl) GetBucketByName(bucketName string, includePrivate bool) (*Bucket, error) {
	var (
		bucket *Bucket
		err    error
	)

	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	if includePrivate {
		err = b.db.Take(&bucket, "bucket_name = ? and removed = false", bucketName).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return bucket, err
	}

	err = b.db.Take(&bucket, "bucket_name = ? and visibility = ? and removed = false", bucketName, types.VISIBILITY_TYPE_PUBLIC_READ.String()).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return bucket, err
}

// GetBucketByID get buckets info by a bucket id
func (b *BsDBImpl) GetBucketByID(bucketID int64, includePrivate bool) (*Bucket, error) {
	var (
		bucket       *Bucket
		err          error
		bucketIDHash common.Hash
	)

	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	bucketIDHash = common.BigToHash(math.NewInt(bucketID).BigInt())
	if includePrivate {
		err = b.db.Take(&bucket, "bucket_id = ? and removed = false", bucketIDHash).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return bucket, err
	}

	err = b.db.Take(&bucket, "bucket_id = ? and visibility = ? and removed = false", bucketIDHash, types.VISIBILITY_TYPE_PUBLIC_READ.String()).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return bucket, err
}

// GetBucketSizeByID get bucket size by a bucket id
func (b *BsDBImpl) GetBucketSizeByID(bucketID uint64) (decimal.Decimal, error) {
	var (
		size         decimal.Decimal
		err          error
		bucketIDHash common.Hash
	)

	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	bucketIDHash = common.BigToHash(math.NewUint(bucketID).BigInt())
	err = b.db.Table((&Bucket{}).TableName()).
		Select("storage_size").
		Take(&size, "bucket_id = ? and removed = false", bucketIDHash).Error
	return size, err
}

// GetUserBucketsCount get buckets count by a user address
func (b *BsDBImpl) GetUserBucketsCount(accountID common.Address, includeRemoved bool) (int64, error) {
	var (
		count int64
		err   error
	)

	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	if includeRemoved {
		err = b.db.Table((&Bucket{}).TableName()).Select("count(1)").Take(&count, "owner = ?", accountID).Error
	} else {
		err = b.db.Table((&Bucket{}).TableName()).Select("count(1)").Take(&count, "owner = ? and removed = false", accountID).Error
	}
	return count, err
}

// ListExpiredBucketsBySp lists expired buckets
func (b *BsDBImpl) ListExpiredBucketsBySp(createAt int64, primarySpID uint32, limit int64) ([]*Bucket, error) {
	var (
		buckets []*Bucket
		err     error
	)

	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	if limit < 1 || limit > ExpiredBucketsDefaultSize {
		limit = ExpiredBucketsDefaultSize
	}

	err = b.db.Raw(`
		SELECT buckets.* 
		FROM buckets 
		INNER JOIN global_virtual_group_families 
		ON buckets.global_virtual_group_family_id = global_virtual_group_families.global_virtual_group_family_id
		WHERE global_virtual_group_families.primary_sp_id = ? 
		AND buckets.status = 'BUCKET_STATUS_CREATED' 
		AND buckets.create_time < ? 
		AND buckets.removed = false
		LIMIT ?`,
		primarySpID, createAt, limit).
		Find(&buckets).Error

	return buckets, err
}

func (b *BsDBImpl) GetBucketMetaByName(bucketName string, includePrivate bool) (*BucketFullMeta, error) {
	var (
		bucketFullMeta *BucketFullMeta
		err            error
	)

	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	if includePrivate {
		err = b.db.Table((&Bucket{}).TableName()).
			Select("*").
			Joins("left join stream_records on buckets.payment_address = stream_records.account").
			Where("buckets.bucket_name = ? and buckets.removed = false", bucketName).
			Take(&bucketFullMeta).Error
	} else {
		err = b.db.Table((&Bucket{}).TableName()).
			Select("*").
			Joins("left join stream_records on buckets.payment_address = stream_records.account").
			Where("buckets.bucket_name = ? and buckets.removed = false and"+
				"buckets.visibility='VISIBILITY_TYPE_PUBLIC_READ'", bucketName).
			Take(&bucketFullMeta).Error
	}

	return bucketFullMeta, err
}

// ListBucketsByIDs list buckets by bucket ids
func (b *BsDBImpl) ListBucketsByIDs(ids []common.Hash, includeRemoved bool) ([]*Bucket, error) {
	var (
		buckets []*Bucket
		err     error
		filters []func(*gorm.DB) *gorm.DB
	)

	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	if !includeRemoved {
		filters = append(filters, RemovedFilter(includeRemoved))
	}

	err = b.db.Table((&Bucket{}).TableName()).
		Select("*").
		Where("bucket_id in (?)", ids).
		Scopes(filters...).
		Find(&buckets).Error
	return buckets, err
}

func (b *BsDBImpl) ListBucketsByVgfID(vgfIDs []uint32, startAfter common.Hash, limit int) ([]*Bucket, error) {
	var (
		buckets []*Bucket
		err     error
		filters []func(*gorm.DB) *gorm.DB
	)

	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	filters = append(filters, ObjectIDStartAfterFilter(startAfter), RemovedFilter(false), WithLimit(limit))
	err = b.db.Table((&Bucket{}).TableName()).
		Select("*").
		Where("global_virtual_group_family_id in (?)", vgfIDs).
		Scopes(filters...).
		Find(&buckets).Error
	return buckets, err
}
