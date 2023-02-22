package blocksyncer

import (
	"context"
	"errors"
	"sync/atomic"

	tomlconfig "github.com/forbole/juno/v4/cmd/migrate/toml"

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
	"github.com/forbole/juno/v4/types"
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
	//var (
	//	metaDB spdb.MetaDB
	//	err    error
	//)
	//
	//metaDB, err = store.NewMetaDB(s.config.MetaDBType,
	//	s.config.MetaLevelDBConfig, s.config.MetaSqlDBConfig)
	//if err != nil {
	//	log.Errorw("failed to init metaDB", "err", err)
	//	return err
	//}
	//s.metaDB = metaDB
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
	s.serve()
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
func (s *BlockSyncer) serve() {
	exportQueue := types.NewQueue(25)
	// Create workers
	workers := make([]parser.Worker, s.config.Parser.Workers)
	for i := range workers {
		workers[i] = parser.NewWorker(s.parserCtx, exportQueue, i)
	}
	//waitGroup := &sync.WaitGroup{}
	//waitGroup.Add(1)

	// Run all the async operations
	for _, module := range s.parserCtx.Modules {
		if module, ok := module.(modules.AsyncOperationsModule); ok {
			go module.RunAsyncOperations()
		}
	}

	// Start each blocking worker in a go-routine where the worker consumes jobs
	// off of the export queue.
	for i, w := range workers {
		log.Debug("starting worker...", "number", i+1)
		go w.Start()
	}

	// Listen for and trap any OS signal to gracefully shutdown and exit
	//trapSignal(s.parserCtx, waitGroup)

	if s.config.Parser.ParseGenesis {
		// Add the genesis to the queue if requested
		exportQueue <- 0
	}

	if s.config.Parser.ParseOldBlocks {
		go enqueueMissingBlocks(exportQueue, s.parserCtx)
	}

	if s.config.Parser.ParseNewBlocks {
		go enqueueNewBlocks(exportQueue, s.parserCtx)
	}

	// Block main process (signal capture will call WaitGroup's Done)
	//waitGroup.Wait()

}
