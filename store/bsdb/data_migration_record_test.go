package bsdb

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

const (
	mockQueryDataMigrationRecordByProcessKey = "SELECT * FROM `data_migration_record` WHERE process_key = ? LIMIT 1"
)

func TestBsDBImpl_GetDataMigrationRecordByProcessKey(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockQueryDataMigrationRecordByProcessKey).
		WithArgs(ProcessKeyUpdateBucketSize).
		WillReturnRows(sqlmock.NewRows([]string{
			"process_key", "is_completed",
		}).AddRow(ProcessKeyUpdateBucketSize, true))

	records, err := s.GetDataMigrationRecordByProcessKey(ProcessKeyUpdateBucketSize)
	assert.Nil(t, err)
	assert.NotNil(t, records)
}
