package bsdb

import (
	"errors"
	"fmt"

	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"
)

// GetGlobalVirtualGroupByGvgID get global virtual group by gvg id
func (b *BsDBImpl) GetGlobalVirtualGroupByGvgID(gvgID uint32) (*GlobalVirtualGroup, error) {
	var (
		gvg     *GlobalVirtualGroup
		filters []func(*gorm.DB) *gorm.DB
		err     error
	)

	filters = append(filters, RemovedFilter(false))
	err = b.db.Table((&VirtualGroupFamily{}).TableName()).
		Select("*").
		Where("gvg_id = ?", gvgID).
		Scopes(filters...).
		Take(&gvg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return gvg, err
}

// ListGvgByPrimarySpID list gvg by primary sp id
func (b *BsDBImpl) ListGvgByPrimarySpID(spID uint32) ([]*GlobalVirtualGroup, error) {
	var (
		gvg     []*GlobalVirtualGroup
		filters []func(*gorm.DB) *gorm.DB
		err     error
	)

	filters = append(filters, RemovedFilter(false))
	err = b.db.Table((&GlobalVirtualGroup{}).TableName()).
		Select("*").
		Where("primary_sp_id = ?", spID).
		Scopes(filters...).
		Find(&gvg).Error

	return gvg, err
}

// ListGvgBySecondarySpID list gvg by secondary sp id
func (b *BsDBImpl) ListGvgBySecondarySpID(spID uint32) ([]*GlobalVirtualGroup, error) {
	var (
		gvg   []*GlobalVirtualGroup
		query string
		err   error
	)
	query = fmt.Sprintf("select * from global_virtual_group where FIND_IN_SET('%d', secondary_sp_ids) > 0 and removed = false; ", spID)
	err = b.db.Raw(query).Find(&gvg).Error

	return gvg, err
}

// GfSpGetGvgByBucketAndLvgID get global virtual group by lvg id and bucket id
func (b *BsDBImpl) GfSpGetGvgByBucketAndLvgID(bucketID common.Hash, lvgID uint32) (*GlobalVirtualGroup, error) {
	var (
		gvg *GlobalVirtualGroup
		lvg *LocalVirtualGroup
		err error
	)
	lvg, err = b.GfSpGetLvgByBucketAndLvgID(bucketID, lvgID)
	if err != nil || lvg == nil {
		return nil, err
	}

	err = b.db.Table((&GlobalVirtualGroup{}).TableName()).
		Select("*").
		Where("global_virtual_group_id = ? and removed = false", lvg.GlobalVirtualGroupId).
		Take(&gvg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return gvg, err
}
