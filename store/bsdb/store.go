package bsdb

import (
	"context"
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
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

func NewStore(cfg *config.SqlDBConfig) (*Store, error) {
	userDB, err := newGORM(cfg)
	if err != nil {
		log.Errorf("fail to new gorm cfg:%v err:%v", cfg, err)
		return nil, err
	}

	return &Store{
		userDB: userDB,
	}, nil
}

func newGORM(cfg *config.SqlDBConfig) (*gorm.DB, error) {
	db, err := InitMetadataServiceDB(cfg)
	if err != nil {
		log.Infof("fail to open database err:%v", err)
		return nil, err
	}

	return db, nil
}

func InitMetadataServiceDB(cfg *config.SqlDBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Passwd,
		cfg.Address,
		cfg.Database)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Errorw("gorm open db failed", "err", err)
		return nil, err
	}

	return db, nil
}
