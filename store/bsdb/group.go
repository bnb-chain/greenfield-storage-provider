package bsdb

import (
	"cosmossdk.io/math"
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
		groups  []*Group
		err     error
		filters []func(*gorm.DB) *gorm.DB
		count   int64
	)

	if sourceType != "" {
		filters = append(filters, SourceTypeFilter(sourceType))
	}

	if !includeRemoved {
		filters = append(filters, RemovedFilter(includeRemoved))
	}

	err = b.db.Table((&Group{}).TableName()).
		Select("*").
		Where("group_name LIKE ? and account_id = ?", prefix+"%"+name+"%", common.HexToAddress(GroupAddress)).
		Scopes(filters...).
		Limit(limit).
		Offset(offset).
		Order("group_id").
		Find(&groups).Error
	if err != nil {
		return nil, 0, err
	}

	err = b.db.Table((&Group{}).TableName()).
		Select("count(*)").
		Where("group_name LIKE ? and account_id = ?", prefix+"%"+name+"%", common.HexToAddress(GroupAddress)).
		Scopes(filters...).
		Take(&count).Error
	return groups, count, err
}

// GetGroupByID get group info by object id
func (b *BsDBImpl) GetGroupByID(groupID int64, includeRemoved bool) (*Group, error) {
	var (
		group       *Group
		groupIDHash common.Hash
		err         error
		filters     []func(*gorm.DB) *gorm.DB
	)

	groupIDHash = common.BigToHash(math.NewInt(groupID).BigInt())
	if !includeRemoved {
		filters = append(filters, RemovedFilter(includeRemoved))
	}

	err = b.db.Table((&Group{}).TableName()).
		Select("*").
		Where("group_id  = ?", groupIDHash).
		Scopes(filters...).
		Find(&group).Error
	return group, err
}
