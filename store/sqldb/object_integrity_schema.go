package sqldb

import (
	"fmt"
)

const (
	IntegrityMetasNumberOfShards = 64
	ReasonableTableSize          = 5_000_000
	ListObjectsDefaultSize       = 1000
)

// PieceHashTable table schema
type PieceHashTable struct {
	ObjectID        uint64 `gorm:"primary_key"`
	SegmentIndex    uint32 `gorm:"primary_key"`
	RedundancyIndex int32  `gorm:"primary_key"`
	PieceChecksum   string
}

// TableName is used to set PieceHashTable schema's table name in database
func (PieceHashTable) TableName() string {
	return PieceHashTableName
}

// IntegrityMetaTable table schema
type IntegrityMetaTable struct {
	ObjectID          uint64 `gorm:"primary_key"`
	RedundancyIndex   int32  `gorm:"primary_key"`
	IntegrityChecksum string
	PieceChecksumList string
}

// TableName is used to set IntegrityMetaTable schema's table name in database
func (IntegrityMetaTable) TableName() string {
	return IntegrityMetaTableName
}

func GetIntegrityMetasTableName(objectID uint64) string {
	return GetIntegrityMetasTableNameByShardNumber(int(GetIntegrityMetasShardNumberByBucketName(objectID)))
}

// GetIntegrityMetasShardNumberByBucketName Allocate each shard table with 5,000,000 continuous entry before using the next shard table
func GetIntegrityMetasShardNumberByBucketName(objectID uint64) uint64 {
	return objectID / ReasonableTableSize % IntegrityMetasNumberOfShards
}

func GetIntegrityMetasTableNameByShardNumber(shard int) string {
	return fmt.Sprintf("%s_%02d", IntegrityMetaTableName, shard)
}
