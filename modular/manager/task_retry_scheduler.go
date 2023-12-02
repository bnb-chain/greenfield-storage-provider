package manager

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfsptqueue"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/prysmaticlabs/prysm/crypto/bls"
)

const (
	// if the failed task in threshold to retry this task, or reject this task.
	defaultRejectUnsealThresholdSecond = 1 * 24 * 60 * 60 // last 1 days

	// prefetchLimit is used to prefetch load task from db.
	prefetchLimit = 100

	// backoffIntervalSecond is used to backoff when something is wrong.
	backoffIntervalSecond = 3 * time.Second

	// retryIntervalSecond is used to control rate limit which avoid too big pressure
	// on other modules, such as db, chain, or other keypoint workflows.
	retryIntervalSecond = 1 * time.Second
)

type RetryTaskType int32

const (
	retryReplicate    RetryTaskType = 0
	retrySeal         RetryTaskType = 1
	retryRejectUnseal RetryTaskType = 2
)

// TaskRetryScheduler is used to schedule background task retry.
type TaskRetryScheduler struct {
	manager                     *ManageModular
	rejectUnsealThresholdSecond uint64
}

// NewTaskRetryScheduler returns a task retry scheduler instance.
func NewTaskRetryScheduler(m *ManageModular) *TaskRetryScheduler {
	rejectUnsealThresholdSecond := m.rejectUnsealThresholdSecond
	if rejectUnsealThresholdSecond == 0 {
		rejectUnsealThresholdSecond = defaultRejectUnsealThresholdSecond
	}
	return &TaskRetryScheduler{
		manager:                     m,
		rejectUnsealThresholdSecond: rejectUnsealThresholdSecond,
	}
}

// Start is used to start the task retry scheduler.
func (s *TaskRetryScheduler) Start() {
	go s.startReplicateTaskRetry()
	go s.startSealTaskRetry()
	go s.startRejectUnsealTaskRetry()
	log.Info("task retry scheduler startup")
}

func (s *TaskRetryScheduler) startReplicateTaskRetry() {
	var (
		iter                   *TaskIterator
		err                    error
		loopNumber             uint64
		currentLoopRetryNumber uint64
		totalRetryNumber       uint64
	)

	for {
		time.Sleep(retryIntervalSecond * 10)
		iter = NewTaskIterator(s.manager.baseApp.GfSpDB(), retryReplicate, int64(s.rejectUnsealThresholdSecond))
		log.Infow("start a new loop to retry replicate", "iterator", iter,
			"loop_number", loopNumber, "total_retry_number", totalRetryNumber)

		for iter.Valid() {
			time.Sleep(retryIntervalSecond)
			// ignore retry error, this task can be retried in next loop.
			err = s.retryReplicateTask(iter.Value())
			currentLoopRetryNumber++
			totalRetryNumber++
			log.Infow("retry replicate task",
				"task", iter.Value(), "loop_number", loopNumber,
				"current_loop_retry_number", currentLoopRetryNumber,
				"total_retry_number", totalRetryNumber,
				"error", err)
			iter.Next()
		}
		loopNumber++
		currentLoopRetryNumber = 0
	}
}

func (s *TaskRetryScheduler) startSealTaskRetry() {
	var (
		iter                   *TaskIterator
		err                    error
		loopNumber             uint64
		currentLoopRetryNumber uint64
		totalRetryNumber       uint64
	)

	for {
		time.Sleep(retryIntervalSecond * 10)
		iter = NewTaskIterator(s.manager.baseApp.GfSpDB(), retrySeal, int64(s.rejectUnsealThresholdSecond))
		log.Infow("start a new loop to retry seal", "iterator", iter,
			"loop_number", loopNumber, "total_retry_number", totalRetryNumber)

		for iter.Valid() {
			time.Sleep(retryIntervalSecond)
			// ignore retry error, this task can be retried in next loop.
			err = s.retrySealTask(iter.Value())
			currentLoopRetryNumber++
			totalRetryNumber++
			log.Infow("retry seal task",
				"task", iter.Value(), "loop_number", loopNumber,
				"current_loop_retry_number", currentLoopRetryNumber,
				"total_retry_number", totalRetryNumber,
				"error", err)
			iter.Next()
		}
		loopNumber++
		currentLoopRetryNumber = 0
	}
}

func (s *TaskRetryScheduler) startRejectUnsealTaskRetry() {
	var (
		iter                   *TaskIterator
		err                    error
		loopNumber             uint64
		currentLoopRetryNumber uint64
		totalRetryNumber       uint64
	)

	for {
		time.Sleep(retryIntervalSecond * 10)
		iter = NewTaskIterator(s.manager.baseApp.GfSpDB(), retryRejectUnseal, int64(s.rejectUnsealThresholdSecond))
		log.Infow("start a new loop to retry reject unseal task", "iterator", iter,
			"loop_number", loopNumber, "total_retry_number", totalRetryNumber)

		for iter.Valid() {
			time.Sleep(retryIntervalSecond)
			// ignore retry error, this task can be retried in next loop.
			err = s.retryRejectUnsealTask(iter.Value())
			currentLoopRetryNumber++
			totalRetryNumber++
			log.Infow("retry reject unseal task",
				"task", iter.Value(), "loop_number", loopNumber,
				"current_loop_retry_number", currentLoopRetryNumber,
				"total_retry_number", totalRetryNumber,
				"error", err)
			iter.Next()
		}
		loopNumber++
		currentLoopRetryNumber = 0
	}
}

func isAlreadyNotFound(err error) bool {
	return strings.Contains(err.Error(), "No such object")
}

// retryReplicateTask is used to push the failed replicate task to task dispatcher,
// and the task will be executed by executor.
func (s *TaskRetryScheduler) retryReplicateTask(t *spdb.UploadObjectMeta) error {
	objectInfo, queryErr := s.manager.baseApp.Consensus().QueryObjectInfoByID(context.Background(), util.Uint64ToString(t.ObjectID))
	if queryErr != nil {
		log.Errorw("failed to query object info", "object_id", t.ObjectID, "error", queryErr)
		if !isAlreadyNotFound(queryErr) {
			time.Sleep(backoffIntervalSecond)
		}
		return queryErr
	}
	if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED {
		log.Infow("object is not in create status", "object_info", objectInfo)
		return fmt.Errorf("object is not in create status")
	}
	storageParams, queryErr := s.manager.baseApp.Consensus().QueryStorageParamsByTimestamp(context.Background(), objectInfo.GetCreateAt())
	if queryErr != nil {
		log.Errorw("failed to query storage param", "object_id", t.ObjectID, "error", queryErr)
		time.Sleep(backoffIntervalSecond)
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
		if errors.Is(pushErr, gfsptqueue.ErrTaskQueueExceed) {
			time.Sleep(backoffIntervalSecond)
		}
		log.Errorw("failed to push replicate piece task to queue", "object_info", objectInfo, "error", pushErr)
		return pushErr
	}
	return nil
}

// retrySealTask is used to send seal tx to chain.
// This task is very lightweight and therefore executed directly inside the scheduler.
func (s *TaskRetryScheduler) retrySealTask(t *spdb.UploadObjectMeta) error {
	objectInfo, queryErr := s.manager.baseApp.Consensus().QueryObjectInfoByID(context.Background(), util.Uint64ToString(t.ObjectID))
	if queryErr != nil {
		log.Errorw("failed to query object info", "object_id", t.ObjectID, "error", queryErr)
		if !isAlreadyNotFound(queryErr) {
			time.Sleep(backoffIntervalSecond)
		}
		return queryErr
	}
	if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED {
		log.Infow("object is not in create status", "object_info", objectInfo)
		return fmt.Errorf("object is not in create status")
	}

	blsSig, err := bls.MultipleSignaturesFromBytes(t.SecondarySignatures)
	if err != nil {
		log.Errorw("failed to get multiple signature", "object_id", t.ObjectID, "error", err)
		return err
	}
	sealMsg := &storagetypes.MsgSealObject{
		Operator:                    s.manager.baseApp.OperatorAddress(),
		BucketName:                  objectInfo.GetBucketName(),
		ObjectName:                  objectInfo.GetObjectName(),
		GlobalVirtualGroupId:        t.GlobalVirtualGroupID,
		SecondarySpBlsAggSignatures: bls.AggregateSignatures(blsSig).Marshal(),
	}
	return sendAndConfirmSealObjectTx(s.manager.baseApp, sealMsg)
}

// retryRejectTask is used to send reject unseal tx to chain.
// This task is very lightweight and therefore executed directly inside the scheduler.
func (s *TaskRetryScheduler) retryRejectUnsealTask(t *spdb.UploadObjectMeta) error {
	objectInfo, queryErr := s.manager.baseApp.Consensus().QueryObjectInfoByID(context.Background(), util.Uint64ToString(t.ObjectID))
	if queryErr != nil {
		log.Errorw("failed to query object info", "object_id", t.ObjectID, "error", queryErr)
		if !isAlreadyNotFound(queryErr) {
			time.Sleep(backoffIntervalSecond)
		}
		return queryErr
	}
	if objectInfo.GetObjectStatus() != storagetypes.OBJECT_STATUS_CREATED {
		log.Infow("object is not in create status", "object_info", objectInfo)
		return fmt.Errorf("object is not in create status")
	}

	rejectUnsealMsg := &storagetypes.MsgRejectSealObject{
		Operator:   s.manager.baseApp.OperatorAddress(),
		BucketName: objectInfo.GetBucketName(),
		ObjectName: objectInfo.GetObjectName(),
	}
	err := sendAndConfirmRejectUnsealObjectTx(s.manager.baseApp, rejectUnsealMsg)
	if err == nil {
		_ = s.manager.baseApp.GfSpDB().DeleteUploadProgress(objectInfo.Id.Uint64())
	}
	return err
}

func sendAndConfirmSealObjectTx(baseApp *gfspapp.GfSpBaseApp, msg *storagetypes.MsgSealObject) error {
	return SendAndConfirmTx(baseApp.Consensus(),
		func() (string, error) {
			var (
				txHash string
				txErr  error
			)
			if txHash, txErr = baseApp.GfSpClient().SealObject(context.Background(), msg); txErr != nil && !isAlreadyExists(txErr) {
				log.Errorw("failed to send seal object", "seal_object_msg", msg, "error", txErr)
				return "", txErr
			}
			return txHash, nil
		})
}

func sendAndConfirmRejectUnsealObjectTx(baseApp *gfspapp.GfSpBaseApp, msg *storagetypes.MsgRejectSealObject) error {
	return SendAndConfirmTx(baseApp.Consensus(),
		func() (string, error) {
			var (
				txHash string
				txErr  error
			)
			if txHash, txErr = baseApp.GfSpClient().RejectUnSealObject(context.Background(), msg); txErr != nil && !isAlreadyExists(txErr) {
				log.Errorw("failed to send seal object", "reject_unseal_object_msg", msg, "error", txErr)
				return "", txErr
			}
			return txHash, nil
		})
}

type PrefetchFunc func(iter *TaskIterator) ([]*spdb.UploadObjectMeta, error)

// TaskIterator is used to load retry task from db.
type TaskIterator struct {
	dbReader             spdb.SPDB
	taskType             RetryTaskType
	startTimeStampSecond int64
	endTimeStampSecond   int64
	prefetchFunc         PrefetchFunc
	cachedValueList      []*spdb.UploadObjectMeta
	currentIndex         int
}

func NewTaskIterator(db spdb.SPDB, taskType RetryTaskType, rejectUnsealThresholdSecond int64) *TaskIterator {
	var (
		startTS      int64
		endTS        int64
		prefetchFunc PrefetchFunc
	)
	switch taskType {
	case retryReplicate:
		startTS = sqldb.GetCurrentUnixTime() - rejectUnsealThresholdSecond
	case retrySeal:
		startTS = sqldb.GetCurrentUnixTime() - rejectUnsealThresholdSecond
	case retryRejectUnseal:
		startTS = sqldb.GetCurrentUnixTime() - 2*rejectUnsealThresholdSecond
		endTS = sqldb.GetCurrentUnixTime() - rejectUnsealThresholdSecond
	}
	prefetchFunc = func(iter *TaskIterator) ([]*spdb.UploadObjectMeta, error) {
		switch iter.taskType {
		case retryReplicate:
			return iter.dbReader.GetUploadMetasToReplicateByStartTS(prefetchLimit, iter.startTimeStampSecond)
		case retrySeal:
			return iter.dbReader.GetUploadMetasToSealByStartTS(prefetchLimit, iter.startTimeStampSecond)
		case retryRejectUnseal:
			return iter.dbReader.GetUploadMetasToRejectByRangeTS(prefetchLimit, iter.startTimeStampSecond, iter.endTimeStampSecond)
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
