package blocksyncer

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	tomlconfig "github.com/forbole/juno/v4/cmd/migrate/toml"
	"github.com/forbole/juno/v4/types"

	"github.com/forbole/juno/v4/cmd"
	parsecmdtypes "github.com/forbole/juno/v4/cmd/parse/types"
	"github.com/forbole/juno/v4/modules/messages"
	"github.com/forbole/juno/v4/modules/registrar"

	"github.com/forbole/juno/v4/types/config"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"

	"github.com/forbole/juno/v4/modules"
	"github.com/forbole/juno/v4/parser"

	"github.com/bnb-chain/greenfield-storage-provider/model"
)

// BlockSyncer synchronizes storage,payment,permission data to db by handling related events
type BlockSyncer struct {
	config            *tomlconfig.TomlConfig
	name              string
	parserCtx         *parser.Context
	running           atomic.Value
	latestBlockHeight uint64
}

var (
	blockMap *sync.Map
	eventMap *sync.Map
)

// NewBlockSyncerService create a BlockSyncer service to index block events data to db
func NewBlockSyncerService(cfg *tomlconfig.TomlConfig) (*BlockSyncer, error) {
	s := &BlockSyncer{
		config: cfg,
		name:   model.BlockSyncerService,
	}
	blockMap = new(sync.Map)
	eventMap = new(sync.Map)
	if err := s.initClient(); err != nil {
		return nil, err
	}
	// init meta db
	if err := s.initDB(); err != nil {
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
	log.Infof("cmd cfg: %+v", cmdCfg)
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
	latestBlockHeight := mustGetLatestHeight(s.parserCtx)
	s.latestBlockHeight = latestBlockHeight
	s.parserCtx.Indexer = NewIndexer(ctx.EncodingConfig.Marshaler,
		ctx.Node,
		ctx.Database,
		ctx.Modules, s.latestBlockHeight)
	return nil
}

// initDB create tables needed by block syncer. It depends on which modules are configured
func (s *BlockSyncer) initDB() error {
	// Prepare tables
	var err error
	for _, module := range s.parserCtx.Modules {
		if module, ok := module.(modules.PrepareTablesModule); ok {
			err = module.PrepareTables()
			if err != nil {
				log.Errorw("failed to prepare tables", "error", err)
				return err
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
	go s.quickFetchBlockData(lastDbBlockHeight+1, s.latestBlockHeight)

	// Start each blocking worker in a go-routine where the worker consumes jobs
	// off of the export queue.
	go worker.Start()

	latestBlockHeight, err := enqueueMissingBlocks(exportQueue, s.parserCtx, lastDbBlockHeight)
	if err != nil {
		log.Errorf("failed to enqueue missing blocks error: %v", err)
	}
	go enqueueNewBlocks(exportQueue, s.parserCtx, latestBlockHeight)
}

func (s *BlockSyncer) quickFetchBlockData(startHeight, endHeight uint64) {
	count := uint64(s.config.Parser.Workers)
	for i := uint64(0); i < count; i++ {
		go func(idx uint64) {
			for cycle := uint64(0); ; cycle++ {
				height := idx + count*cycle + startHeight
				if height > endHeight {
					return
				}
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
				blockMap.Store(height, block)
				eventMap.Store(height, events)
			}
		}(i)
	}
}
