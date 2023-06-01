package spdb

// IntegrityMeta defines the payload integrity hash and piece checksum with objectID.
type IntegrityMeta struct {
	ObjectID          uint64
	IntegrityChecksum []byte
	PieceChecksumList [][]byte
	Signature         []byte
}
