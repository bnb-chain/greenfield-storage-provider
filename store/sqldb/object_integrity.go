package sqldb

import (
	"encoding/hex"
	"errors"
	"fmt"

	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// GetObjectIntegrity returns the integrity hash info
func (s *SpDBImpl) GetObjectIntegrity(objectID uint64) (*corespdb.IntegrityMeta, error) {
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
	integrityChecksum, err := hex.DecodeString(queryReturn.IntegrityChecksum)
	if err != nil {
		return nil, err
	}
	signature, err := hex.DecodeString(queryReturn.Signature)
	if err != nil {
		return nil, err
	}

	meta := &corespdb.IntegrityMeta{
		ObjectID:          queryReturn.ObjectID,
		IntegrityChecksum: integrityChecksum,
		Signature:         signature,
	}
	meta.PieceChecksumList, err = util.StringToBytesSlice(queryReturn.PieceChecksumList)
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

// SetObjectIntegrity puts(overwrites) integrity hash info to db
func (s *SpDBImpl) SetObjectIntegrity(meta *corespdb.IntegrityMeta) error {
	insertIntegrityMetaRecord := &IntegrityMetaTable{
		ObjectID:          meta.ObjectID,
		PieceChecksumList: util.BytesSliceToString(meta.PieceChecksumList),
		IntegrityChecksum: hex.EncodeToString(meta.IntegrityChecksum),
		Signature:         hex.EncodeToString(meta.Signature),
	}
	result := s.db.Create(insertIntegrityMetaRecord)
	if result.Error != nil && MysqlErrCode(result.Error) == ErrDuplicateEntryCode {
		return nil
	}
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("failed to insert integrity meta record: %s", result.Error)
	}
	return nil
}

// DeleteObjectIntegrity deletes integrity meta info.
func (s *SpDBImpl) DeleteObjectIntegrity(objectID uint64) error {
	return s.db.Delete(&IntegrityMetaTable{
		ObjectID: objectID, // should be the primary key
	}).Error
}

// GetReplicatePieceChecksum gets replicate piece checksum.
func (s *SpDBImpl) GetReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceIdx uint32) ([]byte, error) {
	var (
		err           error
		queryReturn   PieceHashTable
		pieceChecksum []byte
	)
	if err = s.db.Model(&PieceHashTable{}).
		Where("object_id = ? and replicate_index = ? and piece_index = ?", objectID, replicateIdx, pieceIdx).
		First(&queryReturn).Error; err != nil {
		return nil, err
	}
	if pieceChecksum, err = hex.DecodeString(queryReturn.PieceChecksum); err != nil {
		return nil, err
	}
	return pieceChecksum, nil
}

// SetReplicatePieceChecksum sets replicate checksum.
func (s *SpDBImpl) SetReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceIdx uint32, checksum []byte) error {
	var (
		result          *gorm.DB
		insertPieceHash *PieceHashTable
	)
	insertPieceHash = &PieceHashTable{
		ObjectID:       objectID,
		ReplicateIndex: replicateIdx,
		PieceIndex:     pieceIdx,
		PieceChecksum:  hex.EncodeToString(checksum),
	}
	result = s.db.Create(insertPieceHash)
	if result.Error != nil && MysqlErrCode(result.Error) == ErrDuplicateEntryCode {
		return nil
	}
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("failed to insert piece hash record: %s", result.Error)
	}
	return nil
}

// DeleteReplicatePieceChecksum deletes piece checksum.
func (s *SpDBImpl) DeleteReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceIdx uint32) error {
	return s.db.Delete(&PieceHashTable{
		ObjectID:       objectID,     // should be the primary key
		ReplicateIndex: replicateIdx, // should be the primary key
		PieceIndex:     pieceIdx,     // should be the primary key
	}).Error
}

// GetAllReplicatePieceChecksum gets all the piece checksums.
func (s *SpDBImpl) GetAllReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceCount uint32) ([][]byte, error) {
	var (
		pieceIndex        uint32
		err               error
		pieceChecksum     []byte
		pieceChecksumList = make([][]byte, 0)
	)
	for pieceIndex < pieceCount {
		pieceChecksum, err = s.GetReplicatePieceChecksum(objectID, replicateIdx, pieceIndex)
		if err != nil {
			return nil, err
		}
		pieceChecksumList = append(pieceChecksumList, pieceChecksum)
		pieceIndex++
	}
	return pieceChecksumList, nil
}

// SetAllReplicatePieceChecksum is unused.
func (s *SpDBImpl) SetAllReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceCount uint32, checksum [][]byte) error {
	return nil
}

// DeleteAllReplicatePieceChecksum deletes all the piece checksum.
func (s *SpDBImpl) DeleteAllReplicatePieceChecksum(objectID uint64, replicateIdx uint32, pieceCount uint32) error {
	var (
		pieceIndex uint32
		err        error
	)
	for pieceIndex < pieceCount {
		err = s.DeleteReplicatePieceChecksum(objectID, replicateIdx, pieceIndex)
		if err != nil {
			return err
		}
		pieceIndex++
	}
	return nil
}
