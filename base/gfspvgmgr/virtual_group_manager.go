package gfspvgmgr

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const (
	VirtualGroupManagerSpace  = "VirtualGroupManager"
	RefreshMetaInterval       = 2 * time.Second
	MaxStorageUsage           = 0.95
	DefaultStakingStorageSize = 64 * 1024 * 1024 * 1024 // 64GB per GVG
)

var (
	ErrFailedPickVGF = gfsperrors.Register(VirtualGroupManagerSpace, http.StatusInternalServerError, 540001,
		"failed to pick virtual group family, need create global virtual group")
	ErrFailedPickGVG = gfsperrors.Register(VirtualGroupManagerSpace, http.StatusInternalServerError, 540002,
		"failed to pick global virtual group, need stake more storage size")
	ErrStaledMetadata = gfsperrors.Register(VirtualGroupManagerSpace, http.StatusInternalServerError, 540003,
		"metadata is staled, need force refresh metadata")
)

type GlobalVirtualGroupMeta struct {
	id                 uint32
	familyID           uint32
	primarySPID        uint32
	secondarySPIDs     []uint32
	usedStorageSize    uint64
	stakingStorageSize uint64 // init by staking deposit / staking price
}

type VirtualGroupFamilyMeta struct {
	id                       uint32
	primarySPID              uint32
	familyUsedStorageSize    uint64 // init by gvgMap
	familyStakingStorageSize uint64 // init by gvgMap
	gvgMap                   map[uint32]*GlobalVirtualGroupMeta
}

// virtualGroupFamilyManager is built by metadata data source.
type virtualGroupFamilyManager struct {
	vgfIDToVgf    map[uint32]*VirtualGroupFamilyMeta // is used to pick VGF
	bucketIDToVgf map[uint64]*VirtualGroupFamilyMeta // is used to pick GVG, bucket:VGL = 1:1.
}

// FreeStorageSizeWeightPicker is used to pick index by storage usage,
// The more free storage usage, the greater the probability of being picked.
type FreeStorageSizeWeightPicker struct {
	freeStorageSizeWeightMap map[uint32]float64
}

func (vgfp *FreeStorageSizeWeightPicker) addVirtualGroupFamily(vgf *VirtualGroupFamilyMeta) {
	if float64(vgf.familyUsedStorageSize) >= MaxStorageUsage*float64(vgf.familyStakingStorageSize) || vgf.familyStakingStorageSize == 0 {
		return
	}
	vgfp.freeStorageSizeWeightMap[vgf.id] = float64(vgf.familyStakingStorageSize-vgf.familyUsedStorageSize) / float64(vgf.familyStakingStorageSize)
}

func (vgfp *FreeStorageSizeWeightPicker) addGlobalVirtualGroup(gvg *GlobalVirtualGroupMeta) {
	if float64(gvg.usedStorageSize) >= MaxStorageUsage*float64(gvg.stakingStorageSize) || gvg.stakingStorageSize == 0 {
		return
	}
	vgfp.freeStorageSizeWeightMap[gvg.id] = float64(gvg.stakingStorageSize-gvg.usedStorageSize) / float64(gvg.stakingStorageSize)
}

func (vgfp *FreeStorageSizeWeightPicker) pickIndex() (uint32, error) {
	var (
		sumWeight     float64
		pickWeight    float64
		tempSumWeight float64
	)
	for _, value := range vgfp.freeStorageSizeWeightMap {
		sumWeight += value
	}
	pickWeight = rand.Float64() * sumWeight
	for key, value := range vgfp.freeStorageSizeWeightMap {
		tempSumWeight += value
		if tempSumWeight > pickWeight {
			return key, nil
		}
	}
	return 0, fmt.Errorf("failed to pick weighted random index")
}

func (vgfm *virtualGroupFamilyManager) pickVirtualGroupFamily() (*VirtualGroupFamilyMeta, error) {
	var (
		picker   FreeStorageSizeWeightPicker
		familyID uint32
		err      error
	)
	for _, f := range vgfm.vgfIDToVgf {
		picker.addVirtualGroupFamily(f)
	}
	if familyID, err = picker.pickIndex(); err != nil {
		log.Errorw("failed to pick vgf", "error", err)
		return nil, ErrFailedPickVGF
	}
	return vgfm.vgfIDToVgf[familyID], nil
}

func (vgfm *virtualGroupFamilyManager) pickGlobalVirtualGroup(bucketID uint64) (*GlobalVirtualGroupMeta, error) {
	var (
		picker               FreeStorageSizeWeightPicker
		globalVirtualGroupID uint32
		err                  error
	)
	if _, existed := vgfm.bucketIDToVgf[bucketID]; !existed {
		return nil, ErrStaledMetadata
	}
	for _, g := range vgfm.bucketIDToVgf[bucketID].gvgMap {
		picker.addGlobalVirtualGroup(g)
	}

	if globalVirtualGroupID, err = picker.pickIndex(); err != nil {
		log.Errorw("failed to pick gvg", "bucket_id", bucketID, "error", err)
		return nil, ErrFailedPickGVG
	}
	return vgfm.bucketIDToVgf[bucketID].gvgMap[globalVirtualGroupID], nil
}

type VirtualGroupManager struct {
	selfOperatorAddress string
	selfSPID            uint32
	metadataClient      *gfspclient.GfSpClient // query VG meta from metadata service
	chainClient         consensus.Consensus    // query VG params from chain
	mutex               sync.RWMutex
	storageStakingPrice uint64
	vgfManager          *virtualGroupFamilyManager
	// TODO: add sp list which is used to generate global virtual group to create new gvg.
	// spManager
}

// NewVirtualGroupManager returns a virtual group manager instance.
func NewVirtualGroupManager(selfOperatorAddress string, metadataClient *gfspclient.GfSpClient,
	chainClient consensus.Consensus) (*VirtualGroupManager, error) {
	vgm := &VirtualGroupManager{
		selfOperatorAddress: selfOperatorAddress,
		metadataClient:      metadataClient,
		chainClient:         chainClient,
	}
	vgm.refreshMeta()
	go func() {
		RefreshMetaTicker := time.NewTicker(RefreshMetaInterval)
		for {
			select {
			case <-RefreshMetaTicker.C:
				log.Info("start to refresh virtual group manager meta")
				vgm.refreshMeta()
				log.Info("finish to refresh virtual group manager meta")
			}
		}
	}()
	return vgm, nil

}

// refreshMetadata is used to refresh virtual group manager metadata in background.
func (vgm *VirtualGroupManager) refreshMeta() {
	vgm.mutex.Lock()
	defer vgm.mutex.Unlock()
	// TODO: depend metadata api.
	// periodic update metadata.
	// TODO: refresh staking price from chain.
}

// PickVirtualGroupFamilyForGetCreateBucketApproval pick a virtual group family(If failed to pick,
// new VGF will be automatically created on the chain) in get create bucket approval workflow.
// TODO: if returns VirtualGroupManagerSpace, the caller need create gvg and force refresh metadata and retry.
func (vgm *VirtualGroupManager) PickVirtualGroupFamilyForGetCreateBucketApproval() (*VirtualGroupFamilyMeta, error) {
	vgm.mutex.RLock()
	defer vgm.mutex.RUnlock()
	return vgm.vgfManager.pickVirtualGroupFamily()
}

// PickGlobalVirtualGroupForReplicateObject picks a global virtual group(If failed to pick,
// new GVG will be created by primary SP) in replicate/seal object workflow.
// return (nil, nil), if there is no gvg in vgm.
// TODO: If returns ErrFailedPickGVG, the caller need re-stake or create gvg and force refresh metadata and retry.
// TODO: if returns ErrStaledMetadata, the caller need force refresh from metadata and retry.
func (vgm *VirtualGroupManager) PickGlobalVirtualGroupForReplicateObject(bucketID uint64) (*GlobalVirtualGroupMeta, error) {
	vgm.mutex.RLock()
	defer vgm.mutex.RUnlock()
	return vgm.vgfManager.pickGlobalVirtualGroup(bucketID)
}

// ForceRefreshMeta is used to query metadata service and refresh the virtual group manager meta.
// if pick func returns ErrStaledMetadata, the caller need force refresh from metadata and retry pick.
func (vgm *VirtualGroupManager) ForceRefreshMeta() {
	vgm.refreshMeta()
}

// GenerateGlobalVirtualGroupMeta is used to generate a new global virtual group meta, the caller need send a tx to chain.
// TODO: support more generate policy.
func (vgm *VirtualGroupManager) GenerateGlobalVirtualGroupMeta() (*GlobalVirtualGroupMeta, error) {
	return nil, nil
}

// TODO: add filter picker to support balance.
