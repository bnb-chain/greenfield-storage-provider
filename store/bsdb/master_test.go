package bsdb

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

const (
	mockGetSwitchDBSignalSQL       = "SELECT * FROM `master_db` LIMIT 1"
	mockGetMysqlVersionSQL         = "SELECT VERSION();"
	mockGetDefaultCharacterSetSQL  = "SELECT DEFAULT_CHARACTER_SET_NAME FROM INFORMATION_SCHEMA.SCHEMATA where SCHEMA_NAME in ('block_syncer');"
	mockGetDefaultCollationNameSQL = "SELECT DEFAULT_COLLATION_NAME FROM INFORMATION_SCHEMA.SCHEMATA where SCHEMA_NAME in ('block_syncer');"
)

func TestBsDBImpl_GetSwitchDBSignalSuccess(t *testing.T) {
	expectedSignal := &MasterDB{OneRowId: true, IsMaster: true}

	s, mock := setupDB(t)
	mock.ExpectQuery(mockGetSwitchDBSignalSQL).WillReturnRows(
		sqlmock.NewRows([]string{"one_row_id", "is_master"}).
			AddRow(expectedSignal.OneRowId, expectedSignal.IsMaster))

	signal, err := s.GetSwitchDBSignal()
	assert.Nil(t, err)
	assert.Equal(t, expectedSignal, signal)
}

func TestBsDBImpl_GetMysqlVersionSuccess(t *testing.T) {
	expectedVersion := "8.0.23"

	s, mock := setupDB(t)
	mock.ExpectQuery(mockGetMysqlVersionSQL).WillReturnRows(
		sqlmock.NewRows([]string{"VERSION()"}).AddRow(expectedVersion))

	version, err := s.GetMysqlVersion()
	assert.Nil(t, err)
	assert.Equal(t, expectedVersion, version)
}

func TestBsDBImpl_GetDefaultCharacterSetSuccess(t *testing.T) {
	expectedCharacterSet := "utf8mb4"

	s, mock := setupDB(t)
	mock.ExpectQuery(mockGetDefaultCharacterSetSQL).WillReturnRows(
		sqlmock.NewRows([]string{"DEFAULT_CHARACTER_SET_NAME"}).AddRow(expectedCharacterSet))

	charSet, err := s.GetDefaultCharacterSet()
	assert.Nil(t, err)
	assert.Equal(t, expectedCharacterSet, charSet)
}

func TestBsDBImpl_GetDefaultCollationNameSuccess(t *testing.T) {
	expectedCollationName := "utf8mb4_unicode_ci"

	s, mock := setupDB(t)
	mock.ExpectQuery(mockGetDefaultCollationNameSQL).WillReturnRows(
		sqlmock.NewRows([]string{"DEFAULT_COLLATION_NAME"}).AddRow(expectedCollationName))

	collation, err := s.GetDefaultCollationName()
	assert.Nil(t, err)
	assert.Equal(t, expectedCollationName, collation)
}
