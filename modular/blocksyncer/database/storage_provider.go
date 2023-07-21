package database

import (
	"context"

	"github.com/forbole/juno/v4/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (db *DB) CreateStorageProvider(ctx context.Context, storageProvider *models.StorageProvider) error {
	return nil
}

func (db *DB) UpdateStorageProvider(ctx context.Context, storageProvider *models.StorageProvider) error {
	return nil
}

func (db *DB) CreateStorageProviderToSQL(ctx context.Context, storageProvider *models.StorageProvider) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.StorageProvider{}).TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sp_id"}},
		UpdateAll: true,
	}).Create(storageProvider).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) UpdateStorageProviderToSQL(ctx context.Context, storageProvider *models.StorageProvider) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.StorageProvider{}).TableName()).Where("sp_id = ? ", storageProvider.SpId).Updates(storageProvider).Statement
	return stat.SQL.String(), stat.Vars
}
