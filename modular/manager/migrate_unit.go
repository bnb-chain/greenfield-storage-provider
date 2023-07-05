package manager

import (
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type MigrateStatus int32

// TODO: refine it.
// sp exit: MigrateInit->WaitForNotifyDestSP->NotifiedDestSP->WaitForMigrate->Migrating->Migrated.
// bucket migrate: MigrateInit->WaitForMigrate->Migrating->Migrated.
var (
	MigrateInit         MigrateStatus = 0
	WaitForNotifyDestSP MigrateStatus = 1
	NotifiedDestSP      MigrateStatus = 2
	WaitForMigrate      MigrateStatus = 3
	Migrating           MigrateStatus = 4
	Migrated            MigrateStatus = 5
)

// GlobalVirtualGroupMigrateExecuteUnit define basic migrate unit, which is used by sp exit and bucket migrate.
type GlobalVirtualGroupMigrateExecuteUnit struct {
	gvg                 *virtualgrouptypes.GlobalVirtualGroup
	redundantIndex      int32 // if < 0, represents migrate primary.
	isConflict          bool  // only be used in sp exit.
	isSecondary         bool  // only be used in sp exit.
	isSrc               bool  // only be used in sp exit.
	srcSP               *sptypes.StorageProvider
	destSP              *sptypes.StorageProvider
	migrateStatus       MigrateStatus
	lastMigrateObjectID uint64 // migrate progress
	checkTimestamp      uint64
	checkStatus         string // only be used in sp exit.
}
