package singer

import (
	"context"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-common/go/hash"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspp2p"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var _ module.Signer = &SingModular{}

type SingModular struct {
	baseApp *gfspapp.GfSpBaseApp
	client  *GreenfieldChainSignClient
}

func (s *SingModular) Name() string {
	return module.SignerModularName
}

func (s *SingModular) Start(ctx context.Context) error {
	return nil
}

func (s *SingModular) Stop(ctx context.Context) error {
	return nil
}

func (s *SingModular) ReserveResource(
	ctx context.Context,
	state *rcmgr.ScopeStat) (
	rcmgr.ResourceScopeSpan, error) {
	return &rcmgr.NullScope{}, nil
}

func (s *SingModular) ReleaseResource(
	ctx context.Context,
	span rcmgr.ResourceScopeSpan) {
	span.Done()
}

func (s *SingModular) SignCreateBucketApproval(
	ctx context.Context,
	bucket *storagetypes.MsgCreateBucket) (
	[]byte, error) {
	msg := bucket.GetApprovalBytes()
	sig, err := s.client.Sign(SignApproval, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SingModular) SignCreateObjectApproval(
	ctx context.Context,
	object *storagetypes.MsgCreateObject) (
	[]byte, error) {
	msg := object.GetApprovalBytes()
	sig, err := s.client.Sign(SignApproval, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SingModular) SignReplicatePieceApproval(
	ctx context.Context,
	task task.ApprovalReplicatePieceTask) (
	[]byte, error) {
	msg := task.GetSignBytes()
	sig, err := s.client.Sign(SignOperator, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SingModular) SignReceivePieceTask(
	ctx context.Context,
	task task.ReceivePieceTask) (
	[]byte, error) {
	msg := task.GetSignBytes()
	sig, err := s.client.Sign(SignOperator, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SingModular) SignIntegrityHash(
	ctx context.Context,
	objectID uint64,
	checksums [][]byte) (
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

func (s *SingModular) SignP2PPingMsg(
	ctx context.Context,
	ping *gfspp2p.GfSpPing) (
	[]byte, error) {
	msg := ping.GetSignBytes()
	sig, err := s.client.Sign(SignOperator, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SingModular) SignP2PPongMsg(
	ctx context.Context,
	pong *gfspp2p.GfSpPong) (
	[]byte, error) {
	msg := pong.GetSignBytes()
	sig, err := s.client.Sign(SignOperator, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *SingModular) SealObject(
	ctx context.Context,
	object *storagetypes.MsgSealObject) error {
	var (
		err       error
		startTime = time.Now()
	)
	defer func() {
		metrics.SealObjectTimeHistogram.WithLabelValues(s.Name()).Observe(time.Since(startTime).Seconds())
		if err != nil {
			metrics.SealObjectFailedCounter.WithLabelValues(s.Name()).Inc()
		} else {
			metrics.SealObjectSucceedCounter.WithLabelValues(s.Name()).Inc()
		}
	}()
	_, err = s.client.SealObject(ctx, SignSeal, object)
	return err
}
