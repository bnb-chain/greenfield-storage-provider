package blocksyncer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"gorm.io/gorm"

	abci "github.com/cometbft/cometbft/abci/types"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/database"
	"github.com/forbole/juno/v4/models"
	"github.com/forbole/juno/v4/modules"
	"github.com/forbole/juno/v4/node"
	"github.com/forbole/juno/v4/parser"
	"github.com/forbole/juno/v4/types"

	localDB "github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/database"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

func NewIndexer(codec codec.Codec, proxy node.Node, db database.Database, modules []modules.Module, serviceName string) parser.Indexer {
	return &Impl{
		codec:           codec,
		Node:            proxy,
		DB:              db,
		Modules:         modules,
		ServiceName:     serviceName,
		ProcessedHeight: 0,
		eventTypeCount:  8,
	}
}

type Impl struct {
	Modules []modules.Module
	codec   codec.Codec
	Node    node.Node
	DB      database.Database

	LatestBlockHeight atomic.Value
	ProcessedHeight   uint64

	eventTypeCount int

	ServiceName string
}

func (i *Impl) ExportEpoch(block *coretypes.ResultBlock) error {
	return nil
}

func (i *Impl) HandleEvent(ctx context.Context, block *coretypes.ResultBlock, txHash common.Hash, event sdk.Event) error {
	return nil
}

// ExportBlock accepts a finalized block and persists then inside the database.
// An error is returned if write fails.
func (i *Impl) ExportBlock(block *coretypes.ResultBlock, events *coretypes.ResultBlockResults, txs []*types.Tx, getTmcValidators modules.GetTmcValidators) error {
	return nil
}

// ExtractEvent accepts the transaction and handles events contained inside the transaction.
func (i *Impl) ExtractEvent(ctx context.Context, block *coretypes.ResultBlock, txHash common.Hash, event sdk.Event) (map[string][]interface{}, error) {
	allSQL := make(map[string][]interface{})
	for _, module := range i.Modules {
		if eventModule, ok := module.(modules.EventModule); ok {
			sqls, err := eventModule.ExtractEvent(ctx, block, txHash, event)
			if err != nil {
				log.Errorw("failed to handle event", "module", module.Name(), "event", event, "error", err)
				return nil, err
			}
			for k, v := range sqls {
				allSQL[k] = v
			}
		}
	}
	return allSQL, nil
}

// Process fetches a block for a given height and associated metadata and export it to a database.
// It returns an error if any export process fails.
func (i *Impl) Process(height uint64) error {
	// log.Debugw("processing block", "height", height)
	var block *coretypes.ResultBlock
	var events *coretypes.ResultBlockResults
	var txs []*types.Tx

	heightKey := fmt.Sprintf("%s-%d", i.GetServiceName(), height)
	blockAny, okb := blockMap.Load(heightKey)
	eventsAny, oke := eventMap.Load(heightKey)
	txsAny, okt := txMap.Load(heightKey)
	block, _ = blockAny.(*coretypes.ResultBlock)
	events, _ = eventsAny.(*coretypes.ResultBlockResults)
	txs, _ = txsAny.([]*types.Tx)
	if !okb || !oke || !okt {
		log.Warnf("failed to get map data height: %d", height)
		return ErrBlockNotFound
	}

	startTime := time.Now().UnixMilli()

	beginBlockEvents := events.BeginBlockEvents
	endBlockEvents := events.EndBlockEvents
	txCount := len(txs)
	eventCount := 0
	allSQL := make([]map[string][]interface{}, 0)

	// 1. handle events in startBlock
	if len(beginBlockEvents) > 0 {
		eventCount += len(beginBlockEvents)
		sqls, err := i.ExportEventsWithoutTx(context.Background(), block, beginBlockEvents)
		if err != nil {
			log.Errorf("failed to export events without tx: %s", err)
			return err
		}
		allSQL = append(allSQL, sqls...)
	}

	// 2. handle events in txs
	sqls, err := i.ExportEventsInTxs(context.Background(), block, txs)
	if err != nil {
		log.Errorf("failed to export events in txs: %s", err)
		return err
	}
	allSQL = append(allSQL, sqls...)

	// 3. handle events in endBlock
	if len(endBlockEvents) > 0 {
		eventCount += len(endBlockEvents)
		sqls, err = i.ExportEventsWithoutTx(context.Background(), block, endBlockEvents)
		if err != nil {
			log.Errorf("failed to export events without tx: %s", err)
			return err
		}
		allSQL = append(allSQL, sqls...)
	}

	sql, val := i.SaveEpoch(block)
	allSQL = append(allSQL, map[string][]interface{}{
		sql: val,
	})

	finalSQL := ""
	finalVal := make([]interface{}, 0)
	for _, m := range allSQL {
		for k, v := range m {
			finalSQL += fmt.Sprintf("%s;   ", k)
			finalVal = append(finalVal, v...)
		}
	}

	//log.Infof("SQL: %s", finalSQL)
	//log.Infof("param: %v", finalVal)
	tx := i.DB.Begin(context.TODO())
	if txErr := tx.Db.Session(&gorm.Session{DryRun: false}).Exec(finalSQL, finalVal...).Error; txErr != nil {
		log.Errorw("failed to exec sql", "error", txErr)
		tx.Rollback()
		return txErr
	}

	if txErr := tx.Commit(); txErr != nil {
		log.Errorw("failed to commit db", "error", txErr)
		return txErr
	}

	log.Infof("handle&write data cost: %d", time.Now().UnixMilli()-startTime)
	log.Infof("height :%d tx count:%d event count:%d", height, txCount, eventCount)

	blockMap.Delete(heightKey)
	eventMap.Delete(heightKey)
	txMap.Delete(heightKey)
	i.ProcessedHeight = height

	cost := time.Now().UnixMilli() - startTime
	log.Infof("total cost: %d", cost)
	metrics.BlockHeightLagGauge.WithLabelValues("blocksyncer").Set(float64(block.Block.Height))
	metrics.BlocksyncerCatchTime.WithLabelValues(fmt.Sprintf("%d", height)).Set(float64(cost))

	return nil
}

// SaveEpoch accept a block result data and persist basic info into db to record current sync progress
func (i *Impl) SaveEpoch(block *coretypes.ResultBlock) (string, []interface{}) {
	log.Infof(common.BytesToHash(block.BlockID.Hash).String())
	return localDB.Cast(i.DB).SaveEpochToSQL(context.Background(), &models.Epoch{
		OneRowId:    true,
		BlockHeight: block.Block.Height,
		BlockHash:   common.BytesToHash(block.BlockID.Hash),
		UpdateTime:  block.Block.Time.Unix(),
	})
}

// ExportTxs accepts a slice of transactions and persists then inside the database.
// An error is returned if write fails.
func (i *Impl) ExportTxs(block *coretypes.ResultBlock, txs []*types.Tx) error {
	return nil
}

// ExportCommit accepts ResultValidators and persists validator commit signatures inside the database.
// An error is returned if write fails.
func (i *Impl) ExportCommit(block *coretypes.ResultBlock, getTmcValidators modules.GetTmcValidators) error {
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

type TxHashEvent struct {
	Event  sdk.Event
	TxHash common.Hash
}

// ExportEventsInTxs accepts a slice of events in tx in order to save in database.
func (i *Impl) ExportEventsInTxs(ctx context.Context, block *coretypes.ResultBlock, txs []*types.Tx) ([]map[string][]interface{}, error) {
	allSQL := make([]map[string][]interface{}, 0)
	for _, tx := range txs {
		txHash := common.HexToHash(tx.TxHash)
		for _, event := range tx.Events {
			sqls, err := i.ExtractEvent(ctx, block, txHash, sdk.Event(event))
			if err != nil {
				return nil, err
			}
			allSQL = append(allSQL, sqls)
		}
	}
	return allSQL, nil
}

// ExportEventsWithoutTx accepts a slice of events not in tx in order to save in database.
// events here don't have txHash
func (i *Impl) ExportEventsWithoutTx(ctx context.Context, block *coretypes.ResultBlock, events []abci.Event) ([]map[string][]interface{}, error) {
	// call the event handlers
	allSQL := make([]map[string][]interface{}, 0)
	for _, event := range events {
		sqls, err := i.ExtractEvent(ctx, block, common.Hash{}, sdk.Event(event))
		if err != nil {
			return nil, err
		}
		allSQL = append(allSQL, sqls)
	}
	return allSQL, nil
}

// HandleGenesis accepts a GenesisDoc and calls all the registered genesis handlers in the order in which they have been registered.
func (i *Impl) HandleGenesis(genesisDoc *tmtypes.GenesisDoc, appState map[string]json.RawMessage) error {
	return nil
}

// HandleBlock accepts block and calls the block handlers.
func (i *Impl) HandleBlock(block *coretypes.ResultBlock, events *coretypes.ResultBlockResults, txs []*types.Tx, getTmcValidators modules.GetTmcValidators) {
	for _, module := range i.Modules {
		if blockModule, ok := module.(modules.BlockModule); ok {
			err := blockModule.HandleBlock(block, events, txs, getTmcValidators)
			if err != nil {
				log.Errorw("error while handling block", "module", module.Name(), "height", block.Block.Height, "err", err)
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
	// log.Infof("epoch height:%d, cur height: %d", ep.BlockHeight, height)
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

func (i *Impl) CreateMasterTable() error {

	return nil
}

func (i *Impl) GetServiceName() string {
	return i.ServiceName
}
