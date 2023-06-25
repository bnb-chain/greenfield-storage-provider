package events

import (
	"context"

	"github.com/forbole/juno/v4/modules"
	"gorm.io/gorm/schema"

	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/database"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

const (
	EventMigrationBucketCompleteModuleName = "bucket_migration_complete"
)

var (
	_ modules.Module              = &EventMigrationBucketCompleteModule{}
	_ modules.PrepareTablesModule = &EventMigrationBucketCompleteModule{}
)

// EventMigrationBucketCompleteModule represents the object module
type EventMigrationBucketCompleteModule struct {
	db *database.DB
}

// NewEventMigrationBucketCompleteModule builds a new Module instance
func NewEventMigrationBucketCompleteModule(db *database.DB) *EventMigrationBucketCompleteModule {
	return &EventMigrationBucketCompleteModule{
		db: db,
	}
}

// Name implements modules.Module
func (m *EventMigrationBucketCompleteModule) Name() string {
	return EventMigrationBucketCompleteModuleName
}

// PrepareTables implements
func (m *EventMigrationBucketCompleteModule) PrepareTables() error {
	return m.db.PrepareTables(context.TODO(), []schema.Tabler{&bsdb.EventCompleteMigrationBucket{}})
}

func (m *EventMigrationBucketCompleteModule) AutoMigrate() error {
	return m.db.AutoMigrate(context.TODO(), []schema.Tabler{&bsdb.EventCompleteMigrationBucket{}})
}
