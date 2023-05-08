package spdb

type IntegrityMeta struct {
	ObjectID      uint64
	Checksum      [][]byte
	IntegrityHash []byte
	Signature     []byte
}

type ObjectIntegrityDB interface {
	GetObjectIntegrity(objectID uint64) (*IntegrityMeta, error)
	SetObjectIntegrity(integrity *IntegrityMeta) error
	DeleteObjectIntegrity(objectID uint64) error
	GetReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceIdx uint32) ([]byte, error)
	SetReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceIdx uint32, checksum []byte) error
	DeleteReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceIdx uint32) error
	GetAllReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceCount uint32) ([][]byte, error)
	SetAllReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceCount uint32, checksum [][]byte) error
	DeleteAllReplicatePieceChecksum(objectID uint64, replicateIdx uint32) error
}
