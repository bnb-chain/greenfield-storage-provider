package bsdb

const (
	ProcessKeyUpdateBucketSize = "key_update_bucket_size"
)

// DataMigrationRecord stores records of data migration processes.
type DataMigrationRecord struct {
	ProcessKey string `gorm:"column:process_key;not null;primaryKey"`
	// IsCompleted defines if corresponding process has been completed.
	IsCompleted bool `gorm:"column:is_completed;"`
}

// TableName is used to set Master table name in database
func (m *DataMigrationRecord) TableName() string {
	return DataMigrationRecordTableName
}
