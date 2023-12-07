package general

import (
	"context"

	"github.com/forbole/juno/v4/modules"
	"gorm.io/gorm/schema"

	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/database"
)

const (
	ModuleName = "general"
)

var (
	_ modules.Module              = &Module{}
	_ modules.PrepareTablesModule = &Module{}
)

// Module represents the bucket module
type Module struct {
	db *database.DB
}

// NewModule builds a new Module instance
func NewModule(db *database.DB) *Module {
	return &Module{
		db: db,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return ModuleName
}

// PrepareTables implements
func (m *Module) PrepareTables() error {
	return m.db.PrepareTables(context.TODO(), []schema.Tabler{})
}

// AutoMigrate implements
func (m *Module) AutoMigrate() error {
	return m.db.AutoMigrate(context.TODO(), []schema.Tabler{})
}
