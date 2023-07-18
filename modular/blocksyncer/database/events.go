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

func (db *DB) SaveEventCancelMigrationBucket(ctx context.Context, eventCancelMigrationBucket *bsdb.EventCancelMigrationBucket) error {
	err := db.Db.WithContext(ctx).Table((&bsdb.EventCancelMigrationBucket{}).TableName()).Create(eventCancelMigrationBucket).Error
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

func (db *DB) SaveEventCancelSwapOut(ctx context.Context, eventCancelSwapOut *bsdb.EventCancelSwapOut) error {
	err := db.Db.WithContext(ctx).Table((&bsdb.EventCancelSwapOut{}).TableName()).Create(eventCancelSwapOut).Error
	if err != nil {
		return err
	}
	return err
}

func (db *DB) SaveEventCompleteSwapOut(ctx context.Context, eventCompleteSwapOut *bsdb.EventCompleteSwapOut) error {
	err := db.Db.WithContext(ctx).Table((&bsdb.EventCompleteSwapOut{}).TableName()).Create(eventCompleteSwapOut).Error
	if err != nil {
		return err
	}
	return err
}

func (db *DB) SaveEventSPExit(ctx context.Context, eventSPExit *bsdb.EventStorageProviderExit) error {
	err := db.Db.WithContext(ctx).Table((&bsdb.EventStorageProviderExit{}).TableName()).Create(eventSPExit).Error
	if err != nil {
		return err
	}
	return err
}

func (db *DB) SaveEventSPCompleteExit(ctx context.Context, eventSpCompleteExit *bsdb.EventCompleteStorageProviderExit) error {
	err := db.Db.WithContext(ctx).Table((&bsdb.EventCompleteStorageProviderExit{}).TableName()).Create(eventSpCompleteExit).Error
	if err != nil {
		return err
	}
	return err
}
