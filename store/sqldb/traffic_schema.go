package sqldb

import (
	"time"
)

// BucketTrafficTable table schema
type BucketTrafficTable struct {
	BucketID uint64 `gorm:"primary_key"`
	Month    string `gorm:"primary_key"`

	BucketName    string
	ReadCostSize  int64
	ReadQuotaSize int64
	ModifiedTime  time.Time
}

// TableName is used to set BucketTraffic Schema's table name in database
func (BucketTrafficTable) TableName() string {
	return BucketTrafficTableName
}

// ReadRecordTable table schema
type ReadRecordTable struct {
	ReadRecordID uint64 `gorm:"primary_key;autoIncrement"`

	BucketID    uint64 `gorm:"index:bucket_to_read_record"`
	ObjectID    uint64 `gorm:"index:object_to_read_record"`
	UserAddress string `gorm:"index:user_to_read_record"`
	ReadTime    int64  `gorm:"index:time_to_read_record"` // second timestamp

	BucketName string
	ObjectName string
	ReadSize   int64
}

// TableName is used to set ReadRecord Schema's table name in database
func (ReadRecordTable) TableName() string {
	return ReadRecordTableName
}
