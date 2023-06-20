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
	PickVirtualGroupFamily() (*VirtualGroupFamilyMeta, error)
	PickGlobalVirtualGroup(vgfID uint32) (*GlobalVirtualGroupMeta, error)
	ForceRefreshMeta() error
	GenerateGlobalVirtualGroupMeta(param *storagetypes.Params) (*GlobalVirtualGroupMeta, error)
}

type NewVirtualGroupManager = func(selfOperatorAddress string, chainClient consensus.Consensus) (VirtualGroupManager, error)
