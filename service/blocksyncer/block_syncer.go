package blocksyncer

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/forbole/juno/v4/cmd"
	tomlconfig "github.com/forbole/juno/v4/cmd/migrate/toml"
	parsecmdtypes "github.com/forbole/juno/v4/cmd/parse/types"
	"github.com/forbole/juno/v4/modules"
	"github.com/forbole/juno/v4/modules/messages"
	"github.com/forbole/juno/v4/modules/registrar"
	"github.com/forbole/juno/v4/parser"
	"github.com/forbole/juno/v4/types"
	"github.com/forbole/juno/v4/types/config"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// BlockSyncer synchronizes storage,payment,permission data to db by handling related events
type BlockSyncer struct {
	config    *tomlconfig.TomlConfig
	name      string
	parserCtx *parser.Context
	running   atomic.Value
}

// Read concurrency required global variables
var (
	blockMap *sync.Map
	eventMap *sync.Map
	txMap    *sync.Map

	LatestBlockHeight atomic.Value
	CatchUpFlag       atomic.Value
)

// NewBlockSyncerService create a BlockSyncer service to index block events data to db
func NewBlockSyncerService(cfg *tomlconfig.TomlConfig) (*BlockSyncer, error) {
	s := &BlockSyncer{
		config: cfg,
		name:   model.BlockSyncerService,
	}
	blockMap = new(sync.Map)
	eventMap = new(sync.Map)
	txMap = new(sync.Map)
	if err := s.initClient(); err != nil {
		return nil, err
	}
	// init meta db
	if err := s.initDB(cfg.RecreateTables); err != nil {
		return nil, err
	}
	return s, nil
}

// initClient initialize a juno client using given configs
func (s *BlockSyncer) initClient() error {
	// JunoConfig the runner
	junoConfig := cmd.NewConfig("juno").
		WithParseConfig(parsecmdtypes.NewConfig().
			WithRegistrar(registrar.NewDefaultRegistrar(
				messages.CosmosMessageAddressesParser,
			)).WithFileType("toml"),
		)
	cmdCfg := junoConfig.GetParseConfig()
	cmdCfg.WithTomlConfig(s.config)

	if readErr := parsecmdtypes.ReadConfigPreRunE(cmdCfg)(nil, nil); readErr != nil {
		log.Infof("readErr: %v", readErr)
		return readErr
	}
	// get DSN from env first
	dsn, envErr := getDBConfigFromEnv(model.DsnBlockSyncer)
	if envErr != nil {
		log.Info("failed to get db config from env, use db config from config file")
	}
	if dsn != "" {
		log.Info("use db config from env")
		config.Cfg.Database.DSN = dsn
	}

	var ctx *parser.Context
	ctx, err := parsecmdtypes.GetParserContext(config.Cfg, cmdCfg)
	if err != nil {
		log.Errorf("failed to GetParserContext err: %v", err)
		return err
	}
	s.parserCtx = ctx
	s.parserCtx.Indexer = NewIndexer(ctx.EncodingConfig.Marshaler,
		ctx.Node,
		ctx.Database,
		ctx.Modules)
	return nil
}

// initDB create tables needed by block syncer. It depends on which modules are configured
func (s *BlockSyncer) initDB(recreateTables bool) error {

	var err error
	// drop tables if needed
	if recreateTables {
		for _, module := range s.parserCtx.Modules {
			if module, ok := module.(modules.PrepareTablesModule); ok {
				err = module.RecreateTables()
				if err != nil {
					log.Errorw("failed to recreate tables", "error", err)
					return err
				}
			}
		}
	} else {
		// Prepare tables without recreate
		for _, module := range s.parserCtx.Modules {
			if module, ok := module.(modules.PrepareTablesModule); ok {
				err = module.PrepareTables()
				if err != nil {
					log.Errorw("failed to prepare tables", "error", err)
					return err
				}
			}
		}
	}
	return nil
}

// Name describes the name of BlockSyncer service
func (s *BlockSyncer) Name() string {
	return s.name
}

// Start running BlockSyncer service
func (s *BlockSyncer) Start(ctx context.Context) error {
	if s.running.Swap(true) == true {
		return errors.New("block syncer hub has already started")
	}
	go s.serve(ctx)
	return nil
}

// Stop running BlockSyncer service
func (s *BlockSyncer) Stop(ctx context.Context) error {
	if s.running.Swap(false) == false {
		return nil
	}
	return nil
}

// serve start BlockSyncer rpc service
func (s *BlockSyncer) serve(ctx context.Context) {
	// Create a queue that will collect, aggregate, and export blocks and metadata
	exportQueue := types.NewQueue(25)

	// Create workers
	worker := parser.NewWorker(s.parserCtx, exportQueue, 0, config.Cfg.Parser.ConcurrentSync)
	worker.SetIndexer(s.parserCtx.Indexer)

	latestBlockHeight := mustGetLatestHeight(s.parserCtx)
	LatestBlockHeight.Store(int64(latestBlockHeight))
	CatchUpFlag.Store(int64(-1))
	go s.getLatestBlockHeight()

	lastDbBlockHeight := uint64(0)
	for {
		epoch, err := s.parserCtx.Database.GetEpoch(context.TODO())
		if err != nil {
			log.Errorw("failed to get last block height from database", "error", err)
			continue
		}
		lastDbBlockHeight = uint64(epoch.BlockHeight)
		break
	}

	// fetch block data
	go s.quickFetchBlockData(lastDbBlockHeight + 1)

	go enqueueNewBlocks(exportQueue, s.parserCtx, lastDbBlockHeight+1)

	// Start each blocking worker in a go-routine where the worker consumes jobs
	// off of the export queue.
	time.Sleep(time.Second)
	go worker.Start()
}

func (s *BlockSyncer) getLatestBlockHeight() {
	for {
		latestBlockHeight, err := s.parserCtx.Node.LatestHeight()
		if err != nil {
			log.Errorw("failed to get last block from RPCConfig client",
				"err", err,
				"retry interval", config.GetAvgBlockTime())
		}
		LatestBlockHeight.Store(latestBlockHeight)

		time.Sleep(config.GetAvgBlockTime())
	}
}

func (s *BlockSyncer) quickFetchBlockData(startHeight uint64) {
	count := uint64(s.config.Parser.Workers)
	for cycle := uint64(0); ; cycle++ {
		latestBlockHeightAny := LatestBlockHeight.Load()
		latestBlockHeight := latestBlockHeightAny.(int64)
		if latestBlockHeight < int64(count*(cycle+1)+startHeight-1) {
			log.Infof("quick fetch ended latestBlockHeight: %d", latestBlockHeight)
			CatchUpFlag.Store(int64(count*cycle + startHeight - 1))
			break
		}
		wg := &sync.WaitGroup{}
		wg.Add(int(count))
		for i := uint64(0); i < count; i++ {
			go func(idx, c uint64) {
				defer wg.Done()
				height := idx + count*c + startHeight
				if height > uint64(latestBlockHeight) {
					return
				}
				for {
					block, err := s.parserCtx.Node.Block(int64(height))
					if err != nil {
						log.Warnf("failed to get block from node: %s", err)
						continue
					}

					events, err := s.parserCtx.Node.BlockResults(int64(height))
					if err != nil {
						log.Warnf("failed to get block results from node: %s", err)
						continue
					}
					txs, err := s.parserCtx.Node.Txs(block)
					if err != nil {
						log.Warnf("failed to get block results from node: %s", err)
						continue
					}
					blockMap.Store(height, block)
					eventMap.Store(height, events)
					txMap.Store(height, txs)
					break
				}
			}(i, cycle)
		}
		wg.Wait()
	}
}
