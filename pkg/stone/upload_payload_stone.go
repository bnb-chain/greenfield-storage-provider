package stone

import (
	"context"
	"errors"

	"github.com/looplab/fsm"

	"github.com/bnb-chain/inscription-storage-provider/util"

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
	gcCh   chan uint64            // the channel of notify StoneHub to delete stone
	jobDB  jobdb.JobDB
	metaDB metadb.MetaDB
}

// NewUploadPayloadStone return the instance of UploadPayloadStone
func NewUploadPayloadStone(ctx context.Context,
	jobContext *types.JobContext, object *types.ObjectInfo,
	jobDB jobdb.JobDB, metaDB metadb.MetaDB,
	jobCh chan StoneJob, gcCh chan uint64) (*UploadPayloadStone, error) {
	jobCtx := NewJobContextWrapper(jobContext, jobDB, metaDB)
	objectCtx := job.NewObjectInfoContext(object, jobDB, metaDB)
	uploadJob, err := job.NewUploadPayloadJob(objectCtx)
	if err != nil {
		return nil, err
	}

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
		jobDB:  jobDB,
		metaDB: metaDB,
	}
	if err := stone.selfActionEvent(ctx); err != nil {
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
	if job.PrimarySPCompleted() {
		state = types.JOB_STATE_UPLOAD_PRIMARY_DONE
	}
	if job.SecondarySPCompleted() {
		state = types.JOB_STATE_UPLOAD_SECONDARY_DONE
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
			log.CtxWarnw(ctx, "ignore self action stone fsm error", "event", event, "error", err)
		}
	}
	if stone.jobCtx.JobErr() != nil {
		actionFsm(ctx, InterruptEvent)
		return stone.jobCtx.JobErr()
	}
	var current string
	var event string
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
			return nil
		}
		actionFsm(ctx, event)
		to := stone.jobFsm.Current()
		log.CtxDebugw(ctx, "self action upload stone fsm", "from",
			util.ReadJobState(current), "to", util.ReadJobState(to))
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
			log.CtxDebugw(ctx, "ignore external action stone fsm error", "event", event, "error", err)
		}
	}
	from := stone.jobFsm.Current()
	actionFsm(ctx, event, args...)
	to := stone.jobFsm.Current()
	log.CtxInfow(ctx, "external action upload stone fsm", "from",
		util.ReadJobState(from), "to", util.ReadJobState(to))
	return stone.selfActionEvent(ctx, args...)
}

// InterruptStone interrupt the fsm and stop the stone
func (stone *UploadPayloadStone) InterruptStone(ctx context.Context, err error) error {
	if stone.jobCtx.JobErr() != nil || stone.jobFsm.Current() == types.JOB_STATE_ERROR {
		log.CtxWarnw(ctx, "interrupt stone fsm params error")
		return stone.jobCtx.JobErr()
	}
	ctx = context.WithValue(ctx, CtxStoneKey, stone)
	stone.jobCtx.SetJobErr(err)
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
func (stone *UploadPayloadStone) StoneKey() uint64 {
	return stone.objCtx.GetObjectID()
}

// GetStoneState return the state of job
func (stone *UploadPayloadStone) GetStoneState() (string, error) {
	return stone.jobCtx.GetJobState()
}

// GetJobContext return the job context
func (stone *UploadPayloadStone) GetJobContext() *types.JobContext {
	return stone.jobCtx.JobContext()
}

// GetObjectInfo return the object info
func (stone *UploadPayloadStone) GetObjectInfo() *types.ObjectInfo {
	return stone.objCtx.GetObjectInfo()
}
