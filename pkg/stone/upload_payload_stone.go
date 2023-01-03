package stone

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/looplab/fsm"

	"github.com/bnb-chain/inscription-storage-provider/pkg/job"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/store/jobdb"
	"github.com/bnb-chain/inscription-storage-provider/store/metadb"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

const (
	// CtxStoneKey defines the key of UploadPayloadStone in context that transfer in fsm
	CtxStoneKey string = "UploadPayloadStone"
)

// StoneJob defines the interface of job that transfer to StoneHub
type StoneJob interface {
}

// UploadPayloadStone maintains the upload payload job and fsm
type UploadPayloadStone struct {
	jobCtx *JobContextWrapper     // job context, goroutine safe
	objCtx *job.ObjectInfoContext // the object info of payload data
	jobFsm *fsm.FSM               // fsm of upload payload
	job    *job.UploadPayloadJob  // records the upload payload job information
	jobCh  chan StoneJob          // the channel of transfer job to StoneHub
	gcCh   chan string            // the channel of notify StoneHub to delete stone
}

// NewUploadPayloadStone return the instance of UploadPayloadStone
func NewUploadPayloadStone(jobContext *types.JobContext, object *types.ObjectInfo, jobDB jobdb.JobDB,
	metaDB metadb.MetaDB, jobCh chan StoneJob, gcCh chan string) (*UploadPayloadStone, error) {
	if jobContext == nil || object == nil || jobCh == nil || gcCh == nil {
		return nil, errors.New("new upload payload stone params error")
	}
	objectCtx := job.NewObjectInfoContext(object, jobDB, metaDB)
	uploadJob, err := job.NewUploadPayloadJob(objectCtx)
	if err != nil {
		return nil, err
	}
	jobCtx := NewJobContextWrapper(jobContext, jobDB)
	state, err := repairState(jobCtx, uploadJob)
	if err != nil {
		log.Error("repair upload payload state error", "error", err)
		return nil, err
	}
	stone := &UploadPayloadStone{
		jobCtx: jobCtx,
		objCtx: objectCtx,
		job:    uploadJob,
		jobFsm: fsm.NewFSM(state, UploadPayloadFsmEvent, UploadPayLoadFsmCallBack),
		jobCh:  jobCh,
		gcCh:   gcCh,
	}
	if err := stone.selfActionEvent(context.Background()); err != nil {
		log.Error("self action stone fsm error", "error", err)
		return nil, err
	}
	return stone, nil
}

// repairState recover the job state according to job completion
func repairState(jobCtx *JobContextWrapper, job *job.UploadPayloadJob) (string, error) {
	state, err := jobCtx.GetJobState()
	if err != nil {
		return state, err
	}
	if state == types.JOB_STATE_SEAL_OBJECT_DONE {
		return state, errors.New("upload payload job has been successfully completed")
	}
	state = types.JOB_STATE_CREATE_OBJECT_DONE
	if job.SecondarySPCompleted() {
		state = types.JOB_STATE_UPLOAD_SECONDARY_DONE
	} else if job.PrimarySPCompleted() {
		state = types.JOB_STATE_UPLOAD_PRIMARY_DONE
	}
	if err := jobCtx.SetJobErr(nil); err != nil {
		return state, err
	}
	if err := jobCtx.SetJobState(state); err != nil {
		return state, err
	}
	return state, nil
}

// selfActionEvent self-propelled fsm execution
func (stone *UploadPayloadStone) selfActionEvent(ctx context.Context, args ...interface{}) error {
	actionFsm := func(ctx context.Context, event string, args ...interface{}) {
		ctx = context.WithValue(ctx, CtxStoneKey, stone)
		if err := stone.jobFsm.Event(ctx, event, args...); err != nil {
			log.Warn("action stone fsm error", "error", err)
		}
	}
	if stone.jobCtx.JobErr() != nil {
		actionFsm(ctx, InterruptEvent)
		return stone.jobCtx.JobErr()
	}
	var current string
	var txHash string
	var event string
	if len(stone.objCtx.TxHash()) > 6 {
		txHash = hex.EncodeToString(stone.objCtx.TxHash())[:6]
	}
	for {
		current = stone.jobFsm.Current()
		switch current {
		case types.JOB_STATE_CREATE_OBJECT_DONE:
			event = UploadPayloadInitEvent
		case types.JOB_STATE_UPLOAD_PRIMARY_INIT:
			event = UploadPrimaryDoingEvent
		case types.JOB_STATE_UPLOAD_PRIMARY_DOING:
			if stone.job.PrimarySPCompleted() {
				event = UploadPrimaryDoneEvent
			} else {
				return nil
			}
		case types.JOB_STATE_UPLOAD_PRIMARY_DONE:
			event = UploadSecondaryInitEvent
		case types.JOB_STATE_UPLOAD_SECONDARY_INIT:
			event = UploadSecondaryDoingEvent
		case types.JOB_STATE_UPLOAD_SECONDARY_DOING:
			if stone.job.SecondarySPCompleted() {
				event = UploadSecondaryDoneEvent
			} else {
				return nil
			}
		case types.JOB_STATE_UPLOAD_SECONDARY_DONE:
			event = SealObjectInitEvent
		case types.JOB_STATE_SEAL_OBJECT_INIT:
			event = SealObjectDoingEvent
		default:
			log.Info("stone fsm self action stop, ", "tx_hash", txHash, " state", current)
			return nil
		}
		log.Info("stone fsm self action, ", "tx_hash: ", txHash, " state: ", current, " event: ", event)
		actionFsm(ctx, event)
		if stone.jobCtx.JobErr() != nil {
			actionFsm(ctx, InterruptEvent)
			return stone.jobCtx.JobErr()
		}
	}
	return nil
}

// ActionEvent receive the event and propelled fsm execution
func (stone *UploadPayloadStone) ActionEvent(ctx context.Context, event string, args ...interface{}) error {
	if stone.jobCtx.JobErr() != nil || stone.jobFsm.Current() == types.JOB_STATE_ERROR {
		// log error
		return stone.jobCtx.JobErr()
	}
	actionFsm := func(ctx context.Context, event string, args ...interface{}) {
		ctx = context.WithValue(ctx, CtxStoneKey, stone)
		if err := stone.jobFsm.Event(ctx, event, args...); err != nil {
			log.Warn("action stone fsm error", "error", err)
		}
	}
	actionFsm(ctx, event, args...)
	return stone.selfActionEvent(ctx, args...)
}

// InterruptStone interrupt the fsm and stop the stone
func (stone *UploadPayloadStone) InterruptStone(ctx context.Context, err error) error {
	if stone.jobCtx.JobErr() != nil || stone.jobFsm.Current() == types.JOB_STATE_ERROR {
		// log error
		return stone.jobCtx.JobErr()
	}
	log.Error("interrupt stone", "hash", stone.StoneKey(), "error", err)
	stone.jobCtx.SetJobErr(err)
	ctx = context.WithValue(ctx, CtxStoneKey, stone)
	stone.jobFsm.Event(ctx, InterruptEvent)
	return nil
}

// PrimarySPJobDone return whether upload primary storage provider completed
func (stone *UploadPayloadStone) PrimarySPJobDone() bool {
	return stone.job.PrimarySPCompleted()
}

// PopPendingPrimarySPJob return the uncompleted upload primary storage provider job
func (stone *UploadPayloadStone) PopPendingPrimarySPJob() *service.PieceJob {
	return stone.job.PopPendingPrimarySPJob()
}

// PopPendingSecondarySPJob return the uncompleted upload secondary storage provider job
func (stone *UploadPayloadStone) PopPendingSecondarySPJob() *service.PieceJob {
	return stone.job.PopPendingSecondarySPJob()
}

// LastModifyTime return the last modify job state time
func (stone *UploadPayloadStone) LastModifyTime() int64 {
	return stone.jobCtx.ModifyTime()
}

// StoneKey return the key of stone, use to index in StoneHub
func (stone *UploadPayloadStone) StoneKey() string {
	return string(stone.objCtx.TxHash())
}

// GetStoneState return the state of job
func (stone *UploadPayloadStone) GetStoneState() (string, error) {
	return stone.jobCtx.GetJobState()
}

// GetJobContext return the job context
func (stone *UploadPayloadStone) GetJobContext() types.JobContext {
	return stone.jobCtx.JobContext()
}

// GetObjectInfo return the object info
func (stone *UploadPayloadStone) GetObjectInfo() types.ObjectInfo {
	return stone.objCtx.GetObjectInfo()
}
