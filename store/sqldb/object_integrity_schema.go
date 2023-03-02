package sqldb

// IntegrityMetaTable table schema
type IntegrityMetaTable struct {
	ObjectID      uint64 `gorm:"primary_key"`
	Checksum      string
	IntegrityHash string
	Signature     string
}

// TableName is used to set IntegrityMetaTable schema's table name in database
func (IntegrityMetaTable) TableName() string {
	return IntegrityMetaTableName
}
