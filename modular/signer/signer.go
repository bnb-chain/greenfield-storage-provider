package signer

import (
	"context"
	"encoding/hex"
	"net/http"

	sdkmath "cosmossdk.io/math"

	sptypes "github.com/evmos/evmos/v12/x/sp/types"
	storagetypes "github.com/evmos/evmos/v12/x/storage/types"
	virtualgrouptypes "github.com/evmos/evmos/v12/x/virtualgroup/types"
	"github.com/zkMeLabs/mechain-storage-provider/base/gfspapp"
	"github.com/zkMeLabs/mechain-storage-provider/base/types/gfsperrors"
	"github.com/zkMeLabs/mechain-storage-provider/base/types/gfspp2p"
	"github.com/zkMeLabs/mechain-storage-provider/base/types/gfsptask"
	"github.com/zkMeLabs/mechain-storage-provider/core/module"
	"github.com/zkMeLabs/mechain-storage-provider/core/rcmgr"
	"github.com/zkMeLabs/mechain-storage-provider/core/task"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/log"
)

var (
	ErrSignMsg                            = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120001, "sign message with private key failed")
	ErrSealObjectOnChain                  = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120002, "send sealObject msg failed")
	ErrRejectUnSealObjectOnChain          = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120003, "send rejectUnSealObject msg failed")
	ErrDiscontinueBucketOnChain           = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120004, "send discontinueBucket msg failed")
	ErrDanglingPointer                    = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120005, "sign or tx msg pointer dangling")
	ErrCreateGVGOnChain                   = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120006, "send create gvg msg failed")
	ErrCompleteMigrateBucketOnChain       = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120007, "send complete migrate bucket failed")
	ErrSwapOutOnChain                     = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120008, "send swap out failed")
	ErrCompleteSwapOutOnChain             = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120009, "send complete swap out failed")
	ErrSPExitOnChain                      = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120010, "send sp exit failed")
	ErrCompleteSPExitOnChain              = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120011, "send complete sp exit failed")
	ErrUpdateSPPriceOnChain               = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120012, "send update sp price failed")
	ErrRejectMigrateBucketOnChain         = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120013, "send reject migrate bucket failed")
	ErrDepositOnChain                     = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120014, "send deposit failed")
	ErrDeleteGVGOnChain                   = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120015, "send delete GVG failed")
	ErrReserveSwapIn                      = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120016, "send reserve swap in failed")
	ErrCompleteSwapIn                     = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120017, "send complete swap in failed")
	ErrCancelSwapIn                       = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120018, "send cancel swap in failed")
	ErrDelegateUpdateObjectContentOnChain = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120019, "send DelegateUpdateObjectContent failed")
	ErrDelegateCreateObjectOnChain        = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120020, "send DelegateCreateObject failed")
)

var _ module.Signer = &SignModular{}

type SignModular struct {
	baseApp *gfspapp.GfSpBaseApp
	client  *GreenfieldChainSignClient
}

func (s *SignModular) Name() string {
	return module.SignModularName
}

func (s *SignModular) Start(ctx context.Context) error {
	return nil
}

func (s *SignModular) Stop(ctx context.Context) error {
	return nil
}

func (s *SignModular) ReserveResource(ctx context.Context, state *rcmgr.ScopeStat) (
	rcmgr.ResourceScopeSpan, error,
) {
	return &rcmgr.NullScope{}, nil
}

func (s *SignModular) ReleaseResource(ctx context.Context, span rcmgr.ResourceScopeSpan) {
	span.Done()
}

func (s *SignModular) SignCreateBucketApproval(ctx context.Context, bucket *storagetypes.MsgCreateBucket) ([]byte, error) {
	msg := bucket.GetApprovalBytes()
	sig, err := s.client.Sign(SignApproval, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SignModular) SignMigrateBucketApproval(ctx context.Context, migrateBucket *storagetypes.MsgMigrateBucket) ([]byte, error) {
	msg := migrateBucket.GetApprovalBytes()
	sig, err := s.client.Sign(SignApproval, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SignModular) SignCreateObjectApproval(ctx context.Context, object *storagetypes.MsgCreateObject) ([]byte, error) {
	msg := object.GetApprovalBytes()
	sig, err := s.client.Sign(SignApproval, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SignModular) SignReplicatePieceApproval(ctx context.Context, task task.ApprovalReplicatePieceTask) ([]byte, error) {
	msg := task.GetSignBytes()
	sig, err := s.client.Sign(SignOperator, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SignModular) SignReceivePieceTask(ctx context.Context, task task.ReceivePieceTask) ([]byte, error) {
	msg := task.GetSignBytes()
	sig, err := s.client.Sign(SignOperator, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SignModular) SignSecondarySealBls(ctx context.Context, objectID uint64, gvgId uint32, checksums [][]byte) ([]byte, error) {
	msg := storagetypes.NewSecondarySpSealObjectSignDoc(s.baseApp.ChainID(), gvgId, sdkmath.NewUint(objectID), storagetypes.GenerateHash(checksums)).GetBlsSignHash()
	sig, err := s.client.blsKm.Sign(msg[:])
	if err != nil {
		return nil, err
	}
	log.Debugw("bls signature", "len", len(sig), "object_id", objectID, "gvg_id", gvgId, "sign_doc", hex.EncodeToString(msg[:]), "pub_key", hex.EncodeToString(s.client.blsKm.PubKey().Bytes()), "sig", hex.EncodeToString(sig))
	return sig, nil
}

func (s *SignModular) SignRecoveryPieceTask(ctx context.Context, task task.RecoveryPieceTask) (
	[]byte, error,
) {
	msg := task.GetSignBytes()
	sig, err := s.client.Sign(SignOperator, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SignModular) SignP2PPingMsg(ctx context.Context, ping *gfspp2p.GfSpPing) ([]byte, error) {
	msg := ping.GetSignBytes()
	sig, err := s.client.Sign(SignOperator, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SignModular) SignP2PPongMsg(ctx context.Context, pong *gfspp2p.GfSpPong) ([]byte, error) {
	msg := pong.GetSignBytes()
	sig, err := s.client.Sign(SignOperator, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SignModular) SealObject(ctx context.Context, object *storagetypes.MsgSealObject) (string, error) {
	return s.client.SealObject(ctx, SignSeal, object)
}

func (s *SignModular) SealObjectEvm(ctx context.Context, object *storagetypes.MsgSealObject) (string, error) {
	return s.client.SealObjectEvm(ctx, SignSeal, object)
}

func (s *SignModular) RejectUnSealObject(ctx context.Context, rejectObject *storagetypes.MsgRejectSealObject) (string, error) {
	return s.client.RejectUnSealObject(ctx, SignSeal, rejectObject)
}

func (s *SignModular) RejectUnSealObjectEvm(ctx context.Context, rejectObject *storagetypes.MsgRejectSealObject) (string, error) {
	return s.client.RejectUnSealObjectEvm(ctx, SignSeal, rejectObject)
}

func (s *SignModular) DiscontinueBucket(ctx context.Context, bucket *storagetypes.MsgDiscontinueBucket) (string, error) {
	return s.client.DiscontinueBucket(ctx, SignGc, bucket)
}

func (s *SignModular) DiscontinueBucketEvm(ctx context.Context, bucket *storagetypes.MsgDiscontinueBucket) (string, error) {
	return s.client.DiscontinueBucketEvm(ctx, SignGc, bucket)
}

func (s *SignModular) CreateGlobalVirtualGroup(ctx context.Context, gvg *virtualgrouptypes.MsgCreateGlobalVirtualGroup) (string, error) {
	return s.client.CreateGlobalVirtualGroup(ctx, SignOperator, gvg)
}

func (s *SignModular) CreateGlobalVirtualGroupEvm(ctx context.Context, gvg *virtualgrouptypes.MsgCreateGlobalVirtualGroup) (string, error) {
	return s.client.CreateGlobalVirtualGroupEvm(ctx, SignOperator, gvg)
}

func (s *SignModular) SignMigrateGVG(ctx context.Context, mp *gfsptask.GfSpMigrateGVGTask) ([]byte, error) {
	sig, err := s.client.Sign(SignOperator, mp.GetSignBytes())
	if err != nil {
		log.Errorw("failed to sign migrate gvg", "error", err)
		return nil, err
	}
	return sig, nil
}

func (s *SignModular) SignBucketMigrationInfo(ctx context.Context, mp *gfsptask.GfSpBucketMigrationInfo) ([]byte, error) {
	sig, err := s.client.Sign(SignOperator, mp.GetSignBytes())
	if err != nil {
		log.Errorw("failed to sign bucket migration info", "error", err)
		return nil, err
	}
	return sig, nil
}

func (s *SignModular) CompleteMigrateBucket(ctx context.Context, migrateBucket *storagetypes.MsgCompleteMigrateBucket) (string, error) {
	return s.client.CompleteMigrateBucket(ctx, SignOperator, migrateBucket)
}

func (s *SignModular) CompleteMigrateBucketEvm(ctx context.Context, migrateBucket *storagetypes.MsgCompleteMigrateBucket) (string, error) {
	return s.client.CompleteMigrateBucketEvm(ctx, SignOperator, migrateBucket)
}

func (s *SignModular) UpdateSPPrice(ctx context.Context, price *sptypes.MsgUpdateSpStoragePrice) (string, error) {
	return s.client.UpdateSPPrice(ctx, SignOperator, price)
}

func (s *SignModular) UpdateSPPriceEvm(ctx context.Context, price *sptypes.MsgUpdateSpStoragePrice) (string, error) {
	return s.client.UpdateSPPriceEvm(ctx, SignOperator, price)
}

func (s *SignModular) SignSecondarySPMigrationBucket(ctx context.Context, signDoc *storagetypes.SecondarySpMigrationBucketSignDoc) ([]byte, error) {
	msg := signDoc.GetBlsSignHash()
	sig, err := s.client.blsKm.Sign(msg[:])
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SignModular) SwapOut(ctx context.Context, swapOut *virtualgrouptypes.MsgSwapOut) (string, error) {
	return s.client.SwapOut(ctx, SignOperator, swapOut)
}

func (s *SignModular) SwapOutEvm(ctx context.Context, swapOut *virtualgrouptypes.MsgSwapOut) (string, error) {
	return s.client.SwapOutEvm(ctx, SignOperator, swapOut)
}

func (s *SignModular) SignSwapOut(ctx context.Context, swapOut *virtualgrouptypes.MsgSwapOut) ([]byte, error) {
	msg := swapOut.GetApprovalBytes()
	sig, err := s.client.Sign(SignApproval, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SignModular) CompleteSwapOut(ctx context.Context, completeSwapOut *virtualgrouptypes.MsgCompleteSwapOut) (string, error) {
	return s.client.CompleteSwapOut(ctx, SignOperator, completeSwapOut)
}

func (s *SignModular) CompleteSwapOutEvm(ctx context.Context, completeSwapOut *virtualgrouptypes.MsgCompleteSwapOut) (string, error) {
	return s.client.CompleteSwapOutEvm(ctx, SignOperator, completeSwapOut)
}

func (s *SignModular) SPExit(ctx context.Context, spExit *virtualgrouptypes.MsgStorageProviderExit) (string, error) {
	return s.client.SPExit(ctx, SignOperator, spExit)
}

func (s *SignModular) SPExitEvm(ctx context.Context, spExit *virtualgrouptypes.MsgStorageProviderExit) (string, error) {
	return s.client.SPExitEvm(ctx, SignOperator, spExit)
}

func (s *SignModular) CompleteSPExit(ctx context.Context, completeSPExit *virtualgrouptypes.MsgCompleteStorageProviderExit) (string, error) {
	return s.client.CompleteSPExit(ctx, SignOperator, completeSPExit)
}

func (s *SignModular) CompleteSPExitEvm(ctx context.Context, completeSPExit *virtualgrouptypes.MsgCompleteStorageProviderExit) (string, error) {
	return s.client.CompleteSPExitEvm(ctx, SignOperator, completeSPExit)
}

func (s *SignModular) RejectMigrateBucket(ctx context.Context, rejectMigrateBucket *storagetypes.MsgRejectMigrateBucket) (string, error) {
	return s.client.RejectMigrateBucket(ctx, SignOperator, rejectMigrateBucket)
}

func (s *SignModular) RejectMigrateBucketEvm(ctx context.Context, rejectMigrateBucket *storagetypes.MsgRejectMigrateBucket) (string, error) {
	return s.client.RejectMigrateBucketEvm(ctx, SignOperator, rejectMigrateBucket)
}

func (s *SignModular) ReserveSwapIn(ctx context.Context, reserveSwapIn *virtualgrouptypes.MsgReserveSwapIn) (string, error) {
	return s.client.ReserveSwapIn(ctx, SignOperator, reserveSwapIn)
}

func (s *SignModular) ReserveSwapInEvm(ctx context.Context, reserveSwapIn *virtualgrouptypes.MsgReserveSwapIn) (string, error) {
	return s.client.ReserveSwapInEvm(ctx, SignOperator, reserveSwapIn)
}

func (s *SignModular) CompleteSwapIn(ctx context.Context, completeSwapIn *virtualgrouptypes.MsgCompleteSwapIn) (string, error) {
	return s.client.CompleteSwapIn(ctx, SignOperator, completeSwapIn)
}

func (s *SignModular) CompleteSwapInEvm(ctx context.Context, completeSwapIn *virtualgrouptypes.MsgCompleteSwapIn) (string, error) {
	return s.client.CompleteSwapInEvm(ctx, SignOperator, completeSwapIn)
}

func (s *SignModular) CancelSwapIn(ctx context.Context, cancelSwapIn *virtualgrouptypes.MsgCancelSwapIn) (string, error) {
	return s.client.CancelSwapIn(ctx, SignOperator, cancelSwapIn)
}

func (s *SignModular) CancelSwapInEvm(ctx context.Context, cancelSwapIn *virtualgrouptypes.MsgCancelSwapIn) (string, error) {
	return s.client.CancelSwapInEvm(ctx, SignOperator, cancelSwapIn)
}

func (s *SignModular) Deposit(ctx context.Context, deposit *virtualgrouptypes.MsgDeposit) (string, error) {
	return s.client.Deposit(ctx, SignOperator, deposit)
}

func (s *SignModular) DepositEvm(ctx context.Context, deposit *virtualgrouptypes.MsgDeposit) (string, error) {
	return s.client.DepositEvm(ctx, SignOperator, deposit)
}

func (s *SignModular) DeleteGlobalVirtualGroup(ctx context.Context, deleteGVG *virtualgrouptypes.MsgDeleteGlobalVirtualGroup) (string, error) {
	return s.client.DeleteGlobalVirtualGroup(ctx, SignOperator, deleteGVG)
}

func (s *SignModular) DeleteGlobalVirtualGroupEvm(ctx context.Context, deleteGVG *virtualgrouptypes.MsgDeleteGlobalVirtualGroup) (string, error) {
	return s.client.DeleteGlobalVirtualGroupEvm(ctx, SignOperator, deleteGVG)
}

func (s *SignModular) DelegateUpdateObjectContent(ctx context.Context, msg *storagetypes.MsgDelegateUpdateObjectContent) (string, error) {
	return s.client.DelegateUpdateObjectContent(ctx, SignOperator, msg)
}

func (s *SignModular) DelegateUpdateObjectContentEvm(ctx context.Context, msg *storagetypes.MsgDelegateUpdateObjectContent) (string, error) {
	return s.client.DelegateUpdateObjectContentEvm(ctx, SignOperator, msg)
}

func (s *SignModular) DelegateCreateObject(ctx context.Context, msg *storagetypes.MsgDelegateCreateObject) (string, error) {
	return s.client.DelegateCreateObject(ctx, SignOperator, msg)
}

func (s *SignModular) DelegateCreateObjectEvm(ctx context.Context, msg *storagetypes.MsgDelegateCreateObject) (string, error) {
	return s.client.DelegateCreateObjectEvm(ctx, SignOperator, msg)
}

func (s *SignModular) SealObjectV2(ctx context.Context, object *storagetypes.MsgSealObjectV2) (string, error) {
	return s.client.SealObjectV2(ctx, SignSeal, object)
}
