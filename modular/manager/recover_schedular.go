package manager

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfsptqueue"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield/x/storage/types"
	types2 "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"gorm.io/gorm"
)

const (
	recoverBatchSize         = 10
	maxRecoveryRetry         = 3
	MaxRecoveryTime          = 50
	primarySPRedundancyIndex = -1

	recoverInterval = 5 * time.Second

	verifyGVGQueryLimit = uint32(50)
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
	recoveryGVG := make([]*spdb.RecoverGVGStats, 0, len(vgf.GetGlobalVirtualGroupIds()))
	gvgSchedulers := make([]*RecoverGVGScheduler, 0, len(vgf.GetGlobalVirtualGroupIds()))
	verifySchedulers := make([]*VerifyGVGScheduler, 0, len(vgf.GetGlobalVirtualGroupIds()))

	for _, gvgID := range vgf.GetGlobalVirtualGroupIds() {
		recoveryGVG = append(recoveryGVG, &spdb.RecoverGVGStats{
			VirtualGroupFamilyID: vgfID,
			VirtualGroupID:       gvgID,
			RedundancyIndex:      primarySPRedundancyIndex,
			StartAfter:           0,
			Limit:                recoverBatchSize,
			ObjectCount:          0,
		})
		gvgScheduler, err := NewRecoverGVGScheduler(m, vgfID, gvgID, primarySPRedundancyIndex)
		if err != nil {
			log.Errorw("failed to create RecoverGVGScheduler")
			return nil, err
		}
		gvgSchedulers = append(gvgSchedulers, gvgScheduler)

		verifyScheduler, err := NewVerifyGVGScheduler(m, vgfID, gvgID, primarySPRedundancyIndex)
		if err != nil {
			log.Errorw("failed to create VerifyGVGScheduler")
			return nil, err
		}
		verifySchedulers = append(verifySchedulers, verifyScheduler)
	}
	err = m.baseApp.GfSpDB().SetRecoverGVGStats(recoveryGVG)
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
	s.mux.Lock()
	defer s.mux.Unlock()
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
	currentBatchObjectIDs map[uint64]struct{}
	vgfID                 uint32
	gvgID                 uint32
	redundancyIndex       int32
	curStartAfter         uint64
}

func NewRecoverGVGScheduler(m *ManageModular, vgfID, gvgID uint32, redundancyIndex int32) (*RecoverGVGScheduler, error) {
	if vgfID == 0 {
		err := m.baseApp.GfSpDB().SetRecoverGVGStats([]*spdb.RecoverGVGStats{{
			VirtualGroupFamilyID: vgfID,
			VirtualGroupID:       gvgID,
			RedundancyIndex:      redundancyIndex,
			StartAfter:           0,
			Limit:                recoverBatchSize, // TODO
		}})
		if err != nil {
			return nil, err
		}
	}
	return &RecoverGVGScheduler{
		manager:               m,
		currentBatchObjectIDs: make(map[uint64]struct{}),
		vgfID:                 vgfID,
		gvgID:                 gvgID,
		redundancyIndex:       redundancyIndex,
	}, nil
}

func (s *RecoverGVGScheduler) Start() {
	defer s.manager.recoverProcessCount.Add(-1)
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
	if gvgStats.Status != spdb.Processing {
		log.Infow("the gvg is already processed")
		return
	}
	recoverTicker := time.NewTicker(recoverInterval)
	defer recoverTicker.Stop()

	startAfter := gvgStats.StartAfter
	limit := gvgStats.Limit

	recoveryCompacity := s.manager.recoveryQueue.Cap()

	objectCount := uint64(0)
	for {
		select {
		case <-recoverTicker.C:
			log.Infow("looping")
			gvgStats, err = s.manager.baseApp.GfSpDB().GetRecoverGVGStats(s.gvgID)
			if err != nil {
				log.Errorw("failed to get gvg stats", "err", err)
				continue
			}
			if gvgStats.Status != spdb.Processing {
				log.Infow("the gvg is already processed", "gvg_id", s.gvgID)
				return
			}

			//if gvgStats.StartAfter == s.curStartAfter && s.curStartAfter != 0 {
			//	log.Debugw("still processing the batch that after object id", "start_after", s.curStartAfter)
			//	continue
			//}
			s.curStartAfter = gvgStats.StartAfter
			startAfter = gvgStats.StartAfter

			//10 == 10
			log.Infow("processing the batch that after object id", "start_after", s.curStartAfter)
			objects, err := s.manager.baseApp.GfSpClient().ListObjectsInGVG(context.Background(), gvgStats.VirtualGroupID, startAfter, uint32(limit))
			if err != nil {
				log.Errorw("failed to list objects in gvg", "start_after_object_id", startAfter, "limit", limit)
				continue
			}

			log.Debugw("list objects in GVG", "start_after", startAfter, "limit", limit, "objects_count", len(objects))

			//objectCount += uint64(len(objects))
			if len(objects) == 0 {
				log.Infow("all objects in gvg have been processed", "start_after_object_id", startAfter, "limit", limit)
				gvgStats.Status = spdb.Processed
				gvgStats.ObjectCount = objectCount
				log.Infow("updating GVG stats status to processed", "gvgStats", gvgStats)
				err = s.manager.baseApp.GfSpDB().UpdateRecoverGVGStats(gvgStats)
				if err != nil {
					log.Error("failed to update GVG stats to processed status", "gvgStats", gvgStats)
					continue
				}
				break
			}

			exceedLimit := false
		out:
			for _, object := range objects {
				objectInfo := object.Object.ObjectInfo
				objectID := objectInfo.Id.Uint64()
				segmentCount := segmentPieceCount(objectInfo.PayloadSize, maxSegmentSize)
				_, ok := s.currentBatchObjectIDs[objectID]
				if ok {
					log.Infow("the object is in processing", "object_id", objectID, "segment_count", segmentCount)
					continue
				}

				log.Infow("starting to recover object", "object_id", objectID, "segment_count", segmentCount)

				curRecoveryTaskNum := s.manager.recoveryQueue.Len()
				if int(segmentCount) >= recoveryCompacity {
					objectCount++
					o := &spdb.RecoverFailedObject{
						ObjectID:        objectID,
						VirtualGroupID:  object.Gvg.Id,
						RedundancyIndex: gvgStats.RedundancyIndex,
					}
					err = s.manager.baseApp.GfSpDB().InsertRecoverFailedObject(o)
					if err != nil {
						log.Errorw("failed to InsertRecoverFailedObject", "error", err)
						break
					}
					log.Infow("inserted recover failed object record to DB", "object_id", o.ObjectID)
					continue
				}

				if curRecoveryTaskNum+int(segmentCount) >= recoveryCompacity {
					log.Errorw("exceeding recovery limit", "cur_task_num", curRecoveryTaskNum, "segment_num", segmentCount, "max", recoveryCompacity)
					exceedLimit = true
					break
				}

				for segmentIdx := uint32(0); segmentIdx < segmentCount; segmentIdx++ {
					task := &gfsptask.GfSpRecoverPieceTask{}
					task.InitRecoverPieceTask(objectInfo, storageParams, coretask.DefaultSmallerPriority, segmentIdx, s.redundancyIndex, maxSegmentSize, MaxRecoveryTime, maxRecoveryRetry)
					task.SetBySuccessorSP(true)
					task.SetGVGID(s.gvgID)
					err = s.manager.recoveryQueue.Push(task)
					if err != nil {
						log.Errorw("failed to push to recovery queue", "object_id", objectInfo.Id, "segmentIdx", segmentIdx, "error", err)
						if errors.Is(err, ErrRepeatedTask) {
							continue
						}
						if errors.Is(err, gfsptqueue.ErrTaskQueueExceed) {
							exceedLimit = true
							break out
						}
					}
				}
				if !s.manager.recoverObjectStats.has(objectID) {
					s.manager.recoverObjectStats.put(objectID, segmentCount)
				}
				s.currentBatchObjectIDs[objectID] = struct{}{}
			}

			// if exceed the queue limit, wait for a while
			if exceedLimit {
				continue
			}

			// once monitoring all objects related recovered piece tasks got response from executor, regardless success or failed,
			// the scheduler will update the StartAfter in recover gvg stats and jump to the next batch of objects to recover
			s.monitorBatch()
		}
	}
}

func (s *RecoverGVGScheduler) monitorBatch() {
	ticker := time.NewTicker(10 * time.Second)
	// todo add a timeout if can get all objects recovered
	for range ticker.C {
		log.Infow("monitoring for current batch objects", "object_ids", s.currentBatchObjectIDs)
		processed := true
		for objectID, _ := range s.currentBatchObjectIDs {
			if !s.manager.recoverObjectStats.isObjectProcessed(objectID) {
				processed = false
				break
			}
		}
		if !processed {
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
			log.Errorw("failed to update recover gvg status")
			continue
		}
		for objectID, _ := range s.currentBatchObjectIDs {
			log.Debugw("removing object stats", "object_id", objectID)
			s.manager.recoverObjectStats.remove(objectID)
		}
		s.currentBatchObjectIDs = make(map[uint64]struct{})
		return
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

	ticker := time.NewTicker(10 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-s.manager.verifyTerminationSignal:
			log.Infow("EEExit", "recover failed", "exit")
			return
		case <-ticker.C:
			recoverFailedObjects, err := s.manager.baseApp.GfSpDB().GetRecoverFailedObjects(3, 3)
			if err != nil {
				log.Errorw("failed to get recover failed object from DB")
				continue
			}
			if len(recoverFailedObjects) == 0 {
				log.Debug("recover failed object not found")
				continue
			}
			log.Debugw("retrieved recover failed object", "recoverFailedObjects", recoverFailedObjects)
		out:
			for _, o := range recoverFailedObjects {
				objectInfo, err := s.manager.baseApp.GfSpClient().GetObjectByID(context.Background(), o.ObjectID)
				if err != nil {
					log.Errorw("failed to get object info", "object_id", o.ObjectID, "err", err)
					continue
				}
				segmentCount := segmentPieceCount(objectInfo.PayloadSize, maxSegmentSize)

				verified, err := verifyIntegrityAndPieceHash(s.manager, objectInfo, o.RedundancyIndex, maxSegmentSize)
				if err != nil {
					log.Errorw("failed to verify Integrity for object", "object_id", o.ObjectID, "err", err)
					continue
				}
				if verified {
					log.Infow("object has been recovered", "object", objectInfo)
					err = s.manager.baseApp.GfSpDB().DeleteRecoverFailedObject(o.ObjectID)
					if err != nil {
						log.Errorw("failed to delete recover failed object entry", "object_id", o.ObjectID)
						continue
					}
					continue
				}

				for segmentIdx := uint32(0); segmentIdx < segmentCount; segmentIdx++ {
					_, err := s.manager.baseApp.GfSpDB().GetReplicatePieceChecksum(objectInfo.Id.Uint64(), segmentIdx, o.RedundancyIndex)
					if err == nil {
						log.Infow("piece is already recovered,", "object_id", objectInfo.Id, "segmentIdx", segmentIdx, "error", err)
						continue
					}
					if err != gorm.ErrRecordNotFound {
						log.Infow("failed to get piece hash fro DB", "object_id", objectInfo.Id, "segmentIdx", segmentIdx, "error", err)
						break out
					}
					task := &gfsptask.GfSpRecoverPieceTask{}
					task.InitRecoverPieceTask(objectInfo, storageParams, coretask.DefaultSmallerPriority, segmentIdx, o.RedundancyIndex, maxSegmentSize, MaxRecoveryTime, maxRecoveryRetry)
					task.SetBySuccessorSP(true)
					task.SetGVGID(o.VirtualGroupID)
					err = s.manager.recoveryQueue.Push(task)
					if err != nil {
						log.Errorw("failed to push to recovery queue", "object_id", objectInfo.Id, "segmentIdx", segmentIdx, "error", err)
						if errors.Is(err, ErrRepeatedTask) {
							continue
						}
						if errors.Is(err, gfsptqueue.ErrTaskQueueExceed) {
							break out
						}
					}
					log.Errorw("pushed piece to recover queue", "object_id", objectInfo.Id, "segmentIdx", segmentIdx)
				}
				o.RetryTime++
				err = s.manager.baseApp.GfSpDB().UpdateRecoverFailedObject(o)
				if err != nil {
					log.Errorw("failed to update the recover failed object", "object_id", objectInfo.Id)
					break
				}
			}
		}
	}

}

// VerifyGVGScheduler Verify that objects in GVG are recovered successfully or not.
//
// verifying the object existence by querying integrate and piece_hash.
// a recover GVG unit is marked as completed from Processed only when all objects pass the verification.
type VerifyGVGScheduler struct {
	manager              *ManageModular
	vgfID                uint32
	gvgID                uint32
	redundancyIndex      int32
	curStartAfter        uint64
	verifyFailedObjects  map[uint64]struct{}
	verifySuccessObjects map[uint64]struct{} // cache
}

func NewVerifyGVGScheduler(m *ManageModular, vgfID, gvgID uint32, redundancyIndex int32) (*VerifyGVGScheduler, error) {
	return &VerifyGVGScheduler{
		manager:              m,
		vgfID:                vgfID,
		gvgID:                gvgID,
		redundancyIndex:      redundancyIndex,
		verifyFailedObjects:  make(map[uint64]struct{}),
		verifySuccessObjects: make(map[uint64]struct{}),
	}, nil
}

func (s *VerifyGVGScheduler) Start() {
	storageParams, err := s.manager.baseApp.Consensus().QueryStorageParams(context.Background())
	if err != nil {
		return
	}
	maxSegmentSize := storageParams.GetMaxSegmentSize()
	verifyTicker := time.NewTicker(10 * time.Second)
	defer func() {
		// todo
		s.manager.verifyTerminationSignal <- true
		log.Infow("finished verify GVG")
		verifyTicker.Stop()
	}()

	for {
		select {
		case <-verifyTicker.C:
			log.Infow("verify gvg scheduler")
			gvgStats, err := s.manager.baseApp.GfSpDB().GetRecoverGVGStats(s.gvgID)
			if err != nil {
				log.Errorw("failed to get recover gvg stats", "err", err)
				continue
			}
			if gvgStats.Status != spdb.Processed {
				log.Infow("Wait for all objects in GVG have been processed")
				continue
			}

			objects, err := s.manager.baseApp.GfSpClient().ListObjectsInGVG(context.Background(), gvgStats.VirtualGroupID, s.curStartAfter, verifyGVGQueryLimit)
			if err != nil {
				log.Errorw("failed to list object in GVG", "err", err)
				continue
			}

			log.Infow("list objects in GVG", "start_after", s.curStartAfter, "limit", verifyGVGQueryLimit)

			// Once iterate all objects in GVG, check all recorded recover-failed objects have been recovered.
			// If there is any object that does not exceed the max retry, will re-start from the beginning object in GVG.
			if len(objects) == 0 {
				recoverFailedObjectsCount := len(s.verifyFailedObjects)
				needDiscontinueCount := 0
				// check if verified failed object is enqueued
				for objectID, _ := range s.verifyFailedObjects {
					recoverFailedObject, err := s.manager.baseApp.GfSpDB().GetRecoverFailedObject(objectID)
					if err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							log.Infow("the object is already recovered", "object_id", objectID)
							delete(s.verifyFailedObjects, objectID)
							recoverFailedObjectsCount--
							continue
						}
						log.Errorw("failed to get recover failed object", "object_id", objectID, "err", err)
						continue
					}
					if recoverFailedObject.RetryTime < maxRecoveryRetry {
						log.Errorw("object has not been recovered yet", "object_id", objectID, "retry", recoverFailedObject.RetryTime, "max_recovery_retry", maxRecoveryRetry)
						s.curStartAfter = 0
						break
					} else {
						// if an object exceeds the max recover retry, will not process further, the SP should manually trigger the discontinue object tx.
						objectMeta, err := s.manager.baseApp.GfSpClient().GetObjectByID(context.Background(), objectID)
						if err != nil {
							continue
						}

						// confirm the object is not recovered.
						verified, err := verifyIntegrityAndPieceHash(s.manager, objectMeta, s.redundancyIndex, maxSegmentSize)
						if err != nil {
							log.Errorw("failed to verify integrity and piece hash", "object", objectMeta, "error", err)
							continue
						}
						if !verified {
							needDiscontinueCount++
						}
					}
				}
				if recoverFailedObjectsCount == 0 {
					msgCompleteSwapIn := &types2.MsgCompleteSwapIn{
						GlobalVirtualGroupFamilyId: s.vgfID,
					}
					txHash, err := s.manager.baseApp.GfSpClient().CompleteSwapIn(context.Background(), msgCompleteSwapIn)
					if err != nil {
						log.Errorw("failed to send complete swap in", "complete_swap_in_msg", msgCompleteSwapIn, "error", err)
						continue
					}
					log.Infow("succeed to complete swap in tx", "complete_swap_in_msg", msgCompleteSwapIn, "tx_hash", txHash)

					time.Sleep(3 * time.Second)

					if s.vgfID != 0 {
						_, err = s.manager.baseApp.Consensus().QuerySwapInInfo(context.Background(), s.vgfID, 0)
					} else {
						_, err = s.manager.baseApp.Consensus().QuerySwapInInfo(context.Background(), 0, s.gvgID)
					}
					log.Debugw("query swapIn info", "vgf_id", s.vgfID, "gvg_id", s.gvgID, "error", err)
					if err == nil {
						continue
					}
					if strings.Contains(err.Error(), "swap in info not exist") {
						gvgStats.Status = spdb.Completed
						err = s.manager.baseApp.GfSpDB().UpdateRecoverGVGStats(gvgStats)
						if err != nil {
							log.Error("failed to update GVG stats to complete status", "gvgStats", gvgStats)
							continue
						}
						return
					}
					continue
				} else if recoverFailedObjectsCount == needDiscontinueCount {
					log.Error("remaining objects need to be discontinue", "objects_count", needDiscontinueCount)
					return
				}
			}

			for _, o := range objects {
				objectInfo := o.Object.ObjectInfo
				objectID := objectInfo.Id.Uint64()
				_, ok := s.verifySuccessObjects[objectID]
				if ok {
					log.Debugw("the object has been verified previously", "object", objectID)
					continue
				}
				verified, err := verifyIntegrityAndPieceHash(s.manager, objectInfo, s.redundancyIndex, maxSegmentSize)
				if err != nil {
					log.Errorw("failed to verify integrity and piece hash", "object", objectInfo, "error", err)
					break
				}
				if !verified {
					s.verifyFailedObjects[objectID] = struct{}{}

					log.Errorw("verified object has not been recovered yet", "object", objectInfo)
					failedObject := &spdb.RecoverFailedObject{
						ObjectID:        objectID,
						VirtualGroupID:  s.gvgID,
						RedundancyIndex: s.redundancyIndex,
					}
					err = s.manager.baseApp.GfSpDB().InsertRecoverFailedObject(failedObject)
					if err != nil {
						break
					}
					continue
				}
				s.verifySuccessObjects[o.Object.ObjectInfo.Id.Uint64()] = struct{}{}
			}

			s.curStartAfter = s.curStartAfter + uint64(verifyGVGQueryLimit)
		}
	}
}

func verifyIntegrityAndPieceHash(m *ManageModular, object *types.ObjectInfo, redundancyIndex int32, maxSegmentSize uint64) (bool, error) {
	_, err := m.baseApp.GfSpDB().GetObjectIntegrity(object.Id.Uint64(), redundancyIndex)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Errorw("failed to verify the integrity, record not exist", "object_id", object.Id)
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
		log.Errorw("failed to verify the piece hash, count mismacth", "expect", segmentCount, "actual", len(pieceChecksums))
		return false, nil
	}
	return true, nil
}

func segmentPieceCount(payloadSize uint64, maxSegmentSize uint64) uint32 {
	count := payloadSize / maxSegmentSize
	if payloadSize%maxSegmentSize > 0 {
		count++
	}
	return uint32(count)
}
