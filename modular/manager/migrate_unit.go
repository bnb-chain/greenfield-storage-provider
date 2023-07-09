package manager

import (
	"fmt"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type MigrateStatus int32

// TODO: refine it, and move to proto.
// migrate: WaitForMigrate(created)->Migrating(schedule success)->Migrated(executor report success).
var (
	WaitForMigrate MigrateStatus = 0
	Migrating      MigrateStatus = 1
	Migrated       MigrateStatus = 2
)

// GlobalVirtualGroupMigrateExecuteUnit define basic migrate unit, which is used by sp exit and bucket migrate.
type GlobalVirtualGroupMigrateExecuteUnit struct {
	gvg                  *virtualgrouptypes.GlobalVirtualGroup
	destGVGID            uint32 // destGVG
	destGVG              *virtualgrouptypes.GlobalVirtualGroup
	redundancyIndex      int32  // if < 0, represents migrate primary.
	swapOutKey           string // only be used in sp exit.
	srcSP                *sptypes.StorageProvider
	destSP               *sptypes.StorageProvider
	migrateStatus        MigrateStatus
	lastMigratedObjectID uint64 // migrate progress
}

// Key is used to as primary key.
func (u *GlobalVirtualGroupMigrateExecuteUnit) Key() string {
	return fmt.Sprintf("SPExit-gvg_id[%d]-vgf_id[%d]-redundancy_idx[%d]",
		u.gvg.GetId(), u.gvg.GetFamilyId(), u.redundancyIndex)
}

func MakeGVGMigrateKey(gvgID uint32, vgfID uint32, redundancyIndex int32) string {
	return fmt.Sprintf("SPExit-gvg_id[%d]-vgf_id[%d]-redundancy_idx[%d]",
		gvgID, vgfID, redundancyIndex)
}

type GlobalVirtualGroupMigrateExecuteUnitByBucket struct {
	bucketID uint64
	GlobalVirtualGroupMigrateExecuteUnit
}

func newGlobalVirtualGroupMigrateExecuteUnitByBucket(bucketID uint64, gvg *virtualgrouptypes.GlobalVirtualGroup, srcSP, destSP *sptypes.StorageProvider,
	migrateStatus MigrateStatus, destGVGID uint32, lastMigrateObjectID uint64, destGVG *virtualgrouptypes.GlobalVirtualGroup) *GlobalVirtualGroupMigrateExecuteUnitByBucket {

	bucketUnit := &GlobalVirtualGroupMigrateExecuteUnitByBucket{}
	bucketUnit.bucketID = bucketID
	bucketUnit.gvg = gvg
	bucketUnit.destGVG = destGVG
	bucketUnit.srcSP = srcSP
	bucketUnit.destSP = destSP
	bucketUnit.migrateStatus = migrateStatus
	bucketUnit.destGVGID = destGVGID

	bucketUnit.redundancyIndex = -1
	bucketUnit.lastMigratedObjectID = lastMigrateObjectID

	return bucketUnit
}

// Key is used to as primary key.
func (ub *GlobalVirtualGroupMigrateExecuteUnitByBucket) Key() string {
	return fmt.Sprintf("MigrateBucket-bucket_id[%d]-gvg_id[%d]", ub.bucketID, ub.gvg.GetId())
}

func MakeBucketMigrateKey(bucketID uint64, gvgID uint32) string {
	return fmt.Sprintf("MigrateBucket-bucket_id[%d]-gvg_id[%d]", bucketID, gvgID)
}
