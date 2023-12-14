package gfspconfig

import (
	"github.com/pelletier/go-toml/v2"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsplimit"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretaskqueue "github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	"github.com/bnb-chain/greenfield-storage-provider/core/vgmgr"
	mwhttp "github.com/bnb-chain/greenfield-storage-provider/pkg/middleware/http"
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
	Env            string     `comment:"optional"`
	AppID          string     `comment:"optional"`
	Server         []string   `comment:"optional"`
	GRPCAddress    string     `comment:"optional"`
	Customize      *Customize `comment:"optional"`
	SpDB           storeconfig.SQLDBConfig
	BsDB           storeconfig.SQLDBConfig
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
	Rcmgr          RcmgrConfig `comment:"optional"`
	Log            LogConfig
	BlockSyncer    BlockSyncerConfig
	APIRateLimiter mwhttp.RateLimiterConfig
	Manager        ManagerConfig
	GC             GCConfig
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
	ChainID                           string   `comment:"required"`
	ChainAddress                      []string `comment:"required"`
	SealGasLimit                      uint64   `comment:"optional"`
	SealFeeAmount                     uint64   `comment:"optional"`
	RejectSealGasLimit                uint64   `comment:"optional"`
	RejectSealFeeAmount               uint64   `comment:"optional"`
	DiscontinueBucketGasLimit         uint64   `comment:"optional"`
	DiscontinueBucketFeeAmount        uint64   `comment:"optional"`
	CreateGlobalVirtualGroupGasLimit  uint64   `comment:"optional"`
	CreateGlobalVirtualGroupFeeAmount uint64   `comment:"optional"`
	CompleteMigrateBucketGasLimit     uint64   `comment:"optional"`
	CompleteMigrateBucketFeeAmount    uint64   `comment:"optional"`
}

type SpAccountConfig struct {
	SpOperatorAddress  string `comment:"required"`
	OperatorPrivateKey string `comment:"required"`
	FundingPrivateKey  string `comment:"optional"`
	SealPrivateKey     string `comment:"required"`
	ApprovalPrivateKey string `comment:"required"`
	GcPrivateKey       string `comment:"required"`
	BlsPrivateKey      string `comment:"required"`
}

type EndpointConfig struct {
	ApproverEndpoint      string `comment:"required"`
	ManagerEndpoint       string `comment:"required"`
	DownloaderEndpoint    string `comment:"required"`
	ReceiverEndpoint      string `comment:"required"`
	MetadataEndpoint      string `comment:"required"`
	UploaderEndpoint      string `comment:"required"`
	P2PEndpoint           string `comment:"required"`
	SignerEndpoint        string `comment:"required"`
	AuthenticatorEndpoint string `comment:"required"`
}

type ApprovalConfig struct {
	BucketApprovalTimeoutHeight uint64 `comment:"optional"`
	ObjectApprovalTimeoutHeight uint64 `comment:"optional"`
	ReplicatePieceTimeoutHeight uint64 `comment:"optional"`
}

type BucketConfig struct {
	AccountBucketNumber    int64  `comment:"optional"`
	MaxListReadQuotaNumber int64  `comment:"optional"`
	MaxPayloadSize         uint64 `comment:"optional"`
}

type GatewayConfig struct {
	DomainName  string `comment:"required"`
	HTTPAddress string `comment:"required"`
}

type ExecutorConfig struct {
	MaxExecuteNumber                int64   `comment:"optional"`
	AskTaskInterval                 int     `comment:"optional"`
	AskReplicateApprovalTimeout     int64   `comment:"optional"`
	AskReplicateApprovalExFactor    float64 `comment:"optional"`
	ListenSealTimeoutHeight         int     `comment:"optional"`
	ListenSealRetryTimeout          int     `comment:"optional"`
	MaxListenSealRetry              int     `comment:"optional"`
	MaxObjectMigrationRetry         int     `comment:"optional"`
	ObjectMigrationRetryTimeout     int     `comment:"optional"`
	EnableSkipFailedToMigrateObject bool    `comment:"optional"`
	BucketTrafficKeepTimeDay        uint64  `comment:"optional"`
	ReadRecordKeepTimeDay           uint64  `comment:"optional"`
	ReadRecordDeleteLimit           uint64  `comment:"optional"`
}

type P2PConfig struct {
	P2PPrivateKey string   `comment:"optional"`
	P2PAddress    string   `comment:"optional"`
	P2PAntAddress string   `comment:"optional"`
	P2PBootstrap  []string `comment:"optional"`
	P2PPingPeriod int      `comment:"optional"`
}

type GCConfig struct {
	// gc object
	GCObjectTimeInterval      int    `comment:"optional"`
	GCObjectBlockInterval     uint64 `comment:"optional"`
	GCObjectSafeBlockDistance uint64 `comment:"optional"`

	// gc zombie
	EnableGCZombie                bool   `comment:"optional"`
	GCZombieSafeObjectIDDistance  uint64 `comment:"optional"`
	GCZombiePieceTimeInterval     int    `comment:"optional"`
	GCZombiePieceObjectIDInterval uint64 `comment:"optional"`

	// gc meta
	EnableGCMeta       bool `comment:"optional"`
	GCMetaTimeInterval int  `comment:"optional"`
}

type ParallelConfig struct {
	GlobalCreateBucketApprovalParallel int `comment:"optional"`
	GlobalCreateObjectApprovalParallel int `comment:"optional"`
	// upload + replicate + seal
	GlobalMaxUploadingParallel int `comment:"optional"`
	// only upload
	GlobalUploadObjectParallel        int    `comment:"optional"`
	GlobalReplicatePieceParallel      int    `comment:"optional"`
	GlobalSealObjectParallel          int    `comment:"optional"`
	GlobalReceiveObjectParallel       int    `comment:"optional"`
	GlobalRecoveryPieceParallel       int    `comment:"optional"`
	GlobalMigrateGVGParallel          int    `comment:"optional"`
	GlobalBackupTaskParallel          int    `comment:"optional"`
	GlobalDownloadObjectTaskCacheSize int    `comment:"optional"`
	GlobalChallengePieceTaskCacheSize int    `comment:"optional"`
	GlobalSyncConsensusInfoInterval   uint64 `comment:"optional"`

	// gc
	GlobalGCObjectParallel          int `comment:"optional"`
	GlobalGCBucketMigrationParallel int `comment:"optional"`
	GlobalGCZombieParallel          int `comment:"optional"`
	GlobalGCMetaParallel            int `comment:"optional"`

	UploadObjectParallelPerNode         int   `comment:"optional"`
	ReceivePieceParallelPerNode         int   `comment:"optional"`
	DownloadObjectParallelPerNode       int   `comment:"optional"`
	ChallengePieceParallelPerNode       int   `comment:"optional"`
	AskReplicateApprovalParallelPerNode int   `comment:"optional"`
	QuerySPParallelPerNode              int64 `comment:"optional"`

	DiscontinueBucketEnabled       bool `comment:"required"`
	DiscontinueBucketTimeInterval  int  `comment:"optional"`
	DiscontinueBucketKeepAliveDays int  `comment:"required"`

	LoadReplicateTimeout int64 `comment:"optional"`
	LoadSealTimeout      int64 `comment:"optional"`
}

type TaskConfig struct {
	UploadTaskSpeed         int64 `comment:"optional"`
	DownloadTaskSpeed       int64 `comment:"optional"`
	ReplicateTaskSpeed      int64 `comment:"optional"`
	ReceiveTaskSpeed        int64 `comment:"optional"`
	SealObjectTaskTimeout   int64 `comment:"optional"`
	SealObjectTaskRetry     int64 `comment:"optional"`
	ReplicateTaskRetry      int64 `comment:"optional"`
	ReceiveConfirmTaskRetry int64 `comment:"optional"`

	// gc
	GcObjectTaskTimeout int64 `comment:"optional"`
	GcZombieTaskTimeout int64 `comment:"optional"`
	GcMetaTaskTimeout   int64 `comment:"optional"`
	GcObjectTaskRetry   int64 `comment:"optional"`
	GcZombieTaskRetry   int64 `comment:"optional"`
	GcMetaTaskRetry     int64 `comment:"optional"`
}

type MonitorConfig struct {
	DisableMetrics     bool   `comment:"required"`
	DisablePProf       bool   `comment:"required"`
	DisableProbe       bool   `comment:"required"`
	MetricsHTTPAddress string `comment:"required"`
	PProfHTTPAddress   string `comment:"required"`
	ProbeHTTPAddress   string `comment:"required"`
}

type RcmgrConfig struct {
	DisableRcmgr bool `comment:"optional"`
	GfSpLimiter  *gfsplimit.GfSpLimiter
}

type LogConfig struct {
	Level string `comment:"optional"`
	Path  string `comment:"optional"`
}

type BlockSyncerConfig struct {
	Modules          []string `comment:"required"`
	Workers          uint     `comment:"required"`
	BsDBWriteAddress string   `comment:"optional"`
}

type MetadataConfig struct {
	// IsMasterDB is used to determine if the master database (BsDBConfig) is currently being used.
	IsMasterDB                 bool  `comment:"required"`
	BsDBSwitchCheckIntervalSec int64 `comment:"optional"`
}

type ManagerConfig struct {
	EnableLoadTask                                 bool     `comment:"optional"`
	EnableHealthyChecker                           bool     `comment:"optional"`
	SubscribeSPExitEventIntervalMillisecond        uint     `comment:"optional"`
	SubscribeSwapOutExitEventIntervalMillisecond   uint     `comment:"optional"`
	SubscribeBucketMigrateEventIntervalMillisecond uint     `comment:"optional"`
	GVGPreferSPList                                []uint32 `comment:"optional"`
	SPBlackList                                    []uint32 `comment:"optional"`

	// EnableTaskRetryScheduler is used to enable task retry scheduler.
	EnableTaskRetryScheduler bool `comment:"optional"`

	// RejectUnsealThresholdSecond is as the following meanings:
	// retry replicate and seal task when the task's create_timestamp + RejectUnsealThresholdSecond > now.time();
	// retry reject unseal when the task's create_timestamp + RejectUnsealThresholdSecond <= now.time() &&
	// create_timestamp + 2*RejectUnsealThresholdSecond > now.time().
	RejectUnsealThresholdSecond uint64 `comment:"optional"`
}
