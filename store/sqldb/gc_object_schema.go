package sqldb

// GCObjectProgressTable table schema
type GCObjectProgressTable struct {
	TaskKey               string `gorm:"primary_key"`
	StartGCBlockID        uint64
	EndGCBlockID          uint64
	CurrentGCBlockID      uint64
	LastDeletedObjectID   uint64
	CreateTimestampSecond int64
	UpdateTimestampSecond int64 `gorm:"index:update_timestamp_index"`
}

// TableName is used to set GCObjectProgressTable Schema's table name in database
func (GCObjectProgressTable) TableName() string {
	return GCObjectProgressTableName
}
