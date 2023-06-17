package vmmgr

import (
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
)

type GlobalVirtualGroupMeta struct {
	ID                 uint32
	FamilyID           uint32
	PrimarySPID        uint32
	SecondarySPIDs     []uint32
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
	PickVirtualGroupFamilyForGetCreateBucketApproval() (*VirtualGroupFamilyMeta, error)
	PickGlobalVirtualGroupForReplicateObject(bucketID uint64) (*GlobalVirtualGroupMeta, error)
}

type NewVirtualGroupManager = func(selfOperatorAddress string, chainClient consensus.Consensus) (VirtualGroupManager, error)
