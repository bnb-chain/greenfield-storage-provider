package sqldb

// UploadEventTable table schema.
type UploadEventTable struct {
	ID          uint64 `gorm:"primary_key;autoIncrement"`
	ObjectID    uint64
	UploadState string
	Description string
	UpdateTime  string
}

// TableName is used to set UploadObjectProgressTable Schema's table name in database.
func (UploadEventTable) TableName() string {
	return UploadEventTableName
}
