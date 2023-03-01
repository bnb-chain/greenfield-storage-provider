package spdb

// import (
//
//	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
//
// )
//
// // IntegrityMeta defines the integrity hash info
//
//	type IntegrityMeta struct {
//		// ObjectID = primary key
//		ObjectID       uint64                `json:"ObjectID"`
//		IsPrimary      bool                  `json:"IsPrimary"`
//		RedundancyType ptypes.RedundancyType `json:"RedundancyType"`
//		EcIdx          uint32                `json:"EcIdx"` // only be used in the secondary sp ec piece
//
//		PieceCount    uint32   `json:"PieceCount"`
//		IntegrityHash []byte   `json:"IntegrityHash"`
//		PieceHash     [][]byte `json:"PieceHash"`
//		Signature     []byte   `json:"Signature"`
//	}
type MetaDB interface{}

//	// SetIntegrityMeta put integrity hash info to db.
//	SetIntegrityMeta(meta *IntegrityMeta) error
//	// GetIntegrityMeta return the integrity hash info.
//	GetIntegrityMeta(objectID uint64) (*IntegrityMeta, error)
//	// Close the low level db
//	Close() error
//}
