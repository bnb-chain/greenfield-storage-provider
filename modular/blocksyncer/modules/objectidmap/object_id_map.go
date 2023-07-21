package objectidmap

import (
	"context"
	"errors"

	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	abci "github.com/cometbft/cometbft/abci/types"
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/log"
)

var (
	EventCreateObject = proto.MessageName(&storagetypes.EventCreateObject{})
)

// buildPrefixTreeEvents maps event types that trigger the creation or deletion of prefix tree nodes.
// If an event type is present and set to true in this map,
// it means that event will result in changes to the prefix tree structure.
var buildPrefixTreeEvents = map[string]bool{
	EventCreateObject: true,
}

func (m *Module) ExtractEvent(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, event sdk.Event) ([]string, error) {
	if !buildPrefixTreeEvents[event.Type] {
		return nil, nil
	}

	typedEvent, err := sdk.ParseTypedEvent(abci.Event(event))
	if err != nil {
		log.Errorw("parse typed events error", "module", m.Name(), "event", event, "err", err)
		return nil, err
	}

	switch event.Type {
	case EventCreateObject:
		createObject, ok := typedEvent.(*storagetypes.EventCreateObject)
		if !ok {
			log.Errorw("type assert error", "type", "EventCreateObject", "event", typedEvent)
			return nil, errors.New("create object event assert error")
		}
		return []string{m.handleCreateObject(ctx, createObject)}, nil
	default:
		return nil, nil
	}
}

// HandleEvent handles the events relevant to the building of the PrefixTree.
// It checks the type of the event and calls the appropriate handler for it.
func (m *Module) HandleEvent(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, event sdk.Event) error {
	return nil
}

// handleCreateObject handles EventCreateObject.
func (m *Module) handleCreateObject(ctx context.Context, createObject *storagetypes.EventCreateObject) string {
	objectIDMap := &bsdb.ObjectIDMap{
		ObjectID:   common.BigToHash(createObject.ObjectId.BigInt()),
		BucketName: createObject.BucketName,
	}

	return m.db.CreateObjectIDMap(ctx, objectIDMap)
}
