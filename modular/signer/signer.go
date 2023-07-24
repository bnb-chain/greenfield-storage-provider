package signer

import (
	"context"
	"net/http"

	sdkmath "cosmossdk.io/math"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsptask"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var (
	ErrSignMsg                      = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120001, "sign message with private key failed")
	ErrSealObjectOnChain            = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120002, "send sealObject msg failed")
	ErrRejectUnSealObjectOnChain    = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120003, "send rejectUnSealObject msg failed")
	ErrDiscontinueBucketOnChain     = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120004, "send discontinueBucket msg failed")
	ErrDanglingPointer              = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120005, "sign or tx msg pointer dangling")
	ErrCreateGVGOnChain             = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120006, "send create gvg msg failed")
	ErrCompleteMigrateBucketOnChain = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120007, "send complete migrate bucket failed")
	ErrSwapOutOnChain               = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120008, "send swap out failed")
	ErrCompleteSwapOutOnChain       = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120009, "send complete swap out failed")
	ErrSPExitOnChain                = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120010, "send sp exit failed")
	ErrCompleteSPExitOnChain        = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120011, "send complete sp exit failed")
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
	rcmgr.ResourceScopeSpan, error) {
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
	// log.Debugw("bls signature length", "len", len(sig), "object_id", objectID, "gvg_id", gvgId, "checksums", checksums)
	return sig, nil
}

func (s *SignModular) SignRecoveryPieceTask(ctx context.Context, task task.RecoveryPieceTask) (
	[]byte, error) {
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

func (s *SignModular) RejectUnSealObject(ctx context.Context, rejectObject *storagetypes.MsgRejectSealObject) (string, error) {
	return s.client.RejectUnSealObject(ctx, SignSeal, rejectObject)
}

func (s *SignModular) DiscontinueBucket(ctx context.Context, bucket *storagetypes.MsgDiscontinueBucket) (string, error) {
	return s.client.DiscontinueBucket(ctx, SignGc, bucket)
}

func (s *SignModular) CreateGlobalVirtualGroup(ctx context.Context, gvg *virtualgrouptypes.MsgCreateGlobalVirtualGroup) error {
	_, err := s.client.CreateGlobalVirtualGroup(ctx, SignOperator, gvg)
	return err
}

func (s *SignModular) SignMigratePiece(ctx context.Context, mp *gfsptask.GfSpMigratePieceTask) ([]byte, error) {
	sig, err := s.client.Sign(SignOperator, mp.GetSignBytes())
	if err != nil {
		log.Errorw("failed to sign migrate piece", "error", err)
		return nil, err
	}
	return sig, nil
}

func (s *SignModular) CompleteMigrateBucket(ctx context.Context, migrateBucket *storagetypes.MsgCompleteMigrateBucket) (string, error) {
	return s.client.CompleteMigrateBucket(ctx, SignOperator, migrateBucket)
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

func (s *SignModular) SPExit(ctx context.Context, spExit *virtualgrouptypes.MsgStorageProviderExit) (string, error) {
	return s.client.SPExit(ctx, SignOperator, spExit)
}

func (s *SignModular) CompleteSPExit(ctx context.Context, completeSPExit *virtualgrouptypes.MsgCompleteStorageProviderExit) (string, error) {
	return s.client.CompleteSPExit(ctx, SignOperator, completeSPExit)
}
