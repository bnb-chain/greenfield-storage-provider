package sqldb

// UploadObjectProgressTable table schema
type UploadObjectProgressTable struct {
	ObjectID             uint64 `gorm:"primary_key"`
	TaskState            int32
	TaskStateDescription string
	ErrorDescription     string
	// TODO: add index for load
	CreateTimestampSecond int64
	UpdateTimestampSecond int64
}

// TableName is used to set UploadObjectProgressTable Schema's table name in database
func (UploadObjectProgressTable) TableName() string {
	return UploadObjectProgressTableName
}
