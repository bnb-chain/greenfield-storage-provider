package events

import (
	"context"

	"github.com/forbole/juno/v4/modules"
	"gorm.io/gorm/schema"

	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/database"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

const (
	SPExitRelatedEventsModuleName = "sp_exit_events"
)

var (
	_ modules.Module              = &Module{}
	_ modules.PrepareTablesModule = &Module{}
)

type Module struct {
	db *database.DB
}

func NewModule(db *database.DB) *Module {
	return &Module{
		db: db,
	}
}

// SetCtx associates a given key with a value in the module's context.
// It takes a key of type string and a value of any type, and stores
// the pair in the context. This is useful for passing data across different
// parts of a module.
func (m *Module) SetCtx(key string, val interface{}) {
}

// GetCtx retrieves the value associated with a given key from the module's context.
// If the key exists in the context, it returns the value; otherwise, it returns nil.
// This is commonly used to access data that was previously stored with Set.
func (m *Module) GetCtx(key string) interface{} {
	return nil
}

// ClearCtx resets the module's context to a new, empty context.
// This effectively removes all key-value pairs previously stored in the context.
// This can be used for cleanup or reinitialization purposes.
func (m *Module) ClearCtx() {
}

// Name implements modules.Module
func (m *Module) Name() string {
	return SPExitRelatedEventsModuleName
}

func (m *Module) PrepareTables() error {
	return m.db.PrepareTables(context.TODO(), []schema.Tabler{
		&bsdb.EventSwapOut{},
		&bsdb.EventCompleteSwapOut{},
		&bsdb.EventCancelSwapOut{},
		&bsdb.EventMigrationBucket{},
		&bsdb.EventCompleteMigrationBucket{},
		&bsdb.EventCancelMigrationBucket{},
		&bsdb.EventStorageProviderExit{},
		&bsdb.EventCompleteStorageProviderExit{},
		&bsdb.EventRejectMigrateBucket{},
	})
}

func (m *Module) AutoMigrate() error {
	return m.db.AutoMigrate(context.TODO(), []schema.Tabler{
		&bsdb.EventSwapOut{},
		&bsdb.EventMigrationBucket{},
		&bsdb.EventCompleteMigrationBucket{},
		&bsdb.EventStorageProviderExit{},
		&bsdb.EventCompleteStorageProviderExit{},
		&bsdb.EventRejectMigrateBucket{},
	})
}
