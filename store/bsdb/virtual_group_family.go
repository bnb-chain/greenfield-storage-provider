package bsdb

import (
	"errors"
	"fmt"

	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"
)

// ListVirtualGroupFamiliesBySpID list virtual group families by sp id
func (b *BsDBImpl) ListVirtualGroupFamiliesBySpID(spID uint32) ([]*VirtualGroupFamily, error) {
	var (
		families []*VirtualGroupFamily
		filters  []func(*gorm.DB) *gorm.DB
		err      error
	)

	filters = append(filters, RemovedFilter(false))
	err = b.db.Table((&VirtualGroupFamily{}).TableName()).
		Select("*").
		Where("sp_id = ?", spID).
		//TODO: BARRY add order by variables
		Order("global_virtual_group_family_id").
		Scopes(filters...).
		Find(&families).Error

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

// GetVirtualGroupFamilyBindingOnBucket get virtual group family binding on bucket
func (b *BsDBImpl) GetVirtualGroupFamilyBindingOnBucket(bucketID common.Hash) (*VirtualGroupFamily, error) {
	var (
		family  *VirtualGroupFamily
		filters []func(*gorm.DB) *gorm.DB
		err     error
	)

	filters = append(filters, RemovedFilter(false))
	err = b.db.Table((&VirtualGroupFamily{}).TableName()).
		Select("*").
		Where("bucket_id = ?", bucketID).
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
	query = fmt.Sprintf("select * from virtual_group_family where (FIND_IN_SET('%d', global_virtual_group_ids) > 0", gvgIDs[0])
	if len(gvgIDs) > 1 {
		for _, id := range gvgIDs[1:] {
			subQuery := fmt.Sprintf("or FIND_IN_SET('%d', global_virtual_group_ids) > 0", id)
			query = query + subQuery
		}
	}
	query = query + ") and removed = false;"
	err = b.db.Raw(query).Find(&families).Error
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
