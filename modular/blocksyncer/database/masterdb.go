package database

import (
	"context"

	"gorm.io/gorm/clause"

	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

// GetMasterDB get master db info
func (db *DB) GetMasterDB(ctx context.Context) (*bsdb.MasterDB, error) {
	var masterDB bsdb.MasterDB

	err := db.Db.Find(&masterDB).Error
	if err != nil && !errIsNotFound(err) {
		return nil, err
	}
	return &masterDB, nil
}

// SetMasterDB set the master db
func (db *DB) SetMasterDB(ctx context.Context, masterDB *bsdb.MasterDB) error {
	err := db.Db.Table((&bsdb.MasterDB{}).TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "one_row_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"is_master"}),
	}).Create(masterDB).Error
	return err
}
