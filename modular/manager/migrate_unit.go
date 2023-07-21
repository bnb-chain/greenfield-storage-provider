package manager

import (
	"fmt"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type MigrateStatus int32

// migrate: WaitForMigrate(created)->Migrating(schedule success)->Migrated(executor report success).
var (
	WaitForMigrate MigrateStatus = 0
	Migrating      MigrateStatus = 1
	Migrated       MigrateStatus = 2
)

type basicGVGMigrateExecuteUnit struct {
	srcGVG               *virtualgrouptypes.GlobalVirtualGroup
	srcSP                *sptypes.StorageProvider
	destSP               *sptypes.StorageProvider // self sp.
	migrateStatus        MigrateStatus
	lastMigratedObjectID uint64
}

// SPExitGVGExecuteUnit is used to record sp exit gvg unit.
type SPExitGVGExecuteUnit struct {
	basicGVGMigrateExecuteUnit
	redundancyIndex int32 // if < 0, represents migrate primary.
	swapOutKey      string
}

// Key is used to as primary key.
func (u *SPExitGVGExecuteUnit) Key() string {
	return fmt.Sprintf("SPExit-gvg_id[%d]-vgf_id[%d]-redundancy_idx[%d]",
		u.srcGVG.GetId(), u.srcGVG.GetFamilyId(), u.redundancyIndex)
}

func MakeGVGMigrateKey(gvgID uint32, vgfID uint32, redundancyIndex int32) string {
	return fmt.Sprintf("SPExit-gvg_id[%d]-vgf_id[%d]-redundancy_idx[%d]",
		gvgID, vgfID, redundancyIndex)
}

// BucketMigrateGVGExecuteUnit is used to record bucket migrate gvg unit.
type BucketMigrateGVGExecuteUnit struct {
	basicGVGMigrateExecuteUnit
	bucketID  uint64
	destGVGID uint32
	destGVG   *virtualgrouptypes.GlobalVirtualGroup
}

func newBucketMigrateGVGExecuteUnit(bucketID uint64, gvg *virtualgrouptypes.GlobalVirtualGroup, srcSP, destSP *sptypes.StorageProvider,
	migrateStatus MigrateStatus, destGVGID uint32, lastMigrateObjectID uint64, destGVG *virtualgrouptypes.GlobalVirtualGroup) *BucketMigrateGVGExecuteUnit {

	bucketUnit := &BucketMigrateGVGExecuteUnit{}
	bucketUnit.bucketID = bucketID
	bucketUnit.srcGVG = gvg
	bucketUnit.destGVG = destGVG
	bucketUnit.srcSP = srcSP
	bucketUnit.destSP = destSP
	bucketUnit.migrateStatus = migrateStatus
	bucketUnit.destGVGID = destGVGID

	bucketUnit.lastMigratedObjectID = lastMigrateObjectID

	return bucketUnit
}

// Key is used to as primary key.
func (ub *BucketMigrateGVGExecuteUnit) Key() string {
	return fmt.Sprintf("MigrateBucket-bucket_id[%d]-gvg_id[%d]", ub.bucketID, ub.srcGVG.GetId())
}

func MakeBucketMigrateKey(bucketID uint64, gvgID uint32) string {
	return fmt.Sprintf("MigrateBucket-bucket_id[%d]-gvg_id[%d]", bucketID, gvgID)
}
