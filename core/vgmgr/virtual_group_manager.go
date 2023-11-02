package vgmgr

import (
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

// GlobalVirtualGroupMeta defines global virtual group meta which is used by sp.
type GlobalVirtualGroupMeta struct {
	ID                   uint32
	FamilyID             uint32
	PrimarySPID          uint32
	SecondarySPIDs       []uint32
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
//
//go:generate mockgen -source=./virtual_group_manager.go -destination=./virtual_group_manager_mock.go -package=vgmgr
type SPPickFilter interface {
	// Check returns true when match pick request condition.
	Check(spID uint32) bool
}

// VGFPickFilter is used to check virtual group families qualification
type VGFPickFilter interface {
	Check(vgfID uint32) bool
}

type PickVGFFilter struct {
	AvailableVgfIDSet map[uint32]struct{}
}

func NewPickVGFFilter(availableVgfIDs []uint32) *PickVGFFilter {
	idSet := make(map[uint32]struct{})
	for _, id := range availableVgfIDs {
		idSet[id] = struct{}{}
	}
	return &PickVGFFilter{AvailableVgfIDSet: idSet}
}

func (p *PickVGFFilter) Check(vgfID uint32) bool {
	_, ok := p.AvailableVgfIDSet[vgfID]
	return ok
}

type PickVGFByGVGFilter struct {
	BlackListSPs map[uint32]struct{}
}

func NewPickVGFByGVGFilter(blackListSPs []uint32) *PickVGFByGVGFilter {
	idSet := make(map[uint32]struct{})
	for _, id := range blackListSPs {
		idSet[id] = struct{}{}
	}
	return &PickVGFByGVGFilter{BlackListSPs: idSet}
}

// Check checks if a family has at least 1 valid GVG(gvg doest not contain blacklist Sp)
func (p *PickVGFByGVGFilter) Check(gvgs map[uint32]*GlobalVirtualGroupMeta) bool {
	validCount := len(gvgs)
	for _, gvg := range gvgs {
		for _, sspID := range gvg.SecondarySPIDs {
			if _, ok := p.BlackListSPs[sspID]; ok {
				validCount--
				break
			}
		}
	}
	return validCount > 0
}

// GVGPickFilter is used to check sp pick condition.
type GVGPickFilter interface {
	// CheckFamily returns true when match pick request condition.
	CheckFamily(familyID uint32) bool
	// CheckGVG returns true when match pick request condition.
	CheckGVG(gvgMeta *GlobalVirtualGroupMeta) bool
}

// GenerateGVGSecondarySPsPolicy is used to generate gvg secondary sp list.
type GenerateGVGSecondarySPsPolicy interface {
	// AddCandidateSP is used to add candidate sp.
	AddCandidateSP(spID uint32)
	// GenerateGVGSecondarySPs returns gvg secondary sp list.
	GenerateGVGSecondarySPs() ([]uint32, error)
}

// ExcludeFilter applies on an ID to check if it should be excluded
type ExcludeFilter interface {
	Apply(id uint32) bool
}

// ExcludeIDFilter The sp freeze pool decides excludeIDFilter ExcludeIDs. When this filter is applied on a sp's id, meeting
// the condition means it is one of freeze sp should be excluded.
type ExcludeIDFilter struct {
	ExcludeIDs map[uint32]struct{}
}

func NewExcludeIDFilter(ids IDSet) ExcludeFilter {
	return &ExcludeIDFilter{
		ExcludeIDs: ids,
	}
}
func (f *ExcludeIDFilter) Apply(id uint32) bool {
	_, ok := f.ExcludeIDs[id]
	return ok
}

// VirtualGroupManager is used to provide virtual group api.
type VirtualGroupManager interface {
	// PickVirtualGroupFamily pick a virtual group family(If failed to pick,
	// new VGF will be automatically created on the chain) in get create bucket approval workflow.
	PickVirtualGroupFamily(filterByGvgList *PickVGFByGVGFilter) (*VirtualGroupFamilyMeta, error)
	// PickGlobalVirtualGroup picks a global virtual group(If failed to pick,
	// new GVG will be created by primary SP) in replicate/seal object workflow.
	PickGlobalVirtualGroup(vgfID uint32, excludeGVGsFilter ExcludeFilter) (*GlobalVirtualGroupMeta, error)
	// PickGlobalVirtualGroupForBucketMigrate picks a global virtual group(If failed to pick,
	// new GVG will be created by primary SP) in replicate/seal object workflow.
	PickGlobalVirtualGroupForBucketMigrate(filter GVGPickFilter) (*GlobalVirtualGroupMeta, error)
	// ForceRefreshMeta is used to query metadata service and refresh the virtual group manager meta.
	// if pick func returns ErrStaledMetadata, the caller need force refresh from metadata and retry pick.
	ForceRefreshMeta() error
	// GenerateGlobalVirtualGroupMeta is used to generate a new global virtual group meta, the caller need send a tx to chain.
	GenerateGlobalVirtualGroupMeta(genPolicy GenerateGVGSecondarySPsPolicy, excludeSPsFilter ExcludeFilter) (*GlobalVirtualGroupMeta, error)
	// PickSPByFilter picks sp which is match pick filter condition.
	PickSPByFilter(filter SPPickFilter) (*sptypes.StorageProvider, error)
	// QuerySPByID returns sp proto.
	QuerySPByID(spID uint32) (*sptypes.StorageProvider, error)
	// FreezeSPAndGVGs puts the secondary SP and its joining Global virtual groups into the freeze pool for a specific period,
	// For those SPs which are in the pool will be skipped when creating a GVG, GVGs in the pool will not be chosen to seal Object
	// until released
	FreezeSPAndGVGs(spID uint32, gvgs []*virtualgrouptypes.GlobalVirtualGroup)
}

// NewVirtualGroupManager is the virtual group manager init api.
type NewVirtualGroupManager = func(selfOperatorAddress string, chainClient consensus.Consensus) (VirtualGroupManager, error)

type IDSet = map[uint32]struct{}
