package bsdb

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/model/metadata"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	_ "github.com/go-sql-driver/mysql"
)

// Store block syncer database, implements IStore interface
type Store struct {
	spDB sqldb.SPDB
}

// IStore contains all the methods required by bs db database
type IStore interface {
	GetUserBuckets(ctx context.Context, accountID string) (ret []*metadata.Bucket, err error)
	ListObjectsByBucketName(ctx context.Context, bucketName string) (ret []*metadata.Object, err error)
}

// NewStore return a database instance
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
