package sqldb

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// GetObjectIntegrity return the integrity hash info
func (s *SpDBImpl) GetObjectIntegrity(objectID uint64) (*IntegrityMeta, error) {
	queryReturn := &IntegrityMetaTable{}
	result := s.db.Model(&IntegrityMetaTable{}).
		Where("object_id = ?", objectID).
		First(queryReturn)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}
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

func MysqlErrCode(err error) int {
	mysqlErr, ok := err.(*mysql.MySQLError)
	if !ok {
		return 0
	}
	return int(mysqlErr.Number)
}

var (
	ErrDuplicateEntryCode = 1062
)

// SetObjectIntegrity put(overwrite) integrity hash info to db
func (s *SpDBImpl) SetObjectIntegrity(meta *IntegrityMeta) error {
	insertIntegrityMetaRecord := &IntegrityMetaTable{
		ObjectID:      meta.ObjectID,
		PieceHashList: util.BytesSliceToString(meta.Checksum),
		IntegrityHash: hex.EncodeToString(meta.IntegrityHash),
		Signature:     hex.EncodeToString(meta.Signature),
	}
	result := s.db.Create(insertIntegrityMetaRecord)
	if result.Error != nil || MysqlErrCode(result.Error) == ErrDuplicateEntryCode {
		return nil
	}
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("failed to insert integrity meta record: %s", result.Error)
	}
	return nil
}
