package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

const (
	// if the failed task in threshold to retry this task, or reject this task.
	defaultRejectSealThresholdSecond = 2 * 24 * 3600 // last 2 days
	defaultPrefetchLimit             = 100
	defaultBackoffIntervalSecond     = 3 * time.Second
)

type RetryTaskType int32

const (
	retryReplicate RetryTaskType = 0
	retrySeal      RetryTaskType = 1
	retryReject    RetryTaskType = 2
)

// TaskRetryScheduler is used to schedule background task retry.
type TaskRetryScheduler struct {
	manager *ManageModular
	// failedReplicateTaskIter *FailedReplicateTaskIterator
	// replicateTaskRetryer    *ReplicateTaskRetryer
}

func NewTaskRetryScheduler(m *ManageModular) *TaskRetryScheduler {
	return &TaskRetryScheduler{
		manager: m,
		// failedReplicateTaskIter: nil,
		// replicateTaskRetryer:    nil,
	}
}

// Start is used to start the task retry scheduler.
func (s *TaskRetryScheduler) Start() {
	// TODO: senity

	go s.startReplicateTaskRetry()
	log.Info("task retry scheduler startup")
}

func (s *TaskRetryScheduler) startReplicateTaskRetry() {
	var (
		loopNumber             uint64
		currentLoopRetryNumber uint64
		totalRetryNumber       uint64
	)

	for {
		time.Sleep(defaultBackoffIntervalSecond * 10)
		iter := NewTaskIterator(s.manager.baseApp.GfSpDB(), retryReplicate)
		log.Infow("start a new loop to retry replicate", "iterator", iter,
			"loop_number", loopNumber, "total_retry_number", totalRetryNumber)
		for iter.Valid() {
			if err := s.retryReplicateTask(iter.Value()); err != nil {
				log.Errorw("failed to retry replicate task", "task", iter.Value(), "error", err)
				continue
			}
			currentLoopRetryNumber++
			totalRetryNumber++
			log.Infow("succeed to retry replicate task",
				"task", iter.Value(), "loop_number", loopNumber,
				"current_loop_retry_number", currentLoopRetryNumber,
				"total_retry_number", totalRetryNumber)
			iter.Next()
		}
		loopNumber++
		currentLoopRetryNumber = 0
	}
}

func (s *TaskRetryScheduler) retryReplicateTask(t *spdb.UploadObjectMeta) error {
	// check object status and add replicate task queue
	objectInfo, queryErr := s.manager.baseApp.Consensus().QueryObjectInfoByID(context.Background(), util.Uint64ToString(t.ObjectID))
	if queryErr != nil {
		log.Errorw("failed to query object info", "object_id", t.ObjectID, "error", queryErr)
		time.Sleep(defaultBackoffIntervalSecond)
		return queryErr
	}
	if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED {
		log.Infow("object is not in create status", "object_info", objectInfo)
		return fmt.Errorf("object is not in create status")
	}
	storageParams, queryErr := s.manager.baseApp.Consensus().QueryStorageParamsByTimestamp(context.Background(), objectInfo.GetCreateAt())
	if queryErr != nil {
		log.Errorw("failed to query storage param", "object_id", t.ObjectID, "error", queryErr)
		time.Sleep(defaultBackoffIntervalSecond)
		return queryErr
	}
	replicateTask := &gfsptask.GfSpReplicatePieceTask{}
	replicateTask.InitReplicatePieceTask(objectInfo, storageParams, s.manager.baseApp.TaskPriority(replicateTask),
		s.manager.baseApp.TaskTimeout(replicateTask, objectInfo.GetPayloadSize()), s.manager.baseApp.TaskMaxRetry(replicateTask))
	replicateTask.SetSecondaryAddresses(t.SecondaryEndpoints)
	replicateTask.SetSecondarySignatures(t.SecondarySignatures)
	replicateTask.GlobalVirtualGroupId = t.GlobalVirtualGroupID
	pushErr := s.manager.replicateQueue.Push(replicateTask)
	if pushErr != nil {
		log.Errorw("failed to push replicate piece task to queue", "object_info", objectInfo, "error", pushErr)
		time.Sleep(defaultBackoffIntervalSecond)
		return pushErr
	}
	return nil
}

type PrefetchFunc func(iter *TaskIterator) ([]*spdb.UploadObjectMeta, error)

// TaskIterator is used to load failed replicate task from db.
type TaskIterator struct {
	dbReader             spdb.SPDB
	taskType             RetryTaskType
	startTimeStampSecond int64
	endTimeStampSecond   int64
	prefetchFunc         PrefetchFunc
	cachedValueList      []*spdb.UploadObjectMeta
	currentIndex         int
}

func NewTaskIterator(db spdb.SPDB, taskType RetryTaskType) *TaskIterator {
	var (
		startTS      int64
		endTS        int64
		prefetchFunc PrefetchFunc
	)
	switch taskType {
	case retryReplicate:
		startTS = sqldb.GetCurrentUnixTime() - defaultRejectSealThresholdSecond
	case retrySeal:
		startTS = sqldb.GetCurrentUnixTime() - defaultRejectSealThresholdSecond
	case retryReject:
		startTS = sqldb.GetCurrentUnixTime() - 2*defaultRejectSealThresholdSecond
		endTS = sqldb.GetCurrentUnixTime() - defaultRejectSealThresholdSecond
	}
	prefetchFunc = func(iter *TaskIterator) ([]*spdb.UploadObjectMeta, error) {
		switch iter.taskType {
		case retryReplicate:
			return iter.dbReader.GetUploadMetasToReplicateByStartTS(defaultPrefetchLimit, iter.startTimeStampSecond)
		case retrySeal:
			return iter.dbReader.GetUploadMetasToSealByStartTS(defaultPrefetchLimit, iter.startTimeStampSecond)
		case retryReject:
			return iter.dbReader.GetUploadMetasToRejectByRangeTS(defaultPrefetchLimit, iter.startTimeStampSecond, iter.endTimeStampSecond)
		}
		return nil, nil
	}
	return &TaskIterator{
		dbReader:             db,
		taskType:             taskType,
		startTimeStampSecond: startTS,
		endTimeStampSecond:   endTS,
		currentIndex:         0,
		prefetchFunc:         prefetchFunc,
	}
}

func (iter *TaskIterator) Valid() bool {
	if iter.currentIndex >= len(iter.cachedValueList) {
		var err error
		iter.cachedValueList, err = iter.prefetchFunc(iter)
		if err != nil {
			log.Errorw("failed to get upload metas to replicate by start timestamp", "error", err)
			return false
		}
		if len(iter.cachedValueList) == 0 {
			return false
		}
		iter.currentIndex = 0
		iter.startTimeStampSecond = iter.cachedValueList[len(iter.cachedValueList)-1].UpdateTimeStamp
	}
	return true
}

func (iter *TaskIterator) Next() {
	iter.currentIndex++
}

func (iter *TaskIterator) Value() *spdb.UploadObjectMeta {
	return iter.cachedValueList[iter.currentIndex]
}
