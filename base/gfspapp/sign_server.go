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
			metrics.ReqCounter.WithLabelValues(SignerFailure).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailure).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccess).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccess).Observe(time.Since(startTime).Seconds())
		}
	}()

	switch t := req.GetRequest().(type) {
	case *gfspserver.GfSpSignRequest_CreateBucketInfo:
		signature, err = g.signer.SignCreateBucketApproval(ctx, t.CreateBucketInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign create bucket approval", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureBucketApproval).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureBucketApproval).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessBucketApproval).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessBucketApproval).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_MigrateBucketInfo:
		signature, err = g.signer.SignMigrateBucketApproval(ctx, t.MigrateBucketInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign migrate bucket approval", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureMigrateBucketApproval).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureMigrateBucketApproval).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessMigrateBucketApproval).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessMigrateBucketApproval).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_CreateObjectInfo:
		signature, err = g.signer.SignCreateObjectApproval(ctx, t.CreateObjectInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign create object approval", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureObjectApproval).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureObjectApproval).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessObjectApproval).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessObjectApproval).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_SealObjectInfo:
		txHash, err = g.signer.SealObject(ctx, t.SealObjectInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to seal object", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureSealObject).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureSealObject).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessSealObject).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessSealObject).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_RejectObjectInfo:
		txHash, err = g.signer.RejectUnSealObject(ctx, t.RejectObjectInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to reject unseal object", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureRejectUnSealObject).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureRejectUnSealObject).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessRejectUnSealObject).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessRejectUnSealObject).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_DiscontinueBucketInfo:
		txHash, err = g.signer.DiscontinueBucket(ctx, t.DiscontinueBucketInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to discontinue bucket", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureDiscontinueBucket).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureDiscontinueBucket).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessDiscontinueBucket).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessDiscontinueBucket).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_SignSecondarySealBls:
		signature, err = g.signer.SignSecondarySealBls(ctx, t.SignSecondarySealBls.ObjectId,
			t.SignSecondarySealBls.GlobalVirtualGroupId, t.SignSecondarySealBls.Checksums)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign secondary bls signature", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureIntegrityHash).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureIntegrityHash).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessIntegrityHash).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessIntegrityHash).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_PingMsg:
		signature, err = g.signer.SignP2PPingMsg(ctx, t.PingMsg)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign p2p ping msg", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailurePing).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailurePing).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessPing).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessPing).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_PongMsg:
		signature, err = g.signer.SignP2PPongMsg(ctx, t.PongMsg)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign p2p pong msg", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailurePong).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailurePong).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessPong).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessPong).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_GfspReceivePieceTask:
		ctx = log.WithValue(ctx, log.CtxKeyTask, t.GfspReceivePieceTask.Key().String())
		signature, err = g.signer.SignReceivePieceTask(ctx, t.GfspReceivePieceTask)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign receive piece task", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureReceiveTask).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureReceiveTask).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessReceiveTask).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessReceiveTask).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_GfspReplicatePieceApprovalTask:
		ctx = log.WithValue(ctx, log.CtxKeyTask, t.GfspReplicatePieceApprovalTask.Key().String())
		signature, err = g.signer.SignReplicatePieceApproval(ctx, t.GfspReplicatePieceApprovalTask)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign replicate piece task", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureReplicateApproval).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureReplicateApproval).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessReplicateApproval).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessReplicateApproval).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_CreateGlobalVirtualGroup:
		txHash, err = g.signer.CreateGlobalVirtualGroup(ctx, &virtualgrouptypes.MsgCreateGlobalVirtualGroup{
			StorageProvider: g.operatorAddress,
			FamilyId:        t.CreateGlobalVirtualGroup.GetVirtualGroupFamilyId(),
			SecondarySpIds:  t.CreateGlobalVirtualGroup.GetSecondarySpIds(),
			Deposit:         *t.CreateGlobalVirtualGroup.GetDeposit(),
		})
		if err != nil {
			log.CtxErrorw(ctx, "failed to create global virtual group", "error", err)
		}
	case *gfspserver.GfSpSignRequest_GfspRecoverPieceTask:
		ctx = log.WithValue(ctx, log.CtxKeyTask, t.GfspRecoverPieceTask.Key().String())
		log.CtxDebugw(ctx, "signing recovery task")
		signature, err = g.signer.SignRecoveryPieceTask(ctx, t.GfspRecoverPieceTask)
		if err != nil {
			metrics.ReqCounter.WithLabelValues(SignerFailureRecoveryTask).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureRecoveryTask).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessRecoveryTask).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessRecoveryTask).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_GfspMigratePieceTask:
		ctx = log.WithValue(ctx, log.CtxKeyTask, t.GfspMigratePieceTask.Key().String())
		signature, err = g.signer.SignMigratePiece(ctx, t.GfspMigratePieceTask)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign migrate piece task", "error", err)
		}
	case *gfspserver.GfSpSignRequest_CompleteMigrateBucket:
		txHash, err = g.signer.CompleteMigrateBucket(ctx, t.CompleteMigrateBucket)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign complete migrate bucket", "error", err)
		}
	case *gfspserver.GfSpSignRequest_SignSecondarySpMigrationBucket:
		signature, err = g.signer.SignSecondarySPMigrationBucket(ctx, t.SignSecondarySpMigrationBucket)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign secondary sp bls migration bucket", "error", err)
		}
	case *gfspserver.GfSpSignRequest_SwapOut:
		txHash, err = g.signer.SwapOut(ctx, t.SwapOut)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign swap out", "error", err)
		}
	case *gfspserver.GfSpSignRequest_SignSwapOut:
		signature, err = g.signer.SignSwapOut(ctx, t.SignSwapOut)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign swap out approval", "error", err)
		}
	case *gfspserver.GfSpSignRequest_CompleteSwapOut:
		txHash, err = g.signer.CompleteSwapOut(ctx, t.CompleteSwapOut)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign complete swap out", "error", err)
		}
	case *gfspserver.GfSpSignRequest_SpExit:
		txHash, err = g.signer.SPExit(ctx, t.SpExit)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign sp exit", "error", err)
		}
	case *gfspserver.GfSpSignRequest_CompleteSpExit:
		txHash, err = g.signer.CompleteSPExit(ctx, t.CompleteSpExit)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign complete sp exit", "error", err)
		}
	case *gfspserver.GfSpSignRequest_SpStoragePrice:
		txHash, err = g.signer.UpdateSPPrice(ctx, t.SpStoragePrice)
		if err != nil {
			log.CtxErrorw(ctx, "failed to update sp price", "error", err)
		}
	case *gfspserver.GfSpSignRequest_GfspMigrateGvgTask:
		ctx = log.WithValue(ctx, log.CtxKeyTask, t.GfspMigrateGvgTask.Key().String())
		signature, err = g.signer.SignMigrateGVG(ctx, t.GfspMigrateGvgTask)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign migrate gvg task", "error", err)
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
		TxHash:    txHash,
	}, nil
}
