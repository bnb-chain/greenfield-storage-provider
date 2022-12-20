package stone

import (
	"context"
	"errors"
	"github.com/bnb-chain/inscription-storage-provider/pkg/job"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	"github.com/looplab/fsm"
)

type UploadSecondaryStone struct {
	jobCtx             *JobContextWrapper
	uploadPayloadStone *UploadPayloadStone
	jobFSM             *fsm.FSM
	uploadSecondaryJob *job.UploadSecondaryJob
}

func NewUploadSecondaryStone(jobCtx *JobContextWrapper, uploadPayloadStone *UploadPayloadStone) *UploadSecondaryStone {
	stone := &UploadSecondaryStone{
		jobCtx:             jobCtx,
		uploadPayloadStone: uploadPayloadStone,
		jobFSM: fsm.NewFSM(
			types.JOB_STATE_INIT_UNSPECIFIED,
			UploadSecondaryFsmEvent,
			UploadSecondaryFsmCallBack),
	}
	return stone
}

func (stone *UploadSecondaryStone) GetStoneState() string {
	return stone.jobFSM.Current()
}

func (stone *UploadSecondaryStone) ActionEvent(ctx context.Context, event string, args ...interface{}) error {
	if stone.jobCtx.JobError() != nil {
		return stone.jobCtx.JobError()
	}
	ctx = context.WithValue(ctx, JobContextKey, stone.jobCtx)
	ctx = context.WithValue(ctx, JobContextUploadPayloadJobStoneKey, stone.uploadPayloadStone)
	if err := stone.jobFSM.Event(ctx, event, args...); err != nil {
		// only log, not return err
	}
	err := ctx.Value(JobContextErrKey).(error)
	if err != nil {
		stone.uploadPayloadStone.ActionEvent(ctx, InterruptEvent)
		return err
	}
	currentState := stone.jobFSM.Current()
	switch currentState {
	case types.JOB_STATE_ALLOC_SECONDARY_INIT:
		return stone.ActionEvent(ctx, AllocSecondaryDoingEvent)
	case types.JOB_STATE_ALLOC_SECONDARY_DONE:
		return stone.ActionEvent(ctx, UploadSecondaryInitEvent)
	case types.JOB_STATE_UPLOAD_PAYLOAD_INIT:
		return stone.ActionEvent(ctx, UploadSecondaryDoingEvent)
	case types.JOB_STATE_UPLOAD_PAYLOAD_DOING:
		if stone.uploadSecondaryJob.IsCompleted() {
			ctx = context.WithValue(ctx, JobContextUploadPayloadJobStoneKey, stone.uploadSecondaryJob)
			return stone.ActionEvent(ctx, UploadSecondaryDoneEvent)
		}
	}
	return nil
}

func CreateAllocSecondaryJob(ctx context.Context, event *fsm.Event) {
	jobCtx := ctx.Value(JobContextKey).(*JobContextWrapper)
	if jobCtx == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	// create alloc secondary job
	jobCtx.GetObjectInfo()
	jobCtx.SendJob(struct{}{})
	return
}

func ConsumeEvent(ctx context.Context, event *fsm.Event) {}

func InitUploadSecondaryJob(ctx context.Context, event *fsm.Event) {
	jobCtx := ctx.Value(JobContextKey).(*JobContextWrapper)
	if jobCtx == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	uploadPayloadJobStone := ctx.Value(JobContextUploadPayloadJobStoneKey).(*UploadPayloadStone)
	if uploadPayloadJobStone == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	secondarySP := event.Args[0].([]string)
	secondaryPrimaryJob, err := job.NewUploadSecondaryJob(jobCtx.GetObjectInfo(), secondarySP)
	if err != nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	uploadPayloadJobStone.SetUploadSecondaryJob(secondaryPrimaryJob)
	return
}

func PopUploadSecondaryJob(ctx context.Context, event *fsm.Event) {
	jobCtx := ctx.Value(JobContextKey).(*JobContextWrapper)
	if jobCtx == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	payloadJobStone := ctx.Value(JobContextUploadPayloadJobStoneKey).(*UploadPayloadStone)
	if payloadJobStone == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	pieceJobs := payloadJobStone.GetUploadPayloadJob().PopAccumulateSecondaryJob()
	for _, pieceJob := range pieceJobs {
		jobCtx.SendJob(pieceJob)
	}
	return
}

func PopUploadSecondaryJobBySegmentId(ctx context.Context, event *fsm.Event) {
	jobCtx := ctx.Value(JobContextKey).(*JobContextWrapper)
	if jobCtx == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	payloadJobStone := ctx.Value(JobContextUploadPayloadJobStoneKey).(*UploadPayloadStone)
	if payloadJobStone == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	pieceKey := event.Args[0].(string)
	// check pieceKey
	pieceJobs := payloadJobStone.GetUploadPayloadJob().GetSecondaryPieceJobBySegmentIdx(pieceKey)
	for _, pieceJob := range pieceJobs {
		jobCtx.SendJob(pieceJob)
	}
	return
}

// DoneSecondaryPieceJob need handler error, different DonePrimaryPieceJob
func DoneSecondaryPieceJob(ctx context.Context, event *fsm.Event) {
	jobCtx := ctx.Value(JobContextKey).(*JobContextWrapper)
	if jobCtx == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	payloadJobStone := ctx.Value(JobContextUploadPayloadJobStoneKey).(*UploadPayloadStone)
	if payloadJobStone == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	pieceKey := event.Args[0].(string)
	secondarySP := event.Args[1].(string)
	pieceErr := event.Args[2].(string)
	if len(pieceErr) > 0 {
		// log
		pieceJobs, err := payloadJobStone.GetUploadPayloadJob().ReplaceSecondaryByBackup(secondarySP)
		if err != nil {
			ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
			return
		}
		for _, pieceJob := range pieceJobs {
			jobCtx.SendJob(pieceJob)
		}
		return
	}
	payloadJobStone.GetUploadPayloadJob().DoneSecondaryPieceJob(pieceKey, secondarySP)
	return
}

func UpdateUploadSecondaryStateToDB(ctx context.Context, event *fsm.Event) {
	jobCtx := ctx.Value(JobContextKey).(*JobContextWrapper)
	if jobCtx == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	payloadJobStone := ctx.Value(JobContextUploadPayloadJobStoneKey).(*UploadPayloadStone)
	if payloadJobStone == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	jobDB := jobCtx.GetJobDB()
	if err := jobDB.SetSecondaryJob(jobCtx.GetJobId(), payloadJobStone.GetUploadPayloadJob().UploadSecondaryJob); err != nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	payloadJobStone.ActionEvent(ctx, UploadPayloadDoneEvent)
}

func InspectUploadSecondaryJobBeforeEvent(ctx context.Context, event *fsm.Event) {}
func InspectUploadSecondaryJobAfterEvent(ctx context.Context, event *fsm.Event)  {}
