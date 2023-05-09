package blocksyncer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/database"
	"github.com/forbole/juno/v4/models"
	"github.com/forbole/juno/v4/modules"
	"github.com/forbole/juno/v4/node"
	"github.com/forbole/juno/v4/parser"
	"github.com/forbole/juno/v4/types"
	abci "github.com/tendermint/tendermint/abci/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

func NewIndexer(codec codec.Codec, proxy node.Node, db database.Database, modules []modules.Module, serviceName string) parser.Indexer {
	return &Impl{
		codec:       codec,
		Node:        proxy,
		DB:          db,
		Modules:     modules,
		ServiceName: serviceName,
	}
}

type Impl struct {
	Modules []modules.Module
	codec   codec.Codec
	Node    node.Node
	DB      database.Database

	LatestBlockHeight atomic.Value
	CatchUpFlag       atomic.Value

	ServiceName string
}

// ExportBlock accepts a finalized block and persists then inside the database.
// An error is returned if write fails.
func (i *Impl) ExportBlock(block *coretypes.ResultBlock, events *coretypes.ResultBlockResults, txs []*types.Tx, vals *coretypes.ResultValidators) error {
	return nil
}

// HandleEvent accepts the transaction and handles events contained inside the transaction.
func (i *Impl) HandleEvent(ctx context.Context, block *coretypes.ResultBlock, txHash common.Hash, event sdk.Event) error {
	for _, module := range i.Modules {
		if eventModule, ok := module.(modules.EventModule); ok {
			err := eventModule.HandleEvent(ctx, block, txHash, event)
			if err != nil {
				log.Errorw("failed to handle event", "module", module.Name(), "event", event, "error", err)
				return err
			}
		}
	}
	return nil
}

// Process fetches a block for a given height and associated metadata and export it to a database.
// It returns an error if any export process fails.
func (i *Impl) Process(height uint64) error {
	log.Debugw("processing block", "height", height)
	var block *coretypes.ResultBlock
	var events *coretypes.ResultBlockResults
	var txs []*types.Tx
	var err error
	flagAny := i.GetCatchUpFlag().Load()
	flag := flagAny.(int64)
	heightKey := fmt.Sprintf("%s-%d", i.GetServiceName(), height)
	if flag == -1 || flag >= int64(height) {
		blockAny, okb := blockMap.Load(heightKey)
		eventsAny, oke := eventMap.Load(heightKey)
		txsAny, okt := txMap.Load(heightKey)
		block, _ = blockAny.(*coretypes.ResultBlock)
		events, _ = eventsAny.(*coretypes.ResultBlockResults)
		txs, _ = txsAny.([]*types.Tx)
		if !okb || !oke || !okt {
			log.Warnf("failed to get map data height: %d", height)
			return errors.ErrBlockNotFound
		}
	} else {
		// get block info
		block, err = i.Node.Block(int64(height))
		if err != nil {
			log.Errorf("failed to get block from node: %s", err)
			return err
		}

		// get txs
		txs, err = i.Node.Txs(block)
		if err != nil {
			log.Errorf("failed to get transactions for block: %s", err)
			return err
		}

		// get block results
		events, err = i.Node.BlockResults(int64(height))
		if err != nil {
			log.Errorf("failed to get block results from node: %s", err)
			return err
		}
	}

	beginBlockEvents := events.BeginBlockEvents
	endBlockEvents := events.EndBlockEvents

	// 1. handle events in startBlock

	if len(beginBlockEvents) > 0 {
		err = i.ExportEventsWithoutTx(context.Background(), block, beginBlockEvents)
		if err != nil {
			log.Errorf("failed to export events without tx: %s", err)
			return err
		}
	}

	// 2. handle events in txs
	err = i.ExportEventsInTxs(context.Background(), block, txs)
	if err != nil {
		log.Errorf("failed to export events in txs: %s", err)
		return err
	}

	// 3. handle events in endBlock
	if len(endBlockEvents) > 0 {
		err = i.ExportEventsWithoutTx(context.Background(), block, endBlockEvents)
		if err != nil {
			log.Errorf("failed to export events without tx: %s", err)
			return err
		}
	}

	err = i.ExportEpoch(block)
	if err != nil {
		log.Errorf("failed to export epoch: %s", err)
		return err
	}

	blockMap.Delete(heightKey)
	eventMap.Delete(heightKey)
	txMap.Delete(heightKey)

	return nil
}

// ExportEpoch accept a block result data and persist basic info into db to record current sync progress
func (i *Impl) ExportEpoch(block *coretypes.ResultBlock) error {
	// Save the block
	err := i.DB.SaveEpoch(context.Background(), &models.Epoch{
		OneRowId:    true,
		BlockHeight: block.Block.Height,
		BlockHash:   common.HexToHash(block.BlockID.Hash.String()),
		UpdateTime:  block.Block.Time.Unix(),
	})
	if err != nil {
		log.Errorf("failed to persist block: %s", err)
		return err
	}

	metrics.BlockHeightLagGauge.WithLabelValues("blocksyncer").Set(float64(block.Block.Height))

	return nil
}

// ExportTxs accepts a slice of transactions and persists then inside the database.
// An error is returned if write fails.
func (i *Impl) ExportTxs(block *coretypes.ResultBlock, txs []*types.Tx) error {
	return nil
}

// ExportValidators accepts ResultValidators and persists validators inside the database.
// An error is returned if write fails.
func (i *Impl) ExportValidators(block *coretypes.ResultBlock, vals *coretypes.ResultValidators) error {
	return nil
}

// ExportCommit accepts ResultValidators and persists validator commit signatures inside the database.
// An error is returned if write fails.
func (i *Impl) ExportCommit(block *coretypes.ResultBlock, vals *coretypes.ResultValidators) error {
	return nil
}

// ExportAccounts accepts a slice of transactions and persists accounts inside the database.
// An error is returned if write fails.
func (i *Impl) ExportAccounts(block *coretypes.ResultBlock, txs []*types.Tx) error {
	return nil
}

// ExportEvents accepts a slice of transactions and get events in order to save in database.
func (i *Impl) ExportEvents(ctx context.Context, block *coretypes.ResultBlock, events *coretypes.ResultBlockResults) error {
	// get all events in order from the txs within the block
	for _, tx := range events.TxsResults {
		// handle all events contained inside the transaction
		// call the event handlers
		for _, event := range tx.Events {
			if err := i.HandleEvent(ctx, block, common.Hash{}, sdk.Event(event)); err != nil {
				return err
			}
		}
	}
	return nil
}

// ExportEventsInTxs accepts a slice of events in tx in order to save in database.
func (i *Impl) ExportEventsInTxs(ctx context.Context, block *coretypes.ResultBlock, txs []*types.Tx) error {
	for _, tx := range txs {
		txHash := common.HexToHash(tx.TxHash)
		for _, event := range tx.Events {
			if err := i.HandleEvent(ctx, block, txHash, sdk.Event(event)); err != nil {
				return err
			}
		}
	}
	return nil
}

// ExportEventsWithoutTx accepts a slice of events not in tx in order to save in database.
// events here don't have txHash
func (i *Impl) ExportEventsWithoutTx(ctx context.Context, block *coretypes.ResultBlock, events []abci.Event) error {
	// call the event handlers
	for _, event := range events {
		if err := i.HandleEvent(ctx, block, common.Hash{}, sdk.Event(event)); err != nil {
			return err
		}
	}
	return nil
}

// HandleGenesis accepts a GenesisDoc and calls all the registered genesis handlers in the order in which they have been registered.
func (i *Impl) HandleGenesis(genesisDoc *tmtypes.GenesisDoc, appState map[string]json.RawMessage) error {
	return nil
}

// HandleBlock accepts block and calls the block handlers.
func (i *Impl) HandleBlock(block *coretypes.ResultBlock, events *coretypes.ResultBlockResults, txs []*types.Tx, vals *coretypes.ResultValidators) {
	for _, module := range i.Modules {
		if blockModule, ok := module.(modules.BlockModule); ok {
			err := blockModule.HandleBlock(block, events, txs, vals)
			if err != nil {
				log.Errorw("failed to handle event", "module", module.Name(), "height", block.Block.Height, "error", err)
			}
		}
	}
}

// HandleTx accepts the transaction and calls the tx handlers.
func (i *Impl) HandleTx(tx *types.Tx) {
	log.Info("HandleTx")
}

// HandleMessage accepts the transaction and handles messages contained inside the transaction.
func (i *Impl) HandleMessage(block *coretypes.ResultBlock, index int, msg sdk.Msg, tx *types.Tx) {
	log.Info("HandleMessage")
}

// Processed tells whether the current Indexer has already processed the given height of Block
// An error is returned if the operation fails.
func (i *Impl) Processed(ctx context.Context, height uint64) (bool, error) {
	ep, err := i.DB.GetEpoch(context.Background())
	if err != nil {
		return false, err
	}
	log.Infof("epoch height:%d, cur height: %d", ep.BlockHeight, height)
	if ep.BlockHeight > int64(height) {
		heightKey := fmt.Sprintf("%s-%d", i.GetServiceName(), height)
		blockMap.Delete(heightKey)
		eventMap.Delete(heightKey)
		txMap.Delete(heightKey)
	}
	return ep.BlockHeight > int64(height), nil
}

// GetBlockRecordNum returns total number of blocks stored in database.
func (i *Impl) GetBlockRecordNum(_ context.Context) int64 {
	return 1
}

// GetLastBlockRecordHeight returns the last block height stored inside the database
func (i *Impl) GetLastBlockRecordHeight(ctx context.Context) (uint64, error) {
	var lastBlockRecordHeight uint64
	currentEpoch, err := i.DB.GetEpoch(ctx)
	if err == nil {
		lastBlockRecordHeight = 0
	} else {
		lastBlockRecordHeight = uint64(currentEpoch.BlockHeight)
	}
	return lastBlockRecordHeight, err
}

func (i *Impl) GetLatestBlockHeight() *atomic.Value {
	return &(i.LatestBlockHeight)

}

func (i *Impl) GetCatchUpFlag() *atomic.Value {
	return &(i.CatchUpFlag)
}

func (i *Impl) CreateMasterTable() error {

	return nil
}

func (i *Impl) GetServiceName() string {
	return i.ServiceName
}
