package bsdb

import (
	"time"

	"cosmossdk.io/math"
	"github.com/forbole/juno/v4/common"
	"gorm.io/gorm"
)

// GetGroupsByGroupIDAndAccount get groups info by group id list and account id
func (b *BsDBImpl) GetGroupsByGroupIDAndAccount(groupIDList []common.Hash, account common.Address, includeRemoved bool) ([]*Group, error) {
	var (
		groups   []*Group
		err      error
		groupIDs []common.Hash
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	// In the "group" table, each group has an account ID of "0x0000000000000000000000000000000000000000" representing the group's creation information.
	// If this group is granted permissions, the non-existent account 0x0000000000000000000000000000000000000000 will gain access, so it's prohibited here
	if account.String() == common.HexToAddress("0").String() {
		return nil, nil
	}

	if includeRemoved {
		// filter those groups already be deleted while group-accounts are still exist
		err = b.db.Table((&Group{}).TableName()).
			Select("group_id").
			Where("group_id in (?) and account_id = ?", groupIDList, common.HexToAddress("0")).
			Find(&groupIDs).Error
		if err != nil {
			return nil, err
		}
		err = b.db.Table((&Group{}).TableName()).
			Select("*").
			Where("group_id in (?) and account_id = ?", groupIDs, account).
			Find(&groups).Error
	} else {
		// filter those groups already be deleted while group-accounts are still exist
		err = b.db.Table((&Group{}).TableName()).
			Select("group_id").
			Where("group_id in (?) and account_id = ? and removed = false", groupIDList, common.HexToAddress("0")).
			Find(&groupIDs).Error
		if err != nil {
			return nil, err
		}
		err = b.db.Table((&Group{}).TableName()).
			Select("*").
			Where("group_id in (?) and account_id = ? and removed = false", groupIDs, account).
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
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

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

// GetGroupMembersCount get the count of group members
func (b *BsDBImpl) GetGroupMembersCount(groupIDs []common.Hash) ([]*GroupCount, error) {
	var (
		results []*GroupCount
		err     error
	)

	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	if len(groupIDs) == 0 {
		return nil, nil
	}

	err = b.db.Table((&Group{}).TableName()).
		Select("group_id, count(*) as count").
		Where("group_id in (?) and account_id != ? and removed = false", groupIDs, common.HexToAddress(GroupAddress)).
		Group("group_id").
		Find(&results).Error
	return results, err
}

// GetGroupByID get group info by object id
func (b *BsDBImpl) GetGroupByID(groupID int64, includeRemoved bool) (*Group, error) {
	var (
		group       *Group
		groupIDHash common.Hash
		err         error
		filters     []func(*gorm.DB) *gorm.DB
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

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

// GetUserGroups get groups info by a user address
func (b *BsDBImpl) GetUserGroups(accountID common.Address, startAfter common.Hash, limit int) ([]*Group, error) {
	var (
		groups  []*Group
		err     error
		filters []func(*gorm.DB) *gorm.DB
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	filters = append(filters, GroupIDStartAfterFilter(startAfter), RemovedFilter(false), WithLimit(limit))
	err = b.db.Table((&Group{}).TableName()).
		Select("*").
		Where("account_id  = ?", accountID).
		Scopes(filters...).
		Order("group_id").
		Find(&groups).Error
	return groups, err
}

// GetGroupMembers get group members by group id
func (b *BsDBImpl) GetGroupMembers(groupID common.Hash, startAfter common.Address, limit int) ([]*Group, error) {
	var (
		members []*Group
		err     error
		filters []func(*gorm.DB) *gorm.DB
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	filters = append(filters, GroupAccountIDStartAfterFilter(startAfter), RemovedFilter(false), WithLimit(limit))
	err = b.db.Table((&Group{}).TableName()).
		Select("*").
		Where("group_id  = ?", groupID).
		Scopes(filters...).
		Order("account_id").
		Find(&members).Error
	return members, err
}

// GetUserOwnedGroups retrieve groups where the user is the owner
func (b *BsDBImpl) GetUserOwnedGroups(accountID common.Address, startAfter common.Hash, limit int) ([]*Group, error) {
	var (
		groups  []*Group
		err     error
		filters []func(*gorm.DB) *gorm.DB
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	filters = append(filters, GroupIDStartAfterFilter(startAfter), RemovedFilter(false), WithLimit(limit))
	err = b.db.Table((&Group{}).TableName()).
		Select("*").
		Where("owner = ? and account_id  = ?", accountID, common.HexToAddress("0")).
		Scopes(filters...).
		Order("group_id").
		Find(&groups).Error
	return groups, err
}

// ListGroupsByIDs list groups by ids
func (b *BsDBImpl) ListGroupsByIDs(ids []common.Hash) ([]*Group, error) {
	var (
		groups []*Group
		err    error
	)
	startTime := time.Now()
	methodName := currentFunction()
	defer func() {
		if err != nil {
			MetadataDatabaseFailureMetrics(err, startTime, methodName)
		} else {
			MetadataDatabaseSuccessMetrics(startTime, methodName)
		}
	}()

	err = b.db.Table((&Group{}).TableName()).
		Select("*").
		Where("group_id in (?) and account_id  = ? and removed = false", ids, common.HexToAddress("0")).
		Order("group_id").
		Find(&groups).Error
	return groups, err
}
