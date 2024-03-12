package sqldb

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestSpDBImpl_GetShadowObjectIntegritySuccess(t *testing.T) {
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
	)
	i := &ShadowIntegrityMetaTable{
		ObjectID:          objectID,
		RedundancyIndex:   redundancyIndex,
		IntegrityChecksum: "1406e05881e299367766d313e26c05564ec91bf721d31726bd6e46e60689539a",
		PieceChecksumList: "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d",
		Version:           1,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `shadow_integrity_meta` WHERE object_id = ? and redundancy_index = ? ORDER BY `shadow_integrity_meta`.`object_id` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"object_id", "redundancy_index", "integrity_checksum", "piece_checksum_list", "version"}).
			AddRow(i.ObjectID, i.RedundancyIndex, i.IntegrityChecksum, i.PieceChecksumList, i.Version))
	result, err := s.GetShadowObjectIntegrity(objectID, redundancyIndex)
	assert.Nil(t, err)
	assert.Equal(t, objectID, result.ObjectID)
}

func TestSpDBImpl_GetShadowObjectIntegrityFailure1(t *testing.T) {
	t.Log("Failure case description: record not found")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `shadow_integrity_meta` WHERE object_id = ? and redundancy_index = ? ORDER BY `shadow_integrity_meta`.`object_id` LIMIT 1").
		WillReturnError(gorm.ErrRecordNotFound)
	result, err := s.GetShadowObjectIntegrity(objectID, redundancyIndex)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	assert.Nil(t, result)
}

func TestSpDBImpl_GetShadowObjectIntegrityFailure2(t *testing.T) {
	t.Log("Failure case description: mock query db returns error")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `shadow_integrity_meta` WHERE object_id = ? and redundancy_index = ? ORDER BY `shadow_integrity_meta`.`object_id` LIMIT 1").
		WillReturnError(mockDBInternalError)
	result, err := s.GetShadowObjectIntegrity(objectID, redundancyIndex)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Nil(t, result)
}

func TestSpDBImpl_GetShadowObjectIntegrityFailure3(t *testing.T) {
	t.Log("Failure case description: hex decode string returns error")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
	)
	i := &ShadowIntegrityMetaTable{
		ObjectID:          objectID,
		RedundancyIndex:   redundancyIndex,
		IntegrityChecksum: "test",
		PieceChecksumList: "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d",
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `shadow_integrity_meta` WHERE object_id = ? and redundancy_index = ? ORDER BY `shadow_integrity_meta`.`object_id` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"object_id", "redundancy_index", "integrity_checksum", "piece_checksum_list"}).
			AddRow(i.ObjectID, i.RedundancyIndex, i.IntegrityChecksum, i.PieceChecksumList))
	result, err := s.GetShadowObjectIntegrity(objectID, redundancyIndex)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestSpDBImpl_GetShadowObjectIntegrityFailure4(t *testing.T) {
	t.Log("Failure case description: covert string to bytes slice returns error")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
	)
	i := &ShadowIntegrityMetaTable{
		ObjectID:          objectID,
		RedundancyIndex:   redundancyIndex,
		IntegrityChecksum: "1406e05881e299367766d313e26c05564ec91bf721d31726bd6e46e60689539a",
		PieceChecksumList: "test",
		Version:           int64(1),
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `shadow_integrity_meta` WHERE object_id = ? and redundancy_index = ? ORDER BY `shadow_integrity_meta`.`object_id` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"object_id", "redundancy_index", "integrity_checksum", "piece_checksum_list", "version"}).
			AddRow(i.ObjectID, i.RedundancyIndex, i.IntegrityChecksum, i.PieceChecksumList, i.Version))
	result, err := s.GetShadowObjectIntegrity(objectID, redundancyIndex)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

func TestSpDBImpl_SetShadowObjectIntegritySuccess(t *testing.T) {
	meta := &corespdb.ShadowIntegrityMeta{
		ObjectID:          10,
		RedundancyIndex:   2,
		IntegrityChecksum: []byte("mockIntegrityChecksum"),
		PieceChecksumList: [][]byte{[]byte("mockPieceChecksumList")},
		Version:           int64(1),
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `shadow_integrity_meta` (`object_id`,`redundancy_index`,`integrity_checksum`,`piece_checksum_list`,`version`) VALUES (?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.SetShadowObjectIntegrity(meta)
	assert.Nil(t, err)
}

func TestSpDBImpl_SetShadowObjectIntegrityFailure1(t *testing.T) {
	t.Log("Failure case description: duplicate entry code")
	meta := &corespdb.ShadowIntegrityMeta{
		ObjectID:          10,
		RedundancyIndex:   2,
		IntegrityChecksum: []byte("mockIntegrityChecksum"),
		PieceChecksumList: [][]byte{[]byte("mockPieceChecksumList")},
		Version:           int64(1),
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `shadow_integrity_meta` (`object_id`,`redundancy_index`,`integrity_checksum`,`piece_checksum_list`,`version`) VALUES (?,?,?,?,?)").
		WillReturnError(&mysql.MySQLError{Number: uint16(ErrDuplicateEntryCode), Message: "duplicate entry code"})
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.SetShadowObjectIntegrity(meta)
	assert.Nil(t, err)
}

func TestSpDBImpl_SetShadowObjectIntegrityFailure2(t *testing.T) {
	t.Log("Failure case description: mock db insert returns error")
	meta := &corespdb.ShadowIntegrityMeta{
		ObjectID:          10,
		RedundancyIndex:   2,
		IntegrityChecksum: []byte("mockIntegrityChecksum"),
		PieceChecksumList: [][]byte{[]byte("mockPieceChecksumList")},
		Version:           int64(1),
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `shadow_integrity_meta` (`object_id`,`redundancy_index`,`integrity_checksum`,`piece_checksum_list`,`version`) VALUES (?,?,?,?,?)").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.SetShadowObjectIntegrity(meta)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_UpdateShadowIntegrityChecksumSuccess(t *testing.T) {
	integrityChecksum := []byte("mockIntegrityChecksum")
	meta := &corespdb.ShadowIntegrityMeta{
		ObjectID:          10,
		RedundancyIndex:   2,
		IntegrityChecksum: integrityChecksum,
		PieceChecksumList: [][]byte{[]byte("mockPieceChecksumList")},
		Version:           int64(1),
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `shadow_integrity_meta` SET `integrity_checksum`=? WHERE object_id = ? and redundancy_index = ?").
		WithArgs(hex.EncodeToString(integrityChecksum), meta.ObjectID, meta.RedundancyIndex).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateShadowIntegrityChecksum(meta)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateShadowIntegrityChecksumFailure(t *testing.T) {
	integrityChecksum := []byte("mockIntegrityChecksum")
	meta := &corespdb.ShadowIntegrityMeta{
		ObjectID:          10,
		RedundancyIndex:   2,
		IntegrityChecksum: integrityChecksum,
		PieceChecksumList: [][]byte{[]byte("mockPieceChecksumList")},
		Version:           int64(1),
	}
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `shadow_integrity_meta` SET `integrity_checksum`=? WHERE object_id = ? and redundancy_index = ?").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.UpdateShadowIntegrityChecksum(meta)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_DeleteShadowObjectIntegritySuccess(t *testing.T) {
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
	)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `shadow_integrity_meta` WHERE (`shadow_integrity_meta`.`object_id`,`shadow_integrity_meta`.`redundancy_index`) IN ((?,?))").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.DeleteShadowObjectIntegrity(objectID, redundancyIndex)
	assert.Nil(t, err)
}

func TestSpDBImpl_DeleteShadowObjectIntegrityFailure(t *testing.T) {
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
	)
	s, mock := setupDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `shadow_integrity_meta` WHERE (`shadow_integrity_meta`.`object_id`,`shadow_integrity_meta`.`redundancy_index`) IN ((?,?))").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.DeleteShadowObjectIntegrity(objectID, redundancyIndex)
	assert.Equal(t, mockDBInternalError, err)
}

func TestSpDBImpl_UpdateShadowPieceChecksumSuccess1(t *testing.T) {
	t.Log("Success case description: get object integrity has data and update table")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
		checksum        = []byte("mockChecksum")
		newChecksum     = "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d,6d6f636b436865636b73756d"
		version         = int64(1)
	)
	i := &ShadowIntegrityMetaTable{
		ObjectID:          objectID,
		RedundancyIndex:   redundancyIndex,
		IntegrityChecksum: "1406e05881e299367766d313e26c05564ec91bf721d31726bd6e46e60689539a",
		PieceChecksumList: "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d",
		Version:           int64(2),
		PieceSize:         1,
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `shadow_integrity_meta` WHERE object_id = ? and redundancy_index = ? ORDER BY `shadow_integrity_meta`.`object_id` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"object_id", "redundancy_index", "integrity_checksum", "piece_checksum_list", "version"}).
			AddRow(i.ObjectID, i.RedundancyIndex, i.IntegrityChecksum, i.PieceChecksumList, i.Version))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `shadow_integrity_meta` SET `piece_checksum_list`=? WHERE object_id = ? and redundancy_index = ?").
		WithArgs(newChecksum, i.ObjectID, i.RedundancyIndex).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateShadowPieceChecksum(objectID, redundancyIndex, checksum, version, 1)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateShadowPieceChecksumSuccess2(t *testing.T) {
	t.Log("Success case description: get object integrity has no data and insert a new row")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
		checksum        = []byte("mockChecksum")
		version         = int64(1)
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `shadow_integrity_meta` WHERE object_id = ? and redundancy_index = ? ORDER BY `shadow_integrity_meta`.`object_id` LIMIT 1").
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `shadow_integrity_meta` (`object_id`,`redundancy_index`,`integrity_checksum`,`piece_checksum_list`,`version`) VALUES (?,?,?,?,?)").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := s.UpdateShadowPieceChecksum(objectID, redundancyIndex, checksum, version, 1)
	assert.Nil(t, err)
}

func TestSpDBImpl_UpdateShadowPieceChecksumFailure1(t *testing.T) {
	t.Log("Failure case description: get object integrity has no data and insert a new row returns error")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
		checksum        = []byte("mockChecksum")
		version         = int64(1)
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `shadow_integrity_meta` WHERE object_id = ? and redundancy_index = ? ORDER BY `shadow_integrity_meta`.`object_id` LIMIT 1").
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `shadow_integrity_meta` (`object_id`,`redundancy_index`,`integrity_checksum`,`piece_checksum_list`,`version`) VALUES (?,?,?,?,?)").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.UpdateShadowPieceChecksum(objectID, redundancyIndex, checksum, version, 1)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_UpdateShadowPieceChecksumFailure2(t *testing.T) {
	t.Log("Failure case description: get object integrity returns non record not found error")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
		checksum        = []byte("mockChecksum")
		version         = int64(1)
	)
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `shadow_integrity_meta` WHERE object_id = ? and redundancy_index = ? ORDER BY `shadow_integrity_meta`.`object_id` LIMIT 1").
		WillReturnError(mockDBInternalError)
	err := s.UpdateShadowPieceChecksum(objectID, redundancyIndex, checksum, version, 1)
	fmt.Println(err)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}

func TestSpDBImpl_UpdateShadowPieceChecksumFailure3(t *testing.T) {
	t.Log("Failure case description: update object integrity returns error")
	var (
		objectID        = uint64(9)
		redundancyIndex = int32(3)
		checksum        = []byte("mockChecksum")
		version         = int64(1)
	)
	i := &ShadowIntegrityMetaTable{
		ObjectID:          objectID,
		RedundancyIndex:   redundancyIndex,
		IntegrityChecksum: "1406e05881e299367766d313e26c05564ec91bf721d31726bd6e46e60689539a",
		PieceChecksumList: "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d",
		Version:           int64(1),
	}
	s, mock := setupDB(t)
	mock.ExpectQuery("SELECT * FROM `shadow_integrity_meta` WHERE object_id = ? and redundancy_index = ? ORDER BY `shadow_integrity_meta`.`object_id` LIMIT 1").
		WillReturnRows(sqlmock.NewRows([]string{"object_id", "redundancy_index", "integrity_checksum", "piece_checksum_list", "version"}).
			AddRow(i.ObjectID, i.RedundancyIndex, i.IntegrityChecksum, i.PieceChecksumList, i.Version))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `shadow_integrity_meta` SET `piece_checksum_list`=? WHERE object_id = ? and redundancy_index = ?").
		WillReturnError(mockDBInternalError)
	mock.ExpectRollback()
	mock.ExpectCommit()
	err := s.UpdateShadowPieceChecksum(objectID, redundancyIndex, checksum, version, 1)
	fmt.Println(err)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}
