package database

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

func (db *DB) SaveEventMigrationBucket(ctx context.Context, eventMigrationBucket *bsdb.EventMigrationBucket) error {
	err := db.Db.WithContext(ctx).Table((&bsdb.EventMigrationBucket{}).TableName()).Create(eventMigrationBucket).Error
	if err != nil {
		return err
	}
	return err
}

func (db *DB) SaveEventCompleteMigrationBucket(ctx context.Context, eventCompleteMigrationBucket *bsdb.EventCompleteMigrationBucket) error {
	err := db.Db.WithContext(ctx).Table((&bsdb.EventCompleteMigrationBucket{}).TableName()).Create(eventCompleteMigrationBucket).Error
	if err != nil {
		return err
	}
	return err
}

func (db *DB) SaveEventSwapOut(ctx context.Context, eventSwapOut *bsdb.EventSwapOut) error {
	err := db.Db.WithContext(ctx).Table((&bsdb.EventSwapOut{}).TableName()).Create(eventSwapOut).Error
	if err != nil {
		return err
	}
	return err
}
