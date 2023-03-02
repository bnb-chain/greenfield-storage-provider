package blocksyncer

import (
	"context"
	"errors"
	"sync/atomic"

	tomlconfig "github.com/forbole/juno/v4/cmd/migrate/toml"
	"github.com/forbole/juno/v4/parser/blocksyncer"
	"github.com/forbole/juno/v4/parser/explorer"
	"github.com/forbole/juno/v4/types"

	"github.com/forbole/juno/v4/cmd"
	parsecmdtypes "github.com/forbole/juno/v4/cmd/parse/types"
	"github.com/forbole/juno/v4/modules/messages"
	"github.com/forbole/juno/v4/modules/registrar"

	"github.com/forbole/juno/v4/types/config"

	"github.com/bnb-chain/greenfield-storage-provider/util/log"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/forbole/juno/v4/modules"
	"github.com/forbole/juno/v4/parser"
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
	log.Info(s.config.Node)
	log.Info(config.Cfg.Node)
	var ctx *parser.Context
	ctx, err := parsecmdtypes.GetParserContext(config.Cfg, cmdCfg)
	if err != nil {
		panic(err)
	}
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
		return errors.New("stone hub has already started")
	}
	go s.serve(ctx)
	return nil
}

// Stop running SyncerService
func (s *BlockSyncer) Stop(ctx context.Context) error {
	if !s.running.Swap(false) {
		return merrors.ErrSyncerStopped
	}
	return nil
}

// serve start syncer rpc service
func (s *BlockSyncer) serve(ctx context.Context) {
	// Create a queue that will collect, aggregate, and export blocks and metadata
	exportQueue := types.NewQueue(25)

	// Create workers
	workers := make([]parser.Worker, config.Cfg.Parser.Workers)
	for i := range workers {
		commonWorker := parser.NewWorker(s.parserCtx, exportQueue, i, config.Cfg.Parser.ConcurrentSync, config.Cfg.Parser.WorkerType)
		switch config.Cfg.Parser.WorkerType {
		case config.BlockSyncerWorkerType:
			workers[i] = &blocksyncer.Worker{CommonWorker: commonWorker}
		case config.ExplorerWorkerType:
			workers[i] = &explorer.Worker{CommonWorker: commonWorker}
		}
	}

	// Start each blocking worker in a go-routine where the worker consumes jobs
	// off of the export queue.
	for i, w := range workers {
		log.Debugw("starting worker...", "number", i+1)
		go w.Start(w)
	}

	if config.Cfg.Parser.ParseOldBlocks {
		if config.Cfg.Parser.ConcurrentSync {
			go enqueueMissingBlocks(exportQueue, s.parserCtx)
		} else {
			enqueueMissingBlocks(exportQueue, s.parserCtx)
		}
	}

	if config.Cfg.Parser.ParseNewBlocks {
		go enqueueNewBlocks(exportQueue, s.parserCtx)
	}
}
