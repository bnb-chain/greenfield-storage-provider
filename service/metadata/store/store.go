package store

import (
	"context"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Store struct {
	userDB *gorm.DB
}

type IStore interface {
	GetUserBuckets(ctx context.Context) (ret []*model.Bucket, err error)
	ListObjectsByBucketName(ctx context.Context, bucketName string) (ret []*model.Object, err error)
}

func (s *Store) getCTXUserDB(ctx context.Context) *gorm.DB {
	return s.userDB.WithContext(ctx)
}

func NewStore(cfg DBConfig) (*Store, error) {
	userDB, err := newGORM(cfg)
	if err != nil {
		log.Errorf("fail to new gorm cfg:%v err:%v", cfg, err)
		return nil, err
	}

	return &Store{
		userDB: userDB,
	}, nil
}

func newGORM(cfg DBConfig) (*gorm.DB, error) {
	opts := []gorm.Option{}
	opts = append(opts, &optionLogger{cfg.EnableLog})

	db, err := gorm.Open(mysql.Open(cfg.ConnectionString.ToString()), opts...)
	if err != nil {
		log.Infof("fail to open database err:%v", err)
		return nil, err
	}

	return db, nil
}
