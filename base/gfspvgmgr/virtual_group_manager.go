package gfspvgmgr

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var _ vgmgr.VirtualGroupManager = &virtualGroupManager{}

const (
	VirtualGroupManagerSpace            = "VirtualGroupManager"
	RefreshMetaInterval                 = 2 * time.Second
	MaxStorageUsageRatio                = 0.95
	DefaultInitialGVGStakingStorageSize = uint64(8) * 1024 * 1024 * 1024 // 8GB per GVG, chain side DefaultMaxStoreSizePerFamily is 64 GB
)

var (
	ErrFailedPickVGF = gfsperrors.Register(VirtualGroupManagerSpace, http.StatusInternalServerError, 540001,
		"failed to pick virtual group family, need create global virtual group")
	ErrFailedPickGVG = gfsperrors.Register(VirtualGroupManagerSpace, http.StatusInternalServerError, 540002,
		"failed to pick global virtual group, need stake more storage size")
	ErrStaledMetadata = gfsperrors.Register(VirtualGroupManagerSpace, http.StatusInternalServerError, 540003,
		"metadata is staled, need force refresh metadata")
	ErrFailedPickDestSP = gfsperrors.Register(VirtualGroupManagerSpace, http.StatusInternalServerError, 540004,
		"failed to pick dest sp due to has conflict, need resolve conflict")
)

// virtualGroupFamilyManager is built by metadata data source.
type virtualGroupFamilyManager struct {
	vgfIDToVgf map[uint32]*vgmgr.VirtualGroupFamilyMeta // is used to pick VGF
}

// FreeStorageSizeWeightPicker is used to pick index by storage usage,
// The more free storage usage, the greater the probability of being picked.
type FreeStorageSizeWeightPicker struct {
	freeStorageSizeWeightMap map[uint32]float64
}

func (vgfp *FreeStorageSizeWeightPicker) addVirtualGroupFamily(vgf *vgmgr.VirtualGroupFamilyMeta) {
	if float64(vgf.FamilyUsedStorageSize) >= MaxStorageUsageRatio*float64(vgf.FamilyStakingStorageSize) || vgf.FamilyStakingStorageSize == 0 {
		return
	}
	vgfp.freeStorageSizeWeightMap[vgf.ID] = float64(vgf.FamilyStakingStorageSize-vgf.FamilyUsedStorageSize) / float64(vgf.FamilyStakingStorageSize)
}

func (vgfp *FreeStorageSizeWeightPicker) addGlobalVirtualGroup(gvg *vgmgr.GlobalVirtualGroupMeta) {
	if float64(gvg.UsedStorageSize) >= MaxStorageUsageRatio*float64(gvg.StakingStorageSize) {
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

func (vgfm *virtualGroupFamilyManager) pickVirtualGroupFamily() (*vgmgr.VirtualGroupFamilyMeta, error) {
	var (
		picker   FreeStorageSizeWeightPicker
		familyID uint32
		err      error
	)
	picker.freeStorageSizeWeightMap = make(map[uint32]float64)
	for _, f := range vgfm.vgfIDToVgf {
		picker.addVirtualGroupFamily(f)
	}
	if familyID, err = picker.pickIndex(); err != nil {
		log.Errorw("failed to pick vgf", "error", err)
		return nil, ErrFailedPickVGF
	}
	return vgfm.vgfIDToVgf[familyID], nil
}

func (vgfm *virtualGroupFamilyManager) pickGlobalVirtualGroup(vgfID uint32) (*vgmgr.GlobalVirtualGroupMeta, error) {
	var (
		picker               FreeStorageSizeWeightPicker
		globalVirtualGroupID uint32
		err                  error
	)
	picker.freeStorageSizeWeightMap = make(map[uint32]float64)
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

func (sm *spManager) generateVirtualGroupMeta(param *storagetypes.Params) (*vgmgr.GlobalVirtualGroupMeta, error) {
	secondarySPNumber := int(param.GetRedundantDataChunkNum() + param.GetRedundantParityChunkNum())
	if sm.primarySP == nil || len(sm.secondarySPs) < secondarySPNumber {
		return nil, fmt.Errorf("no enough sp")
	}
	secondarySPIDs := make([]uint32, 0)
	for i, sp := range sm.secondarySPs {
		if i < secondarySPNumber {
			secondarySPIDs = append(secondarySPIDs, sp.GetId())
		}
	}
	return &vgmgr.GlobalVirtualGroupMeta{
		PrimarySPID:        sm.primarySP.Id,
		SecondarySPIDs:     secondarySPIDs,
		StakingStorageSize: DefaultInitialGVGStakingStorageSize,
	}, nil
}

func (sm *spManager) pickSPByFilter(filter vgmgr.PickFilter) (*sptypes.StorageProvider, error) {
	for _, secondarySP := range sm.secondarySPs {
		if pickSucceed := filter.Check(secondarySP.GetId()); pickSucceed {
			return secondarySP, nil
		}
	}
	return nil, ErrFailedPickDestSP
}

// TODO: add metadata service client.
type virtualGroupManager struct {
	selfOperatorAddress string
	chainClient         consensus.Consensus // query VG params from chain
	mutex               sync.RWMutex
	selfSPID            uint32
	spManager           *spManager // is used to generate a new gvg
	vgParams            *virtualgrouptypes.Params
	vgfManager          *virtualGroupFamilyManager
}

// NewVirtualGroupManager returns a virtual group manager interface.
func NewVirtualGroupManager(selfOperatorAddress string, chainClient consensus.Consensus) (vgmgr.VirtualGroupManager, error) {
	vgm := &virtualGroupManager{
		selfOperatorAddress: selfOperatorAddress,
		chainClient:         chainClient,
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
		spMap           map[uint32]*sptypes.StorageProvider
	)

	spMap = make(map[uint32]*sptypes.StorageProvider)
	toSPEndpoints := func(spIDs []uint32) []string {
		spInfoEndpoints := make([]string, 0)
		for _, id := range spIDs {
			spInfoEndpoints = append(spInfoEndpoints, spMap[id].GetEndpoint())
		}
		return spInfoEndpoints
	}
	// query meta
	if spList, err = vgm.chainClient.ListSPs(context.Background()); err != nil {
		log.Errorw("failed to list sps", "error", err)
		return
	}
	for _, sp := range spList {
		spMap[sp.Id] = sp
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
	// log.Infow("list sp info", "primary_sp", primarySP, "secondary_sps", secondarySPList, "sp_map", spMap)

	if spID == 0 {
		log.Error("failed to refresh due to current sp is not in sp list")
		return
	}

	if vgParams, err = vgm.chainClient.QueryVirtualGroupParams(context.Background()); err != nil {
		log.Errorw("failed to query virtual group params", "error", err)
		return
	}
	// log.Infow("query virtual group params", "params", vgParams)

	vgfm = &virtualGroupFamilyManager{
		vgfIDToVgf: make(map[uint32]*vgmgr.VirtualGroupFamilyMeta),
	}
	if vgfList, err = vgm.chainClient.ListVirtualGroupFamilies(context.Background(), spID); err != nil {
		log.Errorw("failed to list virtual group family", "error", err)
		return
	}

	// log.Infow("list virtual group family info", "vgf_list", vgfList)
	for _, vgf := range vgfList {
		vgfm.vgfIDToVgf[vgf.Id] = &vgmgr.VirtualGroupFamilyMeta{
			ID:          vgf.Id,
			PrimarySPID: spID,
			GVGMap:      make(map[uint32]*vgmgr.GlobalVirtualGroupMeta),
		}
		for _, gvgID := range vgf.GlobalVirtualGroupIds {
			var gvg *virtualgrouptypes.GlobalVirtualGroup
			if gvg, err = vgm.chainClient.QueryGlobalVirtualGroup(context.Background(), gvgID); err != nil {
				log.Errorw("failed to query global virtual group", "error", err)
				return
			}
			gvgMeta := &vgmgr.GlobalVirtualGroupMeta{
				ID:                   gvg.GetId(),
				FamilyID:             vgf.Id,
				PrimarySPID:          spID,
				SecondarySPIDs:       gvg.GetSecondarySpIds(),
				SecondarySPEndpoints: toSPEndpoints(gvg.GetSecondarySpIds()),
				UsedStorageSize:      gvg.GetStoredSize(),
				StakingStorageSize:   vgm.GetTotalStakingStoreSizeOfGVG(gvg),
			}
			log.Infow("query global virtual group info", "gvg_info", gvg, "gvg_meta", gvgMeta)
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

// PickVirtualGroupFamily pick a virtual group family(If failed to pick,
// new VGF will be automatically created on the chain) in get create bucket approval workflow.
func (vgm *virtualGroupManager) PickVirtualGroupFamily() (*vgmgr.VirtualGroupFamilyMeta, error) {
	vgm.mutex.RLock()
	defer vgm.mutex.RUnlock()
	return vgm.vgfManager.pickVirtualGroupFamily()
}

// PickGlobalVirtualGroup picks a global virtual group(If failed to pick,
// new GVG will be created by primary SP) in replicate/seal object workflow.
func (vgm *virtualGroupManager) PickGlobalVirtualGroup(vgfID uint32) (*vgmgr.GlobalVirtualGroupMeta, error) {
	vgm.mutex.RLock()
	defer vgm.mutex.RUnlock()
	return vgm.vgfManager.pickGlobalVirtualGroup(vgfID)
}

// ForceRefreshMeta is used to query metadata service and refresh the virtual group manager meta.
// if pick func returns ErrStaledMetadata, the caller need force refresh from metadata and retry pick.
func (vgm *virtualGroupManager) ForceRefreshMeta() error {
	// sleep 2 seconds for waiting a new block
	time.Sleep(2 * time.Second)
	vgm.refreshMeta()
	return nil
}

// GenerateGlobalVirtualGroupMeta is used to generate a gvg meta.
// TODO: support more generation strategies.
func (vgm *virtualGroupManager) GenerateGlobalVirtualGroupMeta(param *storagetypes.Params) (*vgmgr.GlobalVirtualGroupMeta, error) {
	vgm.mutex.RLock()
	defer vgm.mutex.RUnlock()
	return vgm.spManager.generateVirtualGroupMeta(param)
}

// TODO: add a generate gvg by filter api for migrate.

// PickSPByFilter is used to pick sp by filter check.
func (vgm *virtualGroupManager) PickSPByFilter(filter vgmgr.PickFilter) (*sptypes.StorageProvider, error) {
	vgm.mutex.RLock()
	defer vgm.mutex.RUnlock()
	return vgm.spManager.pickSPByFilter(filter)
}

func (vgm *virtualGroupManager) GetTotalStakingStoreSizeOfGVG(gvg *virtualgrouptypes.GlobalVirtualGroup) uint64 {
	total := gvg.TotalDeposit.Quo(vgm.vgParams.GvgStakingPerBytes)
	if !total.IsUint64() {
		return math.MaxUint64
	}
	return total.Uint64()
}
