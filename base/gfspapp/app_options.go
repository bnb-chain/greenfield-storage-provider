package gfspapp

import (
	"math"
	"os"
	"strings"

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
	"github.com/bnb-chain/greenfield-storage-provider/store/bsdb"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	piecestoreclient "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
)

const (
	// DefaultGfSpAppIDPrefix defines the default app id prefix.
	DefaultGfSpAppIDPrefix = "gfsp"
	// DefaultGRPCAddress defines the default gRPC address.
	DefaultGRPCAddress = "localhost:9333"
	// DefaultMetricsAddress defines the default metrics service address.
	DefaultMetricsAddress = "localhost:24367"
	// DefaultPProfAddress defines the default pprof service address.
	DefaultPProfAddress = "localhost:24368"

	// DefaultChainID defines the default greenfield chainID.
	DefaultChainID = "greenfield_9000-121"
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
)

func DefaultStaticOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if len(cfg.Server) == 0 {
		cfg.Server = GetRegisterModulus()
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
	app.authorizer = &coremodule.NullModular{}
	app.downloader = &coremodule.NilModular{}
	app.executor = &coremodule.NilModular{}
	app.gater = &coremodule.NullModular{}
	app.manager = &coremodule.NullModular{}
	app.p2p = &coremodule.NilModular{}
	app.receiver = &coremodule.NullReceiveModular{}
	app.signer = &coremodule.NilModular{}
	app.metrics = &coremodule.NilModular{}
	app.pprof = &coremodule.NilModular{}
	app.newRpcServer()
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
	if cfg.Endpoint.AuthorizerEndpoint == "" {
		cfg.Endpoint.AuthorizerEndpoint = cfg.GRPCAddress
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
		cfg.Endpoint.AuthorizerEndpoint,
		!cfg.Monitor.DisableMetrics)
	return nil
}

func DefaultGfSpDBOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Customize.GfSpDB != nil {
		app.gfSpDB = cfg.Customize.GfSpDB
		return nil
	}
	if val, ok := os.LookupEnv(sqldb.SpDBUser); ok {
		cfg.SpDB.User = val
	}
	if val, ok := os.LookupEnv(sqldb.SpDBPasswd); ok {
		cfg.SpDB.Passwd = val
	}
	if val, ok := os.LookupEnv(sqldb.SpDBAddress); ok {
		cfg.SpDB.Address = val
	}
	if val, ok := os.LookupEnv(sqldb.SpDBDataBase); ok {
		cfg.SpDB.Database = val
	}
	if cfg.SpDB.ConnMaxLifetime == 0 {
		cfg.SpDB.ConnMaxLifetime = sqldb.DefaultConnMaxLifetime
	}
	if cfg.SpDB.ConnMaxIdleTime == 0 {
		cfg.SpDB.ConnMaxIdleTime = sqldb.DefaultConnMaxIdleTime
	}
	if cfg.SpDB.MaxIdleConns == 0 {
		cfg.SpDB.MaxIdleConns = sqldb.DefaultMaxIdleConns
	}
	if cfg.SpDB.MaxOpenConns == 0 {
		cfg.SpDB.MaxOpenConns = sqldb.DefaultMaxOpenConns
	}
	if cfg.SpDB.User == "" {
		cfg.SpDB.User = "root"
	}
	if cfg.SpDB.Passwd == "" {
		cfg.SpDB.User = "test"
	}
	if cfg.SpDB.Address == "" {
		cfg.SpDB.User = "127.0.0.1:3306"
	}
	if cfg.SpDB.Database == "" {
		cfg.SpDB.Database = "storage_provider_db"
	}
	dbCfg := &cfg.SpDB
	db, err := sqldb.NewSpDB(dbCfg)
	if err != nil {
		log.Warnw("if not use spdb, please ignore: failed to new spdb", "error", err)
		return nil
	}
	app.gfSpDB = db
	return nil
}

func DefaultGfBsDBOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if val, ok := os.LookupEnv(bsdb.BsDBUser); ok {
		cfg.BsDB.User = val
	}
	if val, ok := os.LookupEnv(bsdb.BsDBPasswd); ok {
		cfg.BsDB.Passwd = val
	}
	if val, ok := os.LookupEnv(bsdb.BsDBAddress); ok {
		cfg.BsDB.Address = val
	}
	if val, ok := os.LookupEnv(bsdb.BsDBDataBase); ok {
		cfg.BsDB.Database = val
	}
	if val, ok := os.LookupEnv(bsdb.BsDBSwitchedUser); ok {
		cfg.BsDBBackup.User = val
	}
	if val, ok := os.LookupEnv(bsdb.BsDBSwitchedPasswd); ok {
		cfg.BsDBBackup.Passwd = val
	}
	if val, ok := os.LookupEnv(bsdb.BsDBSwitchedAddress); ok {
		cfg.BsDBBackup.Address = val
	}
	if val, ok := os.LookupEnv(bsdb.BsDBSwitchedDataBase); ok {
		cfg.BsDBBackup.Database = val
	}

	DefaultGfBsDB(&cfg.BsDB)
	DefaultGfBsDB(&cfg.BsDBBackup)

	bsDBBlockSyncerMaster, err := bsdb.NewBsDB(cfg, false)
	if err != nil {
		log.Warnw("if not use bsdb, please ignore: failed to new bsdb", "error", err)
		return nil
	}

	bsDBBlockSyncerBackUp, err := bsdb.NewBsDB(cfg, true)
	if err != nil {
		log.Warnw("if not use bsdb, please ignore: failed to new bsdb", "error", err)
		return nil
	}

	app.gfBsDBMaster = bsDBBlockSyncerMaster
	app.gfBsDBBackup = bsDBBlockSyncerBackUp

	return nil
}

// DefaultGfBsDB cast block syncer db connections, user and password if not loaded from env vars
func DefaultGfBsDB(config *config.SQLDBConfig) {

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
		config.User = "test"
	}
	if config.Address == "" {
		config.User = "127.0.0.1:3306"
	}
	if config.Database == "" {
		config.Database = "block_syncer_db"
	}
}

func DefaultGfSpPieceStoreOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Customize.PieceStore != nil {
		app.pieceStore = cfg.Customize.PieceStore
		return nil
	}
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
	pieceStore, err := piecestoreclient.NewStoreClient(&cfg.PieceStore)
	if err != nil {
		log.Warnw("if not use piece store, please ignore: failed to new piece store", "error", err)
		return nil
	}
	app.pieceStore = pieceStore
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

func DefaultGfSpModulusOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
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
		case coremodule.AuthorizationModularName:
			app.authorizer = module.(coremodule.Authorizer)
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
		case coremodule.SignerModularName:
			app.signer = module.(coremodule.Signer)
		case coremodule.UploadModularName:
			app.uploader = module.(coremodule.Uploader)
		}
	}
	return nil
}

func DefaultGfSpMetricOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Monitor.DisableMetrics {
		app.metrics = &coremodule.NullModular{}
	}
	if cfg.Monitor.MetricsHTTPAddress == "" {
		cfg.Monitor.MetricsHTTPAddress = DefaultMetricsAddress
	}
	app.metrics = metrics.NewMetrics(cfg.Monitor.MetricsHTTPAddress)
	app.RegisterServices(app.metrics)
	return nil
}

func DefaultGfSpPprofOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Monitor.DisablePProf {
		app.pprof = &coremodule.NullModular{}
	}
	if cfg.Monitor.PProfHTTPAddress == "" {
		cfg.Monitor.PProfHTTPAddress = DefaultPProfAddress
	}
	app.pprof = pprof.NewPProf(cfg.Monitor.PProfHTTPAddress)
	app.RegisterServices(app.pprof)
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
	DefaultGfSpModulusOption,
	DefaultGfSpMetricOption,
	DefaultGfSpPprofOption,
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
