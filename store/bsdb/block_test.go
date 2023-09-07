package bsdb

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

const (
	mockBlockHeightQuerySQL = "SELECT block_height FROM `epoch` LIMIT 1"
)

func TestBsDBImpl_GetLatestBlockNumberSuccess(t *testing.T) {
	expectedBlockHeight := int64(12345)

	s, mock := setupDB(t)
	mock.ExpectQuery(mockBlockHeightQuerySQL).
		WillReturnRows(
			sqlmock.NewRows([]string{"block_height"}).
				AddRow(expectedBlockHeight))

	blockNum, err := s.GetLatestBlockNumber()
	assert.Nil(t, err)
	assert.Equal(t, expectedBlockHeight, blockNum)
}

func TestBsDBImpl_GetLatestBlockNumberNoRecords(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockBlockHeightQuerySQL).WillReturnError(gorm.ErrRecordNotFound)

	blockNum, err := s.GetLatestBlockNumber()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	assert.Equal(t, int64(0), blockNum)
}

func TestBsDBImpl_GetLatestBlockNumberDBError(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockBlockHeightQuerySQL).WillReturnError(mockDBInternalError)

	blockNum, err := s.GetLatestBlockNumber()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.Equal(t, int64(0), blockNum)
}

func TestBsDBImpl_GetLatestBlockNumberMultipleRecords(t *testing.T) {
	s, mock := setupDB(t)
	mock.ExpectQuery(mockBlockHeightQuerySQL).WillReturnRows(
		sqlmock.NewRows([]string{"block_height"}).AddRow(int64(12345)).AddRow(int64(54321)),
	)

	blockNum, err := s.GetLatestBlockNumber()
	assert.Nil(t, err)
	assert.Equal(t, int64(54321), blockNum)
}
