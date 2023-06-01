package sqldb

// GCObjectProgressTable table schema
type GCObjectProgressTable struct {
	TaskKey                string `gorm:"primary_key"`
	CurrentDeletingBlockID uint64
	LastDeletedObjectID    uint64
	// TODO: timestamp
}

// TableName is used to set GCObjectProgressTable Schema's table name in database
func (GCObjectProgressTable) TableName() string {
	return GCObjectProgressTableName
}
