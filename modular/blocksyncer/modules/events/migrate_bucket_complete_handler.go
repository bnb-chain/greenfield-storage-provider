package events

import (
	"context"
	"errors"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	abci "github.com/cometbft/cometbft/abci/types"
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/log"

	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

var (
	EventCompleteMigrationBucket = proto.MessageName(&storagetypes.EventCompleteMigrationBucket{})
)

var completeMigrationBucketEvents = map[string]bool{
	EventCompleteMigrationBucket: true,
}

func (m *EventMigrationBucketCompleteModule) HandleEvent(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, event sdk.Event) error {
	if !migrationBucketEvents[event.Type] {
		return nil
	}

	typedEvent, err := sdk.ParseTypedEvent(abci.Event(event))
	if err != nil {
		log.Errorw("parse typed events error", "module", m.Name(), "event", event, "err", err)
		return err
	}

	switch event.Type {
	case EventCompleteMigrationBucket:
		migrationBucket, ok := typedEvent.(*storagetypes.EventCompleteMigrationBucket)
		if !ok {
			log.Errorw("type assert error", "type", "EventCompleteMigrationBucket", "event", typedEvent)
			return errors.New("migration bucket complete event assert error")
		}
		return m.handleCompleteMigrationBucket(ctx, block, txHash, migrationBucket)
	default:
		return nil
	}
}

func (m *EventMigrationBucketCompleteModule) handleCompleteMigrationBucket(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, completeMigrationBucket *storagetypes.EventCompleteMigrationBucket) error {

	eventCompleteMigrateBucketItem := &bsdb.EventCompleteMigrationBucket{
		BucketID:                   common.BigToHash(completeMigrationBucket.BucketId.BigInt()),
		Operator:                   common.HexToAddress(completeMigrationBucket.Operator),
		BucketName:                 completeMigrationBucket.BucketName,
		GlobalVirtualGroupFamilyId: completeMigrationBucket.GlobalVirtualGroupFamilyId,

		CreateAt:     block.Block.Height,
		CreateTxHash: txHash,
		CreateTime:   block.Block.Time.UTC().Unix(),
	}

	return m.db.SaveEventCompleteMigrationBucket(ctx, eventCompleteMigrateBucketItem)

}
