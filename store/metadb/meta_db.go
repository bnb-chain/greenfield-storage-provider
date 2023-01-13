package metadb

import types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"

// IntegrityMeta defines the integrity hash info
type IntegrityMeta struct {
	ObjectID       uint64               `json:"ObjectID"`
	PieceIdx       uint32               `json:"PieceIdx"` // only use for ec piece and secondary
	PieceCount     uint32               `json:"PieceCount"`
	IsPrimary      bool                 `json:"IsPrimary"`
	RedundancyType types.RedundancyType `json:"RedundancyType"`

	IntegrityHash []byte   `json:"IntegrityHash"`
	PieceHash     [][]byte `json:"PieceHash"`
}

// UploadPayloadAskingMeta defines the payload asking info
type UploadPayloadAskingMeta struct {
	BucketName string `json:"BucketName"`
	ObjectName string `json:"ObjectName"`
	Timeout    int64  `json:"Timeout"`
}

type MetaDB interface {
	// SetIntegrityMeta put integrity hash info to db.
	SetIntegrityMeta(meta *IntegrityMeta) error
	// GetIntegrityMeta return the integrity hash info.
	GetIntegrityMeta(objectID uint64) (*IntegrityMeta, error)
	// SetUploadPayloadAskingMeta put payload asking info to db.
	SetUploadPayloadAskingMeta(meta *UploadPayloadAskingMeta) error
	// GetUploadPayloadAskingMeta return the payload asking info.
	GetUploadPayloadAskingMeta(bucket, object string) (*UploadPayloadAskingMeta, error)
	// Close the low level db
	Close() error
}
