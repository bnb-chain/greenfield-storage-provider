package database

import (
	"context"
	"gorm.io/gorm"

	"gorm.io/gorm/clause"

	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

// CreateObjectIDMap create object id map table entry
func (db *DB) CreateObjectIDMap(ctx context.Context, objectIDMap *bsdb.ObjectIDMap) string {
	return db.Db.WithContext(ctx).ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "object_id"}},
			UpdateAll: true,
		}).Create(objectIDMap)
	})
}
