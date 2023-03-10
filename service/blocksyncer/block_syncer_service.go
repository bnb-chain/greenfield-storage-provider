package blocksyncer

import (
	"context"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	eventutil "github.com/forbole/juno/v4/types/event"

	"github.com/forbole/juno/v4/parser"
	"github.com/forbole/juno/v4/types"
	"github.com/forbole/juno/v4/types/config"
	"github.com/forbole/juno/v4/types/utils"
	abci "github.com/tendermint/tendermint/abci/types"
)

// enqueueMissingBlocks enqueues jobs (block heights) for missed blocks starting
// at the startHeight up until the latest known height.
func enqueueMissingBlocks(exportQueue types.HeightQueue, ctx *parser.Context) (uint64, error) {
	// Get the config
	cfg := config.Cfg.Parser

	// Get the latest height
	latestBlockHeight := mustGetLatestHeight(ctx)

	epoch, err := ctx.Database.GetEpoch(context.TODO())
	if err != nil {
		log.Errorw("failed to get last block height from database", "error", err)
		return 0, err
	}
	lastDbBlockHeight := uint64(epoch.BlockHeight)
	// Get the start height, default to the config's height
	startHeight := cfg.StartHeight

	// Set startHeight to the latest height in database
	// if is not set inside config.yaml file
	if startHeight == 0 {
		startHeight = utils.MaxUint64(0, lastDbBlockHeight)
	}

	log.Infow("syncing missing blocks...", "latest_block_height", latestBlockHeight)
	for i := lastDbBlockHeight + 1; i <= latestBlockHeight; i++ {
		log.Debugw("enqueueing missing block", "height", i)
		exportQueue <- i
	}

	return latestBlockHeight + 1, nil
}

// enqueueNewBlocks enqueues new block heights onto the provided queue.
func enqueueNewBlocks(exportQueue types.HeightQueue, ctx *parser.Context, currHeight uint64) {
	// Enqueue upcoming heights
	for {
		latestBlockHeight := mustGetLatestHeight(ctx)

		// Enqueue all heights from the current height up to the latest height
		for ; currHeight <= latestBlockHeight; currHeight++ {
			log.Debugw("enqueueing new block", "height", currHeight)
			exportQueue <- currHeight
		}
		time.Sleep(config.GetAvgBlockTime())
	}
}

// mustGetLatestHeight tries getting the latest height from the RPC client.
// If after 50 tries no latest height can be found, it returns 0.
func mustGetLatestHeight(ctx *parser.Context) uint64 {
	for retryCount := 0; retryCount < 50; retryCount++ {
		latestBlockHeight, err := ctx.Node.LatestHeight()
		if err == nil {
			return uint64(latestBlockHeight)
		}

		log.Errorw("failed to get last block from RPCConfig client",
			"err", err,
			"retry interval", config.GetAvgBlockTime(),
			"retry count", retryCount)

		time.Sleep(config.GetAvgBlockTime())
	}

	return 0
}

func filterEventsType(tx *abci.ResponseDeliverTx) []sdk.Event {
	filteredEvents := make([]sdk.Event, 0)
	for _, event := range tx.Events {
		if _, ok := eventutil.EventProcessedMap[event.Type]; ok {
			filteredEvents = append(filteredEvents, sdk.Event(event))
		}
	}
	return filteredEvents
}
