package sqldb

import (
	"time"
)

// BucketTrafficTable table schema
type BucketTrafficTable struct {
	BucketID                     uint64 `gorm:"primary_key"`
	Month                        string `gorm:"primary_key"`
	BucketName                   string
	ReadConsumedSize             uint64 // indicates the consumed chargedQuota of this month
	FreeQuotaConsumedSize        uint64 // indicates the consumed free quota size of this month
	MonthlyFreeQuotaConsumedSize uint64 // indicates the consumed monthly free quota size of this month
	FreeQuotaSize                uint64 // indicate the remained free quota
	ChargedQuotaSize             uint64 // indicate the greenfield chain bucket charged quota
	MonthlyQuotaSize             uint64 // indicate the remained monthly free quota
	ModifiedTime                 time.Time
}

// TableName is used to set BucketTraffic Schema's table name in database
func (BucketTrafficTable) TableName() string {
	return BucketTrafficTableName
}

// ReadRecordTable table schema
type ReadRecordTable struct {
	ReadRecordID    uint64 `gorm:"primary_key;autoIncrement"`
	BucketID        uint64 `gorm:"index:bucket_to_read_record"`
	ObjectID        uint64 `gorm:"index:object_to_read_record"`
	UserAddress     string `gorm:"index:user_to_read_record"`
	ReadTimestampUs int64  `gorm:"index:time_to_read_record"` // microsecond timestamp
	BucketName      string
	ObjectName      string
	ReadSize        uint64
}

// TableName is used to set ReadRecord Schema's table name in database
func (ReadRecordTable) TableName() string {
	return ReadRecordTableName
}
