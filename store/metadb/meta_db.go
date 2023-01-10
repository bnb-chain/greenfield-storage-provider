package metadb

import types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"

// IntegrityMeta defines the integrity hash info
type IntegrityMeta struct {
	ObjectID       uint64               `json:"ObjectID"`
	PieceIdx       uint32               `json:"PieceIdx"`
	PieceCount     uint32               `json:"PieceCount"`
	IsPrimary      bool                 `json:"IsPrimary"`
	RedundancyType types.RedundancyType `json:"RedundancyType"`

	IntegrityHash []byte            `json:"IntegrityHash"`
	PieceHash     map[string][]byte `json:"PieceHash"`
}

type MetaDB interface {
	// SetIntegrityMeta put integrity hash info to db.
	SetIntegrityMeta(meta *IntegrityMeta) error
	// GetIntegrityMeta return the integrity hash info
	GetIntegrityMeta(objectID uint64) (*IntegrityMeta, error)
	// Close the low level db
	Close() error
}
