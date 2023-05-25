package blocksyncer

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/forbole/juno/v4/cmd"
	tomlconfig "github.com/forbole/juno/v4/cmd/migrate/toml"
	parsecmdtypes "github.com/forbole/juno/v4/cmd/parse/types"
	databaseconfig "github.com/forbole/juno/v4/database/config"
	loggingconfig "github.com/forbole/juno/v4/log/config"
	"github.com/forbole/juno/v4/models"
	"github.com/forbole/juno/v4/modules"
	"github.com/forbole/juno/v4/modules/messages"
	"github.com/forbole/juno/v4/modules/registrar"
	"github.com/forbole/juno/v4/node/remote"
	"github.com/forbole/juno/v4/parser"
	parserconfig "github.com/forbole/juno/v4/parser/config"
	"github.com/forbole/juno/v4/types"
	"github.com/forbole/juno/v4/types/config"
	"gorm.io/gorm/schema"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

func NewBlockSyncerModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	junoCfg := makeBlockSyncerConfig(cfg)
	MainService = &BlockSyncerModular{
		config:  junoCfg,
		name:    BlockSyncerModularName,
		baseApp: app,
	}
	blockMap = new(sync.Map)
	eventMap = new(sync.Map)
	txMap = new(sync.Map)
	NeedBackup = junoCfg.EnableDualDB
	if err := MainService.initClient(); err != nil {
		return nil, err
	}

	//prepare master table
	FlagDB = MainService.parserCtx.Database
	MainService.prepareMasterFlagTable()
	mainServiceDB, _ := FlagDB.GetMasterDB(context.TODO())
	mainDBIsMaster := mainServiceDB.IsMaster

	// init main service db, if main service DB is not current master then recreate tables
	if err := MainService.initDB(!mainDBIsMaster || junoCfg.RecreateTables); err != nil {
		return nil, err
	}

	// when NeedBackup config true Or backup db is current master DB, init backup service
	if NeedBackup || !mainServiceDB.IsMaster {
		//create backup block syncer
		if blockSyncerBackup, err := newBackupBlockSyncerService(junoCfg, mainDBIsMaster); err != nil {
			return nil, err
		} else {
			BackupService = blockSyncerBackup
		}
	}

	return MainService, nil
}

// initClient initialize a juno client using given configs
func (s *BlockSyncerModular) initClient() error {
	// JunoConfig the runner
	junoConfig := cmd.NewConfig("juno").
		WithParseConfig(parsecmdtypes.NewConfig().
			WithRegistrar(registrar.NewDefaultRegistrar(
				messages.CosmosMessageAddressesParser,
			)).WithFileType("toml"),
		)
	cmdCfg := junoConfig.GetParseConfig()
	cmdCfg.WithTomlConfig(s.config)

	//set toml config to juno config
	if readErr := parsecmdtypes.ReadConfigPreRunE(cmdCfg)(nil, nil); readErr != nil {
		log.Infof("readErr: %v", readErr)
		return readErr
	}

	// get DSN from env first
	var dbEnv string
	if s.Name() == BlockSyncerModularName {
		dbEnv = model.DsnBlockSyncer
	} else {
		dbEnv = model.DsnBlockSyncerSwitched
	}

	dsn, envErr := getDBConfigFromEnv(dbEnv)
	if envErr != nil {
		log.Info("failed to get db config from env, use db config from config file")
		config.Cfg.Database.DSN = s.config.Database.DSN
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
		ctx.Modules,
		s.Name())
	return nil
}

// initDB create tables needed by block syncer. It depends on which modules are configured
func (s *BlockSyncerModular) initDB(recreateTables bool) error {

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

// serve start BlockSyncer rpc service
func (b *BlockSyncerModular) serve(ctx context.Context) {
	// Create a queue that will collect, aggregate, and export blocks and metadata
	exportQueue := types.NewQueue(25)

	// Create workers
	worker := parser.NewWorker(b.parserCtx, exportQueue, 0, config.Cfg.Parser.ConcurrentSync)
	worker.SetIndexer(b.parserCtx.Indexer)

	latestBlockHeight := mustGetLatestHeight(b.parserCtx)
	Cast(b.parserCtx.Indexer).GetLatestBlockHeight().Store(int64(latestBlockHeight))
	Cast(b.parserCtx.Indexer).GetCatchUpFlag().Store(int64(-1))
	go b.getLatestBlockHeight(ctx)

	lastDbBlockHeight := uint64(0)
	for {
		epoch, err := b.parserCtx.Database.GetEpoch(context.TODO())
		if err != nil {
			log.Errorw("failed to get last block height from database", "error", err)
			continue
		}
		lastDbBlockHeight = uint64(epoch.BlockHeight)
		break
	}

	// fetch block data
	go b.quickFetchBlockData(lastDbBlockHeight + 1)

	go b.enqueueNewBlocks(ctx, exportQueue, lastDbBlockHeight+1)

	// Start each blocking worker in a go-routine where the worker consumes jobs
	// off of the export queue.
	Cast(b.parserCtx.Indexer).ProcessedQueue <- uint64(0) // init ProcessedQueue
	go worker.Start(ctx)
}

// enqueueNewBlocks enqueues new block heights onto the provided queue.
func (b *BlockSyncerModular) enqueueNewBlocks(context context.Context, exportQueue types.HeightQueue, currHeight uint64) {
	// Enqueue upcoming heights
	for {
		select {
		case <-context.Done():
			log.Infof("Receive cancel signal, enqueueNewBlocks routine will stop")
			//close channel
			close(exportQueue)
			return
		default:
			{
				latestBlockHeightAny := Cast(b.parserCtx.Indexer).GetLatestBlockHeight().Load()
				latestBlockHeight := latestBlockHeightAny.(int64)
				// Enqueue all heights from the current height up to the latest height
				for ; currHeight <= uint64(latestBlockHeight); currHeight++ {
					log.Debugw("enqueueing new block", "height", currHeight)
					exportQueue <- currHeight
				}
				time.Sleep(config.GetAvgBlockTime())
			}
		}
	}
}

func (b *BlockSyncerModular) getLatestBlockHeight(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Infof("Receive cancel signal, getLatestBlockHeight routine will stop")
			return
		default:
			{
				latestBlockHeight, err := b.parserCtx.Node.LatestHeight()
				if err != nil {
					log.Errorw("failed to get last block from RPCConfig client",
						"error", err,
						"retry interval", config.GetAvgBlockTime())
					continue
				}
				Cast(b.parserCtx.Indexer).GetLatestBlockHeight().Store(latestBlockHeight)

				time.Sleep(config.GetAvgBlockTime())
			}
		}
	}
}

func (b *BlockSyncerModular) quickFetchBlockData(startHeight uint64) {
	count := uint64(b.config.Parser.Workers)
	cycle := uint64(0)
	for {
		latestBlockHeightAny := Cast(b.parserCtx.Indexer).GetLatestBlockHeight().Load()
		latestBlockHeight := latestBlockHeightAny.(int64)
		if latestBlockHeight < int64(count*(cycle+1)+startHeight-1) {
			log.Infof("quick fetch ended latestBlockHeight: %d", latestBlockHeight)
			Cast(b.parserCtx.Indexer).GetCatchUpFlag().Store(int64(count*cycle + startHeight - 1))
			break
		}
		processedHeight, ok := <-Cast(b.parserCtx.Indexer).ProcessedQueue
		if !ok {
			log.Warnf("ProcessedQueue is closed")
			return
		}
		if processedHeight != 0 && count*cycle+startHeight-processedHeight > 4*count {
			continue
		}
		b.fetchData(count, cycle, startHeight, latestBlockHeight)
		cycle++
	}
}

func (b *BlockSyncerModular) fetchData(count, cycle, startHeight uint64, latestBlockHeight int64) {
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
				block, err := b.parserCtx.Node.Block(int64(height))
				if err != nil {
					log.Warnf("failed to get block from node: %s", err)
					continue
				}

				events, err := b.parserCtx.Node.BlockResults(int64(height))
				if err != nil {
					log.Warnf("failed to get block results from node: %s", err)
					continue
				}
				txs, err := b.parserCtx.Node.Txs(block)
				if err != nil {
					log.Warnf("failed to get block results from node: %s", err)
					continue
				}
				heightKey := fmt.Sprintf("%s-%d", b.Name(), height)
				blockMap.Store(heightKey, block)
				eventMap.Store(heightKey, events)
				txMap.Store(heightKey, txs)
				break
			}
		}(i, cycle)
	}
	wg.Wait()
}

func (b *BlockSyncerModular) prepareMasterFlagTable() error {
	if err := b.parserCtx.Database.
		PrepareTables(context.TODO(), []schema.Tabler{&models.MasterDB{}}); err != nil {
		return err
	}
	masterRecord, err := b.parserCtx.Database.GetMasterDB(context.TODO())
	if err != nil {
		return err
	}
	//not exist
	if !masterRecord.OneRowId {
		if err = b.parserCtx.Database.SetMasterDB(context.TODO(), &models.MasterDB{
			OneRowId: true,
			IsMaster: true,
		}); err != nil {
			return err
		}
	}
	return nil
}

// makeBlockSyncerConfig make block syncer service config from StorageProviderConfig
func makeBlockSyncerConfig(cfg *gfspconfig.GfSpConfig) *tomlconfig.TomlConfig {
	rpcAddress := cfg.Chain.ChainAddress[0]

	return &tomlconfig.TomlConfig{
		Chain: config.ChainConfig{
			Bech32Prefix: "cosmos",
			Modules:      cfg.BlockSyncer.Modules,
		},
		Node: tomlconfig.NodeConfig{
			Type: "remote",
			RPC: &remote.RPCConfig{
				ClientName: "juno",
				Address:    rpcAddress,
			},
		},
		Parser: parserconfig.Config{
			Workers: int64(cfg.BlockSyncer.Workers),
		},
		Database: databaseconfig.Config{
			Type:               "mysql",
			DSN:                cfg.BlockSyncer.Dsn,
			PartitionBatchSize: 10_000,
			MaxIdleConnections: 10,
			MaxOpenConnections: 30,
		},
		Logging: loggingconfig.Config{
			Level: "debug",
		},
		RecreateTables: cfg.BlockSyncer.RecreateTables,
		EnableDualDB:   cfg.BlockSyncer.EnableDualDB,
		DsnSwitched:    cfg.BlockSyncer.DsnSwitched,
	}
}

func newBackupBlockSyncerService(cfg *tomlconfig.TomlConfig, mainDBIsMaster bool) (*BlockSyncerModular, error) {
	backUpConfig, err := generateConfigForBackup(cfg)
	if err != nil {
		return nil, err
	}

	BackupService = &BlockSyncerModular{
		config: backUpConfig,
		name:   BlockSyncerModularBackupName,
	}

	if err = BackupService.initClient(); err != nil {
		return nil, err
	}

	// init meta db, if mainService db is not current master, backup is master, don't recreate
	if err = BackupService.initDB(mainDBIsMaster || backUpConfig.RecreateTables); err != nil {
		return nil, err
	}
	return BackupService, nil
}

func generateConfigForBackup(cfg *tomlconfig.TomlConfig) (*tomlconfig.TomlConfig, error) {
	configBackup := new(tomlconfig.TomlConfig)
	if err := DeepCopyByGob(cfg, configBackup); err != nil {
		return nil, err
	}

	configBackup.Database.DSN = configBackup.DsnSwitched

	return configBackup, nil
}

func DeepCopyByGob(src, dst interface{}) error {
	var buffer bytes.Buffer
	if err := gob.NewEncoder(&buffer).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(&buffer).Decode(dst)
}

func getDBConfigFromEnv(dsn string) (string, error) {
	dsnVal, ok := os.LookupEnv(dsn)
	if !ok {
		return "", errors.ErrDSNNotSet
	}
	return dsnVal, nil
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
		if epochMaster.BlockHeight-epochSlave.BlockHeight < DefaultBlockHeightDiff {
			SwitchMasterDBFlag()
			StopMainService()
			break
		}
		time.Sleep(time.Minute * DefaultCheckDiffPeriod)
	}
}

func SwitchMasterDBFlag() error {
	masterFlag, err := FlagDB.GetMasterDB(context.TODO())
	if err != nil {
		return err
	}

	//switch flag
	masterFlag.IsMaster = !masterFlag.IsMaster
	if err = FlagDB.SetMasterDB(context.TODO(), masterFlag); err != nil {
		return err
	}
	log.Infof("DB switched")
	return nil
}

func StopMainService() error {
	CancelMain()
	return nil
}

// mustGetLatestHeight tries getting the latest height from the RPC client.
// If no latest height can be found after MaxRetryCount, it returns 0.
func mustGetLatestHeight(ctx *parser.Context) uint64 {
	for retryCount := 0; retryCount < MaxRetryCount; retryCount++ {
		latestBlockHeight, err := ctx.Node.LatestHeight()
		if err == nil {
			return uint64(latestBlockHeight)
		}

		log.Errorw("failed to get last block from RPCConfig client", "error", err, "retry interval", config.GetAvgBlockTime(), "retry count", retryCount)

		time.Sleep(config.GetAvgBlockTime())
	}

	return 0
}
