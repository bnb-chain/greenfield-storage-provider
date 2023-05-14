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

func NewManageModular(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig,
	opts ...gfspapp.Option) (
	coremodule.Modular, error) {
	if cfg.Manager != nil {
		app.SetManager(cfg.Manager)
		return cfg.Manager, nil
	}
	manager := &ManageModular{baseApp: app}
	opts = append(opts, manager.DefaultManagerOptions)
	for _, opt := range opts {
		if err := opt(app, cfg); err != nil {
			return nil, err
		}
	}
	app.SetManager(manager)
	return manager, nil
}

func (m *ManageModular) DefaultManagerOptions(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig) error {
	if cfg.GlobalMaxUploadingNumber == 0 {
		cfg.GlobalMaxUploadingNumber = DefaultGlobalMaxUploadingNumber
	}
	if cfg.GlobalUploadObjectParallel == 0 {
		cfg.GlobalUploadObjectParallel = DefaultGlobalUploadObjectParallel
	}
	if cfg.GlobalReplicatePieceParallel == 0 {
		cfg.GlobalReplicatePieceParallel = DefaultGlobalReplicatePieceParallel
	}
	if cfg.GlobalSealObjectParallel == 0 {
		cfg.GlobalSealObjectParallel = DefaultGlobalSealObjectParallel
	}
	if cfg.GlobalReceiveObjectParallel == 0 {
		cfg.GlobalReceiveObjectParallel = DefaultGlobalReceiveObjectParallel
	}
	if cfg.GlobalGCObjectParallel == 0 {
		cfg.GlobalGCObjectParallel = DefaultGlobalGCObjectParallel
	}
	if cfg.GlobalGCZombieParallel == 0 {
		cfg.GlobalGCZombieParallel = DefaultGlobalGCZombieParallel
	}
	if cfg.GlobalGCMetaParallel == 0 {
		cfg.GlobalGCMetaParallel = DefaultGlobalGCMetaParallel
	}
	if cfg.GlobalDownloadObjectTaskCacheSize == 0 {
		cfg.GlobalDownloadObjectTaskCacheSize = DefaultGlobalDownloadObjectTaskCacheSize
	}
	if cfg.GlobalChallengePieceTaskCacheSize == 0 {
		cfg.GlobalChallengePieceTaskCacheSize = DefaultGlobalChallengePieceTaskCacheSize
	}
	if cfg.GlobalBatchGcObjectTimeInterval == 0 {
		cfg.GlobalBatchGcObjectTimeInterval = DefaultGlobalBatchGcObjectTimeInterval
	}
	if cfg.GlobalGcObjectBlockInterval == 0 {
		cfg.GlobalGcObjectBlockInterval = DefaultGlobalGcObjectBlockInterval
	}
	if cfg.GlobalGcObjectSafeBlockDistance == 0 {
		cfg.GlobalGcObjectSafeBlockDistance = DefaultGlobalGcObjectSafeBlockDistance
	}
	if cfg.GlobalSyncConsensusInfoInterval == 0 {
		cfg.GlobalSyncConsensusInfoInterval = DefaultGlobalSyncConsensusInfoInterval
	}
	m.maxUploadObjectNumber = cfg.GlobalMaxUploadingNumber
	m.gcObjectTimeInterval = cfg.GlobalBatchGcObjectTimeInterval
	m.gcObjectBlockInterval = cfg.GlobalGcObjectBlockInterval
	m.gcSafeBlockDistance = cfg.GlobalGcObjectSafeBlockDistance
	m.syncConsensusInfoInterval = cfg.GlobalSyncConsensusInfoInterval
	m.uploadQueue = cfg.NewStrategyTQueueFunc(
		m.Name()+"-upload-object", cfg.GlobalUploadObjectParallel)
	m.replicateQueue = cfg.NewStrategyTQueueWithLimitFunc(
		m.Name()+"-replicate-piece", cfg.GlobalReplicatePieceParallel)
	m.sealQueue = cfg.NewStrategyTQueueWithLimitFunc(
		m.Name()+"-seal-object", cfg.GlobalSealObjectParallel)
	m.receiveQueue = cfg.NewStrategyTQueueWithLimitFunc(
		m.Name()+"-confirm-receive-piece", cfg.GlobalReceiveObjectParallel)
	m.gcObjectQueue = cfg.NewStrategyTQueueWithLimitFunc(
		m.Name()+"-gc-object", cfg.GlobalGCObjectParallel)
	m.gcZombieQueue = cfg.NewStrategyTQueueWithLimitFunc(
		m.Name()+"-gc-zombie", cfg.GlobalGCZombieParallel)
	m.gcMetaQueue = cfg.NewStrategyTQueueWithLimitFunc(
		m.Name()+"-gc-meta", cfg.GlobalGCMetaParallel)

	m.downloadQueue = cfg.NewStrategyTQueueFunc(
		m.Name()+"-cache-download-object", cfg.GlobalDownloadObjectTaskCacheSize)
	m.challengeQueue = cfg.NewStrategyTQueueFunc(
		m.Name()+"-cache-challenge-piece", cfg.GlobalChallengePieceTaskCacheSize)
	return nil
}
