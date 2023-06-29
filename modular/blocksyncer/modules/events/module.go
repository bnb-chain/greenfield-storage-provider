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

// Name implements modules.Module
func (m *Module) Name() string {
	return SPExitRelatedEventsModuleName
}

func (m *Module) PrepareTables() error {
	return m.db.PrepareTables(context.TODO(), []schema.Tabler{
		&bsdb.EventSwapOut{},
		&bsdb.EventMigrationBucket{},
		&bsdb.EventCompleteMigrationBucket{},
		&bsdb.EventStorageProviderExit{},
		&bsdb.EventCompleteStorageProviderExit{},
	})
}

func (m *Module) AutoMigrate() error {
	return m.db.AutoMigrate(context.TODO(), []schema.Tabler{
		&bsdb.EventSwapOut{},
		&bsdb.EventMigrationBucket{},
		&bsdb.EventCompleteMigrationBucket{},
		&bsdb.EventStorageProviderExit{},
		&bsdb.EventCompleteStorageProviderExit{}})
}
