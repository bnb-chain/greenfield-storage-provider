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
	"github.com/bnb-chain/greenfield-storage-provider/base/gnfd"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsplimit"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/pprof"
	piecestoreclient "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
)

const (
	DefaultGfSpAppIDPrefix = "gfsp"
	DefaultGrpcAddress     = "localhost:9333"
	DefaultMetricsAddress  = "localhost:24367"
	DefaultPprofAddress    = "localhost:24368"

	// SpDBUser defines env variable name for sp db user name
	SpDBUser = "SP_DB_USER"
	// SpDBPasswd defines env variable name for sp db user passwd
	SpDBPasswd = "SP_DB_PASSWORD"
	// SpDBAddress defines env variable name for sp db address
	SpDBAddress = "SP_DB_ADDRESS"
	// SpDBDataBase defines env variable name for sp db database
	SpDBDataBase = "SP_DB_DATABASE"

	// DefaultConnMaxLifetime defines the default max liveness time of connection
	DefaultConnMaxLifetime = 60
	// DefaultConnMaxIdleTime defines the default max idle time of connection
	DefaultConnMaxIdleTime = 30
	// DefaultMaxIdleConns defines the default max number of idle connections
	DefaultMaxIdleConns = 16
	// DefaultMaxOpenConns defines the default max number of open connections
	DefaultMaxOpenConns = 32

	DefaultMemoryLimit     = 8 * 1024 * 1024 * 1024
	DefaultTaskTotalLimit  = 10240
	DefaultHighTaskLimit   = 128
	DefaultMediumTaskLimit = 1024
	DefaultLowTaskLimit    = 16
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
	if cfg.GrpcAddress == "" {
		cfg.GrpcAddress = DefaultGrpcAddress
	}
	app.grpcAddress = cfg.GrpcAddress
	app.operateAddress = cfg.SpAccount.SpOperateAddress
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
	app.newRpcServer()
	return nil
}

func DefaultGfSpClientOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Endpoint.ApproverEndpoint == "" {
		cfg.Endpoint.ApproverEndpoint = DefaultGrpcAddress
	}
	if cfg.Endpoint.ManagerEndpoint == "" {
		cfg.Endpoint.ManagerEndpoint = DefaultGrpcAddress
	}
	if cfg.Endpoint.DownloaderEndpoint == "" {
		cfg.Endpoint.DownloaderEndpoint = DefaultGrpcAddress
	}
	if cfg.Endpoint.ReceiverEndpoint == "" {
		cfg.Endpoint.ReceiverEndpoint = DefaultGrpcAddress
	}
	if cfg.Endpoint.MetadataEndpoint == "" {
		cfg.Endpoint.MetadataEndpoint = DefaultGrpcAddress
	}
	if cfg.Endpoint.RetrieverEndpoint == "" {
		cfg.Endpoint.RetrieverEndpoint = DefaultGrpcAddress
	}
	if cfg.Endpoint.UploaderEndpoint == "" {
		cfg.Endpoint.UploaderEndpoint = DefaultGrpcAddress
	}
	if cfg.Endpoint.P2PEndpoint == "" {
		cfg.Endpoint.P2PEndpoint = DefaultGrpcAddress
	}
	if cfg.Endpoint.SingerEndpoint == "" {
		cfg.Endpoint.SingerEndpoint = DefaultGrpcAddress
	}
	if cfg.Endpoint.AuthorizerEndpoint == "" {
		cfg.Endpoint.AuthorizerEndpoint = DefaultGrpcAddress
	}
	app.client = gfspclient.NewGfSpClient(
		cfg.Endpoint.ApproverEndpoint,
		cfg.Endpoint.ManagerEndpoint,
		cfg.Endpoint.DownloaderEndpoint,
		cfg.Endpoint.ReceiverEndpoint,
		cfg.Endpoint.MetadataEndpoint,
		cfg.Endpoint.RetrieverEndpoint,
		cfg.Endpoint.UploaderEndpoint,
		cfg.Endpoint.P2PEndpoint,
		cfg.Endpoint.SingerEndpoint,
		cfg.Endpoint.AuthorizerEndpoint,
		!cfg.Monitor.DisableMetrics)
	return nil
}

func DefaultGfSpDBOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Customize.GfSpDB != nil {
		app.gfSpDB = cfg.Customize.GfSpDB
		return nil
	}
	if val, ok := os.LookupEnv(SpDBUser); ok {
		cfg.SpDB.User = val
	}
	if val, ok := os.LookupEnv(SpDBPasswd); ok {
		cfg.SpDB.Passwd = val
	}
	if val, ok := os.LookupEnv(SpDBAddress); ok {
		cfg.SpDB.Address = val
	}
	if val, ok := os.LookupEnv(SpDBDataBase); ok {
		cfg.SpDB.Database = val
	}
	if cfg.SpDB.ConnMaxLifetime == 0 {
		cfg.SpDB.ConnMaxLifetime = DefaultConnMaxLifetime
	}
	if cfg.SpDB.ConnMaxIdleTime == 0 {
		cfg.SpDB.ConnMaxIdleTime = DefaultConnMaxIdleTime
	}
	if cfg.SpDB.MaxIdleConns == 0 {
		cfg.SpDB.MaxIdleConns = DefaultMaxIdleConns
	}
	if cfg.SpDB.MaxOpenConns == 0 {
		cfg.SpDB.MaxOpenConns = DefaultMaxOpenConns
	}
	dbCfg := &cfg.SpDB
	db, err := sqldb.NewSpDB(dbCfg)
	if err != nil {
		return err
	}
	app.gfSpDB = db
	return nil
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
		return err
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
		return nil
	}
	if cfg.Customize.NewStrategyTQueueWithLimitFunc == nil {
		cfg.Customize.NewStrategyTQueueWithLimitFunc = gfsptqueue.NewGfSpTQueueWithLimit
		return nil
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
	app.rcmgr = cfg.Customize.Rcmgr
	return nil
}

func DefaultGfSpConsensusOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Customize.Consensus != nil {
		app.chain = cfg.Customize.Consensus
		return nil
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
		newFunc := GetNewModularFunc(modular)
		module, err := newFunc(app, cfg)
		if err != nil {
			log.Errorw("failed to new modular instance", "name", modular)
		}
		RegisterModularInstance(module)
	}
	return nil
}

func DefaultGfSpMetricOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Monitor.DisableMetrics {
		app.metrics = &coremodule.NullModular{}
	}
	if cfg.Customize.Metrics != nil {
		app.metrics = cfg.Customize.Metrics
	} else {
		if cfg.Monitor.MetricsHttpAddress == "" {
			cfg.Monitor.MetricsHttpAddress = DefaultPprofAddress
		}
		app.metrics = metrics.NewMetrics(cfg.Monitor.MetricsHttpAddress)
	}
	RegisterModularInstance(app.metrics)
	return nil
}

func DefaultGfSpPprofOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Monitor.DisablePProf {
		app.pprof = &coremodule.NullModular{}
	}
	if cfg.Customize.PProf != nil {
		app.pprof = cfg.Customize.PProf
	} else {
		if cfg.Monitor.PProfHttpAddress == "" {
			cfg.Monitor.PProfHttpAddress = DefaultMetricsAddress
		}
		app.pprof = pprof.NewPProf(cfg.Monitor.PProfHttpAddress)
	}
	RegisterModularInstance(app.pprof)
	return nil
}

var gfspBaseAppDefaultOptions = []Option{
	DefaultStaticOption,
	DefaultGfSpClientOption,
	DefaultGfSpDBOption,
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
			return nil, err
		}
	}
	return app, nil
}
