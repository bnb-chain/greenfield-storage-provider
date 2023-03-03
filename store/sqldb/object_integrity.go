package sqldb

import (
	"encoding/hex"
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/util"
)

// GetObjectIntegrity return the integrity hash info
func (s *SpDBImpl) GetObjectIntegrity(objectID uint64) (*IntegrityMeta, error) {
	queryReturn := &IntegrityMetaTable{}
	result := s.db.Model(&IntegrityMetaTable{}).
		Where("object_id = ?", objectID).
		First(queryReturn)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query integrity meta record: %s", result.Error)
	}
	integrityHash, err := hex.DecodeString(queryReturn.IntegrityHash)
	if err != nil {
		return nil, err
	}
	signature, err := hex.DecodeString(queryReturn.Signature)
	if err != nil {
		return nil, err
	}

	meta := &IntegrityMeta{
		ObjectID:      queryReturn.ObjectID,
		IntegrityHash: integrityHash,
		Signature:     signature,
	}
	meta.Checksum, err = util.StringToBytesSlice(queryReturn.PieceHashList)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

// SetObjectIntegrity put(overwrite) integrity hash info to db
func (s *SpDBImpl) SetObjectIntegrity(meta *IntegrityMeta) error {
	insertIntegrityMetaRecord := &IntegrityMetaTable{
		ObjectID:      meta.ObjectID,
		PieceHashList: util.BytesSliceToString(meta.Checksum),
		IntegrityHash: hex.EncodeToString(meta.IntegrityHash),
		Signature:     hex.EncodeToString(meta.Signature),
	}
	result := s.db.Create(insertIntegrityMetaRecord)
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("failed to insert integrity meta record: %s", result.Error)
	}
	return nil
}
