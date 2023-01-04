package stone

import (
	"context"
	"errors"

	"github.com/looplab/fsm"

	"github.com/bnb-chain/inscription-storage-provider/pkg/job"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

/*
 * upload_payload_callback.go implement the callback of fsm
 * fsm support 8 kins callback:
 * 1. before_<EVENT> - called before event named <EVENT>
 * 2. before_event - called before all events
 * 3. leave_<OLD_STATE> - called before leaving <OLD_STATE>
 * 4. leave_state - called before leaving all states
 * 5. enter_<NEW_STATE> - called after entering <NEW_STATE>
 * 6. enter_state - called after entering all states
 * 7. after_<EVENT> - called after event named <EVENT>
 * 8. after_event - called after all events
 */

// EnterStateUploadPrimaryInit is called when enter JOB_STATE_UPLOAD_PRIMARY_INIT
func EnterStateUploadPrimaryInit(ctx context.Context, event *fsm.Event) {
	return
}

// EnterStateUploadPrimaryDoing is called when enter JOB_STATE_UPLOAD_PRIMARY_DOING
func EnterStateUploadPrimaryDoing(ctx context.Context, event *fsm.Event) {
	return
}

// AfterUploadPrimaryPieceDone is called when primary piece job is done,
// and update the job state to the DB
func AfterUploadPrimaryPieceDone(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	if len(event.Args) < 1 {
		stone.jobCtx.SetJobErr(errors.New("miss piece job params"))
		log.CtxErrorw(ctx, "primary piece done miss params")
		return
	}
	pieceInfo := event.Args[0].(*service.PieceJob)
	if err := stone.job.DonePrimarySPJob(pieceInfo); err != nil {
		stone.jobCtx.SetJobErr(errors.New("primary piece job error"))
		log.CtxErrorw(ctx, "done primary piece job error", "piece info", pieceInfo, "error", err)
		return
	}
	return
}

// EnterUploadPrimaryDone is called when upload primary storage provider is completed,
// and update the job state to the DB
func EnterUploadPrimaryDone(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	if err := stone.jobCtx.SetJobState(types.JOB_STATE_UPLOAD_PRIMARY_DONE); err != nil {
		stone.jobCtx.SetJobErr(err)
		log.CtxErrorw(ctx, "update primary done job state error", "error", err)
		return
	}
	return
}

// EnterUploadSecondaryInit is called when enter JOB_STATE_UPLOAD_SECONDARY_INIT
func EnterUploadSecondaryInit(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	secondaryJob := stone.job.PopPendingSecondarySPJob()
	if secondaryJob == nil {
		return
	}
	stone.jobCh <- secondaryJob
	return
}

// EnterUploadSecondaryDoing is called when enter JOB_STATE_UPLOAD_SECONDARY_DOING
func EnterUploadSecondaryDoing(ctx context.Context, event *fsm.Event) {
	return
}

// AfterUploadSecondaryPieceDone is called when secondary piece job is done,
// and update the job state to the DB
func AfterUploadSecondaryPieceDone(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	if len(event.Args) < 1 {
		stone.jobCtx.SetJobErr(errors.New("mis piece job params"))
		log.CtxErrorw(ctx, "secondary piece job done miss params")
		return
	}
	pieceInfo := event.Args[0].(*service.PieceJob)
	if err := stone.job.DoneSecondarySPJob(pieceInfo); err != nil {
		stone.jobCtx.SetJobErr(errors.New("secondary piece job error"))
		log.CtxErrorw(ctx, "done secondary piece job error", "piece info", pieceInfo, "error", err)
		return
	}
	return
}

// EnterUploadSecondaryDone is called when upload secondary storage providers is completed,
// and update the job state to the DB
func EnterUploadSecondaryDone(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	if err := stone.jobCtx.SetJobState(types.JOB_STATE_UPLOAD_SECONDARY_DONE); err != nil {
		stone.jobCtx.SetJobErr(err)
		log.CtxErrorw(ctx, "update primary done job state error", "error", err)
		return
	}
	return
}

// SealObjectJob defines the job to transfer StoneHub
type SealObjectJob struct {
	StoneKey          string
	BucketName        string
	ObjectName        string
	PrimarySealInfo   *job.SealInfo
	SecondarySealInfo []*job.SealInfo
}

// EnterSealObjectInit is called when enter JOB_STATE_SEAL_OBJECT_INIT,
// and sent SealObjectJob to StoneHub
func EnterSealObjectInit(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	primarySealInfo, err := stone.job.PrimarySPSealInfo()
	if err != nil {
		stone.jobCtx.SetJobErr(err)
		log.CtxErrorw(ctx, "get primary seal info error", "error", err)
		return
	}
	secondarySealInfo, err := stone.job.SecondarySPSealInfo()
	if err != nil {
		stone.jobCtx.SetJobErr(err)
		log.CtxErrorw(ctx, "get secondary seal info error", "error", err)
		return
	}
	object := stone.objCtx.GetObjectInfo()
	job := &SealObjectJob{
		StoneKey:          stone.StoneKey(),
		BucketName:        object.BucketName,
		ObjectName:        object.ObjectName,
		PrimarySealInfo:   primarySealInfo,
		SecondarySealInfo: secondarySealInfo,
	}
	stone.jobCh <- job
	return
}

// EnterSealObjectDoing is called when enter JOB_STATE_SEAL_OBJECT_DOING,
func EnterSealObjectDoing(ctx context.Context, event *fsm.Event) {
	return
}

// EnterSealObjectDone is called when enter JOB_STATE_SEAL_OBJECT_DONE,
// and update the job state to the DB
func EnterSealObjectDone(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	if err := stone.jobCtx.SetJobState(types.JOB_STATE_SEAL_OBJECT_DONE); err != nil {
		stone.jobCtx.SetJobErr(err)
		log.CtxErrorw(ctx, "update seal object done job state error", "error", err)
		return
	}
	return
}

// AfterInterrupt is called when call InterruptStone,
// and send the stone key to gc
func AfterInterrupt(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	log.CtxErrorw(ctx, "interrupt stone fsm", "error", stone.jobCtx.JobErr())
	stone.gcCh <- stone.StoneKey()
	return
}

// ShowStoneInfo is call before and after event,
// TBO::use for debugging, inspect, statistics, etc.
func ShowStoneInfo(ctx context.Context, event *fsm.Event) {
	return
}
