package bsdb

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/forbole/juno/v4/common"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

const (
	mockSelectSPByAddressSQL = "SELECT * FROM `storage_providers` WHERE operator_address = ? and removed = false LIMIT 1"
)

func TestBsDBImpl_GetSPByAddressSuccess(t *testing.T) {
	mockOperatorAddress := common.HexToAddress("0")
	expectedSP := &StorageProvider{
		SpId:            1,
		OperatorAddress: mockOperatorAddress,
	}

	s, mock := setupDB(t)
	mock.ExpectQuery(mockSelectSPByAddressSQL).
		WithArgs(mockOperatorAddress.Bytes()).
		WillReturnRows(sqlmock.NewRows([]string{
			"sp_id", "operator_address",
		}).
			AddRow(expectedSP.SpId, expectedSP.OperatorAddress.Bytes()))

	sp, err := s.GetSPByAddress(mockOperatorAddress)
	assert.Nil(t, err)
	assert.Equal(t, expectedSP, sp)
}

// For TestBsDBImpl_GetSPByAddressNotFound
func TestBsDBImpl_GetSPByAddressNotFound(t *testing.T) {
	mockOperatorAddress := common.HexToAddress("0")

	s, mock := setupDB(t)
	mock.ExpectQuery(mockSelectSPByAddressSQL).
		WithArgs(mockOperatorAddress.Bytes()).
		WillReturnError(gorm.ErrRecordNotFound)

	sp, err := s.GetSPByAddress(mockOperatorAddress)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	assert.NotNil(t, sp)
	assert.Equal(t, uint64(0), sp.ID)
}

// For TestBsDBImpl_GetSPByAddressDBError
func TestBsDBImpl_GetSPByAddressDBError(t *testing.T) {
	mockOperatorAddress := common.HexToAddress("0")

	s, mock := setupDB(t)
	mock.ExpectQuery(mockSelectSPByAddressSQL).
		WithArgs(mockOperatorAddress.Bytes()).
		WillReturnError(mockDBInternalError)

	sp, err := s.GetSPByAddress(mockOperatorAddress)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
	assert.NotNil(t, sp)
	assert.Equal(t, uint64(0), sp.ID)
}
