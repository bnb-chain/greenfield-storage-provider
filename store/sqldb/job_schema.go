package sqldb

import (
	"time"
)

// JobTable table schema
type JobTable struct {
	JobID        uint64 `gorm:"primary_key;autoIncrement"`
	JobType      int32
	JobState     int32
	JobErrorCode uint32
	CreatedTime  time.Time
	ModifiedTime time.Time
}

// TableName is used to set JobTable Schema's table name in database
func (JobTable) TableName() string {
	return JobTableName
}

// ObjectTable table schema
type ObjectTable struct {
	ObjectID             uint64 `gorm:"primary_key"`
	JobID                uint64 `gorm:"index:job_to_object"` // Job.JobID
	Owner                string
	BucketName           string
	ObjectName           string
	PayloadSize          uint64
	IsPublic             bool
	ContentType          string
	CreatedAtHeight      int64
	ObjectStatus         int32
	RedundancyType       int32
	SourceType           int32
	SpIntegrityHash      string
	SecondarySpAddresses string
}

// TableName is used to set ObjectTable Schema's table name in database
func (ObjectTable) TableName() string {
	return ObjectTableName
}
