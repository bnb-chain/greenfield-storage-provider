package database

import (
	"context"

	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

func (db *DB) SaveEventMigrationBucket(ctx context.Context, eventMigrationBucket *bsdb.EventMigrationBucket) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&bsdb.EventMigrationBucket{}).TableName()).Create(eventMigrationBucket).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) SaveEventCompleteMigrationBucket(ctx context.Context, eventCompleteMigrationBucket *bsdb.EventCompleteMigrationBucket) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&bsdb.EventCompleteMigrationBucket{}).TableName()).Create(eventCompleteMigrationBucket).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) SaveEventCancelMigrationBucket(ctx context.Context, eventCancelMigrationBucket *bsdb.EventCancelMigrationBucket) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&bsdb.EventCancelMigrationBucket{}).TableName()).Create(eventCancelMigrationBucket).Statement
	return stat.SQL.String(), stat.Vars

}

func (db *DB) SaveEventSwapOut(ctx context.Context, eventSwapOut *bsdb.EventSwapOut) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&bsdb.EventSwapOut{}).TableName()).Create(eventSwapOut).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) SaveEventCancelSwapOut(ctx context.Context, eventCancelSwapOut *bsdb.EventCancelSwapOut) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&bsdb.EventCancelSwapOut{}).TableName()).Create(eventCancelSwapOut).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) SaveEventCompleteSwapOut(ctx context.Context, eventCompleteSwapOut *bsdb.EventCompleteSwapOut) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{DryRun: true}).Table((&bsdb.EventCompleteSwapOut{}).TableName()).Create(eventCompleteSwapOut).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) SaveEventSPExit(ctx context.Context, eventSPExit *bsdb.EventStorageProviderExit) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{
		DryRun: true,
	}).Table((&bsdb.EventStorageProviderExit{}).TableName()).Create(eventSPExit).Statement
	return stat.SQL.String(), stat.Vars
}

func (db *DB) SaveEventSPCompleteExit(ctx context.Context, eventSpCompleteExit *bsdb.EventCompleteStorageProviderExit) (string, []interface{}) {
	stat := db.Db.Session(&gorm.Session{
		DryRun: true,
	}).Table((&bsdb.EventCompleteStorageProviderExit{}).TableName()).Create(eventSpCompleteExit).Statement
	return stat.SQL.String(), stat.Vars
}
