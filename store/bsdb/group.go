package bsdb

import (
	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"
)

// GetGroupsByGroupIDAndAccount get groups info by group id list and account id
func (b *BsDBImpl) GetGroupsByGroupIDAndAccount(groupIDList []common.Hash, account common.Hash) ([]*Group, error) {
	var (
		groups []*Group
		err    error
	)

	err = b.db.Table((&Group{}).TableName()).
		Select("*").
		Where("group_id in ? and account_id = ?", groupIDList, account).
		Find(&groups).Error
	return groups, err
}

// ListGroupsByNameAndSourceType get groups list by specific parameters
func (b *BsDBImpl) ListGroupsByNameAndSourceType(name, prefix, sourceType string, limit, offset int) ([]*Group, error) {
	var (
		groups   []*Group
		err      error
		filters  []func(*gorm.DB) *gorm.DB
		subQuery *gorm.DB
	)
	if sourceType != "" {
		filters = append(filters, SourceTypeFilter(sourceType))
	}
	subQuery = b.db.Table((&Group{}).TableName()).
		Select("DISTINCT(group_id)").
		Where("group_name LIKE ?", prefix+"%"+name+"%").
		Scopes(filters...)

	err = b.db.Table((&Group{}).TableName()).
		Select("*").
		Where("group_id IN (?)", subQuery).
		Limit(limit).
		Offset(offset).
		Find(&groups).
		Error
	return groups, err
}
