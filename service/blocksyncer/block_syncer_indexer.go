package blocksyncer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"

	"github.com/ethereum/go-ethereum/common"

	"github.com/forbole/juno/v4/models"

	"github.com/forbole/juno/v4/database"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/juno/v4/parser"
	"github.com/forbole/juno/v4/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/juno/v4/modules"
	"github.com/forbole/juno/v4/node"
)

func NewIndexer(codec codec.Codec, proxy node.Node, db database.Database, modules []modules.Module) parser.Indexer {
	return &Impl{
		codec:   codec,
		Node:    proxy,
		DB:      db,
		Modules: modules,
	}
}

type Impl struct {
	Ctx context.Context

	Modules []modules.Module

	codec codec.Codec

	Node node.Node
	DB   database.Database
}

func (i *Impl) ExportBlock(block *coretypes.ResultBlock, events *coretypes.ResultBlockResults, txs []*types.Tx, vals *coretypes.ResultValidators) error {
	return nil
}

func (i *Impl) HandleEvent(ctx context.Context, block *coretypes.ResultBlock, index int, event sdk.Event) {

}

func (i *Impl) Process(height uint64) error {
	log.Debugw("processing block", "height", height)

	block, err := i.Node.Block(int64(height))
	if err != nil {
		return fmt.Errorf("failed to get block from node: %s", err)
	}

	events, err := i.Node.BlockResults(int64(height))
	if err != nil {
		return fmt.Errorf("failed to get block results from node: %s", err)
	}

	err = i.ExportEvents(i.Ctx, block, events)
	if err != nil {
		return fmt.Errorf("failed to ExportEvents: %s", err)
	}

	err = i.ExportEpoch(block)
	if err != nil {
		return fmt.Errorf("failed to ExportEpoch: %s", err)
	}

	return nil
}

func (i *Impl) ExportEpoch(block *coretypes.ResultBlock) error {
	// Save the block
	err := i.DB.SaveEpoch(i.Ctx, &models.Epoch{
		ID:          1,
		BlockHeight: block.Block.Height,
		BlockHash:   common.HexToHash(block.BlockID.Hash.String()),
		UpdateTime:  block.Block.Time.Unix(),
	})
	if err != nil {
		return fmt.Errorf("failed to persist block: %s", err)
	}

	return nil
}

func (i *Impl) ExportTxs(txs []*types.Tx) error {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) ExportValidators(block *coretypes.ResultBlock, vals *coretypes.ResultValidators) error {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) ExportCommit(block *coretypes.ResultBlock, vals *coretypes.ResultValidators) error {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) ExportAccounts(block *coretypes.ResultBlock, txs []*types.Tx) error {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) ExportEvents(ctx context.Context, block *coretypes.ResultBlock, events *coretypes.ResultBlockResults) error {
	// get all events in order from the txs within the block
	for _, tx := range events.TxsResults {
		// handle all events contained inside the transaction
		events := filterEventsType(tx)
		var eventModule modules.EventModule
		for _, module := range i.Modules {
			if _, ok := module.(modules.EventModule); ok {
				eventModule = module.(modules.EventModule)
			}
		}
		// call the event handlers
		for idx, event := range events {
			err := eventModule.HandleEvent(ctx, block, idx, event)
			if err != nil {
				log.Errorw("error while handling event", "event", event, "err", err)
			}
		}
	}
	return nil
}

func (i *Impl) HandleGenesis(genesisDoc *tmtypes.GenesisDoc, appState map[string]json.RawMessage) error {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) HandleBlock(block *coretypes.ResultBlock, events *coretypes.ResultBlockResults, txs []*types.Tx, vals *coretypes.ResultValidators) {
	for _, module := range i.Modules {
		if blockModule, ok := module.(modules.BlockModule); ok {
			err := blockModule.HandleBlock(block, events, txs, vals)
			if err != nil {
				log.Errorw("error while handling block", "module", module.Name(), "height", block.Block.Height, "err", err)
			}
		}
	}
}

func (i *Impl) HandleTx(tx *types.Tx) {
	//TODO implement me
	panic("implement me")
}

func (i *Impl) HandleMessage(index int, msg sdk.Msg, tx *types.Tx) {
	//TODO implement me
	panic("implement me")
}
