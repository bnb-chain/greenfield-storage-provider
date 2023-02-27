package store

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type Store struct {
	userDB *gorm.DB
}

type IStore interface {
	GetUserBuckets(ctx context.Context, accountID string) (ret []*model.Bucket, err error)
	ListObjectsByBucketName(ctx context.Context, bucketName string) (ret []*model.Object, err error)
}

func NewStore(cfg *config.SQLDBConfig) (*Store, error) {
	userDB, err := newGORM(cfg)
	if err != nil {
		log.Errorf("fail to new gorm cfg:%v err:%v", cfg, err)
		return nil, err
	}

	return &Store{
		userDB: userDB,
	}, nil
}

func newGORM(cfg *config.SQLDBConfig) (*gorm.DB, error) {
	db, err := sqldb.InitDB(cfg)

	if err != nil {
		log.Infof("fail to open database err:%v", err)
		return nil, err
	}

	return db, nil
}
