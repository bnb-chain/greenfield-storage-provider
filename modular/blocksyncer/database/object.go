package database

import (
	"context"
	"github.com/forbole/juno/v4/models"
	"gorm.io/gorm/clause"

	"github.com/forbole/juno/v4/common"
)

func (db *DB) SaveObject(ctx context.Context, object *models.Object) error {
	err := db.Db.WithContext(ctx).Table((&models.Object{}).TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "object_id"}},
		UpdateAll: true,
	}).Create(object).Error
	return err
}

func (db *DB) UpdateObject(ctx context.Context, object *models.Object) error {
	err := db.Db.WithContext(ctx).Table((&models.Object{}).TableName()).Where("object_id = ?", object.ObjectID).Updates(object).Error
	return err
}

func (db *DB) GetObject(ctx context.Context, objectId common.Hash) (*models.Object, error) {
	var object models.Object

	err := db.Db.WithContext(ctx).Where(
		"object_id = ? AND removed IS NOT TRUE", objectId).Find(&object).Error
	if err != nil {
		return nil, err
	}
	return &object, nil
}
