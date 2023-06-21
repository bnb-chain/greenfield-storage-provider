package sqldb

// UploadObjectProgressTable table schema.
type UploadObjectProgressTable struct {
	ObjectID              uint64 `gorm:"primary_key"`
	TaskState             int32  `gorm:"index:state_index"`
	GlobalVirtualGroupID  uint32
	TaskStateDescription  string
	ErrorDescription      string
	SecondaryAddresses    string
	SecondarySignatures   string
	CreateTimestampSecond int64
	UpdateTimestampSecond int64 `gorm:"index:update_timestamp_index"`
}

// TableName is used to set UploadObjectProgressTable Schema's table name in database.
func (UploadObjectProgressTable) TableName() string {
	return UploadObjectProgressTableName
}
