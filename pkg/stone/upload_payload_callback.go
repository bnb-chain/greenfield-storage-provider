package stone

import (
	"context"

	"github.com/looplab/fsm"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
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
}

// EnterStateUploadPrimaryDoing is called when enter JOB_STATE_UPLOAD_PRIMARY_DOING
func EnterStateUploadPrimaryDoing(ctx context.Context, event *fsm.Event) {
}

// AfterUploadPrimaryPieceDone is called when primary piece job is done,
// and update the job state to the DB
func AfterUploadPrimaryPieceDone(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	var err error
	defer func() {
		if err != nil {
			stone.jobCtx.SetJobErr(err)
			log.CtxErrorw(ctx, "done primary piece job error", "error", err)
		}
	}()
	if len(event.Args) < 1 {
		err = merrors.ErrPieceJobMissing
		return
	}
	pieceInfo, ok := event.Args[0].(*stypes.PieceJob)
	if !ok {
		err = merrors.ErrPieceJobMissing
		return
	}
	err = stone.job.DonePrimarySPJob(pieceInfo)
}

// EnterUploadPrimaryDone is called when upload primary storage provider is completed,
// and update the job state to the DB
func EnterUploadPrimaryDone(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	if err := stone.jobCtx.SetJobState(ptypes.JOB_STATE_UPLOAD_PRIMARY_DONE); err != nil {
		stone.jobCtx.SetJobErr(err)
		log.CtxErrorw(ctx, "update primary done job state error", "error", err)
		return
	}
}

// EnterUploadSecondaryInit is called when enter JOB_STATE_UPLOAD_SECONDARY_INIT
func EnterUploadSecondaryInit(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	secondaryJob := stone.job.PopPendingSecondarySPJob()
	if secondaryJob == nil {
		return
	}
	stone.jobCh <- secondaryJob
}

// EnterUploadSecondaryDoing is called when enter JOB_STATE_UPLOAD_SECONDARY_DOING
func EnterUploadSecondaryDoing(ctx context.Context, event *fsm.Event) {
}

// AfterUploadSecondaryPieceDone is called when secondary piece job is done,
// and update the job state to the DB
func AfterUploadSecondaryPieceDone(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	var interruptErr error
	defer func() {
		if interruptErr != nil {
			stone.jobCtx.SetJobErr(interruptErr)
			log.CtxErrorw(ctx, "done secondary piece job failed", "error", interruptErr)
		}
	}()
	if len(event.Args) < 1 {
		interruptErr = merrors.ErrPieceJobMissing
		return
	}
	pieceInfo, ok := event.Args[0].(*stypes.PieceJob)
	if !ok {
		interruptErr = merrors.ErrPieceJobMissing
		return
	}
	if interruptErr = stone.job.DoneSecondarySPJob(pieceInfo); interruptErr != nil {
		return
	}
}

// EnterUploadSecondaryDone is called when upload secondary storage providers is completed,
// and update the job state to the DB
func EnterUploadSecondaryDone(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	if err := stone.jobCtx.SetJobState(ptypes.JOB_STATE_UPLOAD_SECONDARY_DONE); err != nil {
		stone.jobCtx.SetJobErr(err)
		log.CtxErrorw(ctx, "update primary done job state error", "error", err)
		return
	}
}

// SealObjectJob defines the job to transfer StoneHub
type SealObjectJob struct {
	ObjectInfo        *ptypes.ObjectInfo
	PrimarySealInfo   *ptypes.StorageProviderInfo
	SecondarySealInfo []*ptypes.StorageProviderInfo
}

// EnterSealObjectInit is called when enter JOB_STATE_SEAL_OBJECT_INIT,
// and sent SealObjectJob to StoneHub
func EnterSealObjectInit(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	var (
		primarySealInfos  []*ptypes.StorageProviderInfo
		secondarySealInfo []*ptypes.StorageProviderInfo
		err               error
	)
	defer func() {
		if err != nil {
			stone.jobCtx.SetJobErr(err)
			log.CtxErrorw(ctx, "seal object init failed", "error", err)
		}
	}()
	primarySealInfos, err = stone.job.PrimarySPSealInfo()
	if err != nil {
		log.Warnw("failed to get primary seal info", "error", err)
		return
	}
	primarySealInfo := primarySealInfos[0]
	if stone.signer != nil {
		primarySealInfo.IntegrityHash, primarySealInfo.Signature, err = stone.signer.SignIntegrityHash(
			context.Background(), primarySealInfo.PieceChecksum)
		if err != nil {
			log.Warnw("failed to sign integrity hash", "error", err)
			return
		}
	}

	// write integrity meta to db
	objectInfo := stone.GetObjectInfo()
	integrityInfo := &spdb.IntegrityMeta{
		ObjectID:       objectInfo.GetObjectId(),
		IsPrimary:      true,
		RedundancyType: objectInfo.GetRedundancyType(),
		IntegrityHash:  primarySealInfo.IntegrityHash,
		PieceHash:      primarySealInfo.GetPieceChecksum(),
		Signature:      primarySealInfo.Signature,
	}
	integrityInfo.PieceCount = uint32(len(integrityInfo.PieceHash))
	if err = stone.metaDB.SetIntegrityMeta(integrityInfo); err != nil {
		return
	}
	secondarySealInfo, err = stone.job.SecondarySPSealInfo()
	if err != nil {
		return
	}

	primarySealInfo.IntegrityHash, primarySealInfo.Signature, err = stone.signer.SignIntegrityHash(
		context.Background(), primarySealInfo.PieceChecksum)
	if err != nil {
		log.Warnw("failed to sign primary integrity hash", "error", err)
		return
	}

	job := &SealObjectJob{
		ObjectInfo:        stone.GetObjectInfo(),
		PrimarySealInfo:   primarySealInfo,
		SecondarySealInfo: secondarySealInfo,
	}
	stone.jobCh <- job
}

// EnterSealObjectDoing is called when enter JOB_STATE_SEAL_OBJECT_DOING,
func EnterSealObjectDoing(ctx context.Context, event *fsm.Event) {
}

// EnterSealObjectDone is called when enter JOB_STATE_SEAL_OBJECT_DONE,
// and update the job state to the DB
func EnterSealObjectDone(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	if err := stone.jobCtx.SetJobState(ptypes.JOB_STATE_SEAL_OBJECT_DONE); err != nil {
		stone.jobCtx.SetJobErr(err)
		log.CtxErrorw(ctx, "update seal object done job state error", "error", err)
		return
	}
}

// AfterInterrupt is called when call InterruptStone,
// and send the stone key to gc
func AfterInterrupt(ctx context.Context, event *fsm.Event) {
	stone := ctx.Value(CtxStoneKey).(*UploadPayloadStone)
	log.CtxErrorw(ctx, "interrupt stone fsm", "error", stone.jobCtx.JobErr())
	stone.gcCh <- stone.StoneKey()
}

// ShowStoneInfo is call before and after event,
// TBO::use for debugging, inspect, statistics, etc.
func ShowStoneInfo(ctx context.Context, event *fsm.Event) {
}
