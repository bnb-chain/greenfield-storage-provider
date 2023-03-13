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

// Syncer synchronizes ec data to piece store
type BlockSyncer struct {
	config    *tomlconfig.TomlConfig
	name      string
	parserCtx *parser.Context
	// store   client.PieceStoreAPI
	// metaDB  spdb.MetaDB // storage provider meta db
	running atomic.Bool
}

// NewSyncerService creates a syncer service to upload piece to piece store
func NewBlockSyncerService(cfg *tomlconfig.TomlConfig) (*BlockSyncer, error) {
	log.Info(cfg.Database.Type)
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

// initClient
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
		log.Info("readErr: %v", readErr)
		return readErr
	}
	// read DSN from env
	dsn, envErr := getDBConfigFromEnv(DsnBlockSyncer)
	if envErr != nil {
		log.Errorf("readErr: %v", envErr)
		return envErr
	}
	if dsn != "" {
		config.Cfg.Database.DSN = dsn
	}
	log.Info(s.config.Node)
	log.Info(config.Cfg.Node)
	var ctx *parser.Context
	ctx, err := parsecmdtypes.GetParserContext(config.Cfg, cmdCfg)
	if err != nil {
		panic(err)
	}
	ctx.Indexer = NewIndexer(ctx.EncodingConfig.Marshaler,
		ctx.Node,
		ctx.Database,
		ctx.Modules)

	s.parserCtx = ctx
	//s.config = config.Cfg.Parser
	return nil
}

// initDB init a meta-db instance
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

// Name describes the name of SyncerService
func (s *BlockSyncer) Name() string {
	return s.name
}

// Start running SyncerService
func (s *BlockSyncer) Start(ctx context.Context) error {
	if s.running.Swap(true) {
		return errors.New("block syncer has already started")
	}
	go s.serve(ctx)
	return nil
}

// Stop running SyncerService
func (s *BlockSyncer) Stop(ctx context.Context) error {
	if !s.running.Swap(false) {
		return nil
	}
	return nil
}

// serve start syncer rpc service
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
		panic(err)
	}
	go enqueueNewBlocks(exportQueue, s.parserCtx, latestBlockHeight)
}
