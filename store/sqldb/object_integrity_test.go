package sqldb

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
)

func TestSpDBImpl_GetObjectIntegritySuccess(t *testing.T) {
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
	)
	i := &IntegrityMetaTable{
		ObjectID:          objectID,
		RedundancyIndex:   redundancyIndex,
		IntegrityChecksum: "1406e05881e299367766d313e26c05564ec91bf721d31726bd6e46e60689539a",
		PieceChecksumList: "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `integrity_meta_00` WHERE object_id = ? and redundancy_index = ? ORDER BY `integrity_meta_00`.`object_id` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"object_id", "redundancy_index", "integrity_checksum", "piece_checksum_list"}).
			AddRow(i.ObjectID, i.RedundancyIndex, i.IntegrityChecksum, i.PieceChecksumList))
	result, err := s.GetObjectIntegrity(objectID, redundancyIndex)
	assert.Nil(t, err)
	assert.Equal(t, objectID, result.ObjectID)
}

func TestSpDBImpl_GetObjectIntegrityFailure1(t *testing.T) {
	t.Log("Failure case description: record not found")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `integrity_meta_00` WHERE object_id = ? and redundancy_index = ? ORDER BY `integrity_meta_00`.`object_id` LIMIT 1").
		WillReturnError(gorm.ErrRecordNotFound)
	result, err := s.GetObjectIntegrity(objectID, redundancyIndex)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	assert.Nil(t, result)
}

func TestSpDBImpl_GetObjectIntegrityFailure2(t *testing.T) {
	t.Log("Failure case description: mock query db returns error")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `integrity_meta_00` WHERE object_id = ? and redundancy_index = ? ORDER BY `integrity_meta_00`.`object_id` LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.GetObjectIntegrity(objectID, redundancyIndex)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_GetObjectIntegrityFailure3(t *testing.T) {
	t.Log("Failure case description: hex decode string returns error")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
	)
	i := &IntegrityMetaTable{
		ObjectID:          objectID,
		RedundancyIndex:   redundancyIndex,
		IntegrityChecksum: "test",
		PieceChecksumList: "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `integrity_meta_00` WHERE object_id = ? and redundancy_index = ? ORDER BY `integrity_meta_00`.`object_id` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"object_id", "redundancy_index", "integrity_checksum", "piece_checksum_list"}).
			AddRow(i.ObjectID, i.RedundancyIndex, i.IntegrityChecksum, i.PieceChecksumList))
	result, err := s.GetObjectIntegrity(objectID, redundancyIndex)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestSpDBImpl_GetObjectIntegrityFailure4(t *testing.T) {
	t.Log("Failure case description: covert string to bytes slice returns error")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
	)
	i := &IntegrityMetaTable{
		ObjectID:          objectID,
		RedundancyIndex:   redundancyIndex,
		IntegrityChecksum: "1406e05881e299367766d313e26c05564ec91bf721d31726bd6e46e60689539a",
		PieceChecksumList: "test",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `integrity_meta_00` WHERE object_id = ? and redundancy_index = ? ORDER BY `integrity_meta_00`.`object_id` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"object_id", "redundancy_index", "integrity_checksum", "piece_checksum_list"}).
			AddRow(i.ObjectID, i.RedundancyIndex, i.IntegrityChecksum, i.PieceChecksumList))
	result, err := s.GetObjectIntegrity(objectID, redundancyIndex)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestMysqlErrCode1(t *testing.T) {
	t.Log("mysql error")
	infraErr := &mysql.MySQLError{Number: 1234, Message: "the server is on fire"}
	result := MysqlErrCode(infraErr)
	assert.Equal(t, 1234, result)
}

func TestMysqlErrCode2(t *testing.T) {
	t.Log("non mysql error")
	result := MysqlErrCode(mockDBInternalError)
	assert.Equal(t, 0, result)
}

func TestSpDBImpl_SetObjectIntegritySuccess(t *testing.T) {
	meta := &corespdb.IntegrityMeta{
		ObjectID:          10,
		RedundancyIndex:   2,
		IntegrityChecksum: []byte("mockIntegrityChecksum"),
		PieceChecksumList: [][]byte{[]byte("mockPieceChecksumList")},
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `integrity_meta_00` (`object_id`,`redundancy_index`,`integrity_checksum`,`piece_checksum_list`) VALUES (?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.SetObjectIntegrity(meta)
	assert.Nil(t, err)
}

func TestSpDBImpl_SetObjectIntegrityFailure1(t *testing.T) {
	t.Log("Failure case description: duplicate entry code")
	meta := &corespdb.IntegrityMeta{
		ObjectID:          10,
		RedundancyIndex:   2,
		IntegrityChecksum: []byte("mockIntegrityChecksum"),
		PieceChecksumList: [][]byte{[]byte("mockPieceChecksumList")},
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `integrity_meta_00` (`object_id`,`redundancy_index`,`integrity_checksum`,`piece_checksum_list`) VALUES (?,?,?,?)").
		WillReturnError(&mysql.MySQLError{Number: uint16(ErrDuplicateEntryCode), Message: "duplicate entry code"})
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.SetObjectIntegrity(meta)
	assert.Nil(t, err)
}

func TestSpDBImpl_SetObjectIntegrityFailure2(t *testing.T) {
	t.Log("Failure case description: mock db insert returns error")
	meta := &corespdb.IntegrityMeta{
		ObjectID:          10,
		RedundancyIndex:   2,
		IntegrityChecksum: []byte("mockIntegrityChecksum"),
		PieceChecksumList: [][]byte{[]byte("mockPieceChecksumList")},
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `integrity_meta_00` (`object_id`,`redundancy_index`,`integrity_checksum`,`piece_checksum_list`) VALUES (?,?,?,?)").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.SetObjectIntegrity(meta)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_UpdateIntegrityChecksumSuccess(t *testing.T) {
	integrityChecksum := []byte("mockIntegrityChecksum")
	meta := &corespdb.IntegrityMeta{
		ObjectID:          10,
		RedundancyIndex:   2,
		IntegrityChecksum: integrityChecksum,
		PieceChecksumList: [][]byte{[]byte("mockPieceChecksumList")},
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `integrity_meta_00` SET `integrity_checksum`=? WHERE object_id = ? and redundancy_index = ?").
		WithArgs(hex.EncodeToString(integrityChecksum), meta.ObjectID, meta.RedundancyIndex).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateIntegrityChecksum(meta)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateIntegrityChecksumFailure(t *testing.T) {
	integrityChecksum := []byte("mockIntegrityChecksum")
	meta := &corespdb.IntegrityMeta{
		ObjectID:          10,
		RedundancyIndex:   2,
		IntegrityChecksum: integrityChecksum,
		PieceChecksumList: [][]byte{[]byte("mockPieceChecksumList")},
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `integrity_meta_00` SET `integrity_checksum`=? WHERE object_id = ? and redundancy_index = ?").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.UpdateIntegrityChecksum(meta)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_DeleteObjectIntegritySuccess(t *testing.T) {
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
	)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `integrity_meta_00` WHERE (`integrity_meta_00`.`object_id`,`integrity_meta_00`.`redundancy_index`) IN ((?,?))").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.DeleteObjectIntegrity(objectID, redundancyIndex)
	assert.Nil(t, err)
}

func TestSpDBImpl_DeleteObjectIntegrityFailure(t *testing.T) {
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
	)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `integrity_meta_00` WHERE (`integrity_meta_00`.`object_id`,`integrity_meta_00`.`redundancy_index`) IN ((?,?))").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.DeleteObjectIntegrity(objectID, redundancyIndex)
	assert.Equal(t, mockDBInternalError, err)
}

func TestSpDBImpl_UpdatePieceChecksumSuccess1(t *testing.T) {
	t.Log("Success case description: get object integrity has data and update table")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
		checksum        = []byte("mockChecksum")
		newChecksum     = "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d,6d6f636b436865636b73756d"
	)
	i := &IntegrityMetaTable{
		ObjectID:          objectID,
		RedundancyIndex:   redundancyIndex,
		IntegrityChecksum: "1406e05881e299367766d313e26c05564ec91bf721d31726bd6e46e60689539a",
		PieceChecksumList: "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `integrity_meta_00` WHERE object_id = ? and redundancy_index = ? ORDER BY `integrity_meta_00`.`object_id` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"object_id", "redundancy_index", "integrity_checksum", "piece_checksum_list"}).
			AddRow(i.ObjectID, i.RedundancyIndex, i.IntegrityChecksum, i.PieceChecksumList))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `integrity_meta_00` SET `piece_checksum_list`=? WHERE object_id = ? and redundancy_index = ?").
		WithArgs(newChecksum, i.ObjectID, i.RedundancyIndex).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdatePieceChecksum(objectID, redundancyIndex, checksum)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdatePieceChecksumSuccess2(t *testing.T) {
	t.Log("Success case description: get object integrity has no data and insert a new row")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
		checksum        = []byte("mockChecksum")
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `integrity_meta_00` WHERE object_id = ? and redundancy_index = ? ORDER BY `integrity_meta_00`.`object_id` LIMIT 1").
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `integrity_meta_00` (`object_id`,`redundancy_index`,`integrity_checksum`,`piece_checksum_list`) VALUES (?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdatePieceChecksum(objectID, redundancyIndex, checksum)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdatePieceChecksumFailure1(t *testing.T) {
	t.Log("Failure case description: get object integrity has no data and insert a new row returns error")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
		checksum        = []byte("mockChecksum")
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `integrity_meta_00` WHERE object_id = ? and redundancy_index = ? ORDER BY `integrity_meta_00`.`object_id` LIMIT 1").
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `integrity_meta_00` (`object_id`,`redundancy_index`,`integrity_checksum`,`piece_checksum_list`) VALUES (?,?,?,?)").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.UpdatePieceChecksum(objectID, redundancyIndex, checksum)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_UpdatePieceChecksumFailure2(t *testing.T) {
	t.Log("Failure case description: get object integrity returns non record not found error")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
		checksum        = []byte("mockChecksum")
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `integrity_meta_00` WHERE object_id = ? and redundancy_index = ? ORDER BY `integrity_meta_00`.`object_id` LIMIT 1").
		WillReturnError(mockDBInternalError)
	err := s.UpdatePieceChecksum(objectID, redundancyIndex, checksum)
	fmt.Println(err)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_UpdatePieceChecksumFailure3(t *testing.T) {
	t.Log("Failure case description: update object integrity returns error")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
		checksum        = []byte("mockChecksum")
	)
	i := &IntegrityMetaTable{
		ObjectID:          objectID,
		RedundancyIndex:   redundancyIndex,
		IntegrityChecksum: "1406e05881e299367766d313e26c05564ec91bf721d31726bd6e46e60689539a",
		PieceChecksumList: "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `integrity_meta_00` WHERE object_id = ? and redundancy_index = ? ORDER BY `integrity_meta_00`.`object_id` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"object_id", "redundancy_index", "integrity_checksum", "piece_checksum_list"}).
			AddRow(i.ObjectID, i.RedundancyIndex, i.IntegrityChecksum, i.PieceChecksumList))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `integrity_meta_00` SET `piece_checksum_list`=? WHERE object_id = ? and redundancy_index = ?").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.UpdatePieceChecksum(objectID, redundancyIndex, checksum)
	fmt.Println(err)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_GetReplicatePieceChecksumSuccess(t *testing.T) {
	var (
		objectID      = uint64(9)
		segmentIdx    = uint32(3)
		redundancyIdx = int32(5)
		pieceChecksum = "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d"
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `piece_hash` WHERE object_id = ? and segment_index = ? and redundancy_index = ? ORDER BY `piece_hash`.`object_id` LIMIT 1").
		WithArgs(objectID, segmentIdx, redundancyIdx).WillReturnRows(sqlmock.NewRows([]string{"object_id", "segment_index",
		"redundancy_index", "piece_checksum"}).AddRow(objectID, segmentIdx, redundancyIdx, pieceChecksum))
	result, err := s.GetReplicatePieceChecksum(objectID, segmentIdx, redundancyIdx)
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestSpDBImpl_GetReplicatePieceChecksumFailure1(t *testing.T) {
	t.Log("Failure case description: query db returns error")
	var (
		objectID      = uint64(9)
		segmentIdx    = uint32(3)
		redundancyIdx = int32(5)
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `piece_hash` WHERE object_id = ? and segment_index = ? and redundancy_index = ? ORDER BY `piece_hash`.`object_id` LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.GetReplicatePieceChecksum(objectID, segmentIdx, redundancyIdx)
	assert.Equal(t, mockDBInternalError, err)
	assert.Nil(t, result)
}

func TestSpDBImpl_GetReplicatePieceChecksumFailure2(t *testing.T) {
	t.Log("Failure case description: hex decode string returns error")
	var (
		objectID      = uint64(9)
		segmentIdx    = uint32(3)
		redundancyIdx = int32(5)
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `piece_hash` WHERE object_id = ? and segment_index = ? and redundancy_index = ? ORDER BY `piece_hash`.`object_id` LIMIT 1").
		WithArgs(objectID, segmentIdx, redundancyIdx).WillReturnRows(sqlmock.NewRows([]string{"object_id", "segment_index",
		"redundancy_index", "piece_checksum"}).AddRow(objectID, segmentIdx, redundancyIdx, "test"))
	result, err := s.GetReplicatePieceChecksum(objectID, segmentIdx, redundancyIdx)
	assert.Equal(t, hex.InvalidByteError(0x74), err)
	assert.Nil(t, result)
}

func TestSpDBImpl_SetReplicatePieceChecksumSuccess(t *testing.T) {
	var (
		objectID      = uint64(9)
		segmentIdx    = uint32(3)
		redundancyIdx = int32(5)
		pieceChecksum = []byte("mock")
	)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `piece_hash` (`object_id`,`segment_index`,`redundancy_index`,`piece_checksum`) VALUES (?,?,?,?)").
		WithArgs(objectID, segmentIdx, redundancyIdx, hex.EncodeToString(pieceChecksum)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.SetReplicatePieceChecksum(objectID, segmentIdx, redundancyIdx, pieceChecksum)
	assert.Nil(t, err)
}

func TestSpDBImpl_SetReplicatePieceChecksumFailure1(t *testing.T) {
	t.Log("Failure case description: duplicate entry code")
	var (
		objectID      = uint64(9)
		segmentIdx    = uint32(3)
		redundancyIdx = int32(5)
		pieceChecksum = []byte("mock")
	)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `piece_hash` (`object_id`,`segment_index`,`redundancy_index`,`piece_checksum`) VALUES (?,?,?,?)").
		WillReturnError(&mysql.MySQLError{Number: uint16(ErrDuplicateEntryCode), Message: "duplicate entry code"})
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.SetReplicatePieceChecksum(objectID, segmentIdx, redundancyIdx, pieceChecksum)
	assert.Nil(t, err)
}

func TestSpDBImpl_SetReplicatePieceChecksumFailure2(t *testing.T) {
	t.Log("Failure case description: mock db insert returns error")
	var (
		objectID      = uint64(9)
		segmentIdx    = uint32(3)
		redundancyIdx = int32(5)
		pieceChecksum = []byte("mock")
	)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `piece_hash` (`object_id`,`segment_index`,`redundancy_index`,`piece_checksum`) VALUES (?,?,?,?)").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.SetReplicatePieceChecksum(objectID, segmentIdx, redundancyIdx, pieceChecksum)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_DeleteReplicatePieceChecksumSuccess(t *testing.T) {
	var (
		objectID      = uint64(9)
		segmentIdx    = uint32(3)
		redundancyIdx = int32(5)
	)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `piece_hash` WHERE (`piece_hash`.`object_id`,`piece_hash`.`segment_index`,`piece_hash`.`redundancy_index`) IN ((?,?,?))").
		WithArgs(objectID, segmentIdx, redundancyIdx).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.DeleteReplicatePieceChecksum(objectID, segmentIdx, redundancyIdx)
	assert.Nil(t, err)
}

func TestSpDBImpl_GetAllReplicatePieceChecksumSuccess(t *testing.T) {
	var (
		objectID      = uint64(9)
		segmentIdx    = uint32(0)
		redundancyIdx = int32(5)
		pieceCount    = uint32(1)
		pieceChecksum = "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d"
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `piece_hash` WHERE object_id = ? and segment_index = ? and redundancy_index = ? ORDER BY `piece_hash`.`object_id` LIMIT 1").
		WithArgs(objectID, segmentIdx, redundancyIdx).WillReturnRows(sqlmock.NewRows([]string{"object_id", "segment_index",
		"redundancy_index", "piece_checksum"}).AddRow(objectID, segmentIdx, redundancyIdx, pieceChecksum))
	result, err := s.GetAllReplicatePieceChecksum(objectID, redundancyIdx, pieceCount)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
}

func TestSpDBImpl_GetAllReplicatePieceChecksumFailure(t *testing.T) {
	var (
		objectID      = uint64(9)
		redundancyIdx = int32(5)
		pieceCount    = uint32(1)
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `piece_hash` WHERE object_id = ? and segment_index = ? and redundancy_index = ? ORDER BY `piece_hash`.`object_id` LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.GetAllReplicatePieceChecksum(objectID, redundancyIdx, pieceCount)
	assert.Equal(t, mockDBInternalError, err)
	assert.Nil(t, result)
}

func TestSpDBImpl_DeleteAllReplicatePieceChecksumSuccess(t *testing.T) {
	var (
		objectID      = uint64(9)
		segmentIdx    = uint32(0)
		redundancyIdx = int32(5)
		pieceCount    = uint32(1)
	)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `piece_hash` WHERE (`piece_hash`.`object_id`,`piece_hash`.`segment_index`,`piece_hash`.`redundancy_index`) IN ((?,?,?))").
		WithArgs(objectID, segmentIdx, redundancyIdx).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.DeleteAllReplicatePieceChecksum(objectID, redundancyIdx, pieceCount)
	assert.Nil(t, err)
}

func TestSpDBImpl_DeleteAllReplicatePieceChecksumFailure(t *testing.T) {
	var (
		objectID      = uint64(9)
		redundancyIdx = int32(5)
		pieceCount    = uint32(1)
	)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `piece_hash` WHERE (`piece_hash`.`object_id`,`piece_hash`.`segment_index`,`piece_hash`.`redundancy_index`) IN ((?,?,?))").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.DeleteAllReplicatePieceChecksum(objectID, redundancyIdx, pieceCount)
	assert.Equal(t, mockDBInternalError, err)
}
