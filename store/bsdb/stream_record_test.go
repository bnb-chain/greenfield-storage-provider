package bsdb

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	mockQueryPrimarySPStreamRecord   = "SELECT s.* , g.global_virtual_group_family_id FROM `stream_records` s join `global_virtual_group_families` g on g.virtual_payment_address = s.account where g.primary_sp_id=? and g.removed = false"
	mockQuerySecondarySPStreamRecord = "SELECT s.* , g.global_virtual_group_id FROM `stream_records` s join `global_virtual_groups` g on g.virtual_payment_address = s.account where FIND_IN_SET(?, g.secondary_sp_ids) > 0 and removed = false;"
)

func TestBsDBImpl_GetPrimarySPStreamRecordBySpID(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockQueryPrimarySPStreamRecord).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
		}).
			AddRow(29265))

	streamRecords, err := s.GetPrimarySPStreamRecordBySpID(1)
	assert.Nil(t, err)
	assert.NotNil(t, streamRecords)
}

func TestBsDBImpl_GetSecondarySPStreamRecordBySpID(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockQuerySecondarySPStreamRecord).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
		}).
			AddRow(29265))

	streamRecords, err := s.GetSecondarySPStreamRecordBySpID(1)
	assert.Nil(t, err)
	assert.NotNil(t, streamRecords)
}
