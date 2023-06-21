package gfspvgmgr

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/vmmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
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
	vgfIDToVgf map[uint32]*vmmgr.VirtualGroupFamilyMeta // is used to pick VGF
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

func (vgfm *virtualGroupFamilyManager) pickGlobalVirtualGroup(vgfID uint32) (*vmmgr.GlobalVirtualGroupMeta, error) {
	var (
		picker               FreeStorageSizeWeightPicker
		globalVirtualGroupID uint32
		err                  error
	)
	if _, existed := vgfm.vgfIDToVgf[vgfID]; !existed {
		return nil, ErrStaledMetadata
	}
	for _, g := range vgfm.vgfIDToVgf[vgfID].GVGMap {
		picker.addGlobalVirtualGroup(g)
	}

	if globalVirtualGroupID, err = picker.pickIndex(); err != nil {
		log.Errorw("failed to pick gvg", "vgf_id", vgfID, "error", err)
		return nil, ErrFailedPickGVG
	}
	return vgfm.vgfIDToVgf[vgfID].GVGMap[globalVirtualGroupID], nil
}

type spManager struct {
	primarySP    *sptypes.StorageProvider
	secondarySPs []*sptypes.StorageProvider
}

func (sm *spManager) generateVirtualGroupMeta(param *storagetypes.Params) (*vmmgr.GlobalVirtualGroupMeta, error) {
	secondarySPNumber := int(param.GetRedundantDataChunkNum() + param.GetRedundantParityChunkNum())
	if sm.primarySP == nil || len(sm.secondarySPs) < secondarySPNumber {
		return nil, fmt.Errorf("no enough sp")
	}
	secondarySPIDs := make([]uint32, 0)
	secondarySPs := make([]*sptypes.StorageProvider, 0)
	for i, sp := range sm.secondarySPs {
		if i < secondarySPNumber {
			secondarySPIDs = append(secondarySPIDs, sp.Id)
			secondarySPs = append(secondarySPs, sp)
		}
	}
	return &vmmgr.GlobalVirtualGroupMeta{
		PrimarySPID:        sm.primarySP.Id,
		SecondarySPIDs:     secondarySPIDs,
		SecondarySPs:       secondarySPs,
		StakingStorageSize: DefaultStakingStorageSize,
	}, nil
}

type virtualGroupManager struct {
	selfOperatorAddress string
	// metadataClient      *gfspclient.GfSpClient // query VG meta from metadata service
	chainClient consensus.Consensus // query VG params from chain
	mutex       sync.RWMutex
	selfSPID    uint32
	spManager   *spManager // is used to generate a new gvg
	vgParams    *virtualgrouptypes.Params
	vgfManager  *virtualGroupFamilyManager
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
		for range RefreshMetaTicker.C {
			log.Info("start to refresh virtual group manager meta")
			vgm.refreshMeta()
			log.Info("finish to refresh virtual group manager meta")
		}
	}()
	return vgm, nil

}

// refreshMetadata is used to refresh virtual group manager metadata in background.
func (vgm *virtualGroupManager) refreshMeta() {
	// TODO: support load from metadata.
	// vgm.refreshMetaByMetaService()
	vgm.refreshMetaByChain()
}

func (vgm *virtualGroupManager) refreshMetaByChain() {
	var (
		err             error
		spList          []*sptypes.StorageProvider
		primarySP       *sptypes.StorageProvider
		secondarySPList []*sptypes.StorageProvider
		spID            uint32
		spm             *spManager
		vgParams        *virtualgrouptypes.Params
		vgfList         []*virtualgrouptypes.GlobalVirtualGroupFamily
		vgfm            *virtualGroupFamilyManager
	)

	// query meta
	if spList, err = vgm.chainClient.ListSPs(context.Background()); err != nil {
		log.Errorw("failed to list sps", "error", err)
		return
	}

	for i, sp := range spList {
		if strings.EqualFold(vgm.selfOperatorAddress, sp.OperatorAddress) {
			spID = sp.Id
			primarySP = sp
			secondarySPList = append(spList[:i], spList[i+1:]...)
		}
	}
	spm = &spManager{
		primarySP:    primarySP,
		secondarySPs: secondarySPList,
	}
	log.Infow("list sp info", "primary_sp", primarySP, "secondary_sps", secondarySPList)

	if spID == 0 {
		log.Error("failed to refresh due to current sp is not in sp list")
		return
	}

	if vgParams, err = vgm.chainClient.QueryVirtualGroupParams(context.Background()); err != nil {
		log.Errorw("failed to query virtual group params", "error", err)
		return
	}

	vgfm = &virtualGroupFamilyManager{
		vgfIDToVgf: make(map[uint32]*vmmgr.VirtualGroupFamilyMeta),
	}
	if vgfList, err = vgm.chainClient.ListVirtualGroupFamilies(context.Background(), spID); err != nil {
		log.Errorw("failed to list virtual group family", "error", err)
		return
	}
	for _, vgf := range vgfList {
		vgfm.vgfIDToVgf[vgf.Id] = &vmmgr.VirtualGroupFamilyMeta{
			ID:          vgf.Id,
			PrimarySPID: spID,
			GVGMap:      make(map[uint32]*vmmgr.GlobalVirtualGroupMeta),
		}
		for _, gvgID := range vgf.GlobalVirtualGroupIds {
			var gvg *virtualgrouptypes.GlobalVirtualGroup
			if gvg, err = vgm.chainClient.QueryGlobalVirtualGroup(context.Background(), gvgID); err != nil {
				log.Errorw("failed to query global virtual group", "error", err)
				return
			}
			gvgMeta := &vmmgr.GlobalVirtualGroupMeta{
				ID:                 gvg.GetId(),
				FamilyID:           vgf.Id,
				PrimarySPID:        spID,
				SecondarySPIDs:     gvg.GetSecondarySpIds(),
				UsedStorageSize:    gvg.GetStoredSize(),
				StakingStorageSize: gvg.TotalDeposit.Uint64() / vgParams.GvgStakingPrice.BigInt().Uint64(), // TODO: refine it.
			}
			vgfm.vgfIDToVgf[vgf.Id].GVGMap[gvg.GetId()] = gvgMeta
			vgfm.vgfIDToVgf[vgf.Id].FamilyUsedStorageSize += gvgMeta.UsedStorageSize
			vgfm.vgfIDToVgf[vgf.Id].FamilyStakingStorageSize += gvgMeta.StakingStorageSize
		}
	}

	// update meta
	vgm.mutex.Lock()
	vgm.selfSPID = spID
	vgm.spManager = spm
	vgm.vgParams = vgParams
	vgm.vgfManager = vgfm
	vgm.mutex.Unlock()
}

/*
func (vgm *virtualGroupManager) refreshMetaByMetaService() {
	// TODO: impl
}
*/

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
// TODO: check storage params.
func (vgm *virtualGroupManager) PickGlobalVirtualGroup(vgfID uint32) (*vmmgr.GlobalVirtualGroupMeta, error) {
	vgm.mutex.RLock()
	defer vgm.mutex.RUnlock()
	return vgm.vgfManager.pickGlobalVirtualGroup(vgfID)
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
func (vgm *virtualGroupManager) GenerateGlobalVirtualGroupMeta(param *storagetypes.Params) (*vmmgr.GlobalVirtualGroupMeta, error) {
	vgm.mutex.RLock()
	defer vgm.mutex.RUnlock()
	return vgm.spManager.generateVirtualGroupMeta(param)
}
