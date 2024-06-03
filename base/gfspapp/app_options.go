package gfspapp

import (
	"math"
	"os"
	"strings"
	"sync"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfsppieceop"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfsprcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfsptqueue"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspvgmgr"
	"github.com/bnb-chain/greenfield-storage-provider/base/gnfd"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsplimit"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/pprof"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/probe"
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	psclient "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
)

const (
	// EnvLocal defines the default environment.
	EnvLocal = "local"
	// EnvDevnet defines the devnet environment.
	EnvDevnet = "devnet"
	// EvnQAnet defines the qanet environment.
	EvnQAnet = "qanet"
	// EvnTestnet defines the testnet environment.
	EvnTestnet = "testnet"
	// EnvMainnet defines the mainnet environment. And as default environment.
	EnvMainnet = "mainnet"

	// DefaultGfSpAppIDPrefix defines the default app id prefix.
	DefaultGfSpAppIDPrefix = "gfsp"
	// DefaultGRPCAddress defines the default gRPC address.
	DefaultGRPCAddress = "localhost:9333"
	// DefaultMetricsAddress defines the default metrics service address.
	DefaultMetricsAddress = "localhost:24367"
	// DefaultPProfAddress defines the default pprof service address.
	DefaultPProfAddress = "localhost:24368"
	// DefaultProbeAddress defines the default probe service address.
	DefaultProbeAddress = "localhost:24369"

	// DefaultChainID defines the default greenfield chainID.
	DefaultChainID = "mechain_1000000-121"
	// DefaultChainAddress defines the default greenfield address.
	DefaultChainAddress = "http://localhost:26750"

	// DefaultMemoryLimit defines the default memory limit for resource manager.
	DefaultMemoryLimit = 8 * 1024 * 1024 * 1024
	// DefaultTaskTotalLimit defines the default total task limit for resource manager.
	DefaultTaskTotalLimit = 10240
	// DefaultHighTaskLimit defines the default high priority task limit for resource manager.
	DefaultHighTaskLimit = 128
	// DefaultMediumTaskLimit defines the default medium priority task limit for resource manager.
	DefaultMediumTaskLimit = 1024
	// DefaultLowTaskLimit defines the default low priority task limit for resource manager.
	DefaultLowTaskLimit = 16

	// DefaultSpMonthlyFreeQuota defines the default value of the free quota sp provides to users each month
	DefaultSpMonthlyFreeQuota = 0
)

const (
	ApproverSuccessGetBucketApproval = "approver_get_bucket_success"
	ApproverFailureGetBucketApproval = "approver_get_bucket_failure"
	ApproverSuccessGetObjectApproval = "approver_get_object_success"
	ApproverFailureGetObjectApproval = "approver_get_object_failure"

	AuthSuccess = "auth_success"
	AuthFailure = "auth_failure"

	DownloaderSuccessGetPiece         = "downloader_get_piece_success"
	DownloaderFailureGetPiece         = "downloader_get_piece_failure"
	DownloaderSuccessGetChallengeInfo = "downloader_get_challenge_info_success"
	DownloaderFailureGetChallengeInfo = "downloader_get_challenge_info_failure"

	ManagerBeginUpload           = "manager_begin_upload_success"
	ManagerFailureBeginUpload    = "manager_begin_upload_failure"
	ManagerSuccessDispatchTask   = "manager_dispatch_task_success"
	ManagerDispatchReplicateTask = "manager_dispatch_replicate_task_success"
	ManagerDispatchSealTask      = "manager_dispatch_seal_task_success"
	ManagerDispatchReceiveTask   = "manager_dispatch_receive_task_success"
	ManagerDispatchGCObjectTask  = "manager_dispatch_gc_object_task_success"
	ManagerDispatchRecoveryTask  = "manager_dispatch_recovery_task_success"
	ManagerNoDispatchTask        = "manager_no_dispatch_task_failure"
	ManagerFailureDispatchTask   = "manager_dispatch_task_failure"
	ManagerReportTask            = "manager_report_task_success"
	ManagerReportUploadTask      = "manager_report_upload_task_success"
	ManagerReportReplicateTask   = "manager_report_replicate_task_success"
	ManagerReportSealTask        = "manager_report_seal_task_success"
	ManagerReportReceiveTask     = "manager_report_receive_task_success"
	ManagerReportGCObjectTask    = "manager_report_gc_object_task_success"
	ManagerReportRecoveryTask    = "manager_report_recovery_task_success"

	ReceiverSuccessReplicatePiece     = "receiver_replicate_piece_success"
	ReceiverFailureReplicatePiece     = "receiver_replicate_piece_failure"
	ReceiverSuccessDoneReplicatePiece = "receiver_done_replicate_piece_success"
	ReceiverFailureDoneReplicatePiece = "receiver_done_replicate_piece_failure"

	SignerSuccess                           = "signer_success"
	SignerFailure                           = "signer_failure"
	SignerSuccessBucketApproval             = "signer_bucket_approval_success"
	SignerFailureBucketApproval             = "signer_bucket_approval_failure"
	SignerSuccessMigrateBucketApproval      = "signer_migrate_bucket_approval_success"
	SignerFailureMigrateBucketApproval      = "signer_migrate_bucket_approval_failure"
	SignerSuccessObjectApproval             = "signer_object_approval_success"
	SignerFailureObjectApproval             = "signer_object_approval_failure"
	SignerSuccessSealObject                 = "signer_seal_object_success"
	SignerFailureSealObject                 = "signer_seal_object_failure"
	SignerSuccessRejectUnSealObject         = "signer_reject_unseal_object_success"
	SignerFailureRejectUnSealObject         = "signer_reject_unseal_object_failure"
	SignerSuccessDiscontinueBucket          = "signer_discontinue_bucket_success"
	SignerFailureDiscontinueBucket          = "signer_discontinue_bucket_failure"
	SignerSuccessIntegrityHash              = "signer_integrity_hash_success"
	SignerFailureIntegrityHash              = "signer_integrity_hash_failure"
	SignerSuccessPing                       = "signer_ping_success"
	SignerFailurePing                       = "signer_ping_failure"
	SignerSuccessPong                       = "signer_pong_success"
	SignerFailurePong                       = "signer_pong_failure"
	SignerSuccessReceiveTask                = "signer_receive_task_success"
	SignerFailureReceiveTask                = "signer_receive_task_failure"
	SignerSuccessReplicateApproval          = "signer_secondary_approval_success"
	SignerFailureReplicateApproval          = "signer_secondary_approval_failure"
	SignerSuccessCreateGVG                  = "signer_create_gvg_success"
	SignerFailureCreateGVG                  = "signer_create_gvg_failure"
	SignerSuccessRecoveryTask               = "signer_recovery_task_success"
	SignerFailureRecoveryTask               = "signer_recovery_task_failure"
	SignerSuccessCompleteMigrateBucket      = "signer_complete_migration_bucket_success"
	SignerFailureCompleteMigrateBucket      = "signer_complete_migration_bucket_failure"
	SignerSuccessSecondarySPMigrationBucket = "signer_secondary_sp_migration_bucket_success"
	SignerFailureSecondarySPMigrationBucket = "signer_secondary_sp_migration_bucket_failure"
	SignerSuccessSwapOut                    = "signer_swap_out_success"
	SignerFailureSwapOut                    = "signer_swap_out_failure"
	SignerSuccessSignSwapOut                = "signer_sign_swap_out_success"
	SignerFailureSignSwapOut                = "signer_sign_swap_out_failure"
	SignerSuccessCompleteSwapOut            = "signer_complete_swap_out_success"
	SignerFailureCompleteSwapOut            = "signer_complete_swap_out_failure"
	SignerSuccessSPExit                     = "signer_sp_exit_success"
	SignerFailureSPExit                     = "signer_sp_exit_failure"
	SignerSuccessCompleteSPExit             = "signer_complete_sp_exit_success"
	SignerFailureCompleteSPExit             = "signer_complete_sp_exit_failure"
	SignerSuccessSPStoragePrice             = "signer_sp_storage_price_success"
	SignerFailureSPStoragePrice             = "signer_sp_storage_price_failure"
	SignerSuccessMigrateGVGTask             = "signer_migrate_gvg_task_success"
	SignerFailureMigrateGVGTask             = "signer_migrate_gvg_task_failure"
	SignerSuccessGfSpBucketMigrateInfo      = "signer_gfsp_bucket_migrate_info_success"
	SignerFailureGfSpBucketMigrateInfo      = "signer_gfsp_bucket_migrate_info_failure"
	SignerSuccessRejectMigrateBucket        = "signer_reject_migrate_bucket_success"
	SignerFailureRejectMigrateBucket        = "signer_reject_migrate_bucket_failure"
	SignerSuccessSwapIn                     = "signer_swap_in_success"
	SignerFailureSwapIn                     = "signer_swap_in_failure"
	SignerSuccessCompleteSwapIn             = "signer_complete_swap_in_success"
	SignerFailureCompleteSwapIn             = "signer_complete_swap_in_failure"
	SignerSuccessCancelSwapIn               = "signer_cancel_swap_in_success"
	SignerFailureCancelSwapIn               = "signer_cancel_swap_in_failure"

	SignerSuccessDeposit = "signer_deposit_success"
	SignerFailureDeposit = "signer_deposit_failure"

	SignerSuccessDeleteGlobalVirtualGroup = "signer_delete_global_virtual_group_success"
	SignerFailureDeleteGlobalVirtualGroup = "signer_delete_global_virtual_group_failure"

	SignerSuccessDelegateUpdateObjectContent = "signer_delegate_update_object_content_success"
	SignerFailureDelegateUpdateObjectContent = "signer_delegate_update_object_content_failure"
	SignerSuccessDelegateCreateObject        = "signer_delegate_create_object_success"
	SignerFailureDelegateCreateObject        = "signer_delegate_create_object_failure"

	UploaderSuccessPutObject = "uploader_put_object_success"
	UploaderFailurePutObject = "uploader_put_object_failure"
)

func DefaultStaticOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Env == "" {
		cfg.Env = EnvMainnet
	}
	if len(cfg.Server) == 0 {
		cfg.Server = GetRegisteredModules()
	}
	if cfg.AppID == "" {
		servers := strings.Join(cfg.Server, `-`)
		cfg.AppID = DefaultGfSpAppIDPrefix + "-" + servers
	}
	app.appID = cfg.AppID
	if cfg.GRPCAddress == "" {
		cfg.GRPCAddress = DefaultGRPCAddress
	}
	app.grpcAddress = cfg.GRPCAddress
	app.operatorAddress = cfg.SpAccount.SpOperatorAddress
	app.chainID = cfg.Chain.ChainID
	app.uploadSpeed = cfg.Task.UploadTaskSpeed
	app.downloadSpeed = cfg.Task.DownloadTaskSpeed
	app.replicateSpeed = cfg.Task.ReplicateTaskSpeed
	app.receiveSpeed = cfg.Task.ReceiveTaskSpeed
	app.sealObjectTimeout = cfg.Task.SealObjectTaskTimeout
	app.gcObjectTimeout = cfg.Task.GcObjectTaskTimeout
	app.gcZombieTimeout = cfg.Task.GcZombieTaskTimeout
	app.gcMetaTimeout = cfg.Task.GcMetaTaskTimeout
	app.sealObjectRetry = cfg.Task.SealObjectTaskRetry
	app.replicateRetry = cfg.Task.ReplicateTaskRetry
	app.receiveConfirmRetry = cfg.Task.ReceiveConfirmTaskRetry
	app.gcObjectRetry = cfg.Task.GcObjectTaskRetry
	app.gcZombieRetry = cfg.Task.GcZombieTaskRetry
	app.gcMetaRetry = cfg.Task.GcMetaTaskRetry
	app.approver = &coremodule.NullModular{}
	app.authenticator = &coremodule.NullModular{}
	app.downloader = &coremodule.NilModular{}
	app.executor = &coremodule.NilModular{}
	app.gater = &coremodule.NullModular{}
	app.manager = &coremodule.NullModular{}
	app.p2p = &coremodule.NilModular{}
	app.receiver = &coremodule.NullReceiveModular{}
	app.signer = &coremodule.NilModular{}
	app.metrics = &coremodule.NilModular{}
	app.pprof = &coremodule.NilModular{}
	app.newRPCServer()
	return nil
}

func DefaultGfSpClientOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Endpoint.ApproverEndpoint == "" {
		cfg.Endpoint.ApproverEndpoint = cfg.GRPCAddress
	}
	if cfg.Endpoint.ManagerEndpoint == "" {
		cfg.Endpoint.ManagerEndpoint = cfg.GRPCAddress
	}
	if cfg.Endpoint.DownloaderEndpoint == "" {
		cfg.Endpoint.DownloaderEndpoint = cfg.GRPCAddress
	}
	if cfg.Endpoint.ReceiverEndpoint == "" {
		cfg.Endpoint.ReceiverEndpoint = cfg.GRPCAddress
	}
	if cfg.Endpoint.MetadataEndpoint == "" {
		cfg.Endpoint.MetadataEndpoint = cfg.GRPCAddress
	}
	if cfg.Endpoint.MetadataEndpoint == "" {
		cfg.Endpoint.MetadataEndpoint = cfg.GRPCAddress
	}
	if cfg.Endpoint.UploaderEndpoint == "" {
		cfg.Endpoint.UploaderEndpoint = cfg.GRPCAddress
	}
	if cfg.Endpoint.P2PEndpoint == "" {
		cfg.Endpoint.P2PEndpoint = cfg.GRPCAddress
	}
	if cfg.Endpoint.SignerEndpoint == "" {
		cfg.Endpoint.SignerEndpoint = cfg.GRPCAddress
	}
	if cfg.Endpoint.AuthenticatorEndpoint == "" {
		cfg.Endpoint.AuthenticatorEndpoint = cfg.GRPCAddress
	}
	app.client = gfspclient.NewGfSpClient(
		cfg.Endpoint.ApproverEndpoint,
		cfg.Endpoint.ManagerEndpoint,
		cfg.Endpoint.DownloaderEndpoint,
		cfg.Endpoint.ReceiverEndpoint,
		cfg.Endpoint.MetadataEndpoint,
		cfg.Endpoint.UploaderEndpoint,
		cfg.Endpoint.P2PEndpoint,
		cfg.Endpoint.SignerEndpoint,
		cfg.Endpoint.AuthenticatorEndpoint,
		!cfg.Monitor.DisableMetrics)
	return nil
}

var spdbOnce = sync.Once{}

func DefaultGfSpDBOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Customize.GfSpDB != nil {
		app.gfSpDB = cfg.Customize.GfSpDB
		return nil
	}
	for _, v := range cfg.Server {
		if v == coremodule.BlockSyncerModularName || v == coremodule.SignModularName || v == coremodule.GateModularName {
			log.Infof("[%s] module doesn't need sp db", v)
			continue
		}
		spdbOnce.Do(func() {
			if val, ok := os.LookupEnv(sqldb.SpDBUser); ok {
				cfg.SpDB.User = val
			}
			if val, ok := os.LookupEnv(sqldb.SpDBPasswd); ok {
				cfg.SpDB.Passwd = val
			}
			if val, ok := os.LookupEnv(sqldb.SpDBAddress); ok {
				cfg.SpDB.Address = val
			}
			if val, ok := os.LookupEnv(sqldb.SpDBDatabase); ok {
				cfg.SpDB.Database = val
			}
			defaultGfSpDB(&cfg.SpDB)
			dbCfg := &cfg.SpDB
			db, err := sqldb.NewSpDB(dbCfg)
			if err != nil {
				log.Panicw("failed to new spdb", "error", err)
				return
			}

			collector, err := db.RegisterStdDBStats()
			if err != nil {
				log.Errorw("failed to register db stats metrics", "error", err)
				return
			}
			metrics.AddMetrics(collector)

			app.gfSpDB = db
		})
	}
	return nil
}

// defaultGfSpDB set default sp db config
func defaultGfSpDB(cfg *config.SQLDBConfig) {
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = sqldb.DefaultConnMaxLifetime
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = sqldb.DefaultConnMaxIdleTime
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = sqldb.DefaultMaxIdleConns
	}
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = sqldb.DefaultMaxOpenConns
	}
	if cfg.User == "" {
		cfg.User = "root"
	}
	if cfg.Passwd == "" {
		cfg.Passwd = "test"
	}
	if cfg.Address == "" {
		cfg.Address = "127.0.0.1:3306"
	}
	if cfg.Database == "" {
		cfg.Database = "storage_provider_db"
	}
	cfg.EnableTracePutEvent = sqldb.DefaultEnableTracePutEvent
}

var bsdbOnce = sync.Once{}

func DefaultGfBsDBOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	for _, v := range cfg.Server {
		if v != coremodule.MetadataModularName {
			log.Infof("[%s] module doesn't need bs db", v)
			continue
		}

		bsdbOnce.Do(func() {
			if val, ok := os.LookupEnv(bsdb.BsDBUser); ok {
				cfg.BsDB.User = val
			}
			if val, ok := os.LookupEnv(bsdb.BsDBPasswd); ok {
				cfg.BsDB.Passwd = val
			}

			defaultGfBsDB(&cfg.BsDB)

			bsDBBlockSyncerMaster, err := bsdb.NewBsDB(cfg)
			if err != nil {
				log.Panicw("failed to new bsdb", "error", err)
				return
			}

			app.gfBsDBMaster = bsDBBlockSyncerMaster
		})
	}
	return nil
}

// defaultGfBsDB cast block syncer db connections, user and password if not loaded from env vars
func defaultGfBsDB(config *config.SQLDBConfig) {
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = sqldb.DefaultConnMaxLifetime
	}
	if config.ConnMaxIdleTime == 0 {
		config.ConnMaxIdleTime = sqldb.DefaultConnMaxIdleTime
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = sqldb.DefaultMaxIdleConns
	}
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = sqldb.DefaultMaxOpenConns
	}
	if config.User == "" {
		config.User = "root"
	}
	if config.Passwd == "" {
		config.Passwd = "test"
	}
	if config.Address == "" {
		config.Address = "127.0.0.1:3306"
	}
	if config.Database == "" {
		config.Database = "block_syncer"
	}
}

var pieceStoreOnce = sync.Once{}

func DefaultGfSpPieceStoreOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Customize.PieceStore != nil {
		app.pieceStore = cfg.Customize.PieceStore
		return nil
	}
	for _, v := range cfg.Server {
		if v == coremodule.ApprovalModularName || v == coremodule.AuthenticationModularName || v == coremodule.SignModularName ||
			v == coremodule.GateModularName || v == coremodule.MetadataModularName || v == coremodule.BlockSyncerModularName {
			log.Infof("[%s] module doesn't need piece store", v)
			continue
		}
		pieceStoreOnce.Do(func() {
			if cfg.PieceStore.Store.Storage == "" {
				cfg.PieceStore.Store.Storage = "file"
			}
			if cfg.PieceStore.Store.BucketURL == "" {
				cfg.PieceStore.Store.BucketURL = "./data"
			}
			if cfg.PieceStore.Store.MaxRetries == 0 {
				cfg.PieceStore.Store.MaxRetries = 5
			}
			if cfg.PieceStore.Store.MinRetryDelay == 0 {
				cfg.PieceStore.Store.MinRetryDelay = 1
			}
			if cfg.PieceStore.Store.IAMType == "" {
				cfg.PieceStore.Store.IAMType = "SA"
			}
			pieceStore, err := psclient.NewStoreClient(&cfg.PieceStore)
			if err != nil {
				log.Panicw("failed to new piece store", "error", err)
				return
			}
			app.pieceStore = pieceStore
		})
	}
	return nil
}

func DefaultGfSpPieceOpOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Customize.PieceOp != nil {
		app.pieceOp = cfg.Customize.PieceOp
		return nil
	}
	app.pieceOp = &gfsppieceop.GfSpPieceOp{}
	return nil
}

func DefaultGfSpTQueueOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Customize.NewStrategyTQueueFunc == nil {
		cfg.Customize.NewStrategyTQueueFunc = gfsptqueue.NewGfSpTQueue
	}
	if cfg.Customize.NewStrategyTQueueWithLimitFunc == nil {
		cfg.Customize.NewStrategyTQueueWithLimitFunc = gfsptqueue.NewGfSpTQueueWithLimit
	}
	if cfg.Customize.NewVirtualGroupManagerFunc == nil {
		cfg.Customize.NewVirtualGroupManagerFunc = gfspvgmgr.NewVirtualGroupManager
	}
	return nil
}

func DefaultGfSpResourceManagerOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Customize.RcLimiter == nil {
		if cfg.Rcmgr.GfSpLimiter != nil {
			cfg.Customize.RcLimiter = cfg.Rcmgr.GfSpLimiter
		} else {
			cfg.Customize.RcLimiter = &gfsplimit.GfSpLimiter{
				System: &gfsplimit.GfSpLimit{
					Memory:              int64(0.9 * float32(DefaultMemoryLimit)),
					Tasks:               DefaultTaskTotalLimit,
					TasksHighPriority:   DefaultHighTaskLimit,
					TasksMediumPriority: DefaultMediumTaskLimit,
					TasksLowPriority:    DefaultLowTaskLimit,
					Fd:                  math.MaxInt32,
					Conns:               math.MaxInt32,
					ConnsInbound:        math.MaxInt32,
					ConnsOutbound:       math.MaxInt32,
				},
			}
		}
	}
	if cfg.Customize.Rcmgr == nil {
		cfg.Customize.Rcmgr = gfsprcmgr.NewResourceManager(cfg.Customize.RcLimiter)
		log.Infow("succeed to init resource manager", "limit", cfg.Customize.RcLimiter.String())
	}
	if !cfg.Rcmgr.DisableRcmgr {
		app.rcmgr = cfg.Customize.Rcmgr
	} else {
		app.rcmgr = &corercmgr.NullResourceManager{}
	}
	return nil
}

func DefaultGfSpConsensusOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Customize.Consensus != nil {
		app.chain = cfg.Customize.Consensus
		return nil
	}
	if cfg.Chain.ChainID == "" {
		cfg.Chain.ChainID = DefaultChainID
	}
	if len(cfg.Chain.ChainAddress) == 0 {
		cfg.Chain.ChainAddress = []string{DefaultChainAddress}
	}
	gnfdCfg := &gnfd.GnfdChainConfig{
		ChainID:      cfg.Chain.ChainID,
		ChainAddress: cfg.Chain.ChainAddress,
	}
	chain, err := gnfd.NewGnfd(gnfdCfg)
	if err != nil {
		return err
	}
	app.chain = chain
	return nil
}

func DefaultGfSpModuleOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	for _, modular := range cfg.Server {
		newFunc := GetNewModularFunc(strings.ToLower(modular))
		module, err := newFunc(app, cfg)
		if err != nil {
			log.Errorw("failed to new modular instance", "name", modular)
			return err
		}
		app.RegisterServices(module)
		switch module.Name() {
		case coremodule.ApprovalModularName:
			app.approver = module.(coremodule.Approver)
		case coremodule.AuthenticationModularName:
			app.authenticator = module.(coremodule.Authenticator)
		case coremodule.DownloadModularName:
			app.downloader = module.(coremodule.Downloader)
		case coremodule.ExecuteModularName:
			app.executor = module.(coremodule.TaskExecutor)
		case coremodule.GateModularName:
			app.gater = module
		case coremodule.ManageModularName:
			app.manager = module.(coremodule.Manager)
		case coremodule.P2PModularName:
			app.p2p = module.(coremodule.P2P)
		case coremodule.ReceiveModularName:
			app.receiver = module.(coremodule.Receiver)
		case coremodule.SignModularName:
			app.signer = module.(coremodule.Signer)
		case coremodule.UploadModularName:
			app.uploader = module.(coremodule.Uploader)
		}
	}
	return nil
}

func DefaultGfSpMetricOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Monitor.DisableMetrics {
		log.Info("disable sp metrics")
		app.metrics = &coremodule.NullModular{}
	} else {
		if cfg.Monitor.MetricsHTTPAddress == "" {
			cfg.Monitor.MetricsHTTPAddress = DefaultMetricsAddress
		}
		app.metrics = metrics.NewMetrics(cfg.Monitor.MetricsHTTPAddress)
		app.RegisterServices(app.metrics)
	}
	return nil
}

func DefaultGfSpPProfOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Monitor.DisablePProf {
		log.Info("disable sp pprof")
		app.pprof = &coremodule.NullModular{}
	} else {
		if cfg.Monitor.PProfHTTPAddress == "" {
			cfg.Monitor.PProfHTTPAddress = DefaultPProfAddress
		}
		app.pprof = pprof.NewPProf(cfg.Monitor.PProfHTTPAddress)
		app.RegisterServices(app.pprof)
	}
	return nil
}

func DefaultGfSpProbeOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Monitor.DisableProbe {
		log.Info("disable sp probe")
		app.probeSvr = &coremodule.NullModular{}
	} else {
		if cfg.Monitor.ProbeHTTPAddress == "" {
			cfg.Monitor.ProbeHTTPAddress = DefaultProbeAddress
		}
		httpProbe := probe.NewHTTPProbe()
		statusProber := probe.Combine(httpProbe, probe.NewInstrumentation())
		app.httpProbe = statusProber

		app.probeSvr = probe.NewProbe(cfg.Monitor.ProbeHTTPAddress, httpProbe)
		app.RegisterServices(app.probeSvr)
	}
	return nil
}

var gfspBaseAppDefaultOptions = []Option{
	DefaultStaticOption,
	DefaultGfSpClientOption,
	DefaultGfSpDBOption,
	DefaultGfBsDBOption,
	DefaultGfSpPieceStoreOption,
	DefaultGfSpPieceOpOption,
	DefaultGfSpResourceManagerOption,
	DefaultGfSpConsensusOption,
	DefaultGfSpTQueueOption,
	DefaultGfSpModuleOption,
	DefaultGfSpMetricOption,
	DefaultGfSpPProfOption,
	DefaultGfSpProbeOption,
}

func NewGfSpBaseApp(cfg *gfspconfig.GfSpConfig, opts ...gfspconfig.Option) (*GfSpBaseApp, error) {
	if cfg.Customize == nil {
		cfg.Customize = &gfspconfig.Customize{}
	}
	if err := cfg.Apply(opts...); err != nil {
		return nil, err
	}
	app := &GfSpBaseApp{}
	for _, opt := range gfspBaseAppDefaultOptions {
		err := opt(app, cfg)
		if err != nil {
			log.Errorw("failed to apply base app opt", "error", err)
			return nil, err
		}
	}
	log.Infof("succeed to init base app, config info: %s", cfg.String())
	return app, nil
}
