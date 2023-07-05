package bsdb

import (
	"errors"

	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"
)

// ListLvgByGvgAndBucketID list lvg by gvg id and bucket id
func (b *BsDBImpl) ListLvgByGvgAndBucketID(bucketID common.Hash, gvgIDs []uint32) ([]*LocalVirtualGroup, error) {
	var (
		groups  []*LocalVirtualGroup
		filters []func(*gorm.DB) *gorm.DB
		err     error
	)

	filters = append(filters, RemovedFilter(false))
	err = b.db.Table((&LocalVirtualGroup{}).TableName()).
		Select("*").
		Where("global_virtual_group_id in (?) and bucket_id = ?", gvgIDs, bucketID).
		Scopes(filters...).
		Find(&groups).Error
	return groups, err
}

// ListLvgByGvgID list lvg by gvg id
func (b *BsDBImpl) ListLvgByGvgID(gvgIDs []uint32) ([]*LocalVirtualGroup, error) {
	var (
		groups  []*LocalVirtualGroup
		filters []func(*gorm.DB) *gorm.DB
		err     error
	)

	filters = append(filters, RemovedFilter(false))
	err = b.db.Table((&LocalVirtualGroup{}).TableName()).
		Select("*").
		Where("global_virtual_group_id in (?)", gvgIDs).
		Scopes(filters...).
		Find(&groups).Error
	return groups, err
}

// GetLvgByBucketAndLvgID get global virtual group by lvg id and bucket id
func (b *BsDBImpl) GetLvgByBucketAndLvgID(bucketID common.Hash, lvgID uint32) (*LocalVirtualGroup, error) {
	var (
		lvg     *LocalVirtualGroup
		filters []func(*gorm.DB) *gorm.DB
		err     error
	)
	filters = append(filters, RemovedFilter(false))
	err = b.db.Table((&LocalVirtualGroup{}).TableName()).
		Select("*").
		Where("bucket_id = ? and local_virtual_group_id = ?", bucketID, lvgID).
		Scopes(filters...).
		Take(&lvg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return lvg, err
}
