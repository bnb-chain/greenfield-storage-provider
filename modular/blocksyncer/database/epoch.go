package database

import (
	"context"

	"github.com/forbole/juno/v4/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (db *DB) SaveEpoch(ctx context.Context, epoch *models.Epoch) error {
	return nil
}

func (db *DB) SaveEpochToSQL(ctx context.Context, epoch *models.Epoch) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.Epoch{}).TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "one_row_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"block_height", "block_hash", "update_time"}),
	}).Create(epoch).Statement
	return stat.SQL.String(), stat.Vars
}
