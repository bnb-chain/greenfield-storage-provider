package group

import (
	"context"
	"fmt"

	"github.com/forbole/juno/v4/models"
	"github.com/forbole/juno/v4/modules"
	"gorm.io/gorm/schema"

	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/database"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const (
	ModuleName = "group"
)

var (
	_ modules.Module              = &Module{}
	_ modules.PrepareTablesModule = &Module{}
)

// Module represents the telemetry module
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
	return m.db.PrepareTables(context.TODO(), []schema.Tabler{&models.Group{}})
}

// In this version, we need to create a unique index for the groups table
// Before creating, we need to remove duplicate data from the table
// This code will be removed in the next version
func (m *Module) temporaryFixGroupData() {
	log.Infof("start fix groups data")
	exist := m.db.Db.Migrator().HasTable("groups")
	if !exist {
		log.Infof("fix groups data success")
		return
	}
	tx := m.db.Db.Begin()
	err := tx.Exec("CREATE TABLE `groups_temp` LIKE `groups`").Error
	if err != nil {
		tx.Rollback()
		panic(fmt.Sprintf("fix groups data failed err: %v", err))
	}
	err = tx.Exec("INSERT INTO `groups_temp` select c1.* FROM `groups` c1 INNER JOIN `groups` c2 WHERE c1.id < c2.id AND c1.account_id = c2.account_id and c1.group_id = c2.group_id order by c1.id").Error
	if err != nil {
		tx.Rollback()
		panic(fmt.Sprintf("fix groups data failed err: %v", err))
	}
	err = tx.Exec("delete from `groups` where id in (select id from `groups_temp`)").Error
	if err != nil {
		tx.Rollback()
		panic(fmt.Sprintf("fix groups data failed err: %v", err))
	}
	err = tx.Exec("DROP TABLE `groups_temp`").Error
	if err != nil {
		tx.Rollback()
		panic(fmt.Sprintf("fix groups data failed err: %v", err))
	}
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		panic(fmt.Sprintf("fix groups data failed err: %v", err))
	}
	log.Infof("fix groups data success")
}

// AutoMigrate implements
func (m *Module) AutoMigrate() error {
	m.temporaryFixGroupData()
	return m.db.AutoMigrate(context.TODO(), []schema.Tabler{&models.Group{}})
}
