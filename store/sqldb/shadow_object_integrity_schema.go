package sqldb

// ShadowIntegrityMetaTable table schema
type ShadowIntegrityMetaTable struct {
	ObjectID          uint64 `gorm:"primary_key"`
	RedundancyIndex   int32  `gorm:"primary_key"`
	IntegrityChecksum string
	PieceChecksumList string
	Version           int64
}

// TableName is used to set ShadowIntegrityMetaTable schema's table name in database
func (ShadowIntegrityMetaTable) TableName() string {
	return ShadowIntegrityMetaTableName
}
