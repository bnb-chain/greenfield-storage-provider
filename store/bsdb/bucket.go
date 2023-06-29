package bsdb

import (
	"errors"

	"cosmossdk.io/math"
	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

// GetUserBuckets get buckets info by a user address
func (b *BsDBImpl) GetUserBuckets(accountID common.Address, includeRemoved bool) ([]*Bucket, error) {
	var (
		buckets []*Bucket
		err     error
	)

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

// GetUserBucketsCount get buckets count by a user address
func (b *BsDBImpl) GetUserBucketsCount(accountID common.Address, includeRemoved bool) (int64, error) {
	var (
		count int64
		err   error
	)

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

	if limit < 1 || limit > ExpiredBucketsDefaultSize {
		limit = ExpiredBucketsDefaultSize
	}

	err = b.db.Table((&Bucket{}).TableName()).
		Select("*").
		Where("primary_sp_id = ? and status = 'BUCKET_STATUS_CREATED' and create_time < ? and removed = false", primarySpID, createAt).
		Limit(int(limit)).
		Order("create_at").
		Find(&buckets).Error

	return buckets, err
}

func (b *BsDBImpl) GetBucketMetaByName(bucketName string, includePrivate bool) (*BucketFullMeta, error) {
	var (
		bucketFullMeta *BucketFullMeta
		err            error
	)

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

// ListBucketsByBucketID list buckets by bucket ids
func (b *BsDBImpl) ListBucketsByBucketID(ids []common.Hash, includeRemoved bool) ([]*Bucket, error) {
	var (
		buckets []*Bucket
		err     error
		filters []func(*gorm.DB) *gorm.DB
	)

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

// ListBucketsBindingOnPrimarySP list buckets by primary sp id
func (b *BsDBImpl) ListBucketsBindingOnPrimarySP(spID uint32, startAfter common.Hash, limit int) ([]*Bucket, error) {
	var (
		groups   []*GlobalVirtualGroup
		families []*VirtualGroupFamily
		buckets  []*Bucket
		gvgIDs   []uint32
		vgfIDs   []uint32
		err      error
	)

	groups, err = b.ListGvgByPrimarySpID(spID)
	if err != nil || groups == nil {
		return nil, err
	}

	gvgIDs = make([]uint32, len(groups))
	for i, group := range groups {
		gvgIDs[i] = group.GlobalVirtualGroupId
	}

	families, err = b.ListVgfByGvgID(gvgIDs)
	if err != nil || families == nil {
		return nil, err
	}

	vgfIDs = make([]uint32, len(families))
	for i, family := range families {
		vgfIDs[i] = family.GlobalVirtualGroupFamilyId
	}

	buckets, err = b.ListBucketsByVgfID(vgfIDs, startAfter, limit)
	return buckets, err
}

// ListBucketsBindingOnSecondarySP list buckets by secondary sp id
func (b *BsDBImpl) ListBucketsBindingOnSecondarySP(spID uint32, startAfter common.Hash, limit int) ([]*Bucket, error) {
	var (
		groups   []*GlobalVirtualGroup
		families []*VirtualGroupFamily
		buckets  []*Bucket
		gvgIDs   []uint32
		vgfIDs   []uint32
		err      error
	)

	groups, err = b.ListGvgBySecondarySpID(spID)
	if err != nil || groups == nil {
		return nil, err
	}

	gvgIDs = make([]uint32, len(groups))
	for i, group := range groups {
		gvgIDs[i] = group.GlobalVirtualGroupId
	}

	families, err = b.ListVgfByGvgID(gvgIDs)
	if err != nil || families == nil {
		return nil, err
	}

	vgfIDs = make([]uint32, len(groups))
	for i, family := range families {
		vgfIDs[i] = family.GlobalVirtualGroupFamilyId
	}

	buckets, err = b.ListBucketsByVgfID(vgfIDs, startAfter, limit)
	return buckets, err
}

// ListBucketsByVgfID list buckets by vgf ids
//
//	func (b *BsDBImpl) ListBucketsByVgfID(vgfIDs []common.Hash, startAfter common.Hash, limit int) ([]*Bucket, error) {
//		var (
//			buckets []*Bucket
//			filters []func(*gorm.DB) *gorm.DB
//			err     error
//		)
//
//		filters = append(filters, RemovedFilter(false), BucketIDStartAfterFilter(startAfter))
//		err = b.db.Table((&Bucket{}).TableName()).
//			Select("*").
//			Where("global_virtual_group_family_id in (?)", vgfIDs).
//			Scopes(filters...).
//			Limit(limit).
//			Find(&buckets).Error
//		return buckets, err
//	}
func (b *BsDBImpl) ListBucketsByVgfID(vgfIDs []uint32, startAfter common.Hash, limit int) ([]*Bucket, error) {
	var (
		buckets []*Bucket
		err     error
		filters []func(*gorm.DB) *gorm.DB
	)
	filters = append(filters, ObjectIDStartAfterFilter(startAfter), RemovedFilter(false), WithLimit(limit))
	err = b.db.Table((&Bucket{}).TableName()).
		Select("*").
		Where("global_virtual_group_family_id in (?)", vgfIDs).
		Scopes(filters...).
		Find(&buckets).Error
	return buckets, err
}
