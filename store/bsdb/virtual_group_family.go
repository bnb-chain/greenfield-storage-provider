package bsdb

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// ListVirtualGroupFamiliesBySpID list virtual group families by sp id
func (b *BsDBImpl) ListVirtualGroupFamiliesBySpID(spID uint32) ([]*VirtualGroupFamily, error) {
	var (
		groups   []*GlobalVirtualGroup
		families []*VirtualGroupFamily
		gvgIDs   []uint32
		filters  []func(*gorm.DB) *gorm.DB
		err      error
	)

	filters = append(filters, RemovedFilter(false))
	err = b.db.Table((&GlobalVirtualGroup{}).TableName()).
		Select("*").
		Where("primary_sp_id = ?", spID).
		Scopes(filters...).
		Find(&groups).Error
	if err != nil || len(groups) == 0 {
		return nil, err
	}

	gvgIDs = make([]uint32, len(groups))
	for i, group := range groups {
		gvgIDs[i] = group.GlobalVirtualGroupId
	}

	families, err = b.ListVgfByGvgID(gvgIDs)
	return families, err
}

// GetVirtualGroupFamiliesByVgfID get virtual group families by vgf id
func (b *BsDBImpl) GetVirtualGroupFamiliesByVgfID(vgfID uint32) (*VirtualGroupFamily, error) {
	var (
		family  *VirtualGroupFamily
		filters []func(*gorm.DB) *gorm.DB
		err     error
	)

	filters = append(filters, RemovedFilter(false))
	err = b.db.Table((&VirtualGroupFamily{}).TableName()).
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
func (b *BsDBImpl) ListVgfByGvgID(gvgIDs []uint32) ([]*VirtualGroupFamily, error) {
	var (
		families []*VirtualGroupFamily
		query    string
		err      error
	)

	//TODO BARRY check the sql logic here by debug log
	if len(gvgIDs) == 0 {
		return nil, nil
	}
	query = fmt.Sprintf("select * from global_virtual_group_families where (FIND_IN_SET('%d', global_virtual_group_ids) > 0", gvgIDs[0])
	if len(gvgIDs) > 1 {
		for _, id := range gvgIDs[1:] {
			subQuery := fmt.Sprintf("or FIND_IN_SET('%d', global_virtual_group_ids) > 0", id)
			query = query + subQuery
		}
	}
	query = query + ") and removed = false;"
	err = b.db.Table((&VirtualGroupFamily{}).TableName()).Raw(query).Find(&families).Error
	//query = "SELECT * FROM virtual_group_family WHERE FIND_IN_SET(?, global_virtual_group_ids) > 0"
	//args := make([]interface{}, len(gvgIDs))
	//
	//for i, id := range gvgIDs {
	//	if i != 0 {
	//		query += " OR FIND_IN_SET(?, global_virtual_group_ids) > 0"
	//	}
	//	args[i] = id
	//}
	//
	//err = b.db.Raw(query, args...).Find(&families).Error

	return families, err
}
