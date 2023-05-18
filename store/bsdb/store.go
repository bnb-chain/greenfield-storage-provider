package bsdb

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
)

var _ BSDB = &BsDBImpl{}

// BsDBImpl block syncer database, implements BSDB interface
type BsDBImpl struct {
	db *gorm.DB
}

// NewBsDB return a block syncer db instance or a block syncer db backup instance based on the isBackup flag
func NewBsDB(config *metadata.MetadataConfig, isBackup bool) (*BsDBImpl, error) {
	//LoadDBConfigFromEnv(config)
	dbConfig := config.BsDBConfig
	if isBackup {
		dbConfig = config.BsDBSwitchedConfig
	}

	db, err := InitDB(dbConfig)
	if err != nil {
		return nil, err
	}

	return &BsDBImpl{db: db}, nil
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
func LoadDBConfigFromEnv(config *metadata.MetadataConfig) {
	if val, ok := os.LookupEnv(model.BsDBUser); ok {
		config.BsDBConfig.User = val
	}
	if val, ok := os.LookupEnv(model.BsDBPasswd); ok {
		config.BsDBConfig.Passwd = val
	}
	if val, ok := os.LookupEnv(model.BsDBAddress); ok {
		config.BsDBConfig.Address = val
	}
	if val, ok := os.LookupEnv(model.BsDBDataBase); ok {
		config.BsDBConfig.Database = val
	}
	if val, ok := os.LookupEnv(model.BsDBSwitchedUser); ok {
		config.BsDBSwitchedConfig.User = val
	}
	if val, ok := os.LookupEnv(model.BsDBSwitchedPasswd); ok {
		config.BsDBSwitchedConfig.Passwd = val
	}
	if val, ok := os.LookupEnv(model.BsDBSwitchedAddress); ok {
		config.BsDBSwitchedConfig.Address = val
	}
	if val, ok := os.LookupEnv(model.BsDBSwitchedDataBase); ok {
		config.BsDBSwitchedConfig.Database = val
	}
}
