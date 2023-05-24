package database

import (
	"context"

	"github.com/forbole/juno/v4/models"
	"gorm.io/gorm/clause"
)

// GetMasterDB get master db info
func (db *DB) GetMasterDB(ctx context.Context) (*models.MasterDB, error) {
	var masterDB models.MasterDB

	err := db.Db.Find(&masterDB).Error
	if err != nil && !errIsNotFound(err) {
		return nil, err
	}
	return &masterDB, nil
}

// SetMasterDB set the master db
func (db *DB) SetMasterDB(ctx context.Context, masterDB *models.MasterDB) error {
	err := db.Db.Table((&models.MasterDB{}).TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "one_row_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"is_master"}),
	}).Create(masterDB).Error
	return err
}
