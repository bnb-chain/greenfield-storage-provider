package bsdb

import (
	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"
)

// GetGroupsByGroupIDAndAccount get groups info by group id list and account id
func (b *BsDBImpl) GetGroupsByGroupIDAndAccount(groupIDList []common.Hash, account common.Address, includeRemoved bool) ([]*Group, error) {
	var (
		groups []*Group
		err    error
	)

	if includeRemoved {
		err = b.db.Table((&Group{}).TableName()).
			Select("*").
			Where("group_id in (?) and account_id = ?", groupIDList, account).
			Find(&groups).Error
	} else {
		err = b.db.Table((&Group{}).TableName()).
			Select("*").
			Where("group_id in (?) and account_id = ? and removed =false", groupIDList, account).
			Find(&groups).Error
	}
	return groups, err
}

// ListGroupsByNameAndSourceType get groups list by specific parameters
func (b *BsDBImpl) ListGroupsByNameAndSourceType(name, prefix, sourceType string, limit, offset int, includeRemoved bool) ([]*Group, int64, error) {
	var (
		groups   []*Group
		err      error
		filters  []func(*gorm.DB) *gorm.DB
		subQuery *gorm.DB
		count    int64
	)
	if sourceType != "" {
		filters = append(filters, SourceTypeFilter(sourceType))
	}
	if includeRemoved {
		subQuery = b.db.Table((&Group{}).TableName()).
			Select("DISTINCT(group_id)").
			Where("group_name LIKE ?", prefix+"%"+name+"%").
			Scopes(filters...)
	} else {
		subQuery = b.db.Table((&Group{}).TableName()).
			Select("DISTINCT(group_id)").
			Where("group_name LIKE ? and removed = false", prefix+"%"+name+"%").
			Scopes(filters...)
	}

	err = b.db.Table((&Group{}).TableName()).
		Select("*").
		Where("group_id IN (?)", subQuery).
		Limit(limit).
		Offset(offset).
		Order("create_time").
		Find(&groups).Error
	if err != nil {
		return nil, 0, err
	}

	if includeRemoved {
		err = b.db.Table((&Group{}).TableName()).
			Select("count(*)").
			Where("group_name LIKE ?", prefix+"%"+name+"%").
			Scopes(filters...).
			Take(&count).Error
	} else {
		err = b.db.Table((&Group{}).TableName()).
			Select("count(*)").
			Where("group_name LIKE ? and removed = false", prefix+"%"+name+"%").
			Scopes(filters...).
			Take(&count).Error
	}
	return groups, count, err
}
