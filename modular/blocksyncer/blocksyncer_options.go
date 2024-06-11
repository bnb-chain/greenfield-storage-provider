package blocksyncer

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	cometbfttypes "github.com/cometbft/cometbft/abci/types"
	"github.com/forbole/juno/v4/cmd"
	parsecmdtypes "github.com/forbole/juno/v4/cmd/parse/types"
	"github.com/forbole/juno/v4/common"
	databaseconfig "github.com/forbole/juno/v4/database/config"
	loggingconfig "github.com/forbole/juno/v4/log/config"
	"github.com/forbole/juno/v4/models"
	"github.com/forbole/juno/v4/modules"
	"github.com/forbole/juno/v4/modules/messages"
	"github.com/forbole/juno/v4/node/remote"
	"github.com/forbole/juno/v4/parser"
	parserconfig "github.com/forbole/juno/v4/parser/config"
	"github.com/forbole/juno/v4/types"
	"github.com/forbole/juno/v4/types/config"
	"github.com/shopspring/decimal"
	"gorm.io/gorm/schema"

	"github.com/zkMeLabs/mechain-storage-provider/base/gfspapp"
	"github.com/zkMeLabs/mechain-storage-provider/base/gfspconfig"
	coremodule "github.com/zkMeLabs/mechain-storage-provider/core/module"
	db "github.com/zkMeLabs/mechain-storage-provider/modular/blocksyncer/database"
	registrar "github.com/zkMeLabs/mechain-storage-provider/modular/blocksyncer/modules"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/log"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/metrics"
	"github.com/zkMeLabs/mechain-storage-provider/store/bsdb"
)

func NewBlockSyncerModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	junoCfg := makeBlockSyncerConfig(cfg)

	MainService = &BlockSyncerModular{
		config:  junoCfg,
		name:    coremodule.BlockSyncerModularName,
		baseApp: app,
	}
	blockMap = new(sync.Map)
	eventMap = new(sync.Map)
	txMap = new(sync.Map)
	txHashMap = new(sync.Map)

	RealTimeStart = &atomic.Bool{}
	RealTimeStart.Store(false)
	CatchEndBlock = &atomic.Int64{}
	CatchEndBlock.Store(-1)

	NeedBackup = junoCfg.EnableDualDB

	if err := MainService.initClient(cfg); err != nil {
		return nil, err
	}

	// prepare master table
	FlagDB = db.Cast(MainService.parserCtx.Database)
	MainService.prepareMasterFlagTable()
	mainServiceDB, _ := FlagDB.GetMasterDB(context.TODO())
	mainDBIsMaster := mainServiceDB.IsMaster

	// init main service db, if main service DB is not current master then recreate tables
	if err := MainService.initDB(false); err != nil {
		return nil, err
	}

	// when NeedBackup config true Or backup db is current master DB, init backup service
	if NeedBackup || !mainServiceDB.IsMaster {
		// create backup block syncer
		if blockSyncerBackup, err := newBackupBlockSyncerService(junoCfg, mainDBIsMaster); err != nil {
			return nil, err
		} else {
			BackupService = blockSyncerBackup
		}
	}

	return MainService, nil
}

// initClient initialize a juno client using given configs
func (b *BlockSyncerModular) initClient(cfg *gfspconfig.GfSpConfig) error {
	// JunoConfig the runner
	junoConfig := cmd.NewConfig("juno").
		WithParseConfig(parsecmdtypes.NewConfig().
			WithRegistrar(registrar.NewBlockSyncerRegistrar(
				messages.CosmosMessageAddressesParser,
			)).WithDBBuilder(db.BlockSyncerDBBuilder).WithFileType("toml"),
		)
	cmdCfg := junoConfig.GetParseConfig()
	cmdCfg.WithTomlConfig(b.config)

	// set toml config to juno config
	if readErr := parsecmdtypes.ReadConfigPreRunE(cmdCfg)(nil, nil); readErr != nil {
		log.Infof("readErr: %v", readErr)
		return readErr
	}

	username, password, envErr := getDBConfigFromEnv(bsdb.BsDBUser, bsdb.BsDBPasswd)
	if envErr != nil {
		log.Infof("failed to get username and password err:%v", envErr)
		username = cfg.BsDB.User
		password = cfg.BsDB.Passwd
	}
	dbAddress := cfg.BsDB.Address
	if cfg.BlockSyncer.BsDBWriteAddress != "" {
		dbAddress = cfg.BlockSyncer.BsDBWriteAddress
	}
	config.Cfg.Database.DSN = fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&multiStatements=true&loc=Local&interpolateParams=true", username, password, dbAddress, cfg.BsDB.Database)

	var ctx *parser.Context
	ctx, err := parsecmdtypes.GetParserContext(config.Cfg, cmdCfg)
	if err != nil {
		log.Errorf("failed to GetParserContext err: %v", err)
		return err
	}
	b.parserCtx = ctx
	log.Infof("blocksyncer dsn : %s", config.Cfg.Database.DSN)
	commitNumber := uint64(CommitNumber)
	if cfg.BlockSyncer.CommitNumber != 0 {
		commitNumber = cfg.BlockSyncer.CommitNumber
	}
	b.parserCtx.Indexer = NewIndexer(ctx.EncodingConfig.Codec,
		ctx.Node,
		ctx.Database,
		ctx.Modules,
		b.Name(), commitNumber)
	return nil
}

// initDB create tables needed by block syncer. It depends on which modules are configured
func (b *BlockSyncerModular) initDB(useMigrate bool) error {
	var err error
	for _, module := range b.parserCtx.Modules {
		if module, ok := module.(modules.PrepareTablesModule); ok {
			if useMigrate {
				err = module.AutoMigrate()
			} else {
				err = module.PrepareTables()
			}
			if err != nil {
				log.Errorw("failed to PrepareTables/AutoMigrate tables", "error", err)
				return err
			}
		}
	}
	return nil
}

func (b *BlockSyncerModular) dataMigration(ctx context.Context) {
	// update bucket size after the bucket_size column is added
	b.syncBucketSize()
	b.syncSlashPrefixTree()
}

// serve start BlockSyncer rpc service
func (b *BlockSyncerModular) serve(ctx context.Context) {
	migrateDBAny := ctx.Value(MigrateDBKey{})
	if migrateDB, ok := migrateDBAny.(bool); ok && migrateDB {
		err := b.initDB(true)
		if err != nil {
			log.Errorw("failed to init DB", "error", err)
			return
		}
	}
	// Create a queue that will collect, aggregate, and export blocks and metadata
	exportQueue := types.NewQueue(100)

	// Create workers
	worker := parser.NewWorker(b.parserCtx, exportQueue, 0, config.Cfg.Parser.ConcurrentSync)
	worker.SetIndexer(b.parserCtx.Indexer)

	b.dataMigration(ctx)

	latestBlockHeight := mustGetLatestHeight(b.parserCtx)
	Cast(b.parserCtx.Indexer).GetLatestBlockHeight().Store(int64(latestBlockHeight))
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
	go b.quickFetchBlockData(ctx, lastDbBlockHeight+1)

	go b.enqueueNewBlocks(ctx, exportQueue, lastDbBlockHeight+1)

	// Start each blocking worker in a go-routine where the worker consumes jobs
	go worker.Start(ctx)
}

// enqueueNewBlocks enqueues new block heights onto the provided queue.
func (b *BlockSyncerModular) enqueueNewBlocks(context context.Context, exportQueue types.HeightQueue, currHeight uint64) {
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	// Enqueue upcoming heights
	for {
		select {
		case <-context.Done():
			log.Infof("Receive cancel signal, enqueueNewBlocks routine will stop")
			// close channel
			close(exportQueue)
			return
		case <-ticker.C:
			{
				latestBlockHeightAny := Cast(b.parserCtx.Indexer).GetLatestBlockHeight().Load()
				latestBlockHeight := latestBlockHeightAny.(int64)
				// Enqueue all heights from the current height up to the latest height
				for ; currHeight <= uint64(latestBlockHeight); currHeight++ {
					// log.Debugw("enqueueing new block", "height", currHeight)
					exportQueue <- currHeight
				}
			}
		}
	}
}

func (b *BlockSyncerModular) getLatestBlockHeight(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Infof("Receive cancel signal, getLatestBlockHeight routine will stop")
			return
		case <-ticker.C:
			{
				latestBlockHeight, err := b.parserCtx.Node.LatestHeight()
				if err != nil {
					log.Errorw("failed to get last block from RPCConfig client",
						"error", err,
						"retry interval", config.GetAvgBlockTime())
					continue
				}
				metrics.ChainLatestHeight.Set(float64(latestBlockHeight))
				metrics.GoRoutineCount.Set(float64(runtime.NumGoroutine()))
				Cast(b.parserCtx.Indexer).GetLatestBlockHeight().Store(latestBlockHeight)
			}
		}
	}
}

func (b *BlockSyncerModular) quickFetchBlockData(ctx context.Context, startHeight uint64) {
	count := uint64(b.config.Parser.Workers)
	cycle := uint64(0)
	startBlock := uint64(0)
	endBlock := uint64(0)
	flag := 0

	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Infof("Receive cancel signal, quickFetchBlockData routine will stop")
			return
		case <-ticker.C:
			latestBlockHeightAny := Cast(b.parserCtx.Indexer).GetLatestBlockHeight().Load()
			latestBlockHeight := latestBlockHeightAny.(int64)
			if latestBlockHeight == int64(endBlock) {
				RealTimeStart.Store(true)
				CatchEndBlock.Store(int64(endBlock))
				return
			}
			if latestBlockHeight > int64(count*(cycle+1)+startHeight-1) {
				startBlock = count*cycle + startHeight
				endBlock = count*(cycle+1) + startHeight - 1
				flag = 1
				processedHeight := Cast(b.parserCtx.Indexer).ProcessedHeight
				if processedHeight != 0 && int64(startBlock)-int64(processedHeight) > int64(MaxHeightGapFactor*count) {
					log.Infof("processedHeight: %d", processedHeight)
					time.Sleep(time.Second)
					continue
				}
				cycle++
			} else if flag != 0 {
				startBlock = endBlock + 1
				if startBlock > uint64(latestBlockHeight) {
					startBlock = uint64(latestBlockHeight)
				}
				endBlock = uint64(latestBlockHeight)
			} else {
				flag = 1
				startBlock = startHeight
				endBlock = uint64(latestBlockHeight)
			}

			b.fetchData(startBlock, endBlock)
		}
	}
}

func (b *BlockSyncerModular) fetchData(start, end uint64) {
	log.Infof("fetch data start:%d end:%d", start, end)
	if start > end {
		return
	}
	wg := &sync.WaitGroup{}
	wg.Add(int(end - start + 1))
	for i := start; i <= end; i++ {
		go func(height uint64) {
			defer wg.Done()

			for {
				rpcStartTime := time.Now()
				block, err := b.parserCtx.Node.Block(int64(height))
				if err != nil {
					log.Warnf("failed to get block from node: %s", err)
					continue
				}
				metrics.ChainRPCTime.Set(float64(time.Since(rpcStartTime).Milliseconds()))
				rpcStartTime = time.Now()
				events, err := b.parserCtx.Node.BlockResults(int64(height))
				if err != nil {
					log.Warnf("failed to get block results from node: %s", err)
					continue
				}
				metrics.ChainRPCTime.Set(float64(time.Since(rpcStartTime).Milliseconds()))
				txs := make(map[common.Hash][]cometbfttypes.Event)
				for idx := 0; idx < len(events.TxsResults); idx++ {
					k := block.Block.Data.Txs[idx]
					v := events.TxsResults[idx].GetEvents()
					txs[common.BytesToHash(k.Hash())] = v
				}

				heightKey := fmt.Sprintf("%s-%d", b.Name(), height)
				blockMap.Store(heightKey, block)
				eventMap.Store(heightKey, events)
				txMap.Store(heightKey, txs)
				txHashMap.Store(heightKey, block.Block.Data.Txs)
				break
			}
		}(i)
	}
	wg.Wait()
}

func (b *BlockSyncerModular) prepareMasterFlagTable() error {
	if err := FlagDB.
		PrepareTables(context.TODO(), []schema.Tabler{&bsdb.MasterDB{}}); err != nil {
		return err
	}
	masterRecord, err := FlagDB.GetMasterDB(context.TODO())
	if err != nil {
		return err
	}
	// not exist
	if !masterRecord.OneRowId {
		if err = FlagDB.SetMasterDB(context.TODO(), &bsdb.MasterDB{
			OneRowId: true,
			IsMaster: true,
		}); err != nil {
			return err
		}
	}
	return nil
}

// makeBlockSyncerConfig make block syncer service config from StorageProviderConfig
func makeBlockSyncerConfig(cfg *gfspconfig.GfSpConfig) *config.TomlConfig {
	rpcAddress := cfg.Chain.ChainAddress[0]

	return &config.TomlConfig{
		Chain: config.ChainConfig{
			Bech32Prefix: "cosmos",
			Modules:      cfg.BlockSyncer.Modules,
		},
		Node: config.NodeConfig{
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
			PartitionBatchSize: 10_000,
			MaxIdleConnections: 10,
			MaxOpenConnections: 30,
		},
		Logging: loggingconfig.Config{
			Level: "debug",
		},
	}
}

func newBackupBlockSyncerService(cfg *config.TomlConfig, mainDBIsMaster bool) (*BlockSyncerModular, error) {
	backUpConfig, err := generateConfigForBackup(cfg)
	if err != nil {
		return nil, err
	}

	BackupService = &BlockSyncerModular{
		config: backUpConfig,
		name:   BlockSyncerModularBackupName,
	}

	if err = BackupService.initClient(nil); err != nil {
		return nil, err
	}

	// init meta db, if mainService db is not current master, backup is master, don't recreate
	if err = BackupService.initDB(false); err != nil {
		return nil, err
	}
	return BackupService, nil
}

func generateConfigForBackup(cfg *config.TomlConfig) (*config.TomlConfig, error) {
	configBackup := new(config.TomlConfig)
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

func getDBConfigFromEnv(usernameKey, passwordKey string) (username, password string, err error) {
	var ok bool
	username, ok = os.LookupEnv(usernameKey)
	if !ok {
		return "", "", ErrDSNNotSet
	}
	password, ok = os.LookupEnv(passwordKey)
	if !ok {
		return "", "", ErrDSNNotSet
	}
	return
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

	// switch flag
	masterFlag.IsMaster = !masterFlag.IsMaster
	if err = FlagDB.SetMasterDB(context.TODO(), masterFlag); err != nil {
		return err
	}
	log.Info("DB switched")
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

func (b *BlockSyncerModular) syncBucketSize() error {
	dataMigrateKey := bsdb.ProcessKeyUpdateBucketSize
	record, err := db.Cast(b.parserCtx.Database).GetDataMigrationRecord(context.Background(), dataMigrateKey)
	if err != nil {
		panic("failed to get the data migration record for " + dataMigrateKey)
	}
	if record != nil && record.IsCompleted {
		log.Infof(dataMigrateKey + "has already been completed. Skip it now.")
		return nil
	}

	log.Infof("start sync bucket size")
	type result struct {
		BucketID    common.Hash
		StorageSize decimal.Decimal
		ChargeSize  decimal.Decimal
	}
	for i := 0; i < ObjectsNumberOfShards; i++ {
		tableName := "objects_"
		if i < 10 {
			tableName += "0"
		}
		tableName += fmt.Sprintf("%d", i)
		var res []*result
		/*
			SELECT A.bucket_id AS bucket_id,
			       SUM(A.payload_size) AS storage_size,
			       SUM(CASE
			               WHEN (CAST((A.payload_size)AS DECIMAL(65, 0)) < 128000) THEN CAST(128000 AS DECIMAL(65, 0))
			               ELSE CAST((A.payload_size) AS DECIMAL(65, 0))
			           END) AS charge_size
			FROM
			  (SELECT bucket_id,
			          payload_size
			   FROM objects_00
			   WHERE removed = FALSE
			     AND `status` = 'OBJECT_STATUS_SEALED'
				) AS A
			GROUP BY bucket_id
		*/
		sql := "select A.bucket_id as bucket_id ,SUM(A.payload_size) as storage_size, SUM(CASE WHEN (CAST((A.payload_size)AS DECIMAL(65,0)) < 128000) THEN CAST(128000 AS DECIMAL(65,0)) ELSE CAST((A.payload_size) AS DECIMAL(65,0)) END) AS charge_size from (select bucket_id, payload_size from " + tableName + " WHERE removed = ? and status = ?) as A group by bucket_id"
		log.Infof("sql:%s", sql)
		err := db.Cast(b.parserCtx.Database).Db.Raw(sql, false, "OBJECT_STATUS_SEALED").Scan(&res).Error
		if err != nil {
			log.Errorw("failed to query size ", "error", err)
			return err
		}
		buckets := make([]*models.Bucket, 0, len(res))
		for _, r := range res {
			buckets = append(buckets, &models.Bucket{
				BucketID:    r.BucketID,
				StorageSize: r.StorageSize,
				ChargeSize:  r.ChargeSize,
			})
		}
		if len(buckets) == 0 {
			continue
		}
		offset := 0
		step := 1000
		for {
			left := offset
			right := offset + step
			if left >= len(buckets) {
				break
			}
			if right > len(buckets) {
				right = len(buckets)
			}
			err = db.Cast(b.parserCtx.Database).BatchUpdateBucketSize(context.Background(), buckets[left:right])
			if err != nil {
				log.Errorw("failed to update size", "error", err)
				return err
			}
			offset = right
		}
	}
	log.Info("succeed to sync bucket size")
	err = db.Cast(b.parserCtx.Database).UpdateDataMigrationRecord(context.Background(), dataMigrateKey, true)
	if err != nil {
		log.Errorw("failed to UpdateDataMigrationRecord", "error", err, "dataMigrateKey", dataMigrateKey)
		return err
	}
	return nil
}

func (b *BlockSyncerModular) syncSlashPrefixTree() error {
	dataMigrateKey := bsdb.ProcessKeyUpdateSlashPrefixTree
	record, err := db.Cast(b.parserCtx.Database).GetDataMigrationRecord(context.Background(), dataMigrateKey)
	if err != nil {
		panic("failed to get the data migration record for " + dataMigrateKey)
	}
	if record != nil && record.IsCompleted {
		log.Infof(dataMigrateKey + "has already been completed. Skip it now.")
		return nil
	}
	type TreeNode struct {
		ID         uint64 `gorm:"id"`
		BucketName string `gorm:"bucket_name"`
		PathName   string `gorm:"path_name"`
		FullName   string `gorm:"full_name"`
	}
	log.Infof("start sync slash prefix tree")
	/*
		Find the folder issue caused by multiple object creations within the same block height.

		SELECT * FROM slash_prefix_tree_nodes_36
		WHERE  id NOT IN  (
			SELECT MIN(id) FROM slash_prefix_tree_nodes_36 GROUP BY bucket_name, path_name, full_name
		)
	*/
	for i := 0; i < ObjectsNumberOfShards; i++ {
		tableName := "slash_prefix_tree_nodes_"
		if i < 10 {
			tableName += "0"
		}
		tableName += fmt.Sprintf("%d", i)
		var res []*TreeNode
		sql := "SELECT * FROM " + tableName + " WHERE id NOT IN (SELECT MIN(id) FROM " + tableName + " GROUP BY bucket_name, path_name, full_name);"
		log.Infof("sql:%s", sql)
		err := db.Cast(b.parserCtx.Database).Db.Raw(sql).Scan(&res).Error
		if err != nil {
			log.Errorw("query slash prefix tree", "error", err)
			return err
		}
		nodes := make([]*bsdb.SlashPrefixTreeNode, 0, len(res))
		for _, r := range res {
			nodes = append(nodes, &bsdb.SlashPrefixTreeNode{
				ID:         r.ID,
				FullName:   r.FullName,
				PathName:   r.PathName,
				BucketName: r.BucketName,
			})
		}
		if len(nodes) == 0 {
			continue
		}
		log.Debugw("start to delete prefix tree node by id", "node lens", len(nodes))
		offset := 0
		step := 1000
		for {
			left := offset
			right := offset + step
			if left >= len(nodes) {
				break
			}
			if right > len(nodes) {
				right = len(nodes)
			}
			err = db.Cast(b.parserCtx.Database).BatchDeletePrefixTreeNodeByID(context.Background(), nodes[left:right])
			if err != nil {
				log.Errorw("failed to delete slash prefix trees", "error", err)
				return err
			}
			offset = right
		}
	}

	/*
		Identify the folder issue caused by multiple object deletions within a single block height not being deleted.

		SELECT *  FROM slash_prefix_tree_nodes_36
		WHERE (bucket_name, path_name) IN(
			SELECT bucket_name, path_name FROM slash_prefix_tree_nodes_36 GROUP BY bucket_name, path_name HAVING COUNT(*) = 1
		)
		AND is_folder =1
	*/
	deletedNodes := make([]*bsdb.SlashPrefixTreeNode, 0)
	for i := 0; i < ObjectsNumberOfShards; i++ {
		tableName := "slash_prefix_tree_nodes_"
		if i < 10 {
			tableName += "0"
		}
		tableName += fmt.Sprintf("%d", i)
		var res []*TreeNode
		// First, query to find the path and bucket name of folders that do not have any objects but only one folder beneath them.
		// Then, perform checks on all the paths.
		sql := "SELECT * FROM " + tableName + " WHERE (bucket_name, path_name) IN (SELECT bucket_name, path_name FROM " + tableName + " GROUP BY bucket_name, path_name HAVING COUNT(*) = 1) AND is_folder = 1;"
		log.Infof("sql:%s", sql)
		err := db.Cast(b.parserCtx.Database).Db.Raw(sql).Scan(&res).Error
		if err != nil {
			log.Errorw("failed to query slash prefix tree", "error", err)
			return err
		}
		nodes := make([]*bsdb.SlashPrefixTreeNode, 0, len(res))
		for _, r := range res {
			nodes = append(nodes, &bsdb.SlashPrefixTreeNode{
				ID:         r.ID,
				FullName:   r.FullName,
				PathName:   r.PathName,
				BucketName: r.BucketName,
			})
		}
		if len(nodes) == 0 {
			continue
		}
		// 1. First, check if all sub-folders also contain only one folder and no objects. If true, delete them and proceed to check the parent folder logic.
		// 2. For parent folders, check if they too contain only one folder and no objects. Recursively perform this check up to the root ('/') folder, and then directly delete it.
		// 3. End this loop.
		for _, node := range nodes {
			// The reason for using full name here is to check if the folder below is not empty
			trees, err := db.Cast(b.parserCtx.Database).GetPrefixTreesByBucketAndPathName(context.Background(), node.BucketName, node.FullName)
			if err != nil {
				log.Errorw("failed to get prefix trees by bucket and path name", "error", err, "bucketName", node.BucketName, "pathName", node.PathName)
				return err
			}
			// In our case, we only consider the value 0, which indicates that there are no subfolders under this minimum subdirectory. This bucket name and path should be deleted.
			if len(trees) == 0 {
				deletedNodes = append(deletedNodes, &bsdb.SlashPrefixTreeNode{PathName: node.PathName, BucketName: node.BucketName, FullName: node.FullName})
				pathParts := strings.Split(node.PathName, "/")
				for j := len(pathParts) - 1; j > 0; j-- {
					path := strings.Join(pathParts[:j], "/") + "/"
					if path != "/" {
						trees, err = db.Cast(b.parserCtx.Database).GetPrefixTreesByBucketAndPathName(context.Background(), node.BucketName, path)
						if err != nil {
							log.Errorw("failed to get prefix trees by bucket and path name", "error", err, "bucketName", node.BucketName, "pathName", node.PathName)
							return err
						}
						if len(trees) <= 1 {
							subPathParts := strings.Split(path, "/")
							subPath := strings.Join(subPathParts[:len(subPathParts)-2], "/") + "/"
							deletedNodes = append(deletedNodes, &bsdb.SlashPrefixTreeNode{PathName: subPath, BucketName: node.BucketName, FullName: path})
						}
					} else {
						deletedNodes = append(deletedNodes, &bsdb.SlashPrefixTreeNode{BucketName: node.BucketName, PathName: path, FullName: strings.Join(pathParts[:j+1], "/") + "/"})
					}
				}
			}
		}
	}
	log.Debugw("start to delete prefix tree by path and full name", "node lens", len(deletedNodes))
	offset := 0
	step := 1000
	for {
		left := offset
		right := offset + step
		if left >= len(deletedNodes) {
			break
		}
		if right > len(deletedNodes) {
			right = len(deletedNodes)
		}
		err = db.Cast(b.parserCtx.Database).BatchDeletePrefixTreeNodeByBucketAndPathAndFullName(context.Background(), deletedNodes[left:right])
		if err != nil {
			log.Errorw("failed to delete slash prefix trees", "error", err)
			return err
		}
		offset = right
	}

	log.Info("succeed to sync slash prefix tree")
	err = db.Cast(b.parserCtx.Database).UpdateDataMigrationRecord(context.Background(), dataMigrateKey, true)
	if err != nil {
		log.Errorw("failed to UpdateDataMigrationRecord", "error", err, "dataMigrateKey", dataMigrateKey)
		return err
	}
	return nil
}
