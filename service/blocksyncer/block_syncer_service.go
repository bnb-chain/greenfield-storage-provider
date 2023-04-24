package blocksyncer

import (
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/model"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"

	"github.com/forbole/juno/v4/parser"
	"github.com/forbole/juno/v4/types"
	"github.com/forbole/juno/v4/types/config"
)

// enqueueNewBlocks enqueues new block heights onto the provided queue.
func enqueueNewBlocks(exportQueue types.HeightQueue, ctx *parser.Context, currHeight uint64) {
	// Enqueue upcoming heights
	for {
		latestBlockHeight := LatestBlockHeight.Load()
		// Enqueue all heights from the current height up to the latest height
		for ; currHeight <= uint64(latestBlockHeight); currHeight++ {
			log.Debugw("enqueueing new block", "height", currHeight)
			exportQueue <- currHeight
		}
		time.Sleep(config.GetAvgBlockTime())
	}
}

// mustGetLatestHeight tries getting the latest height from the RPC client.
// If no latest height can be found after MaxRetryCount, it returns 0.
func mustGetLatestHeight(ctx *parser.Context) uint64 {
	for retryCount := 0; retryCount < model.MaxRetryCount; retryCount++ {
		latestBlockHeight, err := ctx.Node.LatestHeight()
		if err == nil {
			return uint64(latestBlockHeight)
		}

		log.Errorw("failed to get last block from RPCConfig client", "error", err, "retry interval", config.GetAvgBlockTime(), "retry count", retryCount)

		time.Sleep(config.GetAvgBlockTime())
	}

	return 0
}
