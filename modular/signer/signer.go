package signer

import (
	"context"
	"net/http"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-common/go/hash"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsperrors"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

var (
	ErrSignMsg                   = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120001, "sign message with private key failed")
	ErrSealObjectOnChain         = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120002, "send sealObject msg failed")
	ErrRejectUnSealObjectOnChain = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120003, "send rejectUnSealObject msg failed")
	ErrDiscontinueBucketOnChain  = gfsperrors.Register(module.SignModularName, http.StatusBadRequest, 120004, "send discontinueBucket msg failed")
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

func (s *SignModular) SignReceivePieceTask(ctx context.Context, task task.ReceivePieceTask) (
	[]byte, error) {
	msg := task.GetSignBytes()
	sig, err := s.client.Sign(SignOperator, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SignModular) SignIntegrityHash(ctx context.Context, objectID uint64, checksums [][]byte) (
	[]byte, []byte, error) {
	integrityHash := hash.GenerateIntegrityHash(checksums)
	opAddr, err := s.client.GetAddr(SignOperator)
	if err != nil {
		return nil, nil, err
	}

	msg := storagetypes.NewSecondarySpSignDoc(opAddr, sdkmath.NewUint(objectID), integrityHash).GetSignBytes()
	sig, err := s.client.Sign(SignApproval, msg)
	if err != nil {
		return nil, nil, err
	}
	return sig, integrityHash, nil
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

func (s *SignModular) SealObject(ctx context.Context, object *storagetypes.MsgSealObject) error {
	var (
		err       error
		startTime = time.Now()
	)
	defer func() {
		metrics.SealObjectTimeHistogram.WithLabelValues(s.Name()).Observe(time.Since(startTime).Seconds())
	}()
	_, err = s.client.SealObject(ctx, SignSeal, object)
	return err
}

func (s *SignModular) RejectUnSealObject(
	ctx context.Context,
	rejectObject *storagetypes.MsgRejectSealObject) error {
	var (
		err       error
		startTime = time.Now()
	)
	defer func() {
		metrics.RejectUnSealObjectTimeHistogram.WithLabelValues(s.Name()).Observe(time.Since(startTime).Seconds())
		if err != nil {
			metrics.RejectUnSealObjectSucceedCounter.WithLabelValues(s.Name()).Inc()
		} else {
			metrics.RejectUnSealObjectFailedCounter.WithLabelValues(s.Name()).Inc()
		}
	}()
	_, err = s.client.RejectUnSealObject(ctx, SignSeal, rejectObject)
	return err
}

func (s *SignModular) DiscontinueBucket(ctx context.Context, bucket *storagetypes.MsgDiscontinueBucket) error {
	var (
		err       error
		startTime = time.Now()
	)
	defer func() {
		metrics.DiscontinueBucketTimeHistogram.WithLabelValues(s.Name()).Observe(time.Since(startTime).Seconds())
		if err != nil {
			metrics.DiscontinueBucketSucceedCounter.WithLabelValues(s.Name()).Inc()
		} else {
			metrics.DiscontinueBucketFailedCounter.WithLabelValues(s.Name()).Inc()
		}
	}()
	_, err = s.client.DiscontinueBucket(ctx, SignGc, bucket)
	return err
}
