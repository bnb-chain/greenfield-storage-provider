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
	// GetUserBuckets get buckets info by a user address
	GetUserBuckets(ctx context.Context, accountID string) ([]*metadata.Bucket, error)
	// ListObjectsByBucketName list objects info by a bucket name
	ListObjectsByBucketName(ctx context.Context, bucketName string) ([]*metadata.Object, error)
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
