package bsdb

import (
	"fmt"
	syslog "log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
)

var _ BSDB = &BsDBImpl{}

// BsDBImpl block syncer database, implements BSDB interface
type BsDBImpl struct {
	db *gorm.DB
}

// NewBsDB return a block syncer db instance or a block syncer db backup instance based on the isBackup flag
func NewBsDB(cfg *gfspconfig.GfSpConfig) (*BsDBImpl, error) {
	//LoadDBConfigFromEnv(config)
	dbConfig := cfg.BsDB

	db, err := InitDB(&dbConfig)
	if err != nil {
		return nil, err
	}

	return &BsDBImpl{db: db}, nil
}

// InitDB init a block syncer db instance
func InitDB(config *config.SQLDBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Passwd, config.Address, config.Database)
	newLogger := logger.New(
		syslog.New(os.Stdout, "\r\n", syslog.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // Disable color
		},
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		log.Errorw("gorm failed to open db", "error", err)
		return nil, err
	}

	return db, nil
}
