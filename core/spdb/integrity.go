package spdb

// IntegrityMeta defines the payload integrity hash and piece checksum with objectID
type IntegrityMeta struct {
	ObjectID          uint64
	IntegrityChecksum []byte
	PieceChecksumList [][]byte
	Signature         []byte
}

/*
// ObjectIntegrityDB abstract object integrity interface
type ObjectIntegrityDB interface {
	// GetObjectIntegrity get integrity meta info by object id
	GetObjectIntegrity(objectID uint64) (*IntegrityMeta, error)
	// SetObjectIntegrity set(maybe overwrite) integrity hash info to db
	SetObjectIntegrity(integrity *IntegrityMeta) error
	DeleteObjectIntegrity(objectID uint64) error

	GetReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceIdx uint32) ([]byte, error)
	SetReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceIdx uint32, checksum []byte) error
	DeleteReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceIdx uint32) error
	GetAllReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceCount uint32) ([][]byte, error)
	SetAllReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceCount uint32, checksum [][]byte) error
	DeleteAllReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceCount uint32) error
}

*/
