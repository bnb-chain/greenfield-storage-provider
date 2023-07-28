package storageprovider

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/database"

	"gorm.io/gorm/schema"

	"github.com/forbole/juno/v4/models"
	"github.com/forbole/juno/v4/modules"
)

const (
	ModuleName = "storage_provider"
)

var (
	_ modules.Module              = &Module{}
	_ modules.PrepareTablesModule = &Module{}
)

// Module represents the storage provider module
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
	return m.db.PrepareTables(context.TODO(), []schema.Tabler{&models.StorageProvider{}})
}

// AutoMigrate implements
func (m *Module) AutoMigrate() error {
	return m.db.AutoMigrate(context.TODO(), []schema.Tabler{&models.StorageProvider{}})
}
