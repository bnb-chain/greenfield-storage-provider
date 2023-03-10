package sqldb

// IntegrityMetaTable table schema
type IntegrityMetaTable struct {
	ObjectID      string `gorm:"primary_key"`
	PieceHashList string
	IntegrityHash string
	Signature     string
}

// TableName is used to set IntegrityMetaTable schema's table name in database
func (IntegrityMetaTable) TableName() string {
	return IntegrityMetaTableName
}
