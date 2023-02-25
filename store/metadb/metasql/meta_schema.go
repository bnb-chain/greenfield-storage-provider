package metasql

// DBIntegrityMeta table schema
type DBIntegrityMeta struct {
	ObjectID       uint64 `gorm:"index:idx_integrity_meta"`
	IsPrimary      bool
	RedundancyType uint32
	EcIdx          uint32

	PieceCount    uint32
	IntegrityHash string // hex encode string
	PieceHash     string
	Signature     string // hex encode string
}

// TableName is used to set Job Schema's table name in database
func (DBIntegrityMeta) TableName() string {
	return "integrity_meta"
}
