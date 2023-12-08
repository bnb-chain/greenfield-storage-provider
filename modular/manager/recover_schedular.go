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
	GvGID             uint32
	Processed         bool
	TotalPiece        int
	FailedPieceCount  vgmgr.IDSet
	SuccessPieceCount vgmgr.IDSet
}

type RecoverScheduler struct {
	manager *ManageModular

	recoverBatchSize uint64 // every time fetch recoverBatchSize of objects in the GVG, and enqueue the recover queue

	mtx                   sync.RWMutex
	currentBatchObjectIDs chan uint64 // record ids of objects that been processed by reported task, either success or failed.

	vgfID              uint32
	gvgIDs             []uint32
	targetSPId         uint32
	isSuccessorPrimary bool
}

func NewRecoverGVGScheduler(m *ManageModular, vgfID, targetSPId uint32) *RecoverScheduler {
	r := &RecoverScheduler{
		manager:               m,
		recoverBatchSize:      recoverBatchSize,
		currentBatchObjectIDs: make(chan uint64, 20),
		vgfID:                 vgfID, // recover a vgf
		gvgIDs:                make([]uint32, 0),
		targetSPId:            targetSPId,
	}

	return r
}

func (s *RecoverScheduler) Init(gvgID uint32) error {
	if gvgID != 0 {
		s.gvgIDs = append(s.gvgIDs, gvgID)
		return s.manager.baseApp.GfSpDB().SetRecoverGVGStats([]*spdb.RecoverGVGStats{
			{
				VirtualGroupFamilyID: s.vgfID,
				VirtualGroupID:       gvgID,
				ExitingSPID:          s.targetSPId,
				StartAfter:           0,
				Limit:                10,
			},
		})
	}

	if s.vgfID != 0 {
		resp, err := s.manager.baseApp.GfSpClient().GetVirtualGroupFamily(context.Background(), s.vgfID)
		if err != nil {
			log.Errorw("failed to GetVirtualGroupFamily", "error", err)
			return err
		}
		recoveryUnit := make([]*spdb.RecoverGVGStats, 0, len(resp.GetGlobalVirtualGroupIds()))
		for _, id := range resp.GetGlobalVirtualGroupIds() {
			recoveryUnit = append(recoveryUnit, &spdb.RecoverGVGStats{
				VirtualGroupFamilyID: s.vgfID,
				VirtualGroupID:       id,
				ExitingSPID:          s.targetSPId,
				StartAfter:           0,
				Limit:                10,
			})
			s.gvgIDs = append(s.gvgIDs, id)
		}
		return s.manager.baseApp.GfSpDB().SetRecoverGVGStats(recoveryUnit)
	}

	return errors.New("param errors")

}

func (s *RecoverScheduler) Start() {
	for _, gvgID := range s.gvgIDs {
		gvgStats, err := s.manager.baseApp.GfSpDB().GetRecoverGVGStats(gvgID)
		if err != nil {
			continue
		}
		if gvgStats.Status != int(spdb.NotDone) {
			continue
		}
		recoverTicker := time.NewTicker(100 * time.Second)
		defer recoverTicker.Stop()
		startAfter := gvgStats.StartAfter
		limit := gvgStats.Limit
		for {
			select {
			case <-recoverTicker.C:
				objects, err := s.manager.baseApp.GfSpClient().ListObjectsInGVG(context.Background(), gvgStats.VirtualGroupID, startAfter, uint32(limit))
				if err != nil {
					log.Errorw("failed to list objects in gvg", "start_after_object_id", startAfter, "limit", limit)
					continue
				}

				if len(objects) == 0 {
					log.Infow("all objects in gvg have been processed", "start_after_object_id", startAfter, "limit", limit)
					// TODO check DB if there is any failed object records
					gvgStats.Status = int(spdb.Processed)
					err = s.manager.baseApp.GfSpDB().UpdateRecoverGVGStats(gvgStats)
					if err != nil {
						log.Error("failed to ", "start_after_object_id", startAfter, "limit", limit)
						return
					}
					break
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
					// todo
					s.currentBatchObjectIDs <- object.GetObject().GetObjectInfo().Id.Uint64()
				}

				startAfter += limit
			}
		}
	}
}

// MonitorBatch and update the next start batch
func (s *RecoverScheduler) MonitorBatch() {
	count := 0
	for {
		select {
		case id := <-s.currentBatchObjectIDs:
			tick := time.NewTicker(10 * time.Millisecond)
			var objectStats *ObjectPieceStats
			var ok bool
			for range tick.C {
				s.mtx.RLock()
				objectStats, ok = s.manager.recoverObjectStats[id]
				s.mtx.Unlock()
				if ok {
					break
				}
			}
			if objectStats.Processed {
				count++
			}
			if count == 10 {
				gvgStats, err := s.manager.baseApp.GfSpDB().GetRecoverGVGStats(objectStats.GvGID)
				if err != nil {
					continue
				}
				gvgStats.StartAfter += gvgStats.Limit
				err = s.manager.baseApp.GfSpDB().UpdateRecoverGVGStats(gvgStats)
				if err != nil {
					continue
				}
			}
		default:
			time.Sleep(20 * time.Millisecond)
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
