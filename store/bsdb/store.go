package bsdb

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
)

var _ BSDB = &BsDBImpl{}

// BsDBImpl block syncer database, implements BSDB interface
type BsDBImpl struct {
	db *gorm.DB
}

// NewBsDB return a block syncer db instance
func NewBsDB(config *config.SQLDBConfig) (*BsDBImpl, error) {
	LoadDBConfigFromEnv(config)
	db, err := InitDB(config)
	if err != nil {
		return nil, err
	}
	return &BsDBImpl{db: db}, err
}

// InitDB init a block syncer db instance
func InitDB(config *config.SQLDBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Passwd, config.Address, config.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Errorw("gorm failed to open db", "error", err)
		return nil, err
	}

	return db, nil
}

// LoadDBConfigFromEnv load block syncer db user and password from env vars
func LoadDBConfigFromEnv(config *config.SQLDBConfig) {
	if val, ok := os.LookupEnv(model.BsDBUser); ok {
		config.User = val
	}
	if val, ok := os.LookupEnv(model.BsDBPasswd); ok {
		config.Passwd = val
	}
	if val, ok := os.LookupEnv(model.BsDBAddress); ok {
		config.Address = val
	}
	if val, ok := os.LookupEnv(model.BsDBDataBase); ok {
		config.Database = val
	}
}
