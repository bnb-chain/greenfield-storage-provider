package database

import (
	"context"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/models"
	"gorm.io/gorm/clause"
)

func (db *DB) SaveObject(ctx context.Context, object *models.Object) error {
	err := db.Db.WithContext(ctx).Scopes(bsdb.ReadObjectsTable(object.BucketName)).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "object_id"}},
		UpdateAll: true,
	}).Create(object).Error
	return err
}

func (db *DB) UpdateObject(ctx context.Context, object *models.Object) error {
	err := db.Db.WithContext(ctx).Scopes(bsdb.ReadObjectsTable(object.BucketName)).Where("object_id = ?", object.ObjectID).Updates(object).Error
	return err
}

func (db *DB) GetObject(ctx context.Context, objectId common.Hash) (*models.Object, error) {
	var object models.Object

	err := db.Db.WithContext(ctx).Scopes(bsdb.ReadObjectsTable(object.BucketName)).Where(
		"object_id = ? AND removed IS NOT TRUE", objectId).Find(&object).Error
	if err != nil {
		return nil, err
	}
	return &object, nil
}
