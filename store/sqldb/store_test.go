package sqldb

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/store/config"
)

const (
	mockUser      = "sp"
	mockPassword  = "greenfield"
	mockDBAddress = "127.0.0.1:3306"
	mockDatabase  = "test_db"
)

var mockDBInternalError = errors.New("db internal error")

type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func setupDB(t *testing.T) (*SpDBImpl, sqlmock.Sqlmock) {
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
	return &SpDBImpl{db: db}, mock
}

func TestNewSpDBFailure(t *testing.T) {
	s, err := NewSpDB(&config.SQLDBConfig{
		User:     mockUser,
		Passwd:   mockPassword,
		Address:  mockDBAddress,
		Database: mockDatabase,
	})
	assert.NotNil(t, err)
	assert.Nil(t, s)
}

func TestInitDBFailure(t *testing.T) {
	db, err := InitDB(&config.SQLDBConfig{
		User:     mockUser,
		Passwd:   mockPassword,
		Address:  mockDBAddress,
		Database: mockDatabase,
	})
	assert.NotNil(t, err)
	assert.Nil(t, db)
}

func TestLoadDBConfigFromEnv(t *testing.T) {
	_ = os.Setenv(SpDBUser, mockUser)
	_ = os.Setenv(SpDBPasswd, mockPassword)
	_ = os.Setenv(SpDBAddress, mockDBAddress)
	_ = os.Setenv(SpDBDatabase, mockDatabase)
	defer os.Unsetenv(SpDBUser)
	defer os.Unsetenv(SpDBPasswd)
	defer os.Unsetenv(SpDBAddress)
	defer os.Unsetenv(SpDBDatabase)
	cfg := &config.SQLDBConfig{}
	LoadDBConfigFromEnv(cfg)
	assert.Equal(t, mockUser, cfg.User)
	assert.Equal(t, mockPassword, cfg.Passwd)
	assert.Equal(t, mockDBAddress, cfg.Address)
	assert.Equal(t, mockDatabase, cfg.Database)
}

func TestOverrideConfigVacancy(t *testing.T) {
	cfg := &config.SQLDBConfig{}
	OverrideConfigVacancy(cfg)
	assert.Equal(t, DefaultConnMaxLifetime, cfg.ConnMaxLifetime)
	assert.Equal(t, DefaultConnMaxIdleTime, cfg.ConnMaxIdleTime)
	assert.Equal(t, DefaultMaxIdleConns, cfg.MaxIdleConns)
	assert.Equal(t, DefaultMaxOpenConns, cfg.MaxOpenConns)
}
