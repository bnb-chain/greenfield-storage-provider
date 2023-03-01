package sqldb

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
)

const (
	// SPDB environment constants
	SPDBUser   = "SP_DB_USER"
	SPDBPasswd = "SP_DB_PASSWORD"
)

var _ SPDB = &SQLStore{}

// SQLStore storage provider database, implements SPDB interface
type SQLStore struct {
	db *gorm.DB
}

// NewSQLStore return a database instance
func NewSQLStore(config *config.SQLDBConfig) (*SQLStore, error) {
	var err error
	config.User, config.Passwd, err = getDBConfigFromEnv(SPDBUser, SPDBPasswd)
	if err != nil {
		return nil, err
	}
	db, err := InitDB(config)
	if err != nil {
		return nil, err
	}
	return &SQLStore{db: db}, err
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
		log.Errorw("failed to create job table", "err", err)
		return nil, err
	}
	if err := db.AutoMigrate(&ObjectTable{}); err != nil {
		log.Errorw("failed to create object table", "err", err)
		return nil, err
	}
	if err := db.AutoMigrate(&SPInfoTable{}); err != nil {
		log.Errorw("failed to create sp info table", "err", err)
		return nil, err
	}
	if err := db.AutoMigrate(&StorageParamsTable{}); err != nil {
		log.Errorw("failed to storage params table", "err", err)
		return nil, err
	}
	if err := db.AutoMigrate(&IntegrityMetaTable{}); err != nil {
		log.Errorw("failed to create integrity meta table", "err", err)
		return nil, err
	}
	if err := db.AutoMigrate(&BucketTrafficTable{}); err != nil {
		log.Warnw("failed to create bucket traffic table", "err", err)
		return nil, err
	}
	if err := db.AutoMigrate(&ReadRecordTable{}); err != nil {
		log.Warnw("failed to create read record table", "err", err)
		return nil, err
	}
	return db, nil
}

func getDBConfigFromEnv(user, passwd string) (string, string, error) {
	userVal, ok := os.LookupEnv(user)
	if !ok {
		return "", "", fmt.Errorf("db %s config is not set in environment", user)
	}
	passwdVal, ok := os.LookupEnv(passwd)
	if !ok {
		return "", "", fmt.Errorf("db %s config is not set in environment", passwd)
	}
	return userVal, passwdVal, nil
}
