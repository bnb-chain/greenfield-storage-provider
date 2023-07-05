package manager

import (
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type MigrateStatus int32

// TODO: should enrich migrate status.
var (
	WaitForMigrate MigrateStatus = 0
	Migrating      MigrateStatus = 1
	Migrated       MigrateStatus = 2
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
