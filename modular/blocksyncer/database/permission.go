package database

import (
	"context"
	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (db *DB) SavePermission(ctx context.Context, permission *models.Permission) error {
	return nil
}

func (db *DB) UpdatePermission(ctx context.Context, permission *models.Permission) error {
	return nil
}

func (db *DB) MultiSaveStatement(ctx context.Context, statements []*models.Statements) error {
	return nil
}

func (db *DB) RemoveStatements(ctx context.Context, policyID common.Hash) error {
	return nil
}

func (db *DB) SavePermissionToSQL(ctx context.Context, permission *models.Permission) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.Permission{}).TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "principal_type"}, {Name: "principal_value"}, {Name: "resource_type"}, {Name: "resource_id"}},
		UpdateAll: true,
	}).Create(permission).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) UpdatePermissionToSQL(ctx context.Context, permission *models.Permission) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.Permission{}).TableName()).Where("policy_id = ?", permission.PolicyID).Updates(permission).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) MultiSaveStatementToSQL(ctx context.Context, statements []*models.Statements) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{
		DryRun: true,
	}).Table((&models.Statements{}).TableName()).Create(statements).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) RemoveStatementsToSQL(ctx context.Context, policyID common.Hash) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.Statements{}).TableName()).Where("policy_id = ?", policyID).Update("removed", true).Statement
	return stat.SQL.String(), stat.Vars
}
