package manager

import (
	"fmt"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type MigrateStatus int32

var (
	// migrate: WaitForMigrate(created)->Migrating(schedule success)->Migrated(executor report success).
	WaitForMigrate MigrateStatus = 0
	Migrating      MigrateStatus = 1
	Migrated       MigrateStatus = 2

	// gc: used for bucket migration
	WaitForGC MigrateStatus = 3
	DoingGC   MigrateStatus = 4
	DoneGC    MigrateStatus = 5
)

type BasicGVGMigrateExecuteUnit struct {
	SrcGVG               *virtualgrouptypes.GlobalVirtualGroup
	SrcSP                *sptypes.StorageProvider
	DestSP               *sptypes.StorageProvider // self sp.
	MigrateStatus        MigrateStatus
	LastMigratedObjectID uint64
}

// SPExitGVGExecuteUnit is used to record sp exit gvg unit.
type SPExitGVGExecuteUnit struct {
	BasicGVGMigrateExecuteUnit
	RedundancyIndex int32 // if < 0, represents migrate primary.
	SwapOutKey      string
}

// Key is used to as primary key.
func (u *SPExitGVGExecuteUnit) Key() string {
	return fmt.Sprintf("SPExit-gvg_id[%d]-vgf_id[%d]-redundancy_idx[%d]",
		u.SrcGVG.GetId(), u.SrcGVG.GetFamilyId(), u.RedundancyIndex)
}

func MakeGVGMigrateKey(gvgID uint32, vgfID uint32, redundancyIndex int32) string {
	return fmt.Sprintf("SPExit-gvg_id[%d]-vgf_id[%d]-redundancy_idx[%d]",
		gvgID, vgfID, redundancyIndex)
}

// BucketMigrateGVGExecuteUnit is used to record bucket migrate gvg unit.
type BucketMigrateGVGExecuteUnit struct {
	BasicGVGMigrateExecuteUnit
	BucketID  uint64
	DestGVGID uint32
	DestGVG   *virtualgrouptypes.GlobalVirtualGroup
}

func newBucketMigrateGVGExecuteUnit(bucketID uint64, gvg *virtualgrouptypes.GlobalVirtualGroup, srcSP, destSP *sptypes.StorageProvider,
	migrateStatus MigrateStatus, destGVGID uint32, lastMigrateObjectID uint64, destGVG *virtualgrouptypes.GlobalVirtualGroup) *BucketMigrateGVGExecuteUnit {

	bucketUnit := &BucketMigrateGVGExecuteUnit{}
	bucketUnit.BucketID = bucketID
	bucketUnit.SrcGVG = gvg
	bucketUnit.DestGVG = destGVG
	bucketUnit.SrcSP = srcSP
	bucketUnit.DestSP = destSP
	bucketUnit.MigrateStatus = migrateStatus
	bucketUnit.DestGVGID = destGVGID

	bucketUnit.LastMigratedObjectID = lastMigrateObjectID

	return bucketUnit
}

// Key is used to as primary key.
func (ub *BucketMigrateGVGExecuteUnit) Key() string {
	return fmt.Sprintf("MigrateBucket-bucket_id[%d]-gvg_id[%d]", ub.BucketID, ub.SrcGVG.GetId())
}

func MakeBucketMigrateKey(bucketID uint64, gvgID uint32) string {
	return fmt.Sprintf("MigrateBucket-bucket_id[%d]-gvg_id[%d]", bucketID, gvgID)
}
