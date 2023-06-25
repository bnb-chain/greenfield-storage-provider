package events

import (
	"context"

	"github.com/forbole/juno/v4/modules"
	"gorm.io/gorm/schema"

	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/database"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

const (
	EventMigrationBucketModuleName = "bucket_migration"
)

var (
	_ modules.Module              = &EventMigrationBucketModule{}
	_ modules.PrepareTablesModule = &EventMigrationBucketModule{}
)

type EventMigrationBucketModule struct {
	db *database.DB
}

// NewEventMigrationBucketModule builds a new Module instance
func NewEventMigrationBucketModule(db *database.DB) *EventMigrationBucketModule {
	return &EventMigrationBucketModule{
		db: db,
	}
}

// Name implements modules.Module
func (m *EventMigrationBucketModule) Name() string {
	return EventMigrationBucketModuleName
}

// PrepareTables implements
func (m *EventMigrationBucketModule) PrepareTables() error {
	return m.db.PrepareTables(context.TODO(), []schema.Tabler{&bsdb.EventMigrationBucket{}})
}

func (m *EventMigrationBucketModule) AutoMigrate() error {
	return m.db.AutoMigrate(context.TODO(), []schema.Tabler{&bsdb.EventMigrationBucket{}})
}
