package stone

import (
	"context"
	"errors"
	"github.com/bnb-chain/inscription-storage-provider/pkg/job"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/store"
	"github.com/looplab/fsm"
)

type UploadPayloadStone struct {
	jobCtx           *JobContextWrapper
	uploadPayloadJob *job.UploadPayloadJob
	jobFSM           *fsm.FSM
	primaryStone     *UploadPrimaryStone
	secondaryStone   *UploadSecondaryStone
	gcCh             chan uint64
}

var (
	JobContextKey                      string = "JobContextKey"
	JobContextErrKey                   string = "JobContextErrKey"
	JobContextUploadPayloadJobStoneKey string = "JobContextUploadPayloadJobStoneKey"
)

func NewUploadPayloadStone(jobCtx *types.JobContext, jobCh chan StoneJob,
	gcCh chan uint64, jobDB store.JobDB, metaDB store.MetaDB) *UploadPayloadStone {
	stone := &UploadPayloadStone{
		jobCtx: NewJobContextWrapper(jobCtx, jobCh, jobDB, metaDB, gcCh),
		jobFSM: fsm.NewFSM(
			types.JOB_STATE_CREATE_OBJECT_DONE,
			UploadPayloadFsmEvent,
			UploadPayLoadFsmCallBack),
	}
	stone.primaryStone = NewUploadPrimaryStone(stone.jobCtx, stone)
	stone.secondaryStone = NewUploadSecondaryStone(stone.jobCtx, stone)
	return stone
}

func (stone *UploadPayloadStone) GetUploadPrimaryStone() *UploadPrimaryStone {
	return stone.primaryStone
}

func (stone *UploadPayloadStone) GetUploadSecondaryStone() *UploadSecondaryStone {
	return stone.secondaryStone
}

func (stone *UploadPayloadStone) GetUploadPayloadJob() *job.UploadPayloadJob {
	return stone.uploadPayloadJob
}

func (stone *UploadPayloadStone) SetUploadPayloadJob(job *job.UploadPayloadJob) {
	stone.uploadPayloadJob = job
}

func (stone *UploadPayloadStone) GetUploadPrimaryJob() *job.UploadPrimaryJob {
	return stone.uploadPayloadJob.UploadPrimaryJob
}

func (stone *UploadPayloadStone) SetUploadPrimaryJob(UploadPrimaryJob *job.UploadPrimaryJob) {
	stone.uploadPayloadJob.UploadPrimaryJob = UploadPrimaryJob
}

func (stone *UploadPayloadStone) GetSecondaryPrimaryJob() *job.UploadSecondaryJob {
	return stone.uploadPayloadJob.UploadSecondaryJob
}

func (stone *UploadPayloadStone) SetUploadSecondaryJob(uploadPayloadJob *job.UploadSecondaryJob) {
	stone.uploadPayloadJob.UploadSecondaryJob = uploadPayloadJob
}

func (stone *UploadPayloadStone) ActionEvent(ctx context.Context, event string, args ...interface{}) error {
	if stone.jobCtx.JobError() != nil {
		return stone.jobCtx.JobError()
	}
	ctx = context.WithValue(ctx, JobContextKey, stone.jobCtx)
	ctx = context.WithValue(ctx, JobContextUploadPayloadJobStoneKey, stone)
	if err := stone.jobFSM.Event(ctx, event, args...); err != nil {
		// only log, not return err
	}
	err := ctx.Value(JobContextErrKey).(error)
	if err != nil {
		stone.jobFSM.Event(ctx, InterruptEvent)
		return err
	}
	currentState := stone.jobFSM.Current()
	stone.jobCtx.SetJobState(currentState)
	switch currentState {
	case types.JOB_STATE_UPLOAD_PAYLOAD_INIT:
		stone.ActionEvent(ctx, UploadPayloadDoingEvent)
	case types.JOB_STATE_UPLOAD_PAYLOAD_DONE:
		stone.ActionEvent(ctx, SealObjectInitEvent)
	case types.JOB_STATE_SEAL_OBJECT_SIGNATURE_DONE:
		stone.ActionEvent(ctx, SealObjectDoingEvent)
	}
	return nil
}

func (stone *UploadPayloadStone) ActionPrimaryEvent(ctx context.Context, event string, args ...interface{}) error {
	return stone.primaryStone.ActionEvent(ctx, event, args...)
}

func (stone *UploadPayloadStone) ActionSecondaryEvent(ctx context.Context, event string, args ...interface{}) error {
	return stone.secondaryStone.ActionEvent(ctx, event, args...)
}

// fsm callbacks

func InitUploadPayloadJobContextFromDB(ctx context.Context, event *fsm.Event) {
	jobCtx := ctx.Value(JobContextKey).(*JobContextWrapper)
	if jobCtx == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	upPayloadJobStone := ctx.Value(JobContextUploadPayloadJobStoneKey).(*UploadPayloadStone)
	if upPayloadJobStone == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	jobDB := jobCtx.GetJobDB()
	jobContext, err := jobDB.GetJobContext(jobCtx.GetJobId())
	if err != nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	jobCtx.SetJobContext(jobContext)
	if jobCtx.GetJobState() != types.JOB_STATE_CREATE_OBJECT_DONE {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	upPayloadJobStone.SetUploadPayloadJob(job.NewUploadPayloadJob())
	return
}

func StartUploadPrimaryAndSecondaryJob(ctx context.Context, event *fsm.Event) {
	upPayloadJobStone := ctx.Value(JobContextUploadPayloadJobStoneKey).(*UploadPayloadStone)
	if upPayloadJobStone == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	upPayloadJobStone.GetUploadPrimaryStone().ActionEvent(ctx, UploadPrimaryInitEvent)
	upPayloadJobStone.GetUploadSecondaryStone().ActionEvent(ctx, AllocSecondaryInitEvent)
}

func UpdateUploadPayloadJobStateDoneToDB(ctx context.Context, event *fsm.Event) {
	jobCtx := ctx.Value(JobContextKey).(*JobContextWrapper)
	if jobCtx == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	jobDB := jobCtx.GetJobDB()
	if err := jobDB.SetUploadJobState(jobCtx.GetJobId(), types.JOB_STATE_UPLOAD_PAYLOAD_DONE); err != nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	return
}

func CreateIntegrityHashJob(ctx context.Context, event *fsm.Event) {
	jobCtx := ctx.Value(JobContextKey).(*JobContextWrapper)
	if jobCtx == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	// create and send integrity hash job
	jobCtx.SendJob(struct{}{})
	return
}

func UpdateIntegrityHashToDB(ctx context.Context, event *fsm.Event) {
	jobCtx := ctx.Value(JobContextKey).(*JobContextWrapper)
	if jobCtx == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	metaDB := jobCtx.GetMetaDB()
	if len(event.Args) <= 2 {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	primarySP := event.Args[0].(types.StorageProviderInfo)
	secondarySP := event.Args[1].([]types.StorageProviderInfo)
	// first write memory, if occurs err, no need write db
	if err := jobCtx.SetIntegrityHash(primarySP, secondarySP); err != nil {
		ctx = context.WithValue(ctx, JobContextErrKey, err)
		return
	}
	if err := metaDB.SetIntegrityHash(primarySP, secondarySP); err != nil {
		ctx = context.WithValue(ctx, JobContextErrKey, err)
		return
	}
	return
}

func CreateSealObjectJob(ctx context.Context, event *fsm.Event) {
	jobCtx := ctx.Value(JobContextKey).(*JobContextWrapper)
	if jobCtx == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	// create and send seal object job
	jobCtx.SendJob(struct{}{})
	return
}

func UpdateUploadPayloadJobStateSealDoneToDB(ctx context.Context, event *fsm.Event) {
	jobCtx := ctx.Value(JobContextKey).(*JobContextWrapper)
	if jobCtx == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	jobDB := jobCtx.GetJobDB()
	if err := jobDB.SetUploadJobState(jobCtx.GetJobId(), types.JOB_STATE_SEAL_OBJECT_DONE); err != nil {
		ctx = context.WithValue(ctx, JobContextErrKey, err)
		return
	}
}

func InterruptUploadPayloadJob(ctx context.Context, event *fsm.Event) {
	jobErr := ctx.Value(JobContextErrKey).(error)
	if jobErr == nil {
		// log warning
	}
	jobCtx := ctx.Value(JobContextKey).(*JobContextWrapper)
	if jobCtx == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	if err := jobCtx.InterruptJob(jobErr); err != nil {
		// log warning
	}
	return
}

func InspectUploadPayloadJobBeforeEvent(ctx context.Context, event *fsm.Event) {}
func InspectUploadPayloadJobAfterEvent(ctx context.Context, event *fsm.Event)  {}
