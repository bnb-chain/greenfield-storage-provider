package metasql

// DBIntegrityMeta table schema
type DBIntegrityMeta struct {
	ObjectID       uint64 `gorm:"index:idx_integrity_meta"`
	IsPrimary      bool   `gorm:"index:idx_integrity_meta"`
	RedundancyType uint32 `gorm:"index:idx_integrity_meta"`
	EcIdx          uint32 `gorm:"index:idx_integrity_meta"`

	PieceCount    uint32
	IntegrityHash string // hex encode string
	PieceHash     string // string(json encode)
}

// TableName is used to set Job Schema's table name in database
func (DBIntegrityMeta) TableName() string {
	return "integrity_meta"
}

// DBUploadPayloadAskingMeta table schema
type DBUploadPayloadAskingMeta struct {
	BucketName string `gorm:"index:idx_upload_payload_asking_meta"`
	ObjectName string `gorm:"index:idx_upload_payload_asking_meta"`

	Timeout int64
}

// TableName is used to set Job Schema's table name in database
func (DBUploadPayloadAskingMeta) TableName() string {
	return "upload_payload_asking_meta"
}
