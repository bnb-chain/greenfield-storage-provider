package database

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

// CreateObjectIDMap create object id map table entry
func (db *DB) CreateObjectIDMap(ctx context.Context, objectIDMap *bsdb.ObjectIDMap) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "object_id"}},
		UpdateAll: true,
	}).Create(objectIDMap).Statement
	return stat.SQL.String(), stat.Vars
}
