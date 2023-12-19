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
	if req == nil || req.GetRequest() == nil {
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
			metrics.ReqCounter.WithLabelValues(SignerFailureCreateGVG).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureCreateGVG).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessCreateGVG).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessCreateGVG).Observe(time.Since(startTime).Seconds())
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
	case *gfspserver.GfSpSignRequest_CompleteMigrateBucket:
		txHash, err = g.signer.CompleteMigrateBucket(ctx, t.CompleteMigrateBucket)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign complete migrate bucket", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureCompleteMigrateBucket).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureCompleteMigrateBucket).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessCompleteMigrateBucket).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessCompleteMigrateBucket).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_SignSecondarySpMigrationBucket:
		signature, err = g.signer.SignSecondarySPMigrationBucket(ctx, t.SignSecondarySpMigrationBucket)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign secondary sp bls migration bucket", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureSecondarySPMigrationBucket).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureSecondarySPMigrationBucket).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessSecondarySPMigrationBucket).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessSecondarySPMigrationBucket).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_SwapOut:
		txHash, err = g.signer.SwapOut(ctx, t.SwapOut)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign swap out", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureSwapOut).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureSwapOut).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessSwapOut).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessSwapOut).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_SignSwapOut:
		signature, err = g.signer.SignSwapOut(ctx, t.SignSwapOut)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign swap out approval", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureSignSwapOut).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureSignSwapOut).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessSignSwapOut).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessSignSwapOut).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_CompleteSwapOut:
		txHash, err = g.signer.CompleteSwapOut(ctx, t.CompleteSwapOut)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign complete swap out", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureCompleteSwapOut).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureCompleteSwapOut).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessCompleteSwapOut).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessCompleteSwapOut).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_SpExit:
		txHash, err = g.signer.SPExit(ctx, t.SpExit)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign sp exit", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureSPExit).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureSPExit).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessSPExit).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessSPExit).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_CompleteSpExit:
		txHash, err = g.signer.CompleteSPExit(ctx, t.CompleteSpExit)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign complete sp exit", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureCompleteSPExit).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureCompleteSPExit).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessCompleteSPExit).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessCompleteSPExit).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_SpStoragePrice:
		txHash, err = g.signer.UpdateSPPrice(ctx, t.SpStoragePrice)
		if err != nil {
			log.CtxErrorw(ctx, "failed to update sp price", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureSPStoragePrice).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureSPStoragePrice).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessSPStoragePrice).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessSPStoragePrice).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_GfspMigrateGvgTask:
		ctx = log.WithValue(ctx, log.CtxKeyTask, t.GfspMigrateGvgTask.Key().String())
		signature, err = g.signer.SignMigrateGVG(ctx, t.GfspMigrateGvgTask)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign migrate gvg task", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureMigrateGVGTask).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureMigrateGVGTask).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessMigrateGVGTask).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessMigrateGVGTask).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_GfspBucketMigrateInfo:
		ctx = log.WithValue(ctx, log.CtxKeyTask, t.GfspBucketMigrateInfo.Key().String())
		signature, err = g.signer.SignBucketMigrationInfo(ctx, t.GfspBucketMigrateInfo)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign bucket migration task", "task", t, "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureGfSpBucketMigrateInfo).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureGfSpBucketMigrateInfo).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessGfSpBucketMigrateInfo).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessGfSpBucketMigrateInfo).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_RejectMigrateBucket:
		txHash, err = g.signer.RejectMigrateBucket(ctx, t.RejectMigrateBucket)
		if err != nil {
			log.CtxErrorw(ctx, "failed to sign reject migrate bucket", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureRejectMigrateBucket).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureRejectMigrateBucket).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessRejectMigrateBucket).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessRejectMigrateBucket).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_ReserveSwapIn:
		txHash, err = g.signer.ReserveSwapIn(ctx, t.ReserveSwapIn)
		if err != nil {
			log.CtxErrorw(ctx, "failed to reserve swap in", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureSwapIn).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureSwapIn).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessSwapIn).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessSwapIn).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_CompleteSwapIn:
		txHash, err = g.signer.CompleteSwapIn(ctx, t.CompleteSwapIn)
		if err != nil {
			log.CtxErrorw(ctx, "failed to complete swap in", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureCompleteSwapIn).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureCompleteSwapIn).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessCompleteSwapIn).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessCompleteSwapIn).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_Deposit:
		txHash, err = g.signer.Deposit(ctx, t.Deposit)
		if err != nil {
			log.CtxErrorw(ctx, "failed to deposit", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureDeposit).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureDeposit).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessDeposit).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessDeposit).Observe(time.Since(startTime).Seconds())
		}
	case *gfspserver.GfSpSignRequest_DeleteGlobalVirtualGroup:
		txHash, err = g.signer.DeleteGlobalVirtualGroup(ctx, t.DeleteGlobalVirtualGroup)
		if err != nil {
			log.CtxErrorw(ctx, "failed to delete global virtual group", "error", err)
			metrics.ReqCounter.WithLabelValues(SignerFailureDeleteGlobalVirtualGroup).Inc()
			metrics.ReqTime.WithLabelValues(SignerFailureDeleteGlobalVirtualGroup).Observe(time.Since(startTime).Seconds())
		} else {
			metrics.ReqCounter.WithLabelValues(SignerSuccessDeleteGlobalVirtualGroup).Inc()
			metrics.ReqTime.WithLabelValues(SignerSuccessDeleteGlobalVirtualGroup).Observe(time.Since(startTime).Seconds())
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
