package manager

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfsptqueue"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretask "github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

const (
	recoverBatchSize = 50
	pauseInterval    = 10 * time.Second
	maxRecoveryRetry = 3
	MaxRecoveryTime  = 50
)

type ObjectPieceStats struct {
	Processed         bool
	TotalPiece        int
	FailedPieceCount  vgmgr.IDSet
	SuccessPieceCount vgmgr.IDSet
}

type RecoverScheduler struct {
	manager *ManageModular

	recoverBatchSize uint64 // every time fetch recoverBatchSize of objects in the GVG, and enqueue the recover queue

	mtx                   sync.RWMutex
	currentBatchObjectIDs map[uint64]struct{} // record ids of objects that been processed by reported task, either success or failed.

	vgfID              uint32
	gvgID              uint32
	isSuccessorPrimary bool
}

func NewRecoverGVGScheduler(m *ManageModular, vgfID, gvgID uint32) *RecoverScheduler {
	return &RecoverScheduler{
		manager:               m,
		recoverBatchSize:      recoverBatchSize,
		currentBatchObjectIDs: make(map[uint64]struct{}, 0),
		vgfID:                 vgfID, // recover a vgf
		gvgID:                 gvgID, // recover a specific gvg as successor secodanry SP
	}
}

func (s *RecoverScheduler) Start() {

	for {
		time.Sleep(100 * time.Second)
		if s.vgfID != 0 {
			gvgsStats, err := s.manager.baseApp.GfSpDB().GetRecoverGVGStatsByFamilyIDAndStatus(s.vgfID, spdb.NotDone)
			if err != nil {
				continue
			}
			if len(gvgsStats) == 0 {
				log.Infow("All GVGs have been processed in this family")
				return
			}
			gvgStats := gvgsStats[0]

			startAfterObjectID := gvgStats.StartAfterObjectID
			limit := gvgStats.Limit

			objects, err := s.manager.baseApp.GfSpClient().ListObjectsInGVG(context.Background(), gvgStats.VirtualGroupID, startAfterObjectID, uint32(limit))
			if err != nil {
				log.Errorw("failed to list objects in gvg", "start_after_object_id", startAfterObjectID, "limit", limit)
				continue
			}

			if len(objects) == 0 {
				log.Infow("all objects in gvg have been processed", "start_after_object_id", startAfterObjectID, "limit", limit)
				// TODO check DB if there is any failed object records
				gvgStats.Status = int(spdb.Processed)
				err = s.manager.baseApp.GfSpDB().UpdateRecoverGVGStats(gvgStats)
				if err != nil {
					log.Error("failed to ", "start_after_object_id", startAfterObjectID, "limit", limit)
					continue
				}

			}

			storageParams, err := s.manager.baseApp.Consensus().QueryStorageParams(context.Background())
			if err != nil {
				return
			}
			maxSegmentSize := storageParams.GetMaxSegmentSize()

			for _, object := range objects {
				objectInfo := object.Object.ObjectInfo
				segmentCount := segmentPieceCount(objectInfo.PayloadSize, maxSegmentSize)

				for segmentIdx := uint32(0); segmentIdx < segmentCount; segmentIdx++ {
					task := &gfsptask.GfSpRecoverPieceTask{}
					task.InitRecoverPieceTask(objectInfo, storageParams, coretask.DefaultSmallerPriority, segmentIdx, int32(-1), maxSegmentSize, MaxRecoveryTime, maxRecoveryRetry)
					task.SetBySuccessorSP(true)

					err = s.manager.recoveryQueue.Push(task)
					if err != nil {
						if errors.Is(err, ErrRepeatedTask) {
							break
						}
						if errors.Is(err, gfsptqueue.ErrTaskQueueExceed) {
							time.Sleep(pauseInterval)
						}
						log.Errorw("failed to push replicate piece task to queue", "object_info", objectInfo, "error", err)
						break
					}
				}
			}
			// todo
			s.currentBatchObjectIDs = append(s.currentBatchObjectIDs, objectID)
		}
	}
}

// MonitorBatch and update the next start batch
func (s *RecoverScheduler) MonitorBatch() {
	ticker := time.NewTicker(20 * time.Second)
	for range ticker.C {
		s.mtx.RLock()
		if uint64(len(s.currentBatchObjectIDs)) == s.recoverBatchSize {
			for id, _ := range s.currentBatchObjectIDs {
				objectStats, ok := s.manager.recoverObjectStats[id]
				if !ok {
					continue
				}
			}
			// if all objects in this batch are proceesd, we can update startAfterObjectID and start next batch.
			if objectStats.Processed {
				gvgStats, err := s.manager.baseApp.GfSpDB().GetRecoverGVGStats(s.gvgID)
				if err != nil {
					continue
				}
				gvgStats.StartAfterObjectID = gvgStats.StartAfterObjectID + gvgStats.Limit
				err = s.manager.baseApp.GfSpDB().UpdateRecoverGVGStats(gvgStats)
				if err != nil {
					continue
				}
			}

		}
		s.mtx.RUnlock()
	}
}

func segmentPieceCount(payloadSize uint64, maxSegmentSize uint64) uint32 {
	count := payloadSize / maxSegmentSize
	if payloadSize%maxSegmentSize > 0 {
		count++
	}
	return uint32(count)
}
