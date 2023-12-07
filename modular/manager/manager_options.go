package manager

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/task"
)

const (
	// DefaultGlobalMaxUploadingNumber defines the default max uploading object number
	// in SP, include: uploading object to primary, replicate object to secondaries,
	// and sealing object on greenfield.
	DefaultGlobalMaxUploadingNumber int = 40960
	// DefaultGlobalUploadObjectParallel defines the default max parallel uploading
	// objects to primary in SP system.
	DefaultGlobalUploadObjectParallel int = 10240
	// DefaultGlobalReplicatePieceParallel defines the default max parallel replicating
	// objects to primary in SP system.
	DefaultGlobalReplicatePieceParallel int = 10240
	// DefaultGlobalSealObjectParallel defines the default max parallel sealing objects
	// on greenfield in SP system.
	DefaultGlobalSealObjectParallel int = 10240
	// DefaultGlobalReceiveObjectParallel defines the default max parallel confirming
	// receive pieces on greenfield in SP system.
	DefaultGlobalReceiveObjectParallel int = 10240 * 10
	// DefaultGlobalBackupTaskParallel defines the default parallel backup tasks for
	// dispatching to task executor
	DefaultGlobalBackupTaskParallel int = 10240 * 100
	// DefaultGlobalGCObjectParallel defines the default max parallel gc objects in SP
	// system.
	DefaultGlobalGCObjectParallel int = 4
	// DefaultGlobalGCZombieParallel defines the default max parallel gc zombie pieces
	// in SP system.
	DefaultGlobalGCZombieParallel int = 1
	// DefaultGlobalGCMetaParallel defines the default max parallel gc meta db in SP
	// system.
	DefaultGlobalGCMetaParallel int = 1
	// DefaultGlobalGCBucketMigrationParallel defines the default max parallel gc bucket migration in SP
	// system.
	DefaultGlobalGCBucketMigrationParallel int = 1
	// 	DefaultGlobalRecoveryPieceParallel defines the default max parallel recovery objects in SP
	// system.
	DefaultGlobalRecoveryPieceParallel int = 7
	// DefaultGlobalMigrateGVGParallel defines the default max parallel migrating gvg in SP system.
	DefaultGlobalMigrateGVGParallel int = 200
	// DefaultGlobalDownloadObjectTaskCacheSize defines the default max cache the download
	// object tasks in manager.
	DefaultGlobalDownloadObjectTaskCacheSize int = 4096
	// DefaultGlobalChallengePieceTaskCacheSize defines the default max cache the challenge
	// piece tasks in manager.
	DefaultGlobalChallengePieceTaskCacheSize int = 4096

	// DefaultGlobalBatchGCObjectTimeInterval defines the default interval for generating
	// gc object task.
	DefaultGlobalBatchGCObjectTimeInterval int = 1 * 60
	// DefaultGlobalGCObjectBlockInterval defines the default blocks number for getting
	// deleted objects.
	DefaultGlobalGCObjectBlockInterval uint64 = 1000
	// DefaultGlobalGCObjectSafeBlockDistance defines the default distance form current block
	// height to gc the deleted object.
	DefaultGlobalGCObjectSafeBlockDistance uint64 = 1000
	// DefaultGlobalGCZombiePieceTimeInterval defines the default interval for generating
	// gc zombie piece task.
	DefaultGlobalGCZombiePieceTimeInterval int = 10 * 60
	// DefaultGlobalGCZombiePieceObjectIDInterval defines the default object id number for getting
	// deleted zombie piece.
	DefaultGlobalGCZombiePieceObjectIDInterval uint64 = 100
	// DefaultGlobalGCZombieSafeObjectIDDistance defines the default distance form current object id
	// to gc the deleted zombie piece.
	DefaultGlobalGCZombieSafeObjectIDDistance uint64 = 1000
	// DefaultGlobalGCMetaTimeInterval defines the default interval for generating
	// gc meta task.
	DefaultGlobalGCMetaTimeInterval int = 10 * 60

	// DefaultGlobalSyncConsensusInfoInterval defines the default interval for sync the sp
	// info list to sp db.
	DefaultGlobalSyncConsensusInfoInterval uint64 = 600
	// DefaultStatisticsOutputInterval defines the default interval for output statistics info,
	// it is used to log and debug.
	DefaultStatisticsOutputInterval int = 60
	// DefaultListenRejectUnSealTimeoutHeight defines the default listen reject unseal object
	// on greenfield timeout height, if after current block height + timeout height, the object
	// is not rejected, it is judged failed to reject unseal object on greenfield.
	DefaultListenRejectUnSealTimeoutHeight int = 10

	// DefaultDiscontinueTimeInterval defines the default interval for starting discontinue
	// buckets task , used for test net.
	DefaultDiscontinueTimeInterval = 3 * 60
	// DefaultDiscontinueBucketKeepAliveDays defines the default bucket keep alive days, after
	// the interval, buckets will be discontinued, used for test net.
	DefaultDiscontinueBucketKeepAliveDays = 7

	// DefaultLoadReplicateTimeout defines the task timeout that load replicate tasks from sp db
	DefaultLoadReplicateTimeout int64 = 60
	// DefaultLoadSealTimeout defines the task timeout that load seal tasks from sp db
	DefaultLoadSealTimeout int64 = 180
	// DefaultSubscribeSPExitEventIntervalMillisecond define the default time interval to subscribe sp exit event from metadata.
	DefaultSubscribeSPExitEventIntervalMillisecond = 2000
	// DefaultSubscribeBucketMigrateEventIntervalMillisecond define the default time interval to subscribe bucket migrate event from metadata.
	DefaultSubscribeBucketMigrateEventIntervalMillisecond = 2000
	// DefaultSubscribeSwapOutEventIntervalMillisecond define the default time interval to subscribe gvg swap out event from metadata.
	DefaultSubscribeSwapOutEventIntervalMillisecond = 2000
)

const (
	ManagerGCBlockNumber           = "manager_gc_block_number"
	ManagerSuccessUpload           = "manager_upload_object_success"
	ManagerFailureUpload           = "manager_upload_object_failure"
	ManagerSuccessReplicate        = "manager_replicate_object_success"
	ManagerFailureReplicate        = "manager_replicate_object_failure"
	ManagerCancelReplicate         = "manager_replicate_object_cancel"
	ManagerSuccessReplicateAndSeal = "manager_replicate_and_seal_object_success"
	ManagerFailureReplicateAndSeal = "manager_replicate_and_seal_object_failure"
	ManagerSuccessSeal             = "manager_seal_object_success"
	ManagerFailureSeal             = "manager_seal_object_failure"
	ManagerCancelSeal              = "manager_seal_object_cancel"
	ManagerSuccessConfirmReceive   = "manager_confirm_receive_success"
	ManagerFailureConfirmReceive   = "manager_confirm_receive_failure"
)

func NewManageModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	manager := &ManageModular{baseApp: app}
	if err := DefaultManagerOptions(manager, cfg); err != nil {
		return nil, err
	}
	return manager, nil
}

func DefaultManagerOptions(manager *ManageModular, cfg *gfspconfig.GfSpConfig) (err error) {
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
	if cfg.Parallel.GlobalGCBucketMigrationParallel == 0 {
		cfg.Parallel.GlobalGCBucketMigrationParallel = DefaultGlobalGCBucketMigrationParallel
	}
	if cfg.Parallel.GlobalRecoveryPieceParallel == 0 {
		cfg.Parallel.GlobalRecoveryPieceParallel = DefaultGlobalRecoveryPieceParallel
	}
	if cfg.Parallel.GlobalMigrateGVGParallel == 0 {
		cfg.Parallel.GlobalMigrateGVGParallel = DefaultGlobalMigrateGVGParallel
	}
	if cfg.Parallel.GlobalBackupTaskParallel == 0 {
		cfg.Parallel.GlobalBackupTaskParallel = DefaultGlobalBackupTaskParallel
	}

	if cfg.Parallel.GlobalDownloadObjectTaskCacheSize == 0 {
		cfg.Parallel.GlobalDownloadObjectTaskCacheSize = DefaultGlobalDownloadObjectTaskCacheSize
	}
	if cfg.Parallel.GlobalChallengePieceTaskCacheSize == 0 {
		cfg.Parallel.GlobalChallengePieceTaskCacheSize = DefaultGlobalChallengePieceTaskCacheSize
	}

	if cfg.GC.GCObjectTimeInterval == 0 {
		cfg.GC.GCObjectTimeInterval = DefaultGlobalBatchGCObjectTimeInterval
	}
	if cfg.GC.GCObjectBlockInterval == 0 {
		cfg.GC.GCObjectBlockInterval = DefaultGlobalGCObjectBlockInterval
	}
	if cfg.GC.GCObjectSafeBlockDistance == 0 {
		cfg.GC.GCObjectSafeBlockDistance = DefaultGlobalGCObjectSafeBlockDistance
	}

	if cfg.GC.GCZombiePieceTimeInterval == 0 {
		cfg.GC.GCZombiePieceTimeInterval = DefaultGlobalGCZombiePieceTimeInterval
	}
	if cfg.GC.GCZombiePieceObjectIDInterval == 0 {
		cfg.GC.GCZombiePieceObjectIDInterval = DefaultGlobalGCZombiePieceObjectIDInterval
	}
	if cfg.GC.GCZombieSafeObjectIDDistance == 0 {
		cfg.GC.GCZombieSafeObjectIDDistance = DefaultGlobalGCZombieSafeObjectIDDistance
	}

	if cfg.GC.GCMetaTimeInterval == 0 {
		cfg.GC.GCMetaTimeInterval = DefaultGlobalGCMetaTimeInterval
	}

	if cfg.Parallel.GlobalSyncConsensusInfoInterval == 0 {
		cfg.Parallel.GlobalSyncConsensusInfoInterval = DefaultGlobalSyncConsensusInfoInterval
	}
	if cfg.Parallel.DiscontinueBucketTimeInterval == 0 {
		cfg.Parallel.DiscontinueBucketTimeInterval = DefaultDiscontinueTimeInterval
	}
	if cfg.Parallel.DiscontinueBucketKeepAliveDays == 0 {
		cfg.Parallel.DiscontinueBucketKeepAliveDays = DefaultDiscontinueBucketKeepAliveDays
	}
	if cfg.Parallel.GlobalRecoveryPieceParallel == 0 {
		cfg.Parallel.GlobalRecoveryPieceParallel = DefaultGlobalRecoveryPieceParallel
	}
	if cfg.Parallel.LoadReplicateTimeout == 0 {
		cfg.Parallel.LoadReplicateTimeout = DefaultLoadReplicateTimeout
	}
	if cfg.Parallel.LoadSealTimeout == 0 {
		cfg.Parallel.LoadSealTimeout = DefaultLoadSealTimeout
	}

	manager.enableLoadTask = cfg.Manager.EnableLoadTask
	manager.enableHealthyChecker = cfg.Manager.EnableHealthyChecker
	manager.loadTaskLimitToReplicate = cfg.Parallel.GlobalReplicatePieceParallel
	manager.loadTaskLimitToSeal = cfg.Parallel.GlobalSealObjectParallel
	manager.loadTaskLimitToGC = cfg.Parallel.GlobalGCObjectParallel

	manager.statisticsOutputInterval = DefaultStatisticsOutputInterval
	manager.maxUploadObjectNumber = cfg.Parallel.GlobalMaxUploadingParallel
	manager.gcObjectTimeInterval = cfg.GC.GCObjectTimeInterval
	manager.gcObjectBlockInterval = cfg.GC.GCObjectBlockInterval
	manager.gcSafeBlockDistance = cfg.GC.GCObjectSafeBlockDistance
	manager.gcZombiePieceEnabled = cfg.GC.EnableGCZombie
	manager.gcZombiePieceTimeInterval = cfg.GC.GCZombiePieceTimeInterval
	manager.gcZombiePieceSafeObjectIDDistance = cfg.GC.GCZombieSafeObjectIDDistance
	manager.gcZombiePieceObjectIDInterval = cfg.GC.GCZombiePieceObjectIDInterval
	manager.gcMetaEnabled = cfg.GC.EnableGCMeta
	manager.gcMetaTimeInterval = cfg.GC.GCMetaTimeInterval
	manager.syncConsensusInfoInterval = cfg.Parallel.GlobalSyncConsensusInfoInterval
	manager.discontinueBucketEnabled = cfg.Parallel.DiscontinueBucketEnabled
	manager.discontinueBucketTimeInterval = cfg.Parallel.DiscontinueBucketTimeInterval
	manager.discontinueBucketKeepAliveDays = cfg.Parallel.DiscontinueBucketKeepAliveDays
	manager.loadReplicateTimeout = cfg.Parallel.LoadReplicateTimeout
	manager.loadSealTimeout = cfg.Parallel.LoadSealTimeout
	manager.taskCh = make(chan task.Task, cfg.Parallel.GlobalBackupTaskParallel)
	manager.uploadQueue = cfg.Customize.NewStrategyTQueueFunc(
		manager.Name()+"-upload-object", cfg.Parallel.GlobalUploadObjectParallel)
	manager.resumableUploadQueue = cfg.Customize.NewStrategyTQueueFunc(
		manager.Name()+"-resumable-upload-object", cfg.Parallel.GlobalUploadObjectParallel)
	manager.replicateQueue = cfg.Customize.NewStrategyTQueueWithLimitFunc(
		manager.Name()+"-replicate-piece", cfg.Parallel.GlobalReplicatePieceParallel)
	manager.recoveryQueue = cfg.Customize.NewStrategyTQueueWithLimitFunc(
		manager.Name()+"-recovery-piece", cfg.Parallel.GlobalRecoveryPieceParallel)
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
	manager.gcBucketMigrationQueue = cfg.Customize.NewStrategyTQueueWithLimitFunc(
		manager.Name()+"-gc-bucket-migration", cfg.Parallel.GlobalGCBucketMigrationParallel)
	manager.migrateGVGQueue = cfg.Customize.NewStrategyTQueueWithLimitFunc(
		manager.Name()+"-migrate-gvg", cfg.Parallel.GlobalMigrateGVGParallel)
	manager.downloadQueue = cfg.Customize.NewStrategyTQueueFunc(
		manager.Name()+"-cache-download-object", cfg.Parallel.GlobalDownloadObjectTaskCacheSize)
	manager.challengeQueue = cfg.Customize.NewStrategyTQueueFunc(
		manager.Name()+"-cache-challenge-piece", cfg.Parallel.GlobalChallengePieceTaskCacheSize)

	if manager.virtualGroupManager, err = cfg.Customize.NewVirtualGroupManagerFunc(manager.baseApp.OperatorAddress(), manager.baseApp.Consensus(), manager.enableHealthyChecker); err != nil {
		return err
	}
	if cfg.Manager.SubscribeSPExitEventIntervalMillisecond == 0 {
		manager.subscribeSPExitEventInterval = DefaultSubscribeSPExitEventIntervalMillisecond
	} else {
		manager.subscribeSPExitEventInterval = cfg.Manager.SubscribeSPExitEventIntervalMillisecond
	}
	if cfg.Manager.SubscribeSwapOutExitEventIntervalMillisecond == 0 {
		manager.subscribeSwapOutEventInterval = DefaultSubscribeSwapOutEventIntervalMillisecond
	} else {
		manager.subscribeSwapOutEventInterval = cfg.Manager.SubscribeSwapOutExitEventIntervalMillisecond
	}
	if cfg.Manager.SubscribeBucketMigrateEventIntervalMillisecond == 0 {
		manager.subscribeBucketMigrateEventInterval = DefaultSubscribeBucketMigrateEventIntervalMillisecond
	} else {
		manager.subscribeBucketMigrateEventInterval = cfg.Manager.SubscribeBucketMigrateEventIntervalMillisecond
	}
	manager.gvgPreferSPList = cfg.Manager.GVGPreferSPList
	manager.recoveryTaskMap = make(map[string]string)

	manager.spBlackList = cfg.Manager.SPBlackList

	manager.recoverObjectStats = make(map[uint64]*ObjectPieceStats)

	return nil
}
