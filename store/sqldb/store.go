package sqldb

import (
	"fmt"
	"os"
	"time"

	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
)

var _ corespdb.SPDB = &SpDBImpl{}

// SpDBImpl storage provider database, implements SPDB interface
type SpDBImpl struct {
	db *gorm.DB
}

// NewSpDB return a database instance
func NewSpDB(config *config.SQLDBConfig) (*SpDBImpl, error) {
	LoadDBConfigFromEnv(config)
	OverrideConfigVacancy(config)
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
	sqlDB, err := db.DB()
	if err != nil {
		log.Errorw("gorm failed to set db params", "error", err)
		return nil, err
	}
	sqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(config.ConnMaxIdleTime) * time.Second)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
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
	if err := db.AutoMigrate(&PieceHashTable{}); err != nil {
		log.Errorw("failed to create piece hash table", "error", err)
		return nil, err
	}
	if err := db.AutoMigrate(&IntegrityMetaTable{}); err != nil {
		log.Errorw("failed to create integrity meta table", "error", err)
		return nil, err
	}
	if err := db.AutoMigrate(&BucketTrafficTable{}); err != nil {
		log.Errorw("failed to create bucket traffic table", "error", err)
		return nil, err
	}
	if err := db.AutoMigrate(&ReadRecordTable{}); err != nil {
		log.Errorw("failed to create read record table", "error", err)
		return nil, err
	}
	if err := db.AutoMigrate(&ServiceConfigTable{}); err != nil {
		log.Errorw("failed to create service config table", "error", err)
		return nil, err
	}
	if err := db.AutoMigrate(&OffChainAuthKeyTable{}); err != nil {
		log.Errorw("failed to create off-chain authKey table", "error", err)
		return nil, err
	}
	return db, nil
}

// LoadDBConfigFromEnv load db user and password from env vars
func LoadDBConfigFromEnv(config *config.SQLDBConfig) {
	if val, ok := os.LookupEnv(model.SpDBUser); ok {
		config.User = val
	}
	if val, ok := os.LookupEnv(model.SpDBPasswd); ok {
		config.Passwd = val
	}
	if val, ok := os.LookupEnv(model.SpDBAddress); ok {
		config.Address = val
	}
	if val, ok := os.LookupEnv(model.SpDBDataBase); ok {
		config.Database = val
	}
}

// OverrideConfigVacancy override the SQLDB param zero value
func OverrideConfigVacancy(config *config.SQLDBConfig) {
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = model.DefaultConnMaxLifetime
	}
	if config.ConnMaxIdleTime == 0 {
		config.ConnMaxIdleTime = model.DefaultConnMaxIdleTime
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = model.DefaultMaxIdleConns
	}
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = model.DefaultMaxOpenConns
	}
}
