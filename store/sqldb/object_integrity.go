package sqldb

import (
	"encoding/hex"
	"errors"

	"gorm.io/gorm"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	errorstypes "github.com/bnb-chain/greenfield-storage-provider/pkg/errors/types"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

// GetObjectIntegrity return the integrity hash info
func (s *SpDBImpl) GetObjectIntegrity(objectID uint64) (*IntegrityMeta, error) {
	queryReturn := &IntegrityMetaTable{}
	result := s.db.Model(&IntegrityMetaTable{}).
		Where("object_id = ?", objectID).
		First(queryReturn)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errorstypes.Error(merrors.DBRecordNotFoundErrCode, result.Error.Error())
	}
	if result.Error != nil {
		return nil, errorstypes.Error(merrors.DBQueryInIntegrityMetaTableErrCode, result.Error.Error())
	}
	integrityHash, err := hex.DecodeString(queryReturn.IntegrityHash)
	if err != nil {
		return nil, errorstypes.Error(merrors.HexDecodeStringErrCode, err.Error())
	}
	signature, err := hex.DecodeString(queryReturn.Signature)
	if err != nil {
		return nil, errorstypes.Error(merrors.HexDecodeStringErrCode, err.Error())
	}

	meta := &IntegrityMeta{
		ObjectID:      queryReturn.ObjectID,
		IntegrityHash: integrityHash,
		Signature:     signature,
	}
	meta.Checksum, err = util.StringToBytesSlice(queryReturn.PieceHashList)
	if err != nil {
		return nil, errorstypes.Error(merrors.StringToByteSliceErrCode, err.Error())
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
		return errorstypes.Error(merrors.DBInsertInIntegrityMetaTableErrCode, result.Error.Error())
	}
	return nil
}
