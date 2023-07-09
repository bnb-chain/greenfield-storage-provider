package sqldb

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
)

const (
	// SpDBUser defines env variable name for sp db user name.
	SpDBUser = "SP_DB_USER"
	// SpDBPasswd defines env variable name for sp db user passwd.
	SpDBPasswd = "SP_DB_PASSWORD"
	// SpDBAddress defines env variable name for sp db address.
	SpDBAddress = "SP_DB_ADDRESS"
	// SpDBDataBase defines env variable name for sp db database.
	SpDBDataBase = "SP_DB_DATABASE"

	// DefaultConnMaxLifetime defines the default max liveliness time of connection.
	DefaultConnMaxLifetime = 2048
	// DefaultConnMaxIdleTime defines the default max idle time of connection.
	DefaultConnMaxIdleTime = 2048
	// DefaultMaxIdleConns defines the default max number of idle connections.
	DefaultMaxIdleConns = 2048
	// DefaultMaxOpenConns defines the default max number of open connections.
	DefaultMaxOpenConns = 2048
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
	if err = db.AutoMigrate(&UploadObjectProgressTable{}); err != nil {
		log.Errorw("failed to upload object progress table", "error", err)
		return nil, err
	}
	if err = db.AutoMigrate(&PutObjectSuccessTable{}); err != nil {
		log.Errorw("failed to successfully put event progress table", "error", err)
		return nil, err
	}
	if err = db.AutoMigrate(&PutObjectEventTable{}); err != nil {
		log.Errorw("failed to put event progress table", "error", err)
		return nil, err
	}
	if err = db.AutoMigrate(&UploadTimeoutTable{}); err != nil {
		log.Errorw("failed to successfully put event progress table", "error", err)
		return nil, err
	}
	if err = db.AutoMigrate(&UploadFailedTable{}); err != nil {
		log.Errorw("failed to put event progress table", "error", err)
		return nil, err
	}
	if err = db.AutoMigrate(&ReplicateTimeoutTable{}); err != nil {
		log.Errorw("failed to successfully put event progress table", "error", err)
		return nil, err
	}
	if err = db.AutoMigrate(&ReplicateFailedTable{}); err != nil {
		log.Errorw("failed to put event progress table", "error", err)
		return nil, err
	}
	if err = db.AutoMigrate(&SealTimeoutTable{}); err != nil {
		log.Errorw("failed to successfully put event progress table", "error", err)
		return nil, err
	}
	if err = db.AutoMigrate(&SealFailedTable{}); err != nil {
		log.Errorw("failed to put event progress table", "error", err)
		return nil, err
	}
	if err = db.AutoMigrate(&GCObjectProgressTable{}); err != nil {
		log.Errorw("failed to gc object progress table", "error", err)
		return nil, err
	}
	if err = db.AutoMigrate(&SpInfoTable{}); err != nil {
		log.Errorw("failed to create sp info table", "error", err)
		return nil, err
	}
	if err = db.AutoMigrate(&PieceHashTable{}); err != nil {
		log.Errorw("failed to create piece hash table", "error", err)
		return nil, err
	}
	for i := 0; i < IntegrityMetasNumberOfShards; i++ {
		shardTableName := fmt.Sprintf(IntegrityMetaTable{}.TableName()+"_%02d", i)
		if err = db.Table(shardTableName).AutoMigrate(&IntegrityMetaTable{}); err != nil {
			log.Errorw("failed to create integrity meta table", "error", err)
			return nil, err
		}
	}
	if err = db.AutoMigrate(&BucketTrafficTable{}); err != nil {
		log.Errorw("failed to create bucket traffic table", "error", err)
		return nil, err
	}
	if err = db.AutoMigrate(&ReadRecordTable{}); err != nil {
		log.Errorw("failed to create read record table", "error", err)
		return nil, err
	}
	if err = db.AutoMigrate(&OffChainAuthKeyTable{}); err != nil {
		log.Errorw("failed to create off-chain authKey table", "error", err)
		return nil, err
	}
	if err = db.AutoMigrate(&MigrateSubscribeProgressTable{}); err != nil {
		log.Errorw("failed to migrate subscribe progress table", "error", err)
		return nil, err
	}
	if err = db.AutoMigrate(&SwapOutTable{}); err != nil {
		log.Errorw("failed to swap out table", "error", err)
		return nil, err
	}
	if err = db.AutoMigrate(&MigrateGVGTable{}); err != nil {
		log.Errorw("failed to migrate gvg table", "error", err)
		return nil, err
	}
	return db, nil
}

// LoadDBConfigFromEnv load db user and password from env vars
func LoadDBConfigFromEnv(config *config.SQLDBConfig) {
	if val, ok := os.LookupEnv(SpDBUser); ok {
		config.User = val
	}
	if val, ok := os.LookupEnv(SpDBPasswd); ok {
		config.Passwd = val
	}
	if val, ok := os.LookupEnv(SpDBAddress); ok {
		config.Address = val
	}
	if val, ok := os.LookupEnv(SpDBDataBase); ok {
		config.Database = val
	}
}

// OverrideConfigVacancy override the SQLDB param zero value
func OverrideConfigVacancy(config *config.SQLDBConfig) {
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = DefaultConnMaxLifetime
	}
	if config.ConnMaxIdleTime == 0 {
		config.ConnMaxIdleTime = DefaultConnMaxIdleTime
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = DefaultMaxIdleConns
	}
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = DefaultMaxOpenConns
	}
}
