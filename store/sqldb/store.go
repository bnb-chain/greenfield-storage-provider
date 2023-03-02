package sqldb

import (
	"fmt"
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
)

var _ SPDB = &SpDBImpl{}

// SpDBImpl storage provider database, implements SPDB interface
type SpDBImpl struct {
	db *gorm.DB
}

// NewSpDB return a database instance
func NewSpDB(config *config.SQLDBConfig) (*SpDBImpl, error) {
	loadDBConfigFromEnv(config)
	db, err := InitDB(config)
	if err != nil {
		return nil, err
	}
	return &SpDBImpl{db: db}, err
}

// InitDB init a db instance
func InitDB(config *config.SQLDBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Passwd, config.Address, config.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Errorw("gorm failed to open db", "error", err)
		return nil, err
	}

	// create if not exist
	if err := db.AutoMigrate(&JobTable{}); err != nil {
		log.Errorw("failed to create job table", "error", err)
		return nil, err
	}
	if err := db.AutoMigrate(&ObjectTable{}); err != nil {
		log.Errorw("failed to create object table", "error", err)
		return nil, err
	}
	if err := db.AutoMigrate(&SpInfoTable{}); err != nil {
		log.Errorw("failed to create sp info table", "error", err)
		return nil, err
	}
	if err := db.AutoMigrate(&StorageParamsTable{}); err != nil {
		log.Errorw("failed to storage params table", "error", err)
		return nil, err
	}
	if err := db.AutoMigrate(&IntegrityMetaTable{}); err != nil {
		log.Errorw("failed to create integrity meta table", "error", err)
		return nil, err
	}
	if err := db.AutoMigrate(&BucketTrafficTable{}); err != nil {
		log.Warnw("failed to create bucket traffic table", "error", err)
		return nil, err
	}
	if err := db.AutoMigrate(&ReadRecordTable{}); err != nil {
		log.Warnw("failed to create read record table", "error", err)
		return nil, err
	}
	return db, nil
}

// loadDBConfigFromEnv load db user and password from env vars
func loadDBConfigFromEnv(config *config.SQLDBConfig) {
	if val, ok := os.LookupEnv(model.SpDBUser); ok {
		config.User = val
	}
	if val, ok := os.LookupEnv(model.SpDBPasswd); ok {
		config.Passwd = val
	}
}
