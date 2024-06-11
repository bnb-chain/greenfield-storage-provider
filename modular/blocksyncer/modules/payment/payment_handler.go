package payment

import (
	"context"
	"errors"

	abci "github.com/cometbft/cometbft/abci/types"
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	paymenttypes "github.com/evmos/evmos/v12/x/payment/types"
	storagetypes "github.com/evmos/evmos/v12/x/storage/types"
	bucketmodule "github.com/zkMeLabs/mechain-storage-provider/modular/blocksyncer/modules/bucket"
	"gorm.io/gorm"

	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/log"
	"github.com/forbole/juno/v4/models"
)

var (
	EventPaymentAccountUpdate      = proto.MessageName(&paymenttypes.EventPaymentAccountUpdate{})
	EventStreamRecordUpdate        = proto.MessageName(&paymenttypes.EventStreamRecordUpdate{})
	EventBucketFlowRateLimitStatus = proto.MessageName(&storagetypes.EventBucketFlowRateLimitStatus{})
)

var PaymentEvents = map[string]bool{
	EventPaymentAccountUpdate:      true,
	EventStreamRecordUpdate:        true,
	EventBucketFlowRateLimitStatus: true,
}

func (m *Module) ExtractEventStatements(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, event sdk.Event) (map[string][]interface{}, error) {
	if !PaymentEvents[event.Type] {
		return nil, nil
	}

	typedEvent, err := sdk.ParseTypedEvent(abci.Event(event))
	if err != nil {
		log.Errorw("parse typed events error", "module", m.Name(), "event", event, "err", err)
		return nil, err
	}

	switch event.Type {
	case EventPaymentAccountUpdate:
		paymentAccountUpdate, ok := typedEvent.(*paymenttypes.EventPaymentAccountUpdate)
		if !ok {
			log.Errorw("type assert error", "type", "EventPaymentAccountUpdate", "event", typedEvent)
			return nil, errors.New("update payment account event assert error")
		}
		return m.handlePaymentAccountUpdate(ctx, block, paymentAccountUpdate), nil
	case EventStreamRecordUpdate:
		streamRecordUpdate, ok := typedEvent.(*paymenttypes.EventStreamRecordUpdate)
		if !ok {
			log.Errorw("type assert error", "type", "EventStreamRecordUpdate", "event", typedEvent)
			return nil, errors.New("update stream record event assert error")
		}
		return m.handleEventStreamRecordUpdate(ctx, streamRecordUpdate), nil
	case EventBucketFlowRateLimitStatus:
		bucketFlowRateLimitStatus, ok := typedEvent.(*storagetypes.EventBucketFlowRateLimitStatus)
		if !ok {
			log.Errorw("type assert error", "type", "EventBucketFlowRateLimitStatus", "event", typedEvent)
			return nil, errors.New("bucket flow rate limit status event assert error")
		}
		return m.handleEventBucketFlowRateLimitStatus(ctx, block, txHash, bucketFlowRateLimitStatus), nil
	}
	return nil, nil
}

func (m *Module) HandleEvent(ctx context.Context, block *tmctypes.ResultBlock, _ common.Hash, event sdk.Event) error {
	return nil
}

func (m *Module) handlePaymentAccountUpdate(ctx context.Context, block *tmctypes.ResultBlock, paymentAccountUpdate *paymenttypes.EventPaymentAccountUpdate) map[string][]interface{} {
	paymentAccount := &models.PaymentAccount{
		Addr:       common.HexToAddress(paymentAccountUpdate.Addr),
		Owner:      common.HexToAddress(paymentAccountUpdate.Owner),
		Refundable: paymentAccountUpdate.Refundable,
		UpdateAt:   block.Block.Height,
		UpdateTime: block.Block.Time.UTC().Unix(),
	}

	k, v := m.db.SavePaymentAccountToSQL(ctx, paymentAccount)
	return map[string][]interface{}{
		k: v,
	}
}

func (m *Module) handleEventStreamRecordUpdate(ctx context.Context, streamRecordUpdate *paymenttypes.EventStreamRecordUpdate) map[string][]interface{} {
	streamRecord := &models.StreamRecord{
		Account:           common.HexToAddress(streamRecordUpdate.Account),
		CrudTimestamp:     streamRecordUpdate.CrudTimestamp,
		NetflowRate:       (*common.Big)(streamRecordUpdate.NetflowRate.BigInt()),
		FrozenNetflowRate: (*common.Big)(streamRecordUpdate.FrozenNetflowRate.BigInt()),
		StaticBalance:     (*common.Big)(streamRecordUpdate.StaticBalance.BigInt()),
		BufferBalance:     (*common.Big)(streamRecordUpdate.BufferBalance.BigInt()),
		LockBalance:       (*common.Big)(streamRecordUpdate.LockBalance.BigInt()),
		Status:            streamRecordUpdate.Status.String(),
		SettleTimestamp:   streamRecordUpdate.SettleTimestamp,
	}

	k, v := m.db.SaveStreamRecordToSQL(ctx, streamRecord)
	return map[string][]interface{}{
		k: v,
	}
}

func (m *Module) handleEventBucketFlowRateLimitStatus(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, bucketFlowRateLimitStatus *storagetypes.EventBucketFlowRateLimitStatus) map[string][]interface{} {
	bucket, err := m.db.GetBucketByBucketName(ctx, bucketFlowRateLimitStatus.BucketName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
	}
	var offChainStatus int
	if bucketFlowRateLimitStatus.IsLimited {
		offChainStatus = bucketmodule.AddStatus(bucket.OffChainStatus, int(bucketmodule.OffChainStatusIsLimited))
	} else {
		offChainStatus = bucketmodule.RemoveStatus(bucket.OffChainStatus, int(bucketmodule.OffChainStatusIsLimited))
	}
	bucketStatus := &models.Bucket{
		BucketName:     bucketFlowRateLimitStatus.BucketName,
		OffChainStatus: offChainStatus,

		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
	}

	k, v := m.db.UpdateBucketOffChainStatus(ctx, bucketStatus)
	return map[string][]interface{}{
		k: v,
	}
}
