package events

import (
	"context"
	"errors"

	vgtypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	abci "github.com/cometbft/cometbft/abci/types"
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/log"

	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
)

var (
	EventSwapOut = proto.MessageName(&vgtypes.EventSwapOut{})
)

var swapOutEvents = map[string]bool{
	EventSwapOut: true,
}

func (m *EventSwapOutModule) HandleEvent(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, event sdk.Event) error {
	if !swapOutEvents[event.Type] {
		return nil
	}

	typedEvent, err := sdk.ParseTypedEvent(abci.Event(event))
	if err != nil {
		log.Errorw("parse typed events error", "module", m.Name(), "event", event, "err", err)
		return err
	}

	switch event.Type {
	case EventSwapOut:
		swapOut, ok := typedEvent.(*vgtypes.EventSwapOut)
		if !ok {
			log.Errorw("type assert error", "type", "EventSwapOut", "event", typedEvent)
			return errors.New("swap out event assert error")
		}
		return m.handleSwapOut(ctx, block, txHash, swapOut)
	default:
		return nil
	}
}

func (m *EventSwapOutModule) handleSwapOut(ctx context.Context, block *tmctypes.ResultBlock, txHash common.Hash, swapOut *vgtypes.EventSwapOut) error {

	eventSwapOut := &bsdb.EventSwapOut{
		StorageProviderId:          swapOut.StorageProviderId,
		GlobalVirtualGroupFamilyId: swapOut.GlobalVirtualGroupFamilyId,
		//GlobalVirtualGroupIds:      swapOut.GlobalVirtualGroupIds,
		SuccessorSpId: swapOut.SuccessorSpId,

		CreateAt:     block.Block.Height,
		CreateTxHash: txHash,
		CreateTime:   block.Block.Time.UTC().Unix(),
	}

	return m.db.SaveEventSwapOut(ctx, eventSwapOut)
}
