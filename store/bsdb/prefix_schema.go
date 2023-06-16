package bsdb

//
//import (
//	"github.com/forbole/juno/v4/common"
//)
//
//// SlashPrefixTreeNode A tree structure based on prefixes
//type SlashPrefixTreeNode struct {
//	ID uint64 `gorm:"column:id;primaryKey"`
//
//	PathName string `gorm:"column:path_name;type:varchar(1024);index:idx_bucket_path,priority:2,length:512"`
//	FullName string `gorm:"column:full_name;type:varchar(1024);index:idx_bucket_full_object,priority:2,length:512"`
//	Name     string `gorm:"column:name;type:varchar(1024)"`
//	IsObject bool   `gorm:"column:is_object;default:false;index:idx_bucket_full_object,priority:3"`
//	IsFolder bool   `gorm:"column:is_folder;default:false"`
//
//	BucketName string      `gorm:"column:bucket_name;type:varchar(64);index:idx_bucket_full_object,priority:1;index:idx_bucket_path,priority:1"`
//	ObjectID   common.Hash `gorm:"column:object_id;type:BINARY(32);index:idx_object_id"`
//	ObjectName string      `gorm:"column:object_name;type:varchar(1024)"`
//}
//
//// TableName is used to set SlashPrefixTreeNode table name in database
//func (*SlashPrefixTreeNode) TableName() string {
//	return PrefixTreeTableName
//}
