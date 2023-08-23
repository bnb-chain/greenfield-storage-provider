package bsdb

import (
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	mockUser      = "block_syncer"
	mockPassword  = "greenfield"
	mockDBAddress = "127.0.0.1:3306"
	mockDatabase  = "test_db"
)

var mockDBInternalError = errors.New("db internal error")

func setupDB(t *testing.T) (*BsDBImpl, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	assert.Nil(t, err)
	assert.NotNil(t, mockDB)
	assert.NotNil(t, mock)
	dia := mysql.New(mysql.Config{
		DriverName: "mysql",
		DSN: fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", mockUser, mockPassword,
			mockDBAddress, mockDatabase),
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dia, &gorm.Config{})
	assert.Nil(t, err)
	assert.NotNil(t, db)
	return &BsDBImpl{db: db}, mock
}
