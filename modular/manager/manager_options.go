package manager

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	DefaultGlobalMaxUploadingNumber          int    = 4096
	DefaultGlobalUploadObjectParallel        int    = 1024
	DefaultGlobalReplicatePieceParallel      int    = 1024
	DefaultGlobalSealObjectParallel          int    = 1024
	DefaultGlobalReceiveObjectParallel       int    = 4096
	DefaultGlobalGCObjectParallel            int    = 4
	DefaultGlobalGCZombieParallel            int    = 1
	DefaultGlobalGCMetaParallel              int    = 1
	DefaultGlobalDownloadObjectTaskCacheSize int    = 4096
	DefaultGlobalChallengePieceTaskCacheSize int    = 4096
	DefaultGlobalBatchGcObjectTimeInterval   int    = 30 * 60
	DefaultGlobalGcObjectBlockInterval       uint64 = 500
	DefaultGlobalGcObjectSafeBlockDistance   uint64 = 1000
	DefaultGlobalSyncConsensusInfoInterval   uint64 = 2
)

func init() {
	gfspapp.RegisterModularInfo(ManageModularName, ManageModularDescription, NewManageModular)
}

func NewManageModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	if cfg.Customize.Manager != nil {
		app.SetManager(cfg.Customize.Manager)
		return cfg.Customize.Manager, nil
	}
	manager := &ManageModular{baseApp: app}
	if err := DefaultManagerOptions(manager, cfg); err != nil {
		return nil, err
	}
	app.SetManager(manager)
	return manager, nil
}

func DefaultManagerOptions(manager *ManageModular, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Parallel.GlobalMaxUploadingParallel == 0 {
		cfg.Parallel.GlobalMaxUploadingParallel = DefaultGlobalMaxUploadingNumber
	}
	if cfg.Parallel.GlobalUploadObjectParallel == 0 {
		cfg.Parallel.GlobalUploadObjectParallel = DefaultGlobalUploadObjectParallel
	}
	if cfg.Parallel.GlobalReplicatePieceParallel == 0 {
		cfg.Parallel.GlobalReplicatePieceParallel = DefaultGlobalReplicatePieceParallel
	}
	if cfg.Parallel.GlobalSealObjectParallel == 0 {
		cfg.Parallel.GlobalSealObjectParallel = DefaultGlobalSealObjectParallel
	}
	if cfg.Parallel.GlobalReceiveObjectParallel == 0 {
		cfg.Parallel.GlobalReceiveObjectParallel = DefaultGlobalReceiveObjectParallel
	}
	if cfg.Parallel.GlobalGCObjectParallel == 0 {
		cfg.Parallel.GlobalGCObjectParallel = DefaultGlobalGCObjectParallel
	}
	if cfg.Parallel.GlobalGCZombieParallel == 0 {
		cfg.Parallel.GlobalGCZombieParallel = DefaultGlobalGCZombieParallel
	}
	if cfg.Parallel.GlobalGCMetaParallel == 0 {
		cfg.Parallel.GlobalGCMetaParallel = DefaultGlobalGCMetaParallel
	}
	if cfg.Parallel.GlobalDownloadObjectTaskCacheSize == 0 {
		cfg.Parallel.GlobalDownloadObjectTaskCacheSize = DefaultGlobalDownloadObjectTaskCacheSize
	}
	if cfg.Parallel.GlobalChallengePieceTaskCacheSize == 0 {
		cfg.Parallel.GlobalChallengePieceTaskCacheSize = DefaultGlobalChallengePieceTaskCacheSize
	}
	if cfg.Parallel.GlobalBatchGcObjectTimeInterval == 0 {
		cfg.Parallel.GlobalBatchGcObjectTimeInterval = DefaultGlobalBatchGcObjectTimeInterval
	}
	if cfg.Parallel.GlobalGcObjectBlockInterval == 0 {
		cfg.Parallel.GlobalGcObjectBlockInterval = DefaultGlobalGcObjectBlockInterval
	}
	if cfg.Parallel.GlobalGcObjectSafeBlockDistance == 0 {
		cfg.Parallel.GlobalGcObjectSafeBlockDistance = DefaultGlobalGcObjectSafeBlockDistance
	}
	if cfg.Parallel.GlobalSyncConsensusInfoInterval == 0 {
		cfg.Parallel.GlobalSyncConsensusInfoInterval = DefaultGlobalSyncConsensusInfoInterval
	}
	manager.maxUploadObjectNumber = cfg.Parallel.GlobalMaxUploadingParallel
	manager.gcObjectTimeInterval = cfg.Parallel.GlobalBatchGcObjectTimeInterval
	manager.gcObjectBlockInterval = cfg.Parallel.GlobalGcObjectBlockInterval
	manager.gcSafeBlockDistance = cfg.Parallel.GlobalGcObjectSafeBlockDistance
	manager.syncConsensusInfoInterval = cfg.Parallel.GlobalSyncConsensusInfoInterval
	manager.uploadQueue = cfg.Customize.NewStrategyTQueueFunc(
		manager.Name()+"-upload-object", cfg.Parallel.GlobalUploadObjectParallel)
	manager.replicateQueue = cfg.Customize.NewStrategyTQueueWithLimitFunc(
		manager.Name()+"-replicate-piece", cfg.Parallel.GlobalReplicatePieceParallel)
	manager.sealQueue = cfg.Customize.NewStrategyTQueueWithLimitFunc(
		manager.Name()+"-seal-object", cfg.Parallel.GlobalSealObjectParallel)
	manager.receiveQueue = cfg.Customize.NewStrategyTQueueWithLimitFunc(
		manager.Name()+"-confirm-receive-piece", cfg.Parallel.GlobalReceiveObjectParallel)
	manager.gcObjectQueue = cfg.Customize.NewStrategyTQueueWithLimitFunc(
		manager.Name()+"-gc-object", cfg.Parallel.GlobalGCObjectParallel)
	manager.gcZombieQueue = cfg.Customize.NewStrategyTQueueWithLimitFunc(
		manager.Name()+"-gc-zombie", cfg.Parallel.GlobalGCZombieParallel)
	manager.gcMetaQueue = cfg.Customize.NewStrategyTQueueWithLimitFunc(
		manager.Name()+"-gc-meta", cfg.Parallel.GlobalGCMetaParallel)
	manager.downloadQueue = cfg.Customize.NewStrategyTQueueFunc(
		manager.Name()+"-cache-download-object", cfg.Parallel.GlobalDownloadObjectTaskCacheSize)
	manager.challengeQueue = cfg.Customize.NewStrategyTQueueFunc(
		manager.Name()+"-cache-challenge-piece", cfg.Parallel.GlobalChallengePieceTaskCacheSize)
	return nil
}
