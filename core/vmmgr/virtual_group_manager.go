package vmmgr

import (
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

type GlobalVirtualGroupMeta struct {
	ID                 uint32
	FamilyID           uint32
	PrimarySPID        uint32
	SecondarySPIDs     []uint32
	SecondarySPs       []*sptypes.StorageProvider
	UsedStorageSize    uint64
	StakingStorageSize uint64 // init by staking deposit / staking price
}

type VirtualGroupFamilyMeta struct {
	ID                       uint32
	PrimarySPID              uint32
	FamilyUsedStorageSize    uint64 // init by gvgMap
	FamilyStakingStorageSize uint64 // init by gvgMap
	GVGMap                   map[uint32]*GlobalVirtualGroupMeta
}

type VirtualGroupManager interface {
	// PickVirtualGroupFamily pick a virtual group family(If failed to pick,
	// new VGF will be automatically created on the chain) in get create bucket approval workflow.
	PickVirtualGroupFamily() (*VirtualGroupFamilyMeta, error)
	// PickGlobalVirtualGroup picks a global virtual group(If failed to pick,
	// new GVG will be created by primary SP) in replicate/seal object workflow.
	PickGlobalVirtualGroup(vgfID uint32) (*GlobalVirtualGroupMeta, error)
	// ForceRefreshMeta is used to query metadata service and refresh the virtual group manager meta.
	// if pick func returns ErrStaledMetadata, the caller need force refresh from metadata and retry pick.
	ForceRefreshMeta() error
	// GenerateGlobalVirtualGroupMeta is used to generate a new global virtual group meta, the caller need send a tx to chain.
	GenerateGlobalVirtualGroupMeta(param *storagetypes.Params) (*GlobalVirtualGroupMeta, error)
}

type NewVirtualGroupManager = func(selfOperatorAddress string, chainClient consensus.Consensus) (VirtualGroupManager, error)
