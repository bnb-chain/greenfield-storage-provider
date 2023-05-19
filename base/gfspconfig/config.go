package gfspconfig

import (
	"github.com/pelletier/go-toml/v2"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsplimit"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretaskqueue "github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
	storeconfig "github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type Option = func(cfg *GfSpConfig) error

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
}

type GfSpConfig struct {
	AppID       string
	Server      []string
	GrpcAddress string
	SpDB        storeconfig.SQLDBConfig
	PieceStore  storage.PieceStoreConfig
	Chain       ChainConfig
	SpAccount   SpAccountConfig
	Endpoint    EndpointConfig
	Approval    ApprovalConfig
	Bucket      BucketConfig
	Gateway     GatewayConfig
	Executor    ExecutorConfig
	P2P         P2PConfig
	Parallel    ParallelConfig
	Task        TaskConfig
	Monitor     MonitorConfig
	Rcmgr       RcmgrConfig
	Customize   *Customize
	BlockSyncer BlockSyncerConfig
}

func (cfg *GfSpConfig) Apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return err
		}
	}
	return nil
}

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
	ChainID      string
	ChainAddress []string
	GasLimit     uint64
}

type SpAccountConfig struct {
	SpOperateAddress   string
	OperatorPrivateKey string
	FundingPrivateKey  string
	SealPrivateKey     string
	ApprovalPrivateKey string
	GcPrivateKey       string
}

type EndpointConfig struct {
	ApproverEndpoint   string
	ManagerEndpoint    string
	DownloaderEndpoint string
	ReceiverEndpoint   string
	MetadataEndpoint   string
	RetrieverEndpoint  string
	UploaderEndpoint   string
	P2PEndpoint        string
	SignerEndpoint     string
	AuthorizerEndpoint string
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
}

type GatewayConfig struct {
	Domain      string
	HttpAddress string
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
	MetricsHttpAddress string
	PProfHttpAddress   string
}

type RcmgrConfig struct {
	DisableRcmgr bool
	GfSpLimiter  *gfsplimit.GfSpLimiter
}

type BlockSyncerConfig struct {
	Modules        []string
	Dsn            string
	DsnSwitched    string
	RecreateTables bool
	Workers        uint
	EnableDualDB   bool
}
