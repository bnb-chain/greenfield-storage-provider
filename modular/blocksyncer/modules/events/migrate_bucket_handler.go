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
	EventMigrationBucket = proto.MessageName(&storagetypes.EventMigrationBucket{})
)

var migrationBucketEvents = map[string]bool{
	EventMigrationBucket: true,
}

func (m *EventMigrationBucketModule) HandleEvent(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, event sdk.Event) error {
	if !migrationBucketEvents[event.Type] {
		return nil
	}

	typedEvent, err := sdk.ParseTypedEvent(abci.Event(event))
	if err != nil {
		log.Errorw("parse typed events error", "module", m.Name(), "event", event, "err", err)
		return err
	}

	switch event.Type {
	case EventMigrationBucket:
		migrationBucket, ok := typedEvent.(*storagetypes.EventMigrationBucket)
		if !ok {
			log.Errorw("type assert error", "type", "EventMigrationBucket", "event", typedEvent)
			return errors.New("migration bucket event assert error")
		}
		return m.handleMigrationBucket(ctx, block, txHash, migrationBucket)
	default:
		return nil
	}
}

func (m *EventMigrationBucketModule) handleMigrationBucket(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, migrationBucket *storagetypes.EventMigrationBucket) error {
	eventMigrateBucketItem := &bsdb.EventMigrationBucket{
		BucketID:       common.BigToHash(migrationBucket.BucketId.BigInt()),
		Operator:       common.HexToAddress(migrationBucket.Operator),
		BucketName:     migrationBucket.BucketName,
		DstPrimarySpId: migrationBucket.DstPrimarySpId,

		CreateAt:     block.Block.Height,
		CreateTxHash: txHash,
		CreateTime:   block.Block.Time.UTC().Unix(),
	}

	return m.db.SaveEventMigrationBucket(ctx, eventMigrateBucketItem)

}
