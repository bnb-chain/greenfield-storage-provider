package database

import (
	"context"
	"gorm.io/gorm"

	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/models"
	"gorm.io/gorm/clause"

	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

func (db *DB) SaveObject(ctx context.Context, object *models.Object) error {
	return nil
}

func (db *DB) UpdateObject(ctx context.Context, object *models.Object) error {
	return nil
}

func (db *DB) GetObject(ctx context.Context, objectId common.Hash) (*models.Object, error) {
	var object models.Object
	bucketName, err := db.GetBucketNameByObjectID(objectId)

	if err != nil {
		return nil, err
	}

	err = db.Db.WithContext(ctx).Table(bsdb.GetObjectsTableName(bucketName)).Where(
		"object_id = ? AND removed IS NOT TRUE", objectId).Find(&object).Error
	if err != nil {
		return nil, err
	}
	return &object, nil
}

// GetBucketNameByObjectID get bucket name info by an object id
func (b *DB) GetBucketNameByObjectID(objectID common.Hash) (string, error) {
	var (
		objectIdMap *bsdb.ObjectIDMap
		err         error
	)

	err = b.Db.Table((&bsdb.ObjectIDMap{}).TableName()).
		Select("*").
		Where("object_id = ?", objectID).
		Take(&objectIdMap).Error

	return objectIdMap.BucketName, err
}

func (db *DB) SaveObjectToSQL(ctx context.Context, object *models.Object) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table(bsdb.GetObjectsTableName(object.BucketName)).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "object_id"}},
		UpdateAll: true,
	}).Create(object).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) UpdateObjectToSQL(ctx context.Context, object *models.Object) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table(bsdb.GetObjectsTableName(object.BucketName)).Where("object_id = ?", object.ObjectID).Updates(object).Statement
	return stat.SQL.String(), stat.Vars
}
