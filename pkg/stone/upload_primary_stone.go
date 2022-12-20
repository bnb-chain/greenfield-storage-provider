package stone

import (
	"context"
	"errors"
	"github.com/bnb-chain/inscription-storage-provider/pkg/job"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	"github.com/looplab/fsm"
)

type UploadPrimaryStone struct {
	jobCtx             *JobContextWrapper
	jobFSM             *fsm.FSM
	uploadPrimaryJob   *job.UploadPrimaryJob
	uploadPayloadStone *UploadPayloadStone
}

func NewUploadPrimaryStone(jobCtx *JobContextWrapper, uploadPayloadStone *UploadPayloadStone) *UploadPrimaryStone {
	stone := &UploadPrimaryStone{
		jobCtx:             jobCtx,
		uploadPayloadStone: uploadPayloadStone,
		jobFSM: fsm.NewFSM(
			types.JOB_STATE_INIT_UNSPECIFIED,
			UploadPrimaryFsmEvent,
			UploadPrimaryFsmCallBack),
	}
	return stone
}

func (stone *UploadPrimaryStone) ActionEvent(ctx context.Context, event string, args ...interface{}) error {
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
	case types.JOB_STATE_UPLOAD_PAYLOAD_INIT:
		stone.ActionEvent(ctx, UploadPrimaryDoingEvent)
	case types.JOB_STATE_UPLOAD_PAYLOAD_DOING:
		if stone.uploadPrimaryJob.IsCompleted() {
			stone.ActionEvent(ctx, UploadPrimaryDoneEvent)
		}
	}
	return nil
}

// fsm callbacks

func InitUploadPrimaryJob(ctx context.Context, event *fsm.Event) {
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
	jobDB := jobCtx.GetJobDB()
	primaryJob, err := jobDB.GetPrimaryJob(jobCtx.GetJobId())
	if err != nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	uploadPayloadJobStone.SetUploadPrimaryJob(&primaryJob)
	return
}

func PopUploadPrimaryJob(ctx context.Context, event *fsm.Event) {
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
	for _, pieceJob := range uploadPayloadJobStone.GetUploadPrimaryJob().PopPieceJob() {
		jobCtx.SendJob(pieceJob)
	}
	return
}

// DonePrimaryPieceJob do not handler PieceJob error, the upper layer will
// action UploadPayloadStone InterruptEvent
func DonePrimaryPieceJob(ctx context.Context, event *fsm.Event) {
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
	piece := event.Args[1].(job.UploadPiece)
	if err := uploadPayloadJobStone.GetUploadPrimaryJob().DonePieceJob(
		piece.UploadPieceKey(), piece.GetSP()); err != nil {
		// only log
		return
	}
	jobDB := jobCtx.GetJobDB()
	if err := jobDB.SetPrimaryPieceJobState(jobCtx.GetJobId(), piece.UploadPieceKey(),
		types.JOB_STATE_UPLOAD_PAYLOAD_DONE); err != nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}

	// notify UploadSecondaryStone
	secondaryJobStone := uploadPayloadJobStone.GetUploadSecondaryStone()
	if secondaryJobStone.GetStoneState() != types.JOB_STATE_UPLOAD_PAYLOAD_DOING {
		return
	}
	secondaryJobStone.ActionEvent(ctx, UploadPieceNotifyEvent, piece.UploadPieceKey())
	return
}

func UpdateUploadPrimaryJobStateToDB(ctx context.Context, event *fsm.Event) {
	jobCtx := ctx.Value(JobContextKey).(*JobContextWrapper)
	if jobCtx == nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
	jobDB := jobCtx.GetJobDB()
	if err := jobDB.SetPrimaryJobState(jobCtx.GetJobId(), types.JOB_STATE_UPLOAD_PAYLOAD_DONE); err != nil {
		ctx = context.WithValue(ctx, JobContextErrKey, errors.New(""))
		return
	}
}

func InspectUploadPrimaryJobBeforeEvent(ctx context.Context, event *fsm.Event) {}
func InspectUploadPrimaryJobAfterEvent(ctx context.Context, event *fsm.Event)  {}
