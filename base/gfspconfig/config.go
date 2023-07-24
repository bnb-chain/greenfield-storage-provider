package gfspconfig

import (
	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	"github.com/pelletier/go-toml/v2"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsplimit"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretaskqueue "github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	localhttp "github.com/bnb-chain/greenfield-storage-provider/pkg/middleware/http"
	storeconfig "github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type Option = func(cfg *GfSpConfig) error

// Customize defines the interface for developer to customize own implement, the GfSp base
// app will call the customized implement.
type Customize struct {
	GfSpDB                         spdb.SPDB
	PieceStore                     piecestore.PieceStore
	PieceOp                        piecestore.PieceOp
	Rcmgr                          corercmgr.ResourceManager
	RcLimiter                      corercmgr.Limiter
	Consensus                      consensus.Consensus
	NewTQueueFunc                  coretaskqueue.NewTQueue
	NewTQueueWithLimit             coretaskqueue.NewTQueueWithLimit
	NewStrategyTQueueFunc          coretaskqueue.NewTQueueOnStrategy
	NewStrategyTQueueWithLimitFunc coretaskqueue.NewTQueueOnStrategyWithLimit
	NewVirtualGroupManagerFunc     vgmgr.NewVirtualGroupManager
}

// GfSpConfig defines the GfSp configuration.
type GfSpConfig struct {
	Env            string
	AppID          string
	Server         []string
	GRPCAddress    string
	Customize      *Customize
	SpDB           storeconfig.SQLDBConfig
	BsDB           storeconfig.SQLDBConfig
	BsDBBackup     storeconfig.SQLDBConfig
	PieceStore     storage.PieceStoreConfig
	Chain          ChainConfig
	SpAccount      SpAccountConfig
	Endpoint       EndpointConfig
	Approval       ApprovalConfig
	Bucket         BucketConfig
	Gateway        GatewayConfig
	Executor       ExecutorConfig
	P2P            P2PConfig
	Parallel       ParallelConfig
	Task           TaskConfig
	Monitor        MonitorConfig
	Rcmgr          RcmgrConfig
	Log            LogConfig
	Metadata       MetadataConfig
	BlockSyncer    BlockSyncerConfig
	APIRateLimiter localhttp.RateLimiterConfig
	Manager        ManagerConfig
}

// Apply sets the customized implement to the GfSp configuration, it will be called
// before init GfSp base app.
func (cfg *GfSpConfig) Apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return err
		}
	}
	return nil
}

// String returns the detail GfSp configuration.
func (cfg *GfSpConfig) String() string {
	customize := cfg.Customize
	cfg.Customize = nil
	bz, err := toml.Marshal(cfg)
	if err != nil {
		return ""
	}
	cfg.Customize = customize
	return string(bz)
}

type ChainConfig struct {
	ChainID                           string
	ChainAddress                      []string
	SealGasLimit                      uint64
	SealFeeAmount                     uint64
	RejectSealGasLimit                uint64
	RejectSealFeeAmount               uint64
	DiscontinueBucketGasLimit         uint64
	DiscontinueBucketFeeAmount        uint64
	CreateGlobalVirtualGroupGasLimit  uint64
	CreateGlobalVirtualGroupFeeAmount uint64
}

type SpAccountConfig struct {
	SpOperatorAddress  string
	OperatorPrivateKey string
	FundingPrivateKey  string
	SealPrivateKey     string
	SealBlsPrivateKey  string
	ApprovalPrivateKey string
	GcPrivateKey       string
}

type EndpointConfig struct {
	ApproverEndpoint      string
	ManagerEndpoint       string
	DownloaderEndpoint    string
	ReceiverEndpoint      string
	MetadataEndpoint      string
	UploaderEndpoint      string
	P2PEndpoint           string
	SignerEndpoint        string
	AuthenticatorEndpoint string
}

type ApprovalConfig struct {
	BucketApprovalTimeoutHeight uint64
	ObjectApprovalTimeoutHeight uint64
	ReplicatePieceTimeoutHeight uint64
}

type BucketConfig struct {
	AccountBucketNumber    int64
	FreeQuotaPerBucket     uint64
	MaxListReadQuotaNumber int64
	MaxPayloadSize         uint64
}

type GatewayConfig struct {
	DomainName  string
	HTTPAddress string
}

type ExecutorConfig struct {
	MaxExecuteNumber             int64
	AskTaskInterval              int
	AskReplicateApprovalTimeout  int64
	AskReplicateApprovalExFactor float64
	ListenSealTimeoutHeight      int
	ListenSealRetryTimeout       int
	MaxListenSealRetry           int
}

type P2PConfig struct {
	P2PPrivateKey string
	P2PAddress    string
	P2PAntAddress string
	P2PBootstrap  []string
	P2PPingPeriod int
}

type ParallelConfig struct {
	GlobalCreateBucketApprovalParallel int
	GlobalCreateObjectApprovalParallel int
	GlobalMaxUploadingParallel         int // upload + replicate + seal
	GlobalUploadObjectParallel         int // only upload
	GlobalReplicatePieceParallel       int
	GlobalSealObjectParallel           int
	GlobalReceiveObjectParallel        int
	GlobalGCObjectParallel             int
	GlobalGCZombieParallel             int
	GlobalGCMetaParallel               int
	GlobalRecoveryPieceParallel        int
	GlobalMigrateGVGParallel           int
	GlobalDownloadObjectTaskCacheSize  int
	GlobalChallengePieceTaskCacheSize  int
	GlobalBatchGcObjectTimeInterval    int
	GlobalGcObjectBlockInterval        uint64
	GlobalGcObjectSafeBlockDistance    uint64
	GlobalSyncConsensusInfoInterval    uint64

	UploadObjectParallelPerNode         int
	ReceivePieceParallelPerNode         int
	DownloadObjectParallelPerNode       int
	ChallengePieceParallelPerNode       int
	AskReplicateApprovalParallelPerNode int
	QuerySPParallelPerNode              int64

	DiscontinueBucketEnabled       bool
	DiscontinueBucketTimeInterval  int
	DiscontinueBucketKeepAliveDays int

	LoadReplicateTimeout int64
	LoadSealTimeout      int64
}

type TaskConfig struct {
	UploadTaskSpeed         int64
	DownloadTaskSpeed       int64
	ReplicateTaskSpeed      int64
	ReceiveTaskSpeed        int64
	SealObjectTaskTimeout   int64
	GcObjectTaskTimeout     int64
	GcZombieTaskTimeout     int64
	GcMetaTaskTimeout       int64
	SealObjectTaskRetry     int64
	ReplicateTaskRetry      int64
	ReceiveConfirmTaskRetry int64
	GcObjectTaskRetry       int64
	GcZombieTaskRetry       int64
	GcMetaTaskRetry         int64
}

type MonitorConfig struct {
	DisableMetrics     bool
	DisablePProf       bool
	MetricsHTTPAddress string
	PProfHTTPAddress   string
}

type RcmgrConfig struct {
	DisableRcmgr bool
	GfSpLimiter  *gfsplimit.GfSpLimiter
}

type LogConfig struct {
	Level string
	Path  string
}

type BlockSyncerConfig struct {
	Modules      []string
	Dsn          string
	DsnSwitched  string
	Workers      uint
	EnableDualDB bool
}

type MetadataConfig struct {
	// IsMasterDB is used to determine if the master database (BsDBConfig) is currently being used.
	IsMasterDB                 bool
	BsDBSwitchCheckIntervalSec int64
}

type ManagerConfig struct {
	EnableLoadTask                         bool
	SubscribeSPExitEventIntervalSec        int
	SubscribeBucketMigrateEventIntervalSec int
	GVGPreferSPList                        []uint32
}
