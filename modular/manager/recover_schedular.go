package manager

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfsptqueue"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield/x/storage/types"
	"gorm.io/gorm"
)

const (
	recoverBatchSize         = 10
	pauseInterval            = 10 * time.Second
	maxRecoveryRetry         = 3
	MaxRecoveryTime          = 50
	primarySPRedundancyIndex = -1
)

type RecoverVGFScheduler struct {
	manager           *ManageModular
	RecoverSchedulers []*RecoverGVGScheduler
	VerifySchedulers  []*VerifyGVGScheduler
}

func NewRecoverVGFScheduler(m *ManageModular, vgfID uint32) (*RecoverVGFScheduler, error) {
	vgf, err := m.baseApp.Consensus().QueryVirtualGroupFamily(context.Background(), vgfID)
	if err != nil {
		log.Errorw("failed to GetVirtualGroupFamily", "error", err)
		return nil, err
	}
	if vgf == nil {
		log.Errorw("vgf not exist")
		return nil, fmt.Errorf("vgf not exist")
	}
	recoveryUnits := make([]*spdb.RecoverGVGStats, 0, len(vgf.GetGlobalVirtualGroupIds()))
	gvgSchedulers := make([]*RecoverGVGScheduler, 0, len(vgf.GetGlobalVirtualGroupIds()))

	verifyGVGProgresses := make([]*spdb.VerifyGVGProgress, 0, len(vgf.GetGlobalVirtualGroupIds()))
	verifySchedulers := make([]*VerifyGVGScheduler, 0, len(vgf.GetGlobalVirtualGroupIds()))

	for _, gvgID := range vgf.GetGlobalVirtualGroupIds() {
		recoveryUnits = append(recoveryUnits, &spdb.RecoverGVGStats{
			VirtualGroupFamilyID: vgfID,
			VirtualGroupID:       gvgID,
			RedundancyIndex:      primarySPRedundancyIndex,
			StartAfter:           0,
			Limit:                recoverBatchSize,
		})
		gvgScheduler, err := NewRecoverGVGScheduler(m, vgfID, gvgID, primarySPRedundancyIndex)
		if err != nil {
			log.Errorw("failed to new RecoverGVGScheduler")
			return nil, err
		}
		gvgSchedulers = append(gvgSchedulers, gvgScheduler)

		verifyGVGProgresses = append(verifyGVGProgresses, &spdb.VerifyGVGProgress{
			VirtualGroupID:  gvgID,
			RedundancyIndex: primarySPRedundancyIndex,
			StartAfter:      0,
			Limit:           recoverBatchSize,
		})
		verifyScheduler, err := NewVerifyGVGScheduler(m, gvgID, vgfID, primarySPRedundancyIndex)
		if err != nil {
			log.Errorw("failed to new VerifyGVGScheduler")
			return nil, err
		}
		verifySchedulers = append(verifySchedulers, verifyScheduler)
	}
	err = m.baseApp.GfSpDB().SetRecoverGVGStats(recoveryUnits)
	if err != nil {
		return nil, err
	}
	err = m.baseApp.GfSpDB().SetVerifyGVGProgress(verifyGVGProgresses)
	if err != nil {
		return nil, err
	}

	return &RecoverVGFScheduler{
		manager:           m,
		RecoverSchedulers: gvgSchedulers,
		VerifySchedulers:  verifySchedulers,
	}, nil
}

func (s *RecoverVGFScheduler) Start() {
	for _, g := range s.RecoverSchedulers {
		log.Infow("started recover scheduler for %d", "gvg_id", g.gvgID)
		go g.Start()
	}
	for _, v := range s.VerifySchedulers {
		log.Infow("started verify gvg scheduler for %d", "gvg_id", v.gvgID)
		go v.Start()
	}
}

type ObjectSegmentsStats struct {
	SegmentCount    int
	FailedSegments  vgmgr.IDSet
	SuccessSegments vgmgr.IDSet
}

type ObjectsSegmentsStats struct {
	mux   sync.RWMutex
	stats map[uint64]*ObjectSegmentsStats
}

func NewObjectsSegmentsStats() *ObjectsSegmentsStats {
	return &ObjectsSegmentsStats{
		stats: make(map[uint64]*ObjectSegmentsStats, 0),
	}
}

func (s *ObjectsSegmentsStats) put(objectID uint64, segmentCount uint32) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.stats[objectID] = &ObjectSegmentsStats{
		SegmentCount:    int(segmentCount),
		FailedSegments:  make(map[uint32]struct{}),
		SuccessSegments: make(map[uint32]struct{}),
	}
}

func (s *ObjectsSegmentsStats) has(objectID uint64) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	_, ok := s.stats[objectID]
	return ok
}

func (s *ObjectsSegmentsStats) get(objectID uint64) (*ObjectSegmentsStats, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	val, ok := s.stats[objectID]
	return val, ok
}

func (s *ObjectsSegmentsStats) remove(objectID uint64) {
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.stats, objectID)
}

func (s *ObjectsSegmentsStats) addSegmentRecord(objectID uint64, success bool, segmentIdx uint32) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	stats, ok := s.stats[objectID]
	if !ok {
		return false
	}
	if success {
		stats.SuccessSegments[segmentIdx] = struct{}{}
	} else {
		stats.FailedSegments[segmentIdx] = struct{}{}
	}
	return true
}

func (s *ObjectsSegmentsStats) isObjectRecovered(objectID uint64) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	stats, ok := s.stats[objectID]
	if !ok {
		return true
	}
	return len(stats.SuccessSegments) == stats.SegmentCount
}

func (s *ObjectsSegmentsStats) isObjectProcessed(objectID uint64) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	stats, ok := s.stats[objectID]
	if !ok {
		return false
	}
	return len(stats.SuccessSegments)+len(stats.FailedSegments) == stats.SegmentCount
}

func (s *ObjectsSegmentsStats) isRecoverFailed(objectID uint64) bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	stats, ok := s.stats[objectID]
	if !ok {
		return true
	}
	return len(stats.SuccessSegments)+len(stats.FailedSegments) == stats.SegmentCount && len(stats.FailedSegments) > 0
}

type RecoverGVGScheduler struct {
	manager               *ManageModular
	mtx                   sync.RWMutex
	currentBatchObjectIDs []uint64 // record ids of objects that been processed by reported task, either success or failed.

	//currentBatchID  uint64 // every 10 objects is a batch
	vgfID           uint32
	gvgID           uint32
	redundancyIndex int32
	curStartAfter   uint64
}

func NewRecoverGVGScheduler(m *ManageModular, vgfID, gvgID uint32, redundancyIndex int32) (*RecoverGVGScheduler, error) {
	if vgfID == 0 {
		err := m.baseApp.GfSpDB().SetRecoverGVGStats([]*spdb.RecoverGVGStats{{
			VirtualGroupFamilyID: vgfID,
			VirtualGroupID:       gvgID,
			RedundancyIndex:      redundancyIndex,
			StartAfter:           0,
			Limit:                recoverBatchSize,
		}})
		if err != nil {
			return nil, err
		}
	}
	return &RecoverGVGScheduler{
		manager:               m,
		currentBatchObjectIDs: make([]uint64, 0),
		vgfID:                 vgfID,
		gvgID:                 gvgID,
		redundancyIndex:       redundancyIndex,
	}, nil
}

func (s *RecoverGVGScheduler) Start() {
	storageParams, err := s.manager.baseApp.Consensus().QueryStorageParams(context.Background())
	if err != nil {
		log.Errorw("failed to get storage params", "err", err)
		return
	}
	maxSegmentSize := storageParams.GetMaxSegmentSize()

	gvgStats, err := s.manager.baseApp.GfSpDB().GetRecoverGVGStats(s.gvgID)
	if err != nil {
		log.Errorw("failed to get gvg stats", "err", err)
		return
	}
	if gvgStats.Status != int(spdb.Starting) {
		log.Infow("the recover gvg unit is not in starting status.")
		return
	}
	recoverTicker := time.NewTicker(10 * time.Second)
	defer recoverTicker.Stop()
	startAfter := gvgStats.StartAfter
	limit := gvgStats.Limit

	for {
		select {
		case <-recoverTicker.C:
			log.Debug("recover gvg loop")
			gvgStats, err = s.manager.baseApp.GfSpDB().GetRecoverGVGStats(s.gvgID)
			if err != nil {
				log.Errorw("failed to get gvg stats", "err", err)
				continue
			}
			if gvgStats.Status != int(spdb.Starting) {
				log.Infow("the recover gvg unit is not in starting status.")
				break
			}
			// upon all reported tasks for a batch objects are done, and finish update the objectSegmentStats, the monitor will update the startAfter.
			if gvgStats.StartAfter == s.curStartAfter && s.curStartAfter != 0 {
				continue
			}
			s.curStartAfter = gvgStats.StartAfter
			startAfter = gvgStats.StartAfter
			objects, err := s.manager.baseApp.GfSpClient().ListObjectsInGVG(context.Background(), gvgStats.VirtualGroupID, startAfter, uint32(limit))
			if err != nil {
				log.Errorw("failed to list objects in gvg", "start_after_object_id", startAfter, "limit", limit)
				continue
			}

			log.Debugw("list objects in GVG", "start_after", startAfter, "limit", limit, "objects_count", len(objects))

			if len(objects) == 0 {
				log.Infow("all objects in gvg have been processed", "start_after_object_id", startAfter, "limit", limit)
				gvgStats.Status = int(spdb.Processed) //it does not mean all objects are safely recovered in this GVG.
				err = s.manager.baseApp.GfSpDB().UpdateRecoverGVGStats(gvgStats)
				if err != nil {
					log.Error("failed to ", "start_after_object_id", startAfter, "limit", limit)
					continue
				}
				break
			}

			for _, object := range objects {
				objectInfo := object.Object.ObjectInfo
				objectID := objectInfo.Id.Uint64()
				segmentCount := segmentPieceCount(objectInfo.PayloadSize, maxSegmentSize)
				objectFailed := false // if there is any segment is failed to send out the recover task, will skip it; there is another recovering-failed object scheduler.
				log.Infow("starting to recover object", "object_info", objectInfo)
				for segmentIdx := uint32(0); segmentIdx < segmentCount; segmentIdx++ {
					task := &gfsptask.GfSpRecoverPieceTask{}
					task.InitRecoverPieceTask(objectInfo, storageParams, coretask.DefaultSmallerPriority, segmentIdx, s.redundancyIndex, maxSegmentSize, MaxRecoveryTime, maxRecoveryRetry)
					task.SetBySuccessorSP(true)
					task.SetGVGID(s.gvgID)
					err = s.manager.recoveryQueue.Push(task)
					if err != nil {
						log.Errorw("failed to push task to recover queue", "task_info", task)
						o := &spdb.RecoverFailedObject{
							ObjectID:        objectID,
							VirtualGroupID:  object.Gvg.Id,
							RedundancyIndex: gvgStats.RedundancyIndex,
						}
						err = s.manager.baseApp.GfSpDB().InsertRecoverFailedObject(o)
						if err != nil {
							log.Errorw("failed to InsertRecoverFailedObject", "error", err)
							return
						}
						objectFailed = true
						break
					}
				}
				if objectFailed {
					continue
				}
				s.manager.recoverObjectStats.put(objectID, segmentCount)
				s.currentBatchObjectIDs = append(s.currentBatchObjectIDs, objectID)
			}
			// once monitoring all objects related recovered piece tasks got response from executor, regardless success or failed,
			// the scheduler will update the StartAfter in recover gvg stats and jump to the next batch of objects to recover
			s.monitorBatch()
		}
	}
}

func segmentPieceCount(payloadSize uint64, maxSegmentSize uint64) uint32 {
	count := payloadSize / maxSegmentSize
	if payloadSize%maxSegmentSize > 0 {
		count++
	}
	return uint32(count)
}

func (s *RecoverGVGScheduler) monitorBatch() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		if len(s.currentBatchObjectIDs) == 0 {
			continue
		}
		log.Infow("monitoring for current batch objects", "object_ids", s.currentBatchObjectIDs)
		processedBatch := true
		for _, objectID := range s.currentBatchObjectIDs {
			if !s.manager.recoverObjectStats.isObjectProcessed(objectID) {
				processedBatch = false
				break
			}
		}
		if !processedBatch {
			continue
		}
		// all objects in the batch are processed.
		gvgStats, err := s.manager.baseApp.GfSpDB().GetRecoverGVGStats(s.gvgID)
		if err != nil {
			continue
		}
		gvgStats.StartAfter = gvgStats.StartAfter + gvgStats.Limit
		err = s.manager.baseApp.GfSpDB().UpdateRecoverGVGStats(gvgStats)
		if err != nil {
			continue
		}
		for _, objectID := range s.currentBatchObjectIDs {
			log.Infow("removing records", "object_id", objectID)
			s.manager.recoverObjectStats.remove(objectID)
		}
		s.currentBatchObjectIDs = make([]uint64, 0)
	}
}

// RecoverFailedObjectScheduler is used to scan the failed_object table for failed object entry, and retry the object recovers
// the entries are inserted from
// (1) RecoverGVGScheduler: Objects failed to recover. Including init the recover task and failed task reported from executor.
// (2) VerifyGVGScheduler: Objects are found to be missed when re-verify by calling api ListObjectsInGVG that
//
//	    verifying the object existence by querying integrate and piece_hash.
//		   A recover GVG unit is marked as completed from Processed only when all objects pass the verification.
type RecoverFailedObjectScheduler struct {
	manager *ManageModular
}

func NewRecoverFailedObjectScheduler(m *ManageModular) *RecoverFailedObjectScheduler {
	return &RecoverFailedObjectScheduler{
		manager: m,
	}
}

func (s *RecoverFailedObjectScheduler) Start() {
	storageParams, err := s.manager.baseApp.Consensus().QueryStorageParams(context.Background())
	if err != nil {
		return
	}
	maxSegmentSize := storageParams.GetMaxSegmentSize()

	ticker := time.NewTicker(50 * time.Second)
	for range ticker.C {

		recoverFailedObjects, err := s.manager.baseApp.GfSpDB().GetRecoverFailedObjects(3, 10)
		if err != nil {
			continue
		}
		if len(recoverFailedObjects) == 0 {
			continue
		}

	out:
		for _, o := range recoverFailedObjects {
			objectInfo, err := s.manager.baseApp.GfSpClient().GetObjectByID(context.Background(), o.ObjectID)
			if err != nil {
				break
			}
			segmentCount := segmentPieceCount(objectInfo.PayloadSize, maxSegmentSize)

			verified, err := verifyIntegrityAndPieceHash(s.manager, objectInfo, o.RedundancyIndex, maxSegmentSize)
			if err != nil {
				break
			}
			if verified {
				log.Infow("object has been recovered", "object", objectInfo)
				err = s.manager.baseApp.GfSpDB().DeleteRecoverFailedObject(o.ObjectID)
				if err != nil {
					break
				}
				continue
			}

			for segmentIdx := uint32(0); segmentIdx < segmentCount; segmentIdx++ {
				task := &gfsptask.GfSpRecoverPieceTask{}
				task.InitRecoverPieceTask(objectInfo, storageParams, coretask.DefaultSmallerPriority, segmentIdx, int32(o.RedundancyIndex), maxSegmentSize, MaxRecoveryTime, maxRecoveryRetry)
				task.SetBySuccessorSP(true)
				task.SetGVGID(o.VirtualGroupID)
				err = s.manager.recoveryQueue.Push(task)
				if err != nil {
					if errors.Is(err, gfsptqueue.ErrTaskQueueExceed) {
						break out
					}
				}
			}
			o.RetryTime++
			err = s.manager.baseApp.GfSpDB().UpdateRecoverFailedObject(o)
			if err != nil {
				break
			}
		}

	}

}

// VerifyGVGScheduler Objects are found to be missed when re-verify by calling api ListObjectsInGVG that
//
// verifying the object existence by querying integrate and piece_hash.
// a recover GVG unit is marked as completed from Processed only when all objects pass the verification.
type VerifyGVGScheduler struct {
	manager         *ManageModular
	vgfID           uint32
	gvgID           uint32
	redundancyIndex int32
	curStartAfter   uint64
}

func NewVerifyGVGScheduler(m *ManageModular, gvgID, vgfID uint32, redundancyIndex int32) (*VerifyGVGScheduler, error) {
	err := m.baseApp.GfSpDB().SetVerifyGVGProgress([]*spdb.VerifyGVGProgress{{
		VirtualGroupID:  gvgID,
		RedundancyIndex: redundancyIndex,
		Limit:           recoverBatchSize,
	}})

	if err != nil {
		return nil, err
	}
	return &VerifyGVGScheduler{
		manager:         m,
		vgfID:           vgfID,
		gvgID:           gvgID,
		redundancyIndex: redundancyIndex,
	}, nil
}

func (s *VerifyGVGScheduler) Start() {
	storageParams, err := s.manager.baseApp.Consensus().QueryStorageParams(context.Background())
	if err != nil {
		return
	}
	maxSegmentSize := storageParams.GetMaxSegmentSize()
	verifyTicker := time.NewTicker(20 * time.Second)
	defer verifyTicker.Stop()

	for {
		select {
		case <-verifyTicker.C:
			log.Infow("verify gvg scheduler")
			gvgStats, err := s.manager.baseApp.GfSpDB().GetRecoverGVGStats(s.gvgID)
			if err != nil {
				log.Errorw("failed to get recover gvg stats", "err", err)
				continue
			}
			if gvgStats.Status != int(spdb.Processed) {
				log.Infow("Wait for all objects in GVG have been processed")
				continue
			}

			verifyGVGProgress, err := s.manager.baseApp.GfSpDB().GetVerifyGVGProgress(s.gvgID)
			if err != nil {
				log.Errorw("failed to verify gvg progress", "err", err)
				continue
			}
			startAfter := verifyGVGProgress.StartAfter
			limit := verifyGVGProgress.Limit

			objects, err := s.manager.baseApp.GfSpClient().ListObjectsInGVG(context.Background(), gvgStats.VirtualGroupID, startAfter, uint32(limit))
			if err != nil {
				log.Errorw("failed to list object in GVG", "err", err)
				continue
			}

			log.Infow("list objects in GVG", "start_after", startAfter, "limit", limit)

			if len(objects) == 0 {
				log.Infow("all objects in gvg have been verified", "start_after_object_id", startAfter, "limit", limit)

				// todo check if there is failed object in DB for such GVG.
				gvgStats.Status = int(spdb.Completed)
				err = s.manager.baseApp.GfSpDB().UpdateRecoverGVGStats(gvgStats)
				if err != nil {
					log.Error("failed to ", "start_after_object_id", startAfter, "limit", limit)
					continue
				}
				break
			}

			for _, o := range objects {
				objectInfo := o.Object.ObjectInfo
				objectID := objectInfo.Id.Uint64()
				verified, err := verifyIntegrityAndPieceHash(s.manager, objectInfo, s.redundancyIndex, maxSegmentSize)
				if err != nil {
					log.Errorw("failed to verify integrity and piece hash", "object", objectInfo, "error", err)
					break
				}
				if !verified {
					log.Errorw("verify object failed", "object", objectInfo)
					failedObject := &spdb.RecoverFailedObject{
						ObjectID:        objectID,
						VirtualGroupID:  s.gvgID,
						RedundancyIndex: s.redundancyIndex,
					}
					err = s.manager.baseApp.GfSpDB().InsertRecoverFailedObject(failedObject)
					if err != nil {
						break
					}
				}
			}
		}
	}
}

func verifyIntegrityAndPieceHash(m *ManageModular, object *types.ObjectInfo, redundancyIndex int32, maxSegmentSize uint64) (bool, error) {
	_, err := m.baseApp.GfSpDB().GetObjectIntegrity(object.Id.Uint64(), redundancyIndex)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		log.Errorw("failed to get object integrity hash", "objectName:", object.ObjectName, "error", err)
		return false, err
	}
	segmentCount := segmentPieceCount(object.PayloadSize, maxSegmentSize)
	pieceChecksums, err := m.baseApp.GfSpDB().GetAllReplicatePieceChecksumOptimized(object.Id.Uint64(), redundancyIndex, segmentCount)
	if err != nil {
		return false, err
	}
	if len(pieceChecksums) != int(segmentCount) {
		return false, nil
	}
	return true, nil
}
