package sqldb

// IntegrityMetaTable table schema
type IntegrityMetaTable struct {
	ObjectID      uint64 `gorm:"index:idx_integrity_meta"`
	Checksum      string
	IntegrityHash string
	Signature     string
}

// TableName is used to set IntegrityMetaTable schema's table name in database
func (IntegrityMetaTable) TableName() string {
	return "integrity_meta"
}
