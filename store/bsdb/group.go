package bsdb

import (
	"github.com/forbole/juno/v4/common"
)

// GetGroupsByGroupIDAndAccount get groups info by group id list and account id
func (b *BsDBImpl) GetGroupsByGroupIDAndAccount(groupIDList []common.Hash, account common.Hash) ([]*Group, error) {
	var groupIDs string
	for _, hash := range groupIDList {
		groupIDs += hash.String() + ","
	}
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
