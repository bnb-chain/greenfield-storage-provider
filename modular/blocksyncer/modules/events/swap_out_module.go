package events

import (
	"context"

	"github.com/forbole/juno/v4/modules"
	"gorm.io/gorm/schema"

	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/database"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

const (
	EventSwapOutModuleName = "swap_out"
)

var (
	_ modules.Module              = &EventSwapOutModule{}
	_ modules.PrepareTablesModule = &EventSwapOutModule{}
)

type EventSwapOutModule struct {
	db *database.DB
}

func NewEventSwapOutModule(db *database.DB) *EventSwapOutModule {
	return &EventSwapOutModule{
		db: db,
	}
}

// Name implements modules.Module
func (m *EventSwapOutModule) Name() string {
	return EventSwapOutModuleName
}

func (m *EventSwapOutModule) PrepareTables() error {
	return m.db.PrepareTables(context.TODO(), []schema.Tabler{&bsdb.EventSwapOut{}})
}

func (m *EventSwapOutModule) AutoMigrate() error {
	return m.db.AutoMigrate(context.TODO(), []schema.Tabler{&bsdb.EventSwapOut{}})
}
