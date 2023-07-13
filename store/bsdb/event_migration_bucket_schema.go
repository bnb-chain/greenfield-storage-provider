package bsdb

import "github.com/forbole/juno/v4/common"

type EventMigrationBucket struct {
	ID             uint64         `gorm:"column:id;primaryKey"`
	BucketID       common.Hash    `gorm:"column:bucket_id;type:BINARY(32);index:idx_bucket_id"`
	Operator       common.Address `gorm:"column:operator;type:BINARY(20)"`
	BucketName     string         `gorm:"column:bucket_name;type:varchar(64);index:idx_bucket_name"`
	DstPrimarySpId uint32         `gorm:"column:dst_primary_sp_id"`

	CreateAt     int64       `gorm:"column:create_at"`
	CreateTxHash common.Hash `gorm:"column:create_tx_hash;type:BINARY(32);not null"`
	CreateTime   int64       `gorm:"column:create_time"` // seconds
}

func (*EventMigrationBucket) TableName() string {
	return EventMigrationTableName
}
