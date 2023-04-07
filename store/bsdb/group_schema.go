package bsdb

import "github.com/forbole/juno/v4/common"

type Group struct {
	ID              uint64         `gorm:"column:id"`
	Owner           common.Address `gorm:"column:owner"`
	GroupID         common.Hash    `gorm:"column:group_id"`
	GroupName       string         `gorm:"column:group_name"`
	SourceType      string         `gorm:"column:source_type"`
	AccountID       common.Hash    `gorm:"column:account_id"`
	OperatorAddress common.Address `gorm:"column:operator_address"`
	CreateAt        int64          `gorm:"column:create_at"`
	CreateTime      int64          `gorm:"column:create_time"`
	UpdateAt        int64          `gorm:"column:update_at"`
	UpdateTime      int64          `gorm:"column:update_time"`
	Removed         bool           `gorm:"column:removed"`
}

// TableName is used to set Group table name in database
func (a *Group) TableName() string {
	return GroupTableName
}
