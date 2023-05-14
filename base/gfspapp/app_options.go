package gfspapp

import (
	"math"
	"os"
	"strings"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspmdmgr"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfsppieceop"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfsprcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfsptqueue"
	"github.com/bnb-chain/greenfield-storage-provider/base/gnfd"
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsplimit"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/pprof"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	piecestoreclient "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
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
		cfg.Server = gfspmdmgr.GetRegisterModulus()
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
	app.operateAddress = cfg.SpOperateAddress
	app.uploadSpeed = cfg.UploadSpeed
	app.downloadSpeed = cfg.DownloadSpeed
	app.replicateSpeed = cfg.ReplicateSpeed
	app.receiveSpeed = cfg.ReceiveSpeed
	app.sealObjectTimeout = cfg.SealObjectTimeout
	app.gcObjectTimeout = cfg.GcObjectTimeout
	app.gcZombieTimeout = cfg.GcZombieTimeout
	app.gcMetaTimeout = cfg.GcMetaTimeout
	app.sealObjectRetry = cfg.SealObjectRetry
	app.replicateRetry = cfg.ReplicateRetry
	app.receiveConfirmRetry = cfg.ReceiveConfirmRetry
	app.gcObjectRetry = cfg.GcObjectRetry
	app.gcZombieRetry = cfg.GcZombieRetry
	app.gcMetaRetry = cfg.GcMetaRetry
	app.newRpcServer()
	return nil
}

func DefaultGfSpClientOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.ApproverEndpoint == "" {
		cfg.ApproverEndpoint = DefaultGrpcAddress
	}
	if cfg.ManagerEndpoint == "" {
		cfg.ManagerEndpoint = DefaultGrpcAddress
	}
	if cfg.DownloaderEndpoint == "" {
		cfg.DownloaderEndpoint = DefaultGrpcAddress
	}
	if cfg.ReceiverEndpoint == "" {
		cfg.ReceiverEndpoint = DefaultGrpcAddress
	}
	if cfg.MetadataEndpoint == "" {
		cfg.MetadataEndpoint = DefaultGrpcAddress
	}
	if cfg.RetrieverEndpoint == "" {
		cfg.RetrieverEndpoint = DefaultGrpcAddress
	}
	if cfg.UploaderEndpoint == "" {
		cfg.UploaderEndpoint = DefaultGrpcAddress
	}
	if cfg.P2PEndpoint == "" {
		cfg.P2PEndpoint = DefaultGrpcAddress
	}
	if cfg.SingerEndpoint == "" {
		cfg.SingerEndpoint = DefaultGrpcAddress
	}
	if cfg.AuthorizerEndpoint == "" {
		cfg.AuthorizerEndpoint = DefaultGrpcAddress
	}
	app.client = gfspclient.NewGfSpClient(
		cfg.ApproverEndpoint,
		cfg.ManagerEndpoint,
		cfg.DownloaderEndpoint,
		cfg.ReceiverEndpoint,
		cfg.MetadataEndpoint,
		cfg.RetrieverEndpoint,
		cfg.UploaderEndpoint,
		cfg.P2PEndpoint,
		cfg.SingerEndpoint,
		cfg.AuthorizerEndpoint)
	return nil
}

func DefaultGfSpDBOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.GfSpDB != nil {
		app.gfSpDB = cfg.GfSpDB
		return nil
	}
	if val, ok := os.LookupEnv(SpDBUser); ok {
		cfg.User = val
	}
	if val, ok := os.LookupEnv(SpDBPasswd); ok {
		cfg.Passwd = val
	}
	if val, ok := os.LookupEnv(SpDBAddress); ok {
		cfg.Address = val
	}
	if val, ok := os.LookupEnv(SpDBDataBase); ok {
		cfg.Database = val
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = DefaultConnMaxLifetime
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = DefaultConnMaxIdleTime
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = DefaultMaxIdleConns
	}
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = DefaultMaxOpenConns
	}
	spdbCfg := &config.SQLDBConfig{
		User:            cfg.User,
		Passwd:          cfg.Passwd,
		Address:         cfg.Address,
		Database:        cfg.Database,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
		MaxIdleConns:    cfg.MaxIdleConns,
		MaxOpenConns:    cfg.MaxOpenConns,
	}
	db, err := sqldb.NewSpDB(spdbCfg)
	if err != nil {
		return err
	}
	app.gfSpDB = db
	return nil
}

func DefaultGfSpPieceStoreOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.PieceStore != nil {
		app.pieceStore = cfg.PieceStore
		return nil
	}
	if cfg.StorageType == "" {
		cfg.StorageType = "file"
	}
	if cfg.BucketURL == "" {
		cfg.BucketURL = "./data"
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 5
	}
	if cfg.MinRetryDelay == 0 {
		cfg.MinRetryDelay = 1
	}
	if cfg.IAMType == "" {
		cfg.IAMType = "SA"
	}
	pieceStoreCfg := &storage.PieceStoreConfig{
		Shards: 0,
		Store: storage.ObjectStorageConfig{
			Storage:               cfg.StorageType,
			BucketURL:             cfg.BucketURL,
			MaxRetries:            cfg.MaxRetries,
			MinRetryDelay:         cfg.MinRetryDelay,
			TLSInsecureSkipVerify: cfg.TLSInsecureSkipVerify,
			IAMType:               cfg.IAMType,
		},
	}
	pieceStore, err := piecestoreclient.NewStoreClient(pieceStoreCfg)
	if err != nil {
		return err
	}
	app.pieceStore = pieceStore
	return nil
}

func DefaultGfSpPieceOpOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.PieceOp != nil {
		app.pieceOp = cfg.PieceOp
		return nil
	}
	app.pieceOp = &gfsppieceop.GfSpPieceOp{}
	return nil
}

func DefaultGfSpTQueueOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.NewStrategyTQueueFunc == nil {
		cfg.NewStrategyTQueueFunc = gfsptqueue.NewGfSpTQueue
		return nil
	}
	if cfg.NewStrategyTQueueWithLimitFunc == nil {
		cfg.NewStrategyTQueueWithLimitFunc = gfsptqueue.NewGfSpTQueueWithLimit
		return nil
	}
	return nil
}

func DefaultGfSpResourceManagerOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Rcmgr != nil {
		app.rcmgr = cfg.Rcmgr
		return nil
	}
	if cfg.DisableRcmgr {
		app.rcmgr = &corercmgr.NullResourceManager{}
		return nil
	}
	if cfg.RcmgrLimiter != nil {
		app.rcmgr = gfsprcmgr.NewResourceManager(cfg.RcmgrLimiter)
		return nil
	}
	if cfg.GfSpLimiter != nil {
		app.rcmgr = gfsprcmgr.NewResourceManager(cfg.GfSpLimiter)
		return nil
	}
	cfg.RcmgrLimiter = &gfsplimit.GfSpLimiter{
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
	app.rcmgr = gfsprcmgr.NewResourceManager(cfg.RcmgrLimiter)
	return nil
}

func DefaultGfSpConsensusOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Chain != nil {
		app.chain = cfg.Chain
		return nil
	}
	gnfdCfg := &gnfd.GnfdChainConfig{
		ChainID:      cfg.ChainID,
		ChainAddress: cfg.ChainAddress,
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
		newFunc := gfspmdmgr.GetNewModularFunc(modular)
		module, err := newFunc(app, cfg)
		if err != nil {
			log.Errorw("failed to new modular instance", "name", modular)
		}
		gfspmdmgr.RegisterModularInstance(module)
	}
	return nil
}

func DefaultGfSpMetricOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.DisableMetrics {
		app.metrics = &coremodule.NullModular{}
	}
	if cfg.Metrics != nil {
		app.metrics = cfg.Metrics
	} else {
		if cfg.MetricsHttpAddress == "" {
			cfg.MetricsHttpAddress = DefaultPprofAddress
		}
		app.metrics = metrics.NewMetrics(cfg.MetricsHttpAddress)
	}
	gfspmdmgr.RegisterModularInstance(app.metrics)
	return nil
}

func DefaultGfSpPprofOption(app *GfSpBaseApp, cfg *gfspconfig.GfSpConfig) error {
	if cfg.DisablePProf {
		app.metrics = &coremodule.NullModular{}
	}
	if cfg.PProf != nil {
		app.pprof = cfg.PProf
	} else {
		if cfg.PProfHttpAddress == "" {
			cfg.PProfHttpAddress = DefaultMetricsAddress
		}
		app.pprof = pprof.NewPProf(cfg.PProfHttpAddress)
	}
	gfspmdmgr.RegisterModularInstance(app.pprof)
	return nil
}

var gfspBaseAppDefaultOptions = []gfspconfig.Option{
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
	app := &GfSpBaseApp{}
	opts = append(opts, gfspBaseAppDefaultOptions...)
	for _, opt := range opts {
		err := opt(app, cfg)
		if err != nil {
			return nil, err
		}
	}
	return app, nil
}
