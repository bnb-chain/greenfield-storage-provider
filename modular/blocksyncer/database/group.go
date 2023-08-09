package database

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/forbole/juno/v4/models"
)

func (db *DB) CreateGroup(ctx context.Context, groupMembers []*models.Group) error {
	return nil
}

func (db *DB) UpdateGroup(ctx context.Context, group *models.Group) error {
	return nil
}

func (db *DB) CreateGroupToSQL(ctx context.Context, groupMembers []*models.Group) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.Group{}).TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}, {Name: "account_id"}},
		UpdateAll: true,
	}).Create(groupMembers).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) UpdateGroupToSQL(ctx context.Context, group *models.Group) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.Group{}).TableName()).Where("group_id = ? AND account_id = ?", group.GroupID, group.AccountID).Updates(group).Statement
	return stat.SQL.String(), stat.Vars
}

//func (db *DB) RenewGroupMemberToSQL(ctx context.Context, group *models.Group) (string, []interface{}) {
//	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.Group{}).TableName()).Where("group_id = ? AND account_id = ?", group.GroupID, group.AccountID).Updates(group).Statement
//	return stat.SQL.String(), stat.Vars
//}
