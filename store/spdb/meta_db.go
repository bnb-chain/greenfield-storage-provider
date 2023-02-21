package spdb

import (
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
)

// IntegrityMeta defines the integrity hash info
type IntegrityMeta struct {
	// ObjectID = primary key
	ObjectID       uint64                `json:"ObjectID"`
	IsPrimary      bool                  `json:"IsPrimary"`
	RedundancyType ptypes.RedundancyType `json:"RedundancyType"`
	EcIdx          uint32                `json:"EcIdx"` // only be used in the secondary sp ec piece

	PieceCount    uint32   `json:"PieceCount"`
	IntegrityHash []byte   `json:"IntegrityHash"`
	PieceHash     [][]byte `json:"PieceHash"`
	Signature     []byte   `json:"Signature"`
}

// UploadPayloadAskingMeta defines the payload asking info
type UploadPayloadAskingMeta struct {
	// BucketName + ObjectName = primary key
	BucketName string `json:"BucketName"`
	ObjectName string `json:"ObjectName"`

	Timeout int64 `json:"Timeout"`
}

type MetaDB interface {
	// SetIntegrityMeta put integrity hash info to db.
	SetIntegrityMeta(meta *IntegrityMeta) error
	// GetIntegrityMeta return the integrity hash info.
	GetIntegrityMeta(objectID uint64) (*IntegrityMeta, error)
	// SetUploadPayloadAskingMeta put payload asking info to db.
	SetUploadPayloadAskingMeta(meta *UploadPayloadAskingMeta) error
	// GetUploadPayloadAskingMeta return the payload asking info.
	GetUploadPayloadAskingMeta(bucketName, objectName string) (*UploadPayloadAskingMeta, error)
	// Close the low level db
	Close() error
}
