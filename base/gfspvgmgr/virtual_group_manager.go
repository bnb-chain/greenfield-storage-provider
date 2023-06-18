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
	"github.com/bnb-chain/greenfield-storage-provider/core/vmmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

var _ vmmgr.VirtualGroupManager = &virtualGroupManager{}

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

// virtualGroupFamilyManager is built by metadata data source.
type virtualGroupFamilyManager struct {
	vgfIDToVgf    map[uint32]*vmmgr.VirtualGroupFamilyMeta // is used to pick VGF
	bucketIDToVgf map[uint64]*vmmgr.VirtualGroupFamilyMeta // is used to pick GVG, bucket:VGL = 1:1.
}

// FreeStorageSizeWeightPicker is used to pick index by storage usage,
// The more free storage usage, the greater the probability of being picked.
type FreeStorageSizeWeightPicker struct {
	freeStorageSizeWeightMap map[uint32]float64
}

func (vgfp *FreeStorageSizeWeightPicker) addVirtualGroupFamily(vgf *vmmgr.VirtualGroupFamilyMeta) {
	if float64(vgf.FamilyUsedStorageSize) >= MaxStorageUsage*float64(vgf.FamilyStakingStorageSize) || vgf.FamilyStakingStorageSize == 0 {
		return
	}
	vgfp.freeStorageSizeWeightMap[vgf.ID] = float64(vgf.FamilyStakingStorageSize-vgf.FamilyUsedStorageSize) / float64(vgf.FamilyStakingStorageSize)
}

func (vgfp *FreeStorageSizeWeightPicker) addGlobalVirtualGroup(gvg *vmmgr.GlobalVirtualGroupMeta) {
	if float64(gvg.UsedStorageSize) >= MaxStorageUsage*float64(gvg.StakingStorageSize) || gvg.StakingStorageSize == 0 {
		return
	}
	vgfp.freeStorageSizeWeightMap[gvg.ID] = float64(gvg.StakingStorageSize-gvg.UsedStorageSize) / float64(gvg.StakingStorageSize)
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

func (vgfm *virtualGroupFamilyManager) pickVirtualGroupFamily() (*vmmgr.VirtualGroupFamilyMeta, error) {
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

func (vgfm *virtualGroupFamilyManager) pickGlobalVirtualGroup(bucketID uint64) (*vmmgr.GlobalVirtualGroupMeta, error) {
	var (
		picker               FreeStorageSizeWeightPicker
		globalVirtualGroupID uint32
		err                  error
	)
	if _, existed := vgfm.bucketIDToVgf[bucketID]; !existed {
		return nil, ErrStaledMetadata
	}
	for _, g := range vgfm.bucketIDToVgf[bucketID].GVGMap {
		picker.addGlobalVirtualGroup(g)
	}

	if globalVirtualGroupID, err = picker.pickIndex(); err != nil {
		log.Errorw("failed to pick gvg", "bucket_id", bucketID, "error", err)
		return nil, ErrFailedPickGVG
	}
	return vgfm.bucketIDToVgf[bucketID].GVGMap[globalVirtualGroupID], nil
}

type spManager struct {
	primarySP    *sptypes.StorageProvider
	secondarySPs []*sptypes.StorageProvider
}

func (sm *spManager) generateVirtualGroupMeta() (*vmmgr.GlobalVirtualGroupMeta, error) {
	// TODO: refine it.
	if sm.primarySP == nil || len(sm.secondarySPs) < 6 {
		return nil, fmt.Errorf("no enough sp")
	}
	secondarySPIDs := make([]uint32, 6)
	for _, sp := range sm.secondarySPs {
		secondarySPIDs = append(secondarySPIDs, sp.Id)
	}
	return &vmmgr.GlobalVirtualGroupMeta{
		PrimarySPID:        sm.primarySP.Id,
		SecondarySPIDs:     secondarySPIDs,
		StakingStorageSize: DefaultStakingStorageSize,
	}, nil
}

type virtualGroupManager struct {
	selfOperatorAddress string
	selfSPID            uint32
	metadataClient      *gfspclient.GfSpClient // query VG meta from metadata service
	chainClient         consensus.Consensus    // query VG params from chain
	mutex               sync.RWMutex
	storageStakingPrice uint64
	// TODO: add storage parms to pick secondary
	vgfManager *virtualGroupFamilyManager
	spManager  *spManager // is used to generate a new gvg
}

// NewVirtualGroupManager returns a virtual group manager interface.
func NewVirtualGroupManager(selfOperatorAddress string, chainClient consensus.Consensus) (vmmgr.VirtualGroupManager, error) {
	vgm := &virtualGroupManager{
		selfOperatorAddress: selfOperatorAddress,
		// metadataClient:      metadataClient,
		chainClient: chainClient,
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
func (vgm *virtualGroupManager) refreshMeta() {
	vgm.mutex.Lock()
	defer vgm.mutex.Unlock()
	// TODO: depend metadata api.
	// periodic update metadata.
	// TODO: refresh staking price from chain.
}

// PickVirtualGroupFamily pick a virtual group family(If failed to pick,
// new VGF will be automatically created on the chain) in get create bucket approval workflow.
// TODO: if returns VirtualGroupManagerSpace, the caller need create gvg and force refresh metadata and retry.
func (vgm *virtualGroupManager) PickVirtualGroupFamily() (*vmmgr.VirtualGroupFamilyMeta, error) {
	vgm.mutex.RLock()
	defer vgm.mutex.RUnlock()
	return vgm.vgfManager.pickVirtualGroupFamily()
}

// PickGlobalVirtualGroup picks a global virtual group(If failed to pick,
// new GVG will be created by primary SP) in replicate/seal object workflow.
// return (nil, nil), if there is no gvg in vgm.
// TODO: If returns ErrFailedPickGVG, the caller need re-stake or create gvg and force refresh metadata and retry.
// TODO: if returns ErrStaledMetadata, the caller need force refresh from metadata and retry.
func (vgm *virtualGroupManager) PickGlobalVirtualGroup(bucketID uint64) (*vmmgr.GlobalVirtualGroupMeta, error) {
	vgm.mutex.RLock()
	defer vgm.mutex.RUnlock()
	return vgm.vgfManager.pickGlobalVirtualGroup(bucketID)
}

// ForceRefreshMeta is used to query metadata service and refresh the virtual group manager meta.
// if pick func returns ErrStaledMetadata, the caller need force refresh from metadata and retry pick.
func (vgm *virtualGroupManager) ForceRefreshMeta() error {
	vgm.refreshMeta()
	return nil
}

// GenerateGlobalVirtualGroupMeta is used to generate a new global virtual group meta, the caller need send a tx to chain.
// TODO: support more generate policy.
// TODO: add filter picker to support balance.
func (vgm *virtualGroupManager) GenerateGlobalVirtualGroupMeta() (*vmmgr.GlobalVirtualGroupMeta, error) {
	vgm.mutex.RLock()
	defer vgm.mutex.RUnlock()
	return vgm.spManager.generateVirtualGroupMeta()
}
