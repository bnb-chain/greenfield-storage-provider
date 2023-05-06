package blocksyncer

import (
	"context"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/model"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"

	"github.com/forbole/juno/v4/parser"
	"github.com/forbole/juno/v4/types"
	"github.com/forbole/juno/v4/types/config"
)

// enqueueNewBlocks enqueues new block heights onto the provided queue.
func (s *BlockSyncer) enqueueNewBlocks(exportQueue types.HeightQueue, ctx *parser.Context, currHeight uint64) {
	// Enqueue upcoming heights
	for {
		latestBlockHeightAny := Cast(s.parserCtx.Indexer).GetLatestBlockHeight().Load()
		latestBlockHeight := latestBlockHeightAny.(int64)
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

func Cast(indexer parser.Indexer) *Impl {
	s, ok := indexer.(*Impl)
	if !ok {
		panic("cannot cast")
	}
	return s
}

func CheckProgress() {
	for {
		epochMaster, err := MainService.parserCtx.Database.GetEpoch(context.TODO())
		if err != nil {
			continue
		}
		epochSlave, err := BackupService.parserCtx.Database.GetEpoch(context.TODO())
		if err != nil {
			continue
		}
		if epochMaster.BlockHeight-epochSlave.BlockHeight < model.DefaultBlockHeightDiff {
			SwitchMasterDBFlag()
			MainService.Stop(context.Background())
			break
		}
		time.Sleep(time.Minute * model.DefaultCheckDiffPeriod)
	}
}

func SwitchMasterDBFlag() error {
	masterFlag, err := FlagDB.GetMasterDB(context.TODO())
	if err != nil {
		return err
	}

	//switch flag
	masterFlag.IsMaster = masterFlag.IsMaster != true
	if err = FlagDB.SetMasterDB(context.TODO(), masterFlag); err != nil {
		return err
	}
	log.Infof("DB switched")
	return nil
}

func determineMainService() error {
	masterFlag, err := FlagDB.GetMasterDB(context.TODO())
	if err != nil {
		return err
	}
	if masterFlag.IsMaster {
		return nil
	} else {
		//switch role
		temp := MainService
		MainService = BackupService
		BackupService = temp
	}
	return nil
}
