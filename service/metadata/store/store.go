package store

import (
	"context"
	"fmt"
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metasql"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Store struct {
	userDB *gorm.DB
}

type IStore interface {
	GetUserBuckets(ctx context.Context, accountID string) (ret []*model.Bucket, err error)
	ListObjectsByBucketName(ctx context.Context, bucketName string) (ret []*model.Object, err error)
}

const (
	MetadataServiceDsn = "Metadata_Service_DSN"
)

func NewStore(cfg *config.SqlDBConfig) (*Store, error) {
	userDB, err := newGORM()
	if err != nil {
		log.Errorf("fail to new gorm cfg:%v err:%v", cfg, err)
		return nil, err
	}

	return &Store{
		userDB: userDB,
	}, nil
}

func newGORM() (*gorm.DB, error) {
	db, err := InitMetaServiceDB()

	if err != nil {
		log.Infof("fail to open database err:%v", err)
		return nil, err
	}

	return db, nil
}

func getDBConfigFromEnv(dsn string) (string, error) {
	dsnVal, ok := os.LookupEnv(dsn)
	if !ok {
		return "", fmt.Errorf("dsn %s config is not set in environment", dsnVal)
	}
	return dsnVal, nil
}

func InitMetaServiceDB() (*gorm.DB, error) {
	var dsnForDB string
	dsn, errOfEnv := getDBConfigFromEnv(MetadataServiceDsn)
	if errOfEnv != nil {
		log.Warn("load metadata service db config from ENV failed, try to use config from file")
	} else {
		log.Infof("Using DB config from ENV")
		dsnForDB = dsn
	}
	db, err := gorm.Open(mysql.Open(dsnForDB), &gorm.Config{})
	if err != nil {
		log.Errorw("gorm open db failed", "err", err)
		return nil, err
	}

	// create if not exist
	if err := db.AutoMigrate(&metasql.DBIntegrityMeta{}); err != nil {
		log.Errorw("failed to create integrity meta table", "err", err)
		return nil, err
	}
	return db, nil
}
