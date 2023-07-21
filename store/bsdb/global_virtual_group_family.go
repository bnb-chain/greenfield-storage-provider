package bsdb

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// ListVirtualGroupFamiliesBySpID list virtual group families by sp id
func (b *BsDBImpl) ListVirtualGroupFamiliesBySpID(spID uint32) ([]*GlobalVirtualGroupFamily, error) {
	var (
		families []*GlobalVirtualGroupFamily
		filters  []func(*gorm.DB) *gorm.DB
		err      error
	)

	filters = append(filters, RemovedFilter(false))
	err = b.db.Table((&GlobalVirtualGroupFamily{}).TableName()).
		Select("*").
		Where("primary_sp_id = ?", spID).
		Scopes(filters...).
		Find(&families).Error

	return families, err
}

// GetVirtualGroupFamiliesByVgfID get virtual group families by vgf id
func (b *BsDBImpl) GetVirtualGroupFamiliesByVgfID(vgfID uint32) (*GlobalVirtualGroupFamily, error) {
	var (
		family  *GlobalVirtualGroupFamily
		filters []func(*gorm.DB) *gorm.DB
		err     error
	)

	filters = append(filters, RemovedFilter(false))
	err = b.db.Table((&GlobalVirtualGroupFamily{}).TableName()).
		Select("*").
		Where("global_virtual_group_family_id = ?", vgfID).
		Scopes(filters...).
		Take(&family).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return family, err
}

// ListVgfByGvgID list vgf by gvg id
func (b *BsDBImpl) ListVgfByGvgID(gvgIDs []uint32) ([]*GlobalVirtualGroupFamily, error) {
	var (
		families []*GlobalVirtualGroupFamily
		query    string
		err      error
	)

	if len(gvgIDs) == 0 {
		return nil, nil
	}
	query = fmt.Sprintf("select * from global_virtual_group_families where (FIND_IN_SET('%d', global_virtual_group_ids) > 0", gvgIDs[0])
	if len(gvgIDs) > 1 {
		for _, id := range gvgIDs[1:] {
			subQuery := fmt.Sprintf(" or FIND_IN_SET('%d', global_virtual_group_ids) > 0", id)
			query = query + subQuery
		}
	}
	query = query + ") and removed = false;"
	err = b.db.Table((&GlobalVirtualGroupFamily{}).TableName()).Raw(query).Find(&families).Error

	return families, err
}
