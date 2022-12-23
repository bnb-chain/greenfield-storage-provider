package stone

import (
	"context"
	"errors"
	"github.com/bnb-chain/inscription-storage-provider/pkg/job"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/store"
	"github.com/looplab/fsm"
)

type StoneJob interface {
}

const (
	CtxStoneKey string = "UploadPayloadStone"
)

type UploadPayloadStone struct {
	jobCtx *JobContextWrapper
	jobFsm *fsm.FSM
	job    *job.UploadPayloadJob
	jobCh  chan StoneJob
	gcCh   chan uint64
}

func NewUploadPayloadStone(jobContext *types.JobContext, jobDB store.JobDB, metaDB store.MetaDB, jobCh chan StoneJob, gcCh chan uint64) (*UploadPayloadStone, error) {
	if jobContext == nil || jobCh == nil || gcCh == nil {
		// return error
		return nil, nil
	}
	jobCtx := NewJobContextWrapper(jobContext, jobDB, metaDB)
	if jobCtx.JobErr() != nil {
		// return error
		return nil, nil
	}
	state, err := jobCtx.GetJobState()
	if err != nil {
		return nil, err
	}
	if !UploadPayLoadState[state] {
		// return error
		return nil, nil
	}
	if state == types.JOB_STATE_SEAL_OBJECT_DONE {
		// return  error
		return nil, nil
	}
	primaryJob, err := jobCtx.GetUploadPrimaryJob()
	if err != nil {
		// return  error
		return nil, nil
	}
	secondaryJob, err := jobCtx.GetUploadSecondaryJob()
	if err != nil {
		// return  error
		return nil, nil
	}
	if secondaryJob.Done() {
		state = types.JOB_STATE_UPLOAD_SECONDARY_DONE
	} else if primaryJob.Done() {
		state = types.JOB_STATE_UPLOAD_PRIMARY_DONE
	}
	stone := &UploadPayloadStone{
		jobCtx: jobCtx,
		jobFsm: fsm.NewFSM(state, UploadPayloadFsmEvent, UploadPayLoadFsmCallBack),
		jobCh:  jobCh,
		gcCh:   gcCh,
	}
	stone.job = &job.UploadPayloadJob{}
	stone.job.SetUploadPrimaryJob(primaryJob)
	stone.job.SetUploadSecondaryJob(secondaryJob)
	if err := stone.selfActionEvent(context.Background()); err != nil {
		return nil, err
	}
	return stone, nil
}

func (stone *UploadPayloadStone) ActionEvent(ctx context.Context, event string, args ...interface{}) error {
	if stone.jobCtx.JobErr() != nil || stone.jobFsm.Current() == types.JOB_STATE_ERROR {
		// log error
		return stone.jobCtx.JobErr()
	}
	ctx = context.WithValue(ctx, CtxStoneKey, stone)
	actionFsm := func(ctx context.Context, event string, args ...interface{}) {
		if err := stone.jobFsm.Event(ctx, event, args...); err != nil {
			// only log warning
		}
	}
	actionFsm(ctx, event, args...)
	return stone.selfActionEvent(ctx, args...)
}

func (stone *UploadPayloadStone) selfActionEvent(ctx context.Context, args ...interface{}) error {
	actionFsm := func(ctx context.Context, event string, args ...interface{}) {
		if err := stone.jobFsm.Event(ctx, event, args...); err != nil {
			// only log warning
		}
	}
	if stone.jobCtx.JobErr() != nil {
		// log error
		actionFsm(ctx, InterruptEvent)
		return stone.jobCtx.JobErr()
	}
	var current string
	for {
		current = stone.jobFsm.Current()
		switch current {
		case types.JOB_STATE_CREATE_OBJECT_DONE:
			actionFsm(ctx, UploadPayloadInitEvent)
		case types.JOB_STATE_UPLOAD_PRIMARY_INIT:
			actionFsm(ctx, UploadPrimaryDoingEvent)
		case types.JOB_STATE_UPLOAD_PRIMARY_DOING:
			if stone.job.IsCompletedPrimaryJob() {
				actionFsm(ctx, UploadPrimaryDoneEvent)
			}
		case types.JOB_STATE_UPLOAD_PRIMARY_DONE:
			actionFsm(ctx, UploadSecondaryInitEvent)
		case types.JOB_STATE_UPLOAD_SECONDARY_INIT:
			actionFsm(ctx, UploadSecondaryDoingEvent)
		case types.JOB_STATE_UPLOAD_SECONDARY_DOING:
			if stone.job.IsCompletedSecondaryJob() {
				actionFsm(ctx, UploadSecondaryDoneEvent)
			}
		case types.JOB_STATE_UPLOAD_SECONDARY_DONE:
			actionFsm(ctx, SealObjectInitEvent)
		case types.JOB_STATE_SEAL_OBJECT_INIT:
			actionFsm(ctx, SealObjectDoingEvent)
		default:
			return nil
		}
		if stone.jobCtx.JobErr() != nil {
			// log error
			actionFsm(ctx, InterruptEvent)
			return stone.jobCtx.JobErr()
		}
	}
	return nil
}

func (stone *UploadPayloadStone) InterruptStone(ctx context.Context, err error) error {
	if stone.jobCtx.JobErr() != nil || stone.jobFsm.Current() == types.JOB_STATE_ERROR {
		// log error
		return stone.jobCtx.JobErr()
	}
	stone.jobCtx.SetJobErr(err)
	stone.jobFsm.Event(ctx, InterruptEvent)
	return nil
}

func (stone *UploadPayloadStone) GetJobState() (string, error) {
	return stone.jobCtx.GetJobState()
}

func EnterStateUploadPrimaryInit(ctx context.Context, event *fsm.Event) {
	return
}

func EnterStateUploadPrimaryDoing(ctx context.Context, event *fsm.Event) {
	return
}

func AfterUploadPrimaryPieceDone(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	if stone == nil {
		stone.jobCtx.SetJobErr(errors.New(""))
		return
	}
	if len(event.Args) < 1 {
		stone.jobCtx.SetJobErr(errors.New(""))
		return
	}
	pieceInfo := event.Args[0].(*service.PieceJob)
	if err := stone.job.DonePrimaryPieceJob(pieceInfo); err != nil {
		// log error
		return
	}
	if err := stone.jobCtx.SetPrimaryPieceJobState(pieceInfo, types.JOB_STATE_UPLOAD_PAYLOAD_DONE); err != nil {
		stone.jobCtx.SetJobErr(errors.New(""))
		return
	}
	return
}

func EnterUploadPrimaryDone(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	if stone == nil {
		stone.jobCtx.SetJobErr(errors.New(""))
		return
	}
	if err := stone.jobCtx.SetJobState(types.JOB_STATE_UPLOAD_PRIMARY_DONE); err != nil {
		stone.jobCtx.SetJobErr(errors.New(""))
		return
	}
	return
}

func EnterUploadSecondaryInit(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	if stone == nil {
		stone.jobCtx.SetJobErr(errors.New(""))
		return
	}
	for _, pieceJob := range stone.job.PopPendingSecondaryJob() {
		// 组装
		stone.jobCh <- pieceJob
	}
	return
}

func EnterUploadSecondaryDoing(ctx context.Context, event *fsm.Event) {
	return
}

func AfterUploadSecondaryPieceDone(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	if stone == nil {
		stone.jobCtx.SetJobErr(errors.New(""))
		return
	}
	if len(event.Args) < 1 {
		stone.jobCtx.SetJobErr(errors.New(""))
		return
	}
	spInfo := event.Args[0].(*service.PieceJob)
	if err := stone.job.DoneSecondaryPieceJob(spInfo); err != nil {
		// log error
		return
	}
	if err := stone.jobCtx.SetSecondaryJobState(spInfo, types.JOB_STATE_UPLOAD_PAYLOAD_DONE); err != nil {
		stone.jobCtx.SetJobErr(err)
		return
	}
	return
}

func EnterUploadSecondaryDone(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	if stone == nil {
		stone.jobCtx.SetJobErr(errors.New(""))
		return
	}
	if err := stone.jobCtx.SetJobState(types.JOB_STATE_UPLOAD_SECONDARY_DONE); err != nil {
		stone.jobCtx.SetJobErr(errors.New(""))
		return
	}
	return
}

func EnterSealObjectInit(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	if stone == nil {
		stone.jobCtx.SetJobErr(errors.New(""))
		return
	}
	// create seal object job
	stone.jobCh <- struct{}{}
	return
}

func EnterSealObjectDoing(ctx context.Context, event *fsm.Event) {
	return
}

func EnterSealObjectDone(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	if stone == nil {
		stone.jobCtx.SetJobErr(errors.New(""))
		return
	}
	if err := stone.jobCtx.SetJobState(types.JOB_STATE_SEAL_OBJECT_DONE); err != nil {
		stone.jobCtx.SetJobErr(errors.New(""))
		return
	}
	return
}

func AfterInterrupt(ctx context.Context, event *fsm.Event) {
	return
}

func ShowJobInfo(ctx context.Context, event *fsm.Event) {
	return
}
