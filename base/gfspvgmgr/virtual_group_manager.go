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
	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var _ vgmgr.VirtualGroupManager = &virtualGroupManager{}

const (
	VirtualGroupManagerSpace            = "VirtualGroupManager"
	RefreshMetaInterval                 = 2 * time.Second
	MaxStorageUsageRatio                = 0.95
	DefaultInitialGVGStakingStorageSize = uint64(8) * 1024 * 1024 * 1024 * 1024 // 8TB per GVG, chain side DefaultMaxStoreSizePerFamily is 64 TB
)

var (
	ErrFailedPickVGF = gfsperrors.Register(VirtualGroupManagerSpace, http.StatusInternalServerError, 540001,
		"failed to pick virtual group family, need creating global virtual group")
	ErrFailedPickGVG = gfsperrors.Register(VirtualGroupManagerSpace, http.StatusInternalServerError, 540002,
		"failed to pick global virtual group, need staking more storage size")
	ErrStaledMetadata = gfsperrors.Register(VirtualGroupManagerSpace, http.StatusInternalServerError, 540003,
		"metadata is staled, need forcing refresh metadata")
	ErrFailedPickDestSP = gfsperrors.Register(VirtualGroupManagerSpace, http.StatusInternalServerError, 540004,
		"failed to pick dest sp")
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

func (picker *FreeStorageSizeWeightPicker) addVirtualGroupFamily(vgf *vgmgr.VirtualGroupFamilyMeta) {
	if float64(vgf.FamilyUsedStorageSize) >= MaxStorageUsageRatio*float64(vgf.FamilyStakingStorageSize) || vgf.FamilyStakingStorageSize == 0 {
		return
	}
	picker.freeStorageSizeWeightMap[vgf.ID] = float64(vgf.FamilyStakingStorageSize-vgf.FamilyUsedStorageSize) / float64(vgf.FamilyStakingStorageSize)
}

func (picker *FreeStorageSizeWeightPicker) addGlobalVirtualGroup(gvg *vgmgr.GlobalVirtualGroupMeta) {
	if float64(gvg.UsedStorageSize) >= MaxStorageUsageRatio*float64(gvg.StakingStorageSize) {
		return
	}
	picker.freeStorageSizeWeightMap[gvg.ID] = float64(gvg.StakingStorageSize-gvg.UsedStorageSize) / float64(gvg.StakingStorageSize)
}

func (picker *FreeStorageSizeWeightPicker) pickIndex() (uint32, error) {
	var (
		sumWeight     float64
		pickWeight    float64
		tempSumWeight float64
	)
	for _, value := range picker.freeStorageSizeWeightMap {
		sumWeight += value
	}
	pickWeight = rand.Float64() * sumWeight
	for key, value := range picker.freeStorageSizeWeightMap {
		tempSumWeight += value
		if tempSumWeight > pickWeight {
			return key, nil
		}
	}
	return 0, fmt.Errorf("failed to pick weighted random index")
}

func (vgfm *virtualGroupFamilyManager) pickVirtualGroupFamily(filter *vgmgr.PickVGFFilter, filterByGvgList *vgmgr.PickVGFByGVGFilter) (*vgmgr.VirtualGroupFamilyMeta, error) {
	var (
		picker   FreeStorageSizeWeightPicker
		familyID uint32
		err      error
	)
	picker.freeStorageSizeWeightMap = make(map[uint32]float64)
	for id, f := range vgfm.vgfIDToVgf {
		if filter != nil && !filter.Check(id) {
			continue
		}
		if !filterByGvgList.Check(f.GVGMap) {
			continue
		}
		picker.addVirtualGroupFamily(f)
	}

	if familyID, err = picker.pickIndex(); err != nil {
		log.Errorw("failed to pick vgf", "error", err)
		return nil, ErrFailedPickVGF
	}
	return vgfm.vgfIDToVgf[familyID], nil
}

func (vgfm *virtualGroupFamilyManager) pickGlobalVirtualGroup(vgfID uint32, filter, excludeGVGsFilter vgmgr.ExcludeFilter) (*vgmgr.GlobalVirtualGroupMeta, error) {
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
		if filter != nil && filter.Apply(g.ID) {
			continue
		}
		if excludeGVGsFilter != nil && excludeGVGsFilter.Apply(g.ID) {
			continue
		}
		picker.addGlobalVirtualGroup(g)
	}

	if globalVirtualGroupID, err = picker.pickIndex(); err != nil {
		log.Errorw("failed to pick gvg", "vgf_id", vgfID, "error", err)
		return nil, ErrFailedPickGVG
	}
	return vgfm.vgfIDToVgf[vgfID].GVGMap[globalVirtualGroupID], nil
}

func (vgfm *virtualGroupFamilyManager) pickGlobalVirtualGroupForBucketMigrate(vgfFilter vgmgr.VGFPickFilter, gvgFilter vgmgr.GVGPickFilter) (*vgmgr.GlobalVirtualGroupMeta, error) {
	var (
		picker               FreeStorageSizeWeightPicker
		globalVirtualGroupID uint32
		err                  error
	)

	for vgfID, vgf := range vgfm.vgfIDToVgf {
		if vgfFilter != nil && !vgfFilter.Check(vgfID) {
			continue
		}
		if !gvgFilter.CheckFamily(vgfID) {
			continue
		}
		picker.freeStorageSizeWeightMap = make(map[uint32]float64)
		for _, gvg := range vgf.GVGMap {
			log.Debugw("prepare to add pickGlobalVirtualGroupForBucketMigrate", "gvg", gvg)
			if gvgFilter.CheckGVG(gvg) {
				picker.addGlobalVirtualGroup(gvg)
				log.Debugw("add pickGlobalVirtualGroupForBucketMigrate", "gvg", gvg)
			}
		}
		if globalVirtualGroupID, err = picker.pickIndex(); err != nil {
			log.Errorw("failed to pick gvg at current vgf", "vgf_id", vgfID, "error", err)
			continue
		}
		return vgfm.vgfIDToVgf[vgfID].GVGMap[globalVirtualGroupID], nil
	}
	log.Errorw("failed to pick gvg for migrate bucket")
	return nil, ErrFailedPickGVG
}

type spManager struct {
	selfSP   *sptypes.StorageProvider
	otherSPs []*sptypes.StorageProvider
}

func (sm *spManager) generateVirtualGroupMeta(genPolicy vgmgr.GenerateGVGSecondarySPsPolicy, filter, excludeSPsFilter vgmgr.ExcludeFilter) (*vgmgr.GlobalVirtualGroupMeta, error) {
	for _, sp := range sm.otherSPs {
		if !sp.IsInService() {
			continue
		}
		if filter != nil && filter.Apply(sp.Id) {
			continue
		}
		if excludeSPsFilter != nil && excludeSPsFilter.Apply(sp.Id) {
			continue
		}
		genPolicy.AddCandidateSP(sp.GetId())
	}
	secondarySPIDs, err := genPolicy.GenerateGVGSecondarySPs()
	if err != nil {
		return nil, err
	}

	return &vgmgr.GlobalVirtualGroupMeta{
		PrimarySPID:        sm.selfSP.Id,
		SecondarySPIDs:     secondarySPIDs,
		StakingStorageSize: DefaultInitialGVGStakingStorageSize,
	}, nil
}

func (sm *spManager) pickSPByFilter(filter vgmgr.SPPickFilter) (*sptypes.StorageProvider, error) {
	for _, destSP := range sm.otherSPs {
		if !destSP.IsInService() {
			continue
		}
		if pickSucceed := filter.Check(destSP.GetId()); pickSucceed {
			return destSP, nil
		}
	}
	return nil, ErrFailedPickDestSP
}

func (sm *spManager) querySPByID(spID uint32) (*sptypes.StorageProvider, error) {
	if sm.selfSP.GetId() == spID {
		return sm.selfSP, nil
	}
	for _, sp := range sm.otherSPs {
		if sp.GetId() == spID {
			return sp, nil
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
	freezeSPPool        *FreezeSPPool
}

// NewVirtualGroupManager returns a virtual group manager interface.
func NewVirtualGroupManager(selfOperatorAddress string, chainClient consensus.Consensus) (vgmgr.VirtualGroupManager, error) {
	vgm := &virtualGroupManager{
		selfOperatorAddress: selfOperatorAddress,
		chainClient:         chainClient,
		freezeSPPool:        NewFreezeSPPool(),
	}
	vgm.refreshMeta()
	go func() {
		RefreshMetaTicker := time.NewTicker(RefreshMetaInterval)
		for range RefreshMetaTicker.C {
			// log.Info("start to refresh virtual group manager meta")
			vgm.refreshMeta()
			// log.Info("finish to refresh virtual group manager meta")
		}
	}()
	go vgm.releaseSPAndGVGLoop()
	return vgm, nil
}

// refreshMetadata is used to refresh virtual group manager metadata in background.
func (vgm *virtualGroupManager) refreshMeta() {
	// TODO: support load from metadata.
	vgm.refreshMetaByChain()
}

func (vgm *virtualGroupManager) refreshMetaByChain() {
	var (
		err         error
		spList      []*sptypes.StorageProvider
		selfSP      *sptypes.StorageProvider
		otherSPList []*sptypes.StorageProvider
		spID        uint32
		spm         *spManager
		vgParams    *virtualgrouptypes.Params
		vgfList     []*virtualgrouptypes.GlobalVirtualGroupFamily
		vgfm        *virtualGroupFamilyManager
		spMap       map[uint32]*sptypes.StorageProvider
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
			selfSP = sp
			otherSPList = append(spList[:i], spList[i+1:]...)
		}
	}
	spm = &spManager{
		selfSP:   selfSP,
		otherSPs: otherSPList,
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
				StakingStorageSize:   util.TotalStakingStoreSizeOfGVG(gvg, vgParams.GvgStakingPerBytes),
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
func (vgm *virtualGroupManager) PickVirtualGroupFamily(filterByGvgList *vgmgr.PickVGFByGVGFilter) (*vgmgr.VirtualGroupFamilyMeta, error) {
	filter, err := vgm.genVgfFilter()
	if err != nil {
		return nil, err
	}
	vgm.mutex.RLock()
	defer vgm.mutex.RUnlock()
	return vgm.vgfManager.pickVirtualGroupFamily(filter, filterByGvgList)
}

// PickGlobalVirtualGroup picks a global virtual group(If failed to pick,
// new GVG will be created by primary SP) in replicate/seal object workflow.
func (vgm *virtualGroupManager) PickGlobalVirtualGroup(vgfID uint32, excludeGVGsFilter vgmgr.ExcludeFilter) (*vgmgr.GlobalVirtualGroupMeta, error) {
	vgm.mutex.RLock()
	defer vgm.mutex.RUnlock()
	return vgm.vgfManager.pickGlobalVirtualGroup(vgfID, vgmgr.NewExcludeIDFilter(vgm.freezeSPPool.GetFreezeGVGsInFamily(vgfID)), excludeGVGsFilter)
}

// PickGlobalVirtualGroupForBucketMigrate picks a global virtual group(If failed to pick,
// new GVG will be created by primary SP) in migrate bucket workflow.
func (vgm *virtualGroupManager) PickGlobalVirtualGroupForBucketMigrate(filter vgmgr.GVGPickFilter) (*vgmgr.GlobalVirtualGroupMeta, error) {
	vgm.mutex.RLock()
	defer vgm.mutex.RUnlock()
	vgfFilter, err := vgm.genVgfFilter()
	if err != nil {
		return nil, err
	}
	return vgm.vgfManager.pickGlobalVirtualGroupForBucketMigrate(vgfFilter, filter)
}

// PickMigrateDestGlobalVirtualGroup picks a global virtual group(If failed to pick,
// new GVG will be created by primary SP) in replicate/seal object workflow.
func (vgm *virtualGroupManager) PickMigrateDestGlobalVirtualGroup(vgfID uint32, excludeGVGsFilter vgmgr.ExcludeFilter) (*vgmgr.GlobalVirtualGroupMeta, error) {
	vgm.mutex.RLock()
	defer vgm.mutex.RUnlock()
	return vgm.vgfManager.pickGlobalVirtualGroup(vgfID, vgmgr.NewExcludeIDFilter(vgm.freezeSPPool.GetFreezeGVGsInFamily(vgfID)), excludeGVGsFilter)
}

// ForceRefreshMeta is used to query metadata service and refresh the virtual group manager meta.
// if pick func returns ErrStaledMetadata, the caller need force refresh from metadata and retry pick.
// TODO: in the future background thread can pre-allocate gvg and reduce impact on foreground thread.
func (vgm *virtualGroupManager) ForceRefreshMeta() error {
	// sleep 5 seconds for waiting a new block
	time.Sleep(5 * time.Second)
	vgm.refreshMeta()
	return nil
}

// GenerateGlobalVirtualGroupMeta is used to generate a gvg meta.
func (vgm *virtualGroupManager) GenerateGlobalVirtualGroupMeta(genPolicy vgmgr.GenerateGVGSecondarySPsPolicy, excludeSPsFilter vgmgr.ExcludeFilter) (*vgmgr.GlobalVirtualGroupMeta, error) {
	vgm.mutex.RLock()
	defer vgm.mutex.RUnlock()
	return vgm.spManager.generateVirtualGroupMeta(genPolicy, vgmgr.NewExcludeIDFilter(vgm.freezeSPPool.GetFreezeSPIDs()), excludeSPsFilter)
}

// PickSPByFilter is used to pick sp by filter check.
func (vgm *virtualGroupManager) PickSPByFilter(filter vgmgr.SPPickFilter) (*sptypes.StorageProvider, error) {
	vgm.mutex.RLock()
	defer vgm.mutex.RUnlock()
	return vgm.spManager.pickSPByFilter(filter)
}

// QuerySPByID return sp info by sp id.
func (vgm *virtualGroupManager) QuerySPByID(spID uint32) (*sptypes.StorageProvider, error) {
	vgm.mutex.RLock()
	sp, err := vgm.spManager.querySPByID(spID)
	vgm.mutex.RUnlock()
	if err == nil {
		return sp, nil
	}
	// query chain if it is not found in memory topology.
	return vgm.chainClient.QuerySPByID(context.Background(), spID)
}

func (vgm *virtualGroupManager) genVgfFilter() (*vgmgr.PickVGFFilter, error) {
	vgm.mutex.RLock()
	vgfIDs := make([]uint32, len(vgm.vgfManager.vgfIDToVgf))
	if len(vgfIDs) == 0 {
		vgm.mutex.RUnlock()
		return nil, ErrFailedPickVGF
	}
	i := 0
	for k := range vgm.vgfManager.vgfIDToVgf {
		vgfIDs[i] = k
		i++
	}
	vgm.mutex.RUnlock()
	availableVgfIDs, err := vgm.chainClient.AvailableGlobalVirtualGroupFamilies(context.Background(), vgfIDs)
	if err != nil {
		log.Errorw("failed to get available virtual group families", "error", err)
		return nil, nil
	}
	return vgmgr.NewPickVGFFilter(availableVgfIDs), nil
}

// FreezeSPAndGVGs freeze a secondary SP and its GVGs
func (vgm *virtualGroupManager) FreezeSPAndGVGs(spID uint32, gvgs []*virtualgrouptypes.GlobalVirtualGroup) {
	vgm.freezeSPPool.FreezeSPAndGVGs(spID, gvgs)
}

// releaseSPAndGVGLoop runs periodically to release SP from the freeze pool
func (vgm *virtualGroupManager) releaseSPAndGVGLoop() {
	ticker := time.NewTicker(ReleaseSPJobInterval)
	for range ticker.C {
		vgm.freezeSPPool.ReleaseSP()

		vgm.mutex.RLock()
		aliveSP := make([]*sptypes.StorageProvider, 0)
		for _, sp := range vgm.spManager.otherSPs {
			if sp.IsInService() {
				aliveSP = append(aliveSP, sp)
			}
		}
		vgm.mutex.RUnlock()
		params, err := vgm.chainClient.QueryStorageParamsByTimestamp(context.Background(), time.Now().Unix())
		if err != nil {
			continue
		}
		if len(aliveSP)-len(vgm.freezeSPPool.GetFreezeSPIDs()) < int(params.GetRedundantDataChunkNum()+params.GetRedundantParityChunkNum()) {
			vgm.freezeSPPool.ReleaseAllSP()
		}
	}
}
