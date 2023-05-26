package sqldb

// GCObjectTaskTable table schema
type GCObjectTaskTable struct {
	TaskKey                string `gorm:"primary_key"`
	CurrentDeletingBlockID uint64
	LastDeletedObjectID    uint64
}

// TableName is used to set JobTable Schema's table name in database
func (GCObjectTaskTable) TableName() string {
	return GCObjectTaskTableName
}
