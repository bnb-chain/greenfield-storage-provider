package bsdb

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

const (
	mockGetBucketInfoByBucketNameQuerySQL = "SELECT * FROM `buckets` WHERE bucket_name = ? LIMIT 1"
)

func TestBsDBImpl_GetBucketInfoByBucketNameSuccess(t *testing.T) {
	expectedBucketName := "test-bucket"

	s, mock := setupDB(t)
	mock.ExpectQuery(mockGetBucketInfoByBucketNameQuerySQL).
		WillReturnRows(
			sqlmock.NewRows([]string{"bucket_name"}).
				AddRow(expectedBucketName))

	bucket, err := s.GetBucketInfoByBucketName(expectedBucketName)
	assert.Nil(t, err)
	assert.Equal(t, expectedBucketName, bucket.BucketName)
}

func TestBsDBImpl_GetBucketInfoByBucketNameNoRecords(t *testing.T) {
	expectedBucketName := "test-bucket"
	s, mock := setupDB(t)
	mock.ExpectQuery(mockGetBucketInfoByBucketNameQuerySQL).WillReturnError(gorm.ErrRecordNotFound)

	_, err := s.GetBucketInfoByBucketName(expectedBucketName)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestBsDBImpl_GetBucketInfoByBucketNameDBError(t *testing.T) {
	expectedBucketName := "test-bucket"
	s, mock := setupDB(t)
	mock.ExpectQuery(mockGetBucketInfoByBucketNameQuerySQL).WillReturnError(mockDBInternalError)

	_, err := s.GetBucketInfoByBucketName(expectedBucketName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), mockDBInternalError.Error())
}
