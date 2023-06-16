package prefixtree

//
//import (
//	"context"
//
//	"github.com/forbole/juno/v4/modules"
//	"gorm.io/gorm/schema"
//
//	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/database"
//	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
//)
//
//const (
//	ModuleName = "prefix_tree"
//)
//
//var (
//	_ modules.Module              = &Module{}
//	_ modules.PrepareTablesModule = &Module{}
//)
//
//// Module represents the object module
//type Module struct {
//	db *database.DB
//}
//
//// NewModule builds a new Module instance
//func NewModule(db *database.DB) *Module {
//	return &Module{
//		db: db,
//	}
//}
//
//// Name implements modules.Module
//func (m *Module) Name() string {
//	return ModuleName
//}
//
//// PrepareTables implements
//func (m *Module) PrepareTables() error {
//	return m.db.PrepareTables(context.TODO(), []schema.Tabler{&bsdb.SlashPrefixTreeNode{}})
//}
//
//func (m *Module) AutoMigrate() error {
//	return m.db.AutoMigrate(context.TODO(), []schema.Tabler{&bsdb.SlashPrefixTreeNode{}})
//}
