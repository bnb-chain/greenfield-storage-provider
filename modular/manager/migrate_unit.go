package manager

import (
	"fmt"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type MigrateStatus int32

// TODO: refine it, and move to proto.
// sp exit: WaitForNotifyDestSP->NotifiedDestSP->WaitForMigrate->Migrating->Migrated.
// bucket migrate: WaitForMigrate(created)->Migrating(schedule success)->Migrated(executor report success).
var (
	WaitForNotifyDestSP MigrateStatus = 0
	NotifiedDestSP      MigrateStatus = 1
	WaitForMigrate      MigrateStatus = 2
	Migrating           MigrateStatus = 3
	Migrated            MigrateStatus = 4
)

// GlobalVirtualGroupMigrateExecuteUnit define basic migrate unit, which is used by sp exit and bucket migrate.
type GlobalVirtualGroupMigrateExecuteUnit struct {
	gvg                  *virtualgrouptypes.GlobalVirtualGroup
	destGVGID            uint32 // destGVG
	redundancyIndex      int32  // if < 0, represents migrate primary.
	isConflicted         bool   // only be used in sp exit.
	isSecondary          bool   // only be used in sp exit.
	isRemoted            bool   // only be used in sp exit.
	srcSP                *sptypes.StorageProvider
	destSP               *sptypes.StorageProvider
	migrateStatus        MigrateStatus
	lastMigratedObjectID uint64 // migrate progress
}

// Key is used to as primary key.
func (u *GlobalVirtualGroupMigrateExecuteUnit) Key() string {
	return fmt.Sprintf("SPExit-gvg_id[%d]-vgf_id[%d]-redundancy_idx[%d]-is_secondary[%t]-is_conflict[%t]-is_remoted[%t]",
		u.gvg.GetId(), u.gvg.GetFamilyId(), u.redundancyIndex,
		u.isSecondary, u.isConflicted, u.isRemoted)
}

func MakeSecondaryGVGMigrateKey(gvgID uint32, vgfID uint32, redundancyIndex int32) string {
	return fmt.Sprintf("SPExit-gvg_id[%d]-vgf_id[%d]-redundancy_idx[%d]-is_secondary[%t]-is_conflict[%t]-is_remoted[%t]",
		gvgID, vgfID, redundancyIndex, true, false, false)
}

type GlobalVirtualGroupMigrateExecuteUnitByBucket struct {
	bucketID uint64
	GlobalVirtualGroupMigrateExecuteUnit
}

// Key is used to as primary key.
func (ub *GlobalVirtualGroupMigrateExecuteUnitByBucket) Key() string {
	return fmt.Sprintf("MigrateBucket-bucket_id[%d]-gvg_id[%d]", ub.bucketID, ub.gvg.GetId())
}

func MakeBucketMigrateKey(bucketID uint64, gvgID uint32) string {
	return fmt.Sprintf("MigrateBucket-bucket_id[%d]-gvg_id[%d]", bucketID, gvgID)
}
