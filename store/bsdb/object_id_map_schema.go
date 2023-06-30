package bsdb

import (
	"github.com/forbole/juno/v4/common"
)

// ObjectIDMap is a mapping table for objectID and bucketName
type ObjectIDMap struct {
	ObjectID   common.Hash `gorm:"column:object_id;type:BINARY(32);index:idx_object_id;primaryKey"`
	BucketName string      `gorm:"column:bucket_name;type:varchar(64);index:idx_bucket_full_object,priority:1;index:idx_bucket_path,priority:1"`
}

// TableName is used to set ObjectIDMap table name in database
func (*ObjectIDMap) TableName() string {
	return ObjectIDMapTableName
}
