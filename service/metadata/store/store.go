package store

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	_ "github.com/go-sql-driver/mysql"
)

type Store struct {
	spDB sqldb.SPDB
}

type IStore interface {
	GetUserBuckets(ctx context.Context, accountID string) (ret []*model.Bucket, err error)
	ListObjectsByBucketName(ctx context.Context, bucketName string) (ret []*model.Object, err error)
}

func NewStore(cfg *config.SQLDBConfig) (*Store, error) {
	spDB, err := sqldb.NewSpDB(cfg)
	if err != nil {
		log.Errorf("fail to new gorm cfg:%v err:%v", cfg, err)
		return nil, err
	}

	return &Store{
		spDB: spDB,
	}, nil
}
