package blocksyncer

import (
	"context"
	"errors"
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
	config    *tomlconfig.TomlConfig
	name      string
	parserCtx *parser.Context
	running   atomic.Bool
}

// NewBlockSyncerService create a BlockSyncer service to index block events data to db
func NewBlockSyncerService(cfg *tomlconfig.TomlConfig) (*BlockSyncer, error) {
	s := &BlockSyncer{
		config: cfg,
		name:   model.BlockSyncerService,
	}
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
	ctx.Indexer = NewIndexer(ctx.EncodingConfig.Marshaler,
		ctx.Node,
		ctx.Database,
		ctx.Modules)

	s.parserCtx = ctx
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
	if s.running.Swap(true) {
		return errors.New("block syncer has already started")
	}
	go s.serve(ctx)
	return nil
}

// Stop running BlockSyncer service
func (s *BlockSyncer) Stop(ctx context.Context) error {
	if !s.running.Swap(false) {
		return nil
	}
	return nil
}

// serve start BlockSyncer rpc service
func (s *BlockSyncer) serve(ctx context.Context) {
	// Create a queue that will collect, aggregate, and export blocks and metadata
	exportQueue := types.NewQueue(25)

	// Create workers
	workers := make([]*parser.Worker, config.Cfg.Parser.Workers)
	for i := range workers {
		workers[i] = parser.NewWorker(s.parserCtx, exportQueue, i, config.Cfg.Parser.ConcurrentSync)
		workers[i].SetIndexer(s.parserCtx.Indexer)
	}

	// Start each blocking worker in a go-routine where the worker consumes jobs
	// off of the export queue.
	for i, w := range workers {
		log.Debugw("starting worker...", "number", i+1)
		go w.Start()
	}
	latestBlockHeight, err := enqueueMissingBlocks(exportQueue, s.parserCtx)
	if err != nil {
		log.Errorf("failed to enqueue missing blocks error: %v", err)
	}
	go enqueueNewBlocks(exportQueue, s.parserCtx, latestBlockHeight)
}
