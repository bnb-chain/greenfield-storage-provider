package database

import (
	"context"
	"github.com/forbole/juno/v4/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (db *DB) SaveBucket(ctx context.Context, bucket *models.Bucket) error {
	return nil
}

func (db *DB) UpdateBucket(ctx context.Context, bucket *models.Bucket) error {
	return nil
}

func (db *DB) SaveBucketToSQL(ctx context.Context, bucket *models.Bucket) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.Bucket{}).TableName()).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "bucket_id"}},
		UpdateAll: true,
	}).Create(bucket).Statement

	return stat.SQL.String(), stat.Vars
}

func (db *DB) UpdateBucketToSQL(ctx context.Context, bucket *models.Bucket) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&models.Bucket{}).TableName()).Where("bucket_id = ?", bucket.BucketID).Updates(bucket).Statement
	return stat.SQL.String(), stat.Vars
}
