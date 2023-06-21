package bsdb

import (
	"errors"
	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"
)

// ListLvgByGvgID list vgf by gvg id
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

// GfSpGetLvgByBucketAndLvgID get global virtual group by lvg id and bucket id
func (b *BsDBImpl) GfSpGetLvgByBucketAndLvgID(bucketID common.Hash, lvgID uint32) (*LocalVirtualGroup, error) {
	var (
		lvg     *LocalVirtualGroup
		filters []func(*gorm.DB) *gorm.DB
		err     error
	)
	filters = append(filters, RemovedFilter(false))
	err = b.db.Table((&LocalVirtualGroup{}).TableName()).
		Select("*").
		Where("bucket_id = ? and global_virtual_group_id = ?", bucketID, lvgID).
		Scopes(filters...).
		Take(&lvg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return lvg, err
}
