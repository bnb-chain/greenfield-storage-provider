package blocksyncer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/juno/v4/common"
	"github.com/forbole/juno/v4/database"
	"github.com/forbole/juno/v4/models"
	"github.com/forbole/juno/v4/modules"
	"github.com/forbole/juno/v4/modules/bucket"
	"github.com/forbole/juno/v4/modules/group"
	"github.com/forbole/juno/v4/modules/object"
	"github.com/forbole/juno/v4/modules/payment"
	"github.com/forbole/juno/v4/modules/permission"
	storageprovider "github.com/forbole/juno/v4/modules/storage_provider"
	virtualgroup "github.com/forbole/juno/v4/modules/virtual_group"
	"github.com/forbole/juno/v4/node"
	"github.com/forbole/juno/v4/parser"
	"github.com/forbole/juno/v4/types"

	localdb "github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/database"
	spExit "github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/modules/events"
	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer/modules/prefixtree"
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
	//LocalDb localDB.DB

	LatestBlockHeight atomic.Value
	ProcessedHeight   uint64

	eventTypeCount int

	ServiceName string
}

// ExportBlock accepts a finalized block and persists then inside the database.
// An error is returned if write fails.
func (i *Impl) ExportBlock(block *coretypes.ResultBlock, events *coretypes.ResultBlockResults, txs []*types.Tx, getTmcValidators modules.GetTmcValidators) error {
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

func (i *Impl) ExtractEvent(ctx context.Context, block *coretypes.ResultBlock, txHash common.Hash, event sdk.Event) (interface{}, error) {
	for _, module := range i.Modules {
		if eventModule, ok := module.(modules.EventModule); ok {
			data, err := eventModule.ExtractEvent(ctx, block, txHash, event)
			if err != nil {
				return nil, err
			}
			if data != nil {
				return data, nil
			}
		}
	}
	return nil, nil
}

// Process fetches a block for a given height and associated metadata and export it to a database.
// It returns an error if any export process fails.
func (i *Impl) Process(height uint64) error {
	// log.Debugw("processing block", "height", height)
	var block *coretypes.ResultBlock
	var events *coretypes.ResultBlockResults
	var txs []*types.Tx
	var err error
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

	// 1. handle events in startBlock

	if len(beginBlockEvents) > 0 {
		eventCount += len(beginBlockEvents)
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
		eventCount += len(endBlockEvents)
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

	log.Infof("handle&write data cost: %d", time.Now().UnixMilli()-startTime)
	log.Infof("height :%d tx count:%d event count:%d", height, txCount, eventCount)

	blockMap.Delete(heightKey)
	eventMap.Delete(heightKey)
	txMap.Delete(heightKey)
	i.ProcessedHeight = height

	cost := time.Now().UnixMilli() - startTime
	log.Infof("total cost: %d", cost)
	metrics.BlocksyncerCatchTime.WithLabelValues(fmt.Sprintf("%d", height)).Set(float64(cost))

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
func (i *Impl) ExportEventsInTxs(ctx context.Context, block *coretypes.ResultBlock, txs []*types.Tx) error {
	bucketEvent := make([]TxHashEvent, 0)
	groupEvent := make([]TxHashEvent, 0)
	permissionEvent := make([]TxHashEvent, 0)
	spEvent := make([]TxHashEvent, 0)
	prefixEvent := make([]TxHashEvent, 0)
	exitEvent := make([]TxHashEvent, 0)
	virtualGroupEvent := make([]TxHashEvent, 0)
	objectEvent := make([]TxHashEvent, 0)
	dataList := make([]interface{}, 0)

	for _, tx := range txs {
		txHash := common.HexToHash(tx.TxHash)
		for _, event := range tx.Events {
			e := TxHashEvent{Event: sdk.Event(event), TxHash: txHash}
			if bucket.BucketEvents[event.Type] {
				bucketEvent = append(bucketEvent, e)
			} else if group.GroupEvents[event.Type] {
				groupEvent = append(groupEvent, e)
			} else if permission.PolicyEvents[event.Type] {
				permissionEvent = append(permissionEvent, e)
			} else if storageprovider.StorageProviderEvents[event.Type] {
				spEvent = append(spEvent, e)
			} else if virtualgroup.VirtualGroupEvents[event.Type] {
				virtualGroupEvent = append(virtualGroupEvent, e)
			} else if spExit.SpExitEvents[event.Type] {
				exitEvent = append(exitEvent, e)
			} else if event.Type == object.EventDeleteObject || event.Type == object.EventCreateObject || payment.PaymentEvents[event.Type] {
				data, err := i.ExtractEvent(ctx, block, common.Hash{}, sdk.Event(event))
				if err != nil {
					return err
				}
				dataList = append(dataList, data)
				prefixEvent = append(prefixEvent, e)
			} else if object.ObjectEvents[event.Type] {
				objectEvent = append(objectEvent, e)
			}
			if prefixtree.BuildPrefixTreeEvents[event.Type] {
				prefixEvent = append(prefixEvent, e)
			}
		}
	}

	if len(dataList) > 0 {
		err := i.BatchHandle(ctx, dataList, block, common.Hash{})
		if err != nil {
			return err
		}
	}

	allEvents := make([][]TxHashEvent, 0)
	allEvents = append(allEvents, bucketEvent)
	allEvents = append(allEvents, groupEvent)
	allEvents = append(allEvents, objectEvent)
	allEvents = append(allEvents, permissionEvent)
	allEvents = append(allEvents, spEvent)
	allEvents = append(allEvents, prefixEvent)
	allEvents = append(allEvents, virtualGroupEvent)
	allEvents = append(allEvents, exitEvent)
	return i.concurrenceHandleEvent(ctx, block, allEvents)
}

func (i *Impl) BatchHandle(ctx context.Context, dataList []interface{}, block *coretypes.ResultBlock, txHash common.Hash) error {
	objects := make([]*models.Object, 0)
	objectIDMap := make(map[common.Hash]bool, 0)
	bucketNameMap := make(map[string][]common.Hash)
	streamRecords := make([]*models.StreamRecord, 0)
	paymentAccount := make([]*models.PaymentAccount, 0)

	for _, data := range dataList {
		switch m := data.(type) {
		case *models.Object:
			objects = append(objects, m)
			objectIDMap[m.ObjectID] = true
			tmp := bucketNameMap[m.BucketName]
			if tmp == nil {
				tmp = make([]common.Hash, 0)
			}
			tmp = append(tmp, m.ObjectID)
			bucketNameMap[m.BucketName] = tmp
		case *models.StreamRecord:
			streamRecords = append(streamRecords, m)
		case *models.PaymentAccount:
			paymentAccount = append(paymentAccount, m)
		}
	}

	localDB := localdb.Cast(i.DB)
	if len(paymentAccount) > 0 {
		if err := localDB.BatchSavePaymentAccount(ctx, paymentAccount); err != nil {
			return err
		}
	}
	if len(streamRecords) > 0 {
		if err := localDB.BatchSaveStreamRecord(ctx, streamRecords); err != nil {
			return err
		}
	}

	insert := make(map[string][]*models.Object)
	update := make(map[string][]*models.Object)

	objResults, err := localDB.GetObjectList(ctx, bucketNameMap)
	if err != nil {
		return err
	}
	resMap := make(map[common.Hash]*models.Object)
	for _, obj := range objResults {
		resMap[obj.ObjectID] = obj
	}

	for _, obj := range objects {
		res, ok := resMap[obj.ObjectID]
		if !ok {
			tmp := insert[obj.BucketName]
			if tmp == nil {
				tmp = make([]*models.Object, 0)
			}
			tmp = append(tmp, obj)
			insert[obj.BucketName] = tmp
			continue
		}
		if obj.Removed {
			res.Removed = true
			res.BucketName = obj.BucketName
			res.ObjectName = obj.ObjectName
			res.ObjectID = obj.ObjectID
			res.LocalVirtualGroupId = obj.LocalVirtualGroupId
			res.UpdateAt = block.Block.Height
			res.UpdateTxHash = txHash
			res.UpdateTime = time.Now().UTC().Unix()
		}
		resMap[obj.ObjectID] = res
	}

	if len(insert) > 0 {
		if err := localDB.BatchSaveObject(ctx, insert); err != nil {
			return err
		}
	}
	for _, v := range resMap {
		tmp := update[v.BucketName]
		if tmp == nil {
			tmp = make([]*models.Object, 0)
		}
		tmp = append(tmp, v)
		update[v.BucketName] = tmp
	}
	if len(update) > 0 {
		if err := localDB.BatchSaveObject(ctx, update); err != nil {
			return err
		}
	}
	return nil
}

func (i *Impl) concurrenceHandleEvent(ctx context.Context, block *coretypes.ResultBlock, allEvents [][]TxHashEvent) error {
	wg := &sync.WaitGroup{}
	wg.Add(len(allEvents))
	var handleErr error
	for _, events := range allEvents {
		go func(event []TxHashEvent) {
			defer wg.Done()
			for _, e := range event {
				if err := i.HandleEvent(ctx, block, e.TxHash, e.Event); err != nil {
					log.Errorf("failed to HandleEvent err:%v", err)
					handleErr = err
					return
				}
			}
		}(events)
	}
	wg.Wait()
	return handleErr
}

// ExportEventsWithoutTx accepts a slice of events not in tx in order to save in database.
// events here don't have txHash
func (i *Impl) ExportEventsWithoutTx(ctx context.Context, block *coretypes.ResultBlock, events []abci.Event) error {
	bucketEvent := make([]TxHashEvent, 0)
	groupEvent := make([]TxHashEvent, 0)
	permissionEvent := make([]TxHashEvent, 0)
	spEvent := make([]TxHashEvent, 0)
	prefixEvent := make([]TxHashEvent, 0)
	exitEvent := make([]TxHashEvent, 0)
	VirtualGroupEvent := make([]TxHashEvent, 0)
	objectEvent := make([]TxHashEvent, 0)
	dataList := make([]interface{}, 0)

	for _, event := range events {
		e := TxHashEvent{Event: sdk.Event(event)}
		if bucket.BucketEvents[event.Type] {
			bucketEvent = append(bucketEvent, e)
		} else if group.GroupEvents[event.Type] {
			groupEvent = append(groupEvent, e)
		} else if permission.PolicyEvents[event.Type] {
			permissionEvent = append(permissionEvent, e)
		} else if storageprovider.StorageProviderEvents[event.Type] {
			spEvent = append(spEvent, e)
		} else if virtualgroup.VirtualGroupEvents[event.Type] {
			VirtualGroupEvent = append(VirtualGroupEvent, e)
		} else if spExit.SpExitEvents[event.Type] {
			exitEvent = append(exitEvent, e)
		} else if event.Type == object.EventDeleteObject || event.Type == object.EventCreateObject || payment.PaymentEvents[event.Type] || virtualgroup.VirtualGroupEvents[event.Type] {
			data, err := i.ExtractEvent(ctx, block, common.Hash{}, e.Event)
			if err != nil {
				return err
			}
			dataList = append(dataList, data)
			prefixEvent = append(prefixEvent, e)
		} else if object.ObjectEvents[event.Type] {
			objectEvent = append(objectEvent, e)
		}
		if prefixtree.BuildPrefixTreeEvents[event.Type] {
			prefixEvent = append(prefixEvent, e)
		}
	}

	if len(dataList) > 0 {
		err := i.BatchHandle(ctx, dataList, block, common.Hash{})
		if err != nil {
			return err
		}
	}

	allEvents := make([][]TxHashEvent, 0)
	allEvents = append(allEvents, bucketEvent)
	allEvents = append(allEvents, groupEvent)
	allEvents = append(allEvents, objectEvent)
	allEvents = append(allEvents, permissionEvent)
	allEvents = append(allEvents, spEvent)
	allEvents = append(allEvents, prefixEvent)
	allEvents = append(allEvents, VirtualGroupEvent)
	allEvents = append(allEvents, exitEvent)

	return i.concurrenceHandleEvent(ctx, block, allEvents)
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
