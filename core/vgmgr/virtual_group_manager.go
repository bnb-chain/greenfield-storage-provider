package vgmgr

import (
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// GlobalVirtualGroupMeta defines global virtual group meta which is used by sp.
type GlobalVirtualGroupMeta struct {
	ID             uint32
	FamilyID       uint32
	PrimarySPID    uint32
	SecondarySPIDs []uint32
	// TODO: refine it, current proto is not compatible
	// SecondarySPs       []*sptypes.StorageProvider
	SecondarySPEndpoints []string
	UsedStorageSize      uint64
	StakingStorageSize   uint64 // init by staking deposit / staking price
}

// VirtualGroupFamilyMeta defines virtual group family meta which is used by sp.
type VirtualGroupFamilyMeta struct {
	ID                       uint32
	PrimarySPID              uint32
	FamilyUsedStorageSize    uint64 // init by gvgMap
	FamilyStakingStorageSize uint64 // init by gvgMap
	GVGMap                   map[uint32]*GlobalVirtualGroupMeta
}

// SPPickFilter is used to check sp pick condition.
type SPPickFilter interface {
	// Check returns true when match pick request condition.
	Check(spID uint32) bool
}

// GVGPickFilter is used to check sp pick condition.
type GVGPickFilter interface {
	// CheckFamily returns true when match pick request condition.
	CheckFamily(familyID uint32) bool
	// CheckGVG returns true when match pick request condition.
	CheckGVG(gvgMeta *GlobalVirtualGroupMeta) bool
}

// VirtualGroupManager is used to provide virtual group api.
type VirtualGroupManager interface {
	// PickVirtualGroupFamily pick a virtual group family(If failed to pick,
	// new VGF will be automatically created on the chain) in get create bucket approval workflow.
	PickVirtualGroupFamily() (*VirtualGroupFamilyMeta, error)
	// PickGlobalVirtualGroup picks a global virtual group(If failed to pick,
	// new GVG will be created by primary SP) in replicate/seal object workflow.
	PickGlobalVirtualGroup(vgfID uint32) (*GlobalVirtualGroupMeta, error)
	// PickGlobalVirtualGroupForBucketMigrate picks a global virtual group(If failed to pick,
	// new GVG will be created by primary SP) in replicate/seal object workflow.
	PickGlobalVirtualGroupForBucketMigrate(filter GVGPickFilter) (*GlobalVirtualGroupMeta, error)
	// ForceRefreshMeta is used to query metadata service and refresh the virtual group manager meta.
	// if pick func returns ErrStaledMetadata, the caller need force refresh from metadata and retry pick.
	ForceRefreshMeta() error
	// GenerateGlobalVirtualGroupMeta is used to generate a new global virtual group meta, the caller need send a tx to chain.
	GenerateGlobalVirtualGroupMeta(param *storagetypes.Params) (*GlobalVirtualGroupMeta, error)
	// PickSPByFilter picks sp which is match pick filter condition.
	PickSPByFilter(filter SPPickFilter) (*sptypes.StorageProvider, error)
	// QuerySPByID returns sp proto.
	QuerySPByID(spID uint32) (*sptypes.StorageProvider, error)
}

// NewVirtualGroupManager is the virtual group manager init api.
type NewVirtualGroupManager = func(selfOperatorAddress string, chainClient consensus.Consensus) (VirtualGroupManager, error)
