package bucket

import (
	"context"
	"errors"

	"gorm.io/gorm"

	abci "github.com/cometbft/cometbft/abci/types"
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/log"
	"github.com/forbole/juno/v4/models"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	EventCreateBucket             = proto.MessageName(&storagetypes.EventCreateBucket{})
	EventDeleteBucket             = proto.MessageName(&storagetypes.EventDeleteBucket{})
	EventUpdateBucketInfo         = proto.MessageName(&storagetypes.EventUpdateBucketInfo{})
	EventDiscontinueBucket        = proto.MessageName(&storagetypes.EventDiscontinueBucket{})
	EventMigrationBucket          = proto.MessageName(&storagetypes.EventMigrationBucket{})
	EventCancelMigrationBucket    = proto.MessageName(&storagetypes.EventCancelMigrationBucket{})
	EventRejectMigrateBucket      = proto.MessageName(&storagetypes.EventRejectMigrateBucket{})
	EventCompleteMigrationBucket  = proto.MessageName(&storagetypes.EventCompleteMigrationBucket{})
	EventToggleSPAsDelegatedAgent = proto.MessageName(&storagetypes.EventToggleSPAsDelegatedAgent{})
)

var BucketEvents = map[string]bool{
	EventCreateBucket:             true,
	EventDeleteBucket:             true,
	EventUpdateBucketInfo:         true,
	EventDiscontinueBucket:        true,
	EventMigrationBucket:          true,
	EventCancelMigrationBucket:    true,
	EventRejectMigrateBucket:      true,
	EventCompleteMigrationBucket:  true,
	EventToggleSPAsDelegatedAgent: true,
}

type OffChainStatus int

const (
	// OffChainStatusIsLimited has the value of 1 (binary: 00000001)
	OffChainStatusIsLimited OffChainStatus = 1 << iota // 1

	// OffChainStatusSpAsDelegatedAgentDisabled has the value of 2 (binary: 00000010)
	OffChainStatusSpAsDelegatedAgentDisabled // 1 << 1
)

// AddStatus updates the current status by adding the specified status to it.
func AddStatus(currentStatus int, statusToAdd int) int {
	return currentStatus | statusToAdd
}

// RemoveStatus removes the specified status from the current status.
func RemoveStatus(currentStatus int, statusToRemove int) int {
	return currentStatus &^ statusToRemove
}

// IsStatusSet checks if the specified status is set in the current status.
func IsStatusSet(currentStatus int, statusToCheck int) bool {
	return currentStatus&statusToCheck == statusToCheck
}

func (m *Module) ExtractEventStatements(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, event sdk.Event) (map[string][]interface{}, error) {
	if !BucketEvents[event.Type] {
		return nil, nil
	}

	typedEvent, err := sdk.ParseTypedEvent(abci.Event(event))
	if err != nil {
		log.Errorw("parse typed events error", "module", m.Name(), "event", event, "err", err)
		return nil, err
	}

	switch event.Type {
	case EventCreateBucket:
		createBucket, ok := typedEvent.(*storagetypes.EventCreateBucket)
		if !ok {
			log.Errorw("type assert error", "type", "EventCreateBucket", "event", typedEvent)
			return nil, errors.New("create bucket event assert error")
		}
		return m.handleCreateBucket(ctx, block, txHash, createBucket), nil
	case EventDeleteBucket:
		deleteBucket, ok := typedEvent.(*storagetypes.EventDeleteBucket)
		if !ok {
			log.Errorw("type assert error", "type", "EventDeleteBucket", "event", typedEvent)
			return nil, errors.New("delete bucket event assert error")
		}
		return m.handleDeleteBucket(ctx, block, txHash, deleteBucket), nil
	case EventUpdateBucketInfo:
		updateBucketInfo, ok := typedEvent.(*storagetypes.EventUpdateBucketInfo)
		if !ok {
			log.Errorw("type assert error", "type", "EventUpdateBucketInfo", "event", typedEvent)
			return nil, errors.New("update bucket event assert error")
		}
		return m.handleUpdateBucketInfo(ctx, block, txHash, updateBucketInfo), nil
	case EventDiscontinueBucket:
		discontinueBucket, ok := typedEvent.(*storagetypes.EventDiscontinueBucket)
		if !ok {
			log.Errorw("type assert error", "type", "EventDiscontinueBucket", "event", typedEvent)
			return nil, errors.New("discontinue bucket event assert error")
		}
		return m.handleDiscontinueBucket(ctx, block, txHash, discontinueBucket), nil
	case EventMigrationBucket:
		migrationBucket, ok := typedEvent.(*storagetypes.EventMigrationBucket)
		if !ok {
			log.Errorw("type assert error", "type", "EventMigrationBucket", "event", typedEvent)
			return nil, errors.New("migration bucket event assert error")
		}
		return m.handleEventMigrationBucket(ctx, block, txHash, migrationBucket), nil
	case EventCancelMigrationBucket:
		cancelMigrationBucket, ok := typedEvent.(*storagetypes.EventCancelMigrationBucket)
		if !ok {
			log.Errorw("type assert error", "type", "EventCancelMigrationBucket", "event", typedEvent)
			return nil, errors.New("cancel migration bucket event assert error")
		}
		return m.handleEventCancelMigrationBucket(ctx, block, txHash, cancelMigrationBucket), nil
	case EventRejectMigrateBucket:
		rejectMigrateBucket, ok := typedEvent.(*storagetypes.EventRejectMigrateBucket)
		if !ok {
			log.Errorw("type assert error", "type", "EventRejectMigrateBucket", "event", typedEvent)
			return nil, errors.New("reject migration bucket event assert error")
		}
		return m.handleEventRejectMigrateBucket(ctx, block, txHash, rejectMigrateBucket), nil
	case EventCompleteMigrationBucket:
		completeMigrationBucket, ok := typedEvent.(*storagetypes.EventCompleteMigrationBucket)
		if !ok {
			log.Errorw("type assert error", "type", "EventCompleteMigrationBucket", "event", typedEvent)
			return nil, errors.New("complete migrate bucket event assert error")
		}
		return m.handleCompleteMigrationBucket(ctx, block, txHash, completeMigrationBucket), nil
	case EventToggleSPAsDelegatedAgent:
		toggleSPAsDelegatedAgent, ok := typedEvent.(*storagetypes.EventToggleSPAsDelegatedAgent)
		if !ok {
			log.Errorw("type assert error", "type", "EventCompleteMigrationBucket", "event", typedEvent)
			return nil, errors.New("complete migrate bucket event assert error")
		}
		return m.handleToggleSPAsDelegatedAgent(ctx, block, txHash, toggleSPAsDelegatedAgent), nil
	}

	return nil, nil
}

func (m *Module) HandleEvent(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, event sdk.Event) error {
	return nil
}

func (m *Module) handleCreateBucket(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, createBucket *storagetypes.EventCreateBucket) map[string][]interface{} {
	bucket := &models.Bucket{
		BucketID:                   common.BigToHash(createBucket.BucketId.BigInt()),
		BucketName:                 createBucket.BucketName,
		Owner:                      common.HexToAddress(createBucket.Owner),
		PaymentAddress:             common.HexToAddress(createBucket.PaymentAddress),
		GlobalVirtualGroupFamilyId: createBucket.GlobalVirtualGroupFamilyId,
		Operator:                   common.HexToAddress(createBucket.Owner),
		SourceType:                 createBucket.SourceType.String(),
		ChargedReadQuota:           createBucket.ChargedReadQuota,
		Visibility:                 createBucket.Visibility.String(),
		Status:                     createBucket.Status.String(),

		Removed:      false,
		CreateAt:     block.Block.Height,
		CreateTxHash: txHash,
		CreateTime:   createBucket.CreateAt,
		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
	}
	k, v := m.db.SaveBucketToSQL(ctx, bucket)
	return map[string][]interface{}{
		k: v,
	}
}

func (m *Module) handleDeleteBucket(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, deleteBucket *storagetypes.EventDeleteBucket) map[string][]interface{} {
	bucket := &models.Bucket{
		BucketID:                   common.BigToHash(deleteBucket.BucketId.BigInt()),
		BucketName:                 deleteBucket.BucketName,
		Owner:                      common.HexToAddress(deleteBucket.Owner),
		GlobalVirtualGroupFamilyId: deleteBucket.GlobalVirtualGroupFamilyId,

		Removed:      true,
		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
	}

	k, v := m.db.UpdateBucketToSQL(ctx, bucket)
	return map[string][]interface{}{
		k: v,
	}
}

func (m *Module) handleDiscontinueBucket(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, discontinueBucket *storagetypes.EventDiscontinueBucket) map[string][]interface{} {
	bucket := &models.Bucket{
		BucketID:     common.BigToHash(discontinueBucket.BucketId.BigInt()),
		BucketName:   discontinueBucket.BucketName,
		DeleteReason: discontinueBucket.Reason,
		DeleteAt:     discontinueBucket.DeleteAt,
		Status:       storagetypes.BUCKET_STATUS_DISCONTINUED.String(),

		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
	}

	k, v := m.db.UpdateBucketToSQL(ctx, bucket)
	return map[string][]interface{}{
		k: v,
	}
}

func (m *Module) handleUpdateBucketInfo(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, updateBucket *storagetypes.EventUpdateBucketInfo) map[string][]interface{} {
	bucket := &models.Bucket{
		BucketName:                 updateBucket.BucketName,
		BucketID:                   common.BigToHash(updateBucket.BucketId.BigInt()),
		ChargedReadQuota:           updateBucket.ChargedReadQuota,
		PaymentAddress:             common.HexToAddress(updateBucket.PaymentAddress),
		Visibility:                 updateBucket.Visibility.String(),
		GlobalVirtualGroupFamilyId: updateBucket.GlobalVirtualGroupFamilyId,

		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
	}

	k, v := m.db.UpdateBucketToSQL(ctx, bucket)
	return map[string][]interface{}{
		k: v,
	}
}

func (m *Module) handleEventMigrationBucket(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, migrationBucket *storagetypes.EventMigrationBucket) map[string][]interface{} {
	bucket := &models.Bucket{
		BucketID:   common.BigToHash(migrationBucket.BucketId.BigInt()),
		BucketName: migrationBucket.BucketName,
		Status:     storagetypes.BUCKET_STATUS_MIGRATING.String(),

		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
	}

	k, v := m.db.UpdateBucketToSQL(ctx, bucket)
	return map[string][]interface{}{
		k: v,
	}
}

func (m *Module) handleEventCancelMigrationBucket(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, cancelMigrationBucket *storagetypes.EventCancelMigrationBucket) map[string][]interface{} {
	bucket := &models.Bucket{
		BucketID:   common.BigToHash(cancelMigrationBucket.BucketId.BigInt()),
		BucketName: cancelMigrationBucket.BucketName,
		Status:     storagetypes.BUCKET_STATUS_CREATED.String(),

		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
	}

	k, v := m.db.UpdateBucketToSQL(ctx, bucket)
	return map[string][]interface{}{
		k: v,
	}
}

func (m *Module) handleEventRejectMigrateBucket(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, rejectMigrateBucket *storagetypes.EventRejectMigrateBucket) map[string][]interface{} {
	bucket := &models.Bucket{
		BucketID:   common.BigToHash(rejectMigrateBucket.BucketId.BigInt()),
		BucketName: rejectMigrateBucket.BucketName,
		Status:     storagetypes.BUCKET_STATUS_CREATED.String(),

		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
	}

	k, v := m.db.UpdateBucketToSQL(ctx, bucket)
	return map[string][]interface{}{
		k: v,
	}
}

func (m *Module) handleCompleteMigrationBucket(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, completeMigrationBucket *storagetypes.EventCompleteMigrationBucket) map[string][]interface{} {
	bucket := &models.Bucket{
		BucketID:                   common.BigToHash(completeMigrationBucket.BucketId.BigInt()),
		BucketName:                 completeMigrationBucket.BucketName,
		GlobalVirtualGroupFamilyId: completeMigrationBucket.GlobalVirtualGroupFamilyId,
		Status:                     storagetypes.BUCKET_STATUS_CREATED.String(),

		UpdateAt:     block.Block.Height,
		UpdateTxHash: txHash,
		UpdateTime:   block.Block.Time.UTC().Unix(),
	}

	k, v := m.db.UpdateBucketToSQL(ctx, bucket)
	return map[string][]interface{}{
		k: v,
	}
}

func (m *Module) handleToggleSPAsDelegatedAgent(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, toggleSPAsDelegatedAgent *storagetypes.EventToggleSPAsDelegatedAgent) map[string][]interface{} {
	bucket, err := m.db.GetBucketByBucketName(ctx, toggleSPAsDelegatedAgent.BucketName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
	}
	var offChainStatus int
	if toggleSPAsDelegatedAgent.SpAsDelegatedAgentDisabled {
		offChainStatus = AddStatus(bucket.OffChainStatus, int(OffChainStatusSpAsDelegatedAgentDisabled))
	} else {
		offChainStatus = RemoveStatus(bucket.OffChainStatus, int(OffChainStatusSpAsDelegatedAgentDisabled))
	}
	bucketStatus := &models.Bucket{
		BucketName:     toggleSPAsDelegatedAgent.BucketName,
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
