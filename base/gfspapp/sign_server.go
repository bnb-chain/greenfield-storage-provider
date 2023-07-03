package gfspapp

import (
	"context"
	"net/http"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var (
	ErrSingTaskDangling = gfsperrors.Register(BaseCodeSpace, http.StatusBadRequest, 991001, "OoooH... request lost")
)

var _ gfspserver.GfSpSignServiceServer = &GfSpBaseApp{}

func (g *GfSpBaseApp) GfSpSign(ctx context.Context, req *gfspserver.GfSpSignRequest) (*gfspserver.GfSpSignResponse, error) {
	if req.GetRequest() == nil {
		log.Error("failed to sign msg due to pointer dangling")
		return &gfspserver.GfSpSignResponse{Err: ErrSingTaskDangling}, nil
	}
	var (
		signature []byte
		txHash    string
		err       error
	)
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.ReqCounter.WithLabelValues(SingerFailure).Inc()
			metrics.ReqTime.WithLabelValues(SingerFailure).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SingerSuccess).Inc()
			metrics.ReqTime.WithLabelValues(SingerSuccess).Observe(time.Since(startTime).Seconds())
		}
	}()

	switch t := req.GetRequest().(type) {
	case *gfspserver.GfSpSignRequest_CreateBucketInfo:
		signature, err = g.signer.SignCreateBucketApproval(ctx, t.CreateBucketInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign create bucket approval", "error", err)
			metrics.ReqCounter.WithLabelValues(SingerFailureBucketApproval).Inc()
			metrics.ReqTime.WithLabelValues(SingerFailureBucketApproval).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SingerSuccessBucketApproval).Inc()
			metrics.ReqTime.WithLabelValues(SingerSuccessBucketApproval).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_MigrateBucketInfo:
		signature, err = g.signer.SignMigrateBucketApproval(ctx, t.MigrateBucketInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign migrate bucket approval", "error", err)
			metrics.ReqCounter.WithLabelValues(SingerFailureMigrateBucketApproval).Inc()
			metrics.ReqTime.WithLabelValues(SingerFailureMigrateBucketApproval).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SingerSuccessMigrateBucketApproval).Inc()
			metrics.ReqTime.WithLabelValues(SingerSuccessMigrateBucketApproval).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_CreateObjectInfo:
		signature, err = g.signer.SignCreateObjectApproval(ctx, t.CreateObjectInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign create object approval", "error", err)
			metrics.ReqCounter.WithLabelValues(SingerFailureObjectApproval).Inc()
			metrics.ReqTime.WithLabelValues(SingerFailureObjectApproval).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SingerSuccessObjectApproval).Inc()
			metrics.ReqTime.WithLabelValues(SingerSuccessObjectApproval).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_SealObjectInfo:
		txHash, err = g.signer.SealObject(ctx, t.SealObjectInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to seal object", "error", err)
			metrics.ReqCounter.WithLabelValues(SingerFailureSealObject).Inc()
			metrics.ReqTime.WithLabelValues(SingerFailureSealObject).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SingerSuccessSealObject).Inc()
			metrics.ReqTime.WithLabelValues(SingerSuccessSealObject).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_RejectObjectInfo:
		txHash, err = g.signer.RejectUnSealObject(ctx, t.RejectObjectInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to reject unseal object", "error", err)
			metrics.ReqCounter.WithLabelValues(SingerFailureRejectUnSealObject).Inc()
			metrics.ReqTime.WithLabelValues(SingerFailureRejectUnSealObject).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SingerSuccessRejectUnSealObject).Inc()
			metrics.ReqTime.WithLabelValues(SingerSuccessRejectUnSealObject).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_DiscontinueBucketInfo:
		txHash, err = g.signer.DiscontinueBucket(ctx, t.DiscontinueBucketInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to discontinue bucket", "error", err)
			metrics.ReqCounter.WithLabelValues(SingerFailureDiscontinueBucket).Inc()
			metrics.ReqTime.WithLabelValues(SingerFailureDiscontinueBucket).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SingerSuccessDiscontinueBucket).Inc()
			metrics.ReqTime.WithLabelValues(SingerSuccessDiscontinueBucket).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_SignSecondaryBls:
		signature, err = g.signer.SignSecondaryBls(ctx, t.SignSecondaryBls.ObjectId,
			t.SignSecondaryBls.GlobalVirtualGroupId, t.SignSecondaryBls.Checksums)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign secondary bls signature", "error", err)
			metrics.ReqCounter.WithLabelValues(SingerFailureIntegrityHash).Inc()
			metrics.ReqTime.WithLabelValues(SingerFailureIntegrityHash).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SingerSuccessIntegrityHash).Inc()
			metrics.ReqTime.WithLabelValues(SingerSuccessIntegrityHash).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_PingMsg:
		signature, err = g.signer.SignP2PPingMsg(ctx, t.PingMsg)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign p2p ping msg", "error", err)
			metrics.ReqCounter.WithLabelValues(SingerFailurePing).Inc()
			metrics.ReqTime.WithLabelValues(SingerFailurePing).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SingerSuccessPing).Inc()
			metrics.ReqTime.WithLabelValues(SingerSuccessPing).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_PongMsg:
		signature, err = g.signer.SignP2PPongMsg(ctx, t.PongMsg)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign p2p pong msg", "error", err)
			metrics.ReqCounter.WithLabelValues(SingerFailurePong).Inc()
			metrics.ReqTime.WithLabelValues(SingerFailurePong).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SingerSuccessPong).Inc()
			metrics.ReqTime.WithLabelValues(SingerSuccessPong).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_GfspReceivePieceTask:
		ctx = log.WithValue(ctx, log.CtxKeyTask, t.GfspReceivePieceTask.Key().String())
		signature, err = g.signer.SignReceivePieceTask(ctx, t.GfspReceivePieceTask)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign receive piece task", "error", err)
			metrics.ReqCounter.WithLabelValues(SingerFailureReceiveTask).Inc()
			metrics.ReqTime.WithLabelValues(SingerFailureReceiveTask).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SingerSuccessReceiveTask).Inc()
			metrics.ReqTime.WithLabelValues(SingerSuccessReceiveTask).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_GfspReplicatePieceApprovalTask:
		ctx = log.WithValue(ctx, log.CtxKeyTask, t.GfspReplicatePieceApprovalTask.Key().String())
		signature, err = g.signer.SignReplicatePieceApproval(ctx, t.GfspReplicatePieceApprovalTask)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign replicate piece task", "error", err)
			metrics.ReqCounter.WithLabelValues(SingerFailureReplicateApproval).Inc()
			metrics.ReqTime.WithLabelValues(SingerFailureReplicateApproval).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SingerSuccessReplicateApproval).Inc()
			metrics.ReqTime.WithLabelValues(SingerSuccessReplicateApproval).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_CreateGlobalVirtualGroup:
		err = g.signer.CreateGlobalVirtualGroup(ctx, &virtualgrouptypes.MsgCreateGlobalVirtualGroup{
			PrimarySpAddress: g.operatorAddress,
			FamilyId:         t.CreateGlobalVirtualGroup.GetVirtualGroupFamilyId(),
			SecondarySpIds:   t.CreateGlobalVirtualGroup.GetSecondarySpIds(),
			Deposit:          *t.CreateGlobalVirtualGroup.GetDeposit(),
		})
		if err != nil {
			log.CtxErrorw(ctx, "failed to create global virtual group", "error", err)
		}
	case *gfspserver.GfSpSignRequest_GfspRecoverPieceTask:
		ctx = log.WithValue(ctx, log.CtxKeyTask, t.GfspRecoverPieceTask.Key().String())
		log.CtxDebugw(ctx, "signing recovery task")
		signature, err = g.signer.SignRecoveryPieceTask(ctx, t.GfspRecoverPieceTask)
		if err != nil {
			metrics.ReqCounter.WithLabelValues(SingerFailureRecoveryTask).Inc()
			metrics.ReqTime.WithLabelValues(SingerFailureRecoveryTask).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SingerSuccessRecoveryTask).Inc()
			metrics.ReqTime.WithLabelValues(SingerSuccessRecoveryTask).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_GfspMigratePieceTask:
		ctx = log.WithValue(ctx, log.CtxKeyTask, t.GfspMigratePieceTask.Key().String())
		signature, err = g.signer.SignMigratePiece(ctx, t.GfspMigratePieceTask)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign migrate piece task", "error", err)
		}
	default:
		log.CtxError(ctx, "unknown gfsp sign request type")
		return &gfspserver.GfSpSignResponse{
			Err:       gfsperrors.MakeGfSpError(ErrUnsupportedTaskType),
			Signature: nil,
		}, nil
	}
	return &gfspserver.GfSpSignResponse{
		Err:       gfsperrors.MakeGfSpError(err),
		Signature: signature,
		TxHash:    txHash}, nil
}
