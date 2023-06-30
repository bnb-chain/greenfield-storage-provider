package database

import (
	"context"

	"gorm.io/gorm/clause"

	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

// CreateObjectIDMap create object id map table entry
func (db *DB) CreateObjectIDMap(ctx context.Context, objectIDMap *bsdb.ObjectIDMap) error {
	err := db.Db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "object_id"}},
		UpdateAll: true,
	}).Create(objectIDMap).Error
	return err
}
