package database

import (
	"context"

	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (db *DB) CreateGroup(ctx context.Context, groupMembers []*models.Group) error {
	return nil
}

func (db *DB) UpdateGroup(ctx context.Context, group *models.Group) error {
	return nil
}

func (db *DB) CreateGroupToSQL(ctx context.Context, groupMembers []*models.Group) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.Group{}).TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "account_id"}, {Name: "group_id"}},
		UpdateAll: true,
	}).Create(groupMembers).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) UpdateGroupToSQL(ctx context.Context, group *models.Group) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.Group{}).TableName()).Where("account_id = ? AND group_id = ? ", group.AccountID, group.GroupID).Updates(group).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) DeleteGroupToSQL(ctx context.Context, group *models.Group) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.Group{}).TableName()).Where("group_id = ?", group.GroupID).Updates(group).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) BatchDeleteGroupMemberToSQL(ctx context.Context, group *models.Group, groupID common.Hash, accountIDs []common.Address) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.Group{}).TableName()).Where("account_id IN ? AND group_id = ? ", accountIDs, groupID).UpdateColumns(group).Statement
	return stat.SQL.String(), stat.Vars
}
