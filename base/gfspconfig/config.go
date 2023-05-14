package gfspconfig

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfsplimit"
	"github.com/bnb-chain/greenfield-storage-provider/core/consensus"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/piecestore"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	coretaskqueue "github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
)

/*
	Local set up necessary configuration:

	SpOperateAddress   string
	User               string
	Passwd             string
	Address            string
	Database           string
	StorageType        string // backend storage type (e.g. s3, file, memory)
	BucketURL          string // the bucket URL of object storage to store data
	ChainID            string
	ChainAddress       []string
	P2PPrivateKey      string
	P2PBootstrap       []string
	OperatorPrivateKey string
	FundingPrivateKey  string
	SealPrivateKey     string
	ApprovalPrivateKey string
	GcPrivateKey       string
*/

/*
	k8s cluster set up need add the following necessary configuration:

	StorageType 	   string
	ApproverEndpoint   string
	ManagerEndpoint    string
	DownloaderEndpoint string
	ReceiverEndpoint   string
	MetadataEndpoint   string
	RetrieverEndpoint  string
	UploaderEndpoint   string
	P2PEndpoint        string
	SingerEndpoint     string
	AuthorizerEndpoint string
*/

/*
	Customizable configuration by implementing the interface

 	GfSpDB                         spdb.SPDB
	PieceStore                     piecestore.PieceStore
	PieceOp                        piecestore.PieceOp
	Rcmgr                          corercmgr.ResourceManager
	RcmgrLimiter                   corercmgr.Limiter
	GfSpLimiter                    *gfsplimit.GfSpLimiter
	Chain                          consensus.Consensus
	Metrics                        module.Modular
	PProf                          module.Modular
	Approver                       module.Approver
	Authorizer                     module.Authorizer
	Downloader                     module.Downloader
	TaskExecutor                   module.TaskExecutor
	Gater                          module.Modular
	Manager                        module.Manager
	P2P                            module.P2P
	Receiver                       module.Receiver
	Retriever                      module.Modular
	Signer                         module.Signer
	Uploader                       module.Uploader
	NewStrategyTQueueFunc          coretaskqueue.NewTQueueOnStrategy
	NewStrategyTQueueWithLimitFunc coretaskqueue.NewTQueueOnStrategyWithLimit
*/

type GfSpConfig struct {
	// gfsp base app configuration
	AppID               string
	SpOperateAddress    string
	Server              []string
	GrpcAddress         string
	UploadSpeed         int64
	DownloadSpeed       int64
	ReplicateSpeed      int64
	ReceiveSpeed        int64
	SealObjectTimeout   int64
	GcObjectTimeout     int64
	GcZombieTimeout     int64
	GcMetaTimeout       int64
	SealObjectRetry     int64
	ReplicateRetry      int64
	ReceiveConfirmRetry int64
	GcObjectRetry       int64
	GcZombieRetry       int64
	GcMetaRetry         int64

	// gfsp app client configuration
	ApproverEndpoint   string
	ManagerEndpoint    string
	DownloaderEndpoint string
	ReceiverEndpoint   string
	MetadataEndpoint   string
	RetrieverEndpoint  string
	UploaderEndpoint   string
	P2PEndpoint        string
	SingerEndpoint     string
	AuthorizerEndpoint string

	// gfsp app db configuration
	GfSpDB          spdb.SPDB
	User            string
	Passwd          string
	Address         string
	Database        string
	ConnMaxLifetime int
	ConnMaxIdleTime int
	MaxIdleConns    int
	MaxOpenConns    int

	// gfsp piece store configuration
	PieceStore            piecestore.PieceStore
	PieceOp               piecestore.PieceOp
	StorageType           string // backend storage type (e.g. s3, file, memory)
	BucketURL             string // the bucket URL of object storage to store data
	MaxRetries            int    // the number of max retries that will be performed
	MinRetryDelay         int64  // the minimum retry delay after which retry will be performed
	TLSInsecureSkipVerify bool   // whether skip the certificate verification of HTTPS requests
	IAMType               string // IAMType is identity and access management type which contains two

	// gfsp resource manager subsystem configuration
	Rcmgr        corercmgr.ResourceManager
	DisableRcmgr bool
	RcmgrLimiter corercmgr.Limiter
	GfSpLimiter  *gfsplimit.GfSpLimiter

	// greenfield chain configuration
	Chain        consensus.Consensus
	ChainID      string
	ChainAddress []string

	// gfsp metrics and pprof configuration
	Metrics            module.Modular
	PProf              module.Modular
	DisableMetrics     bool
	DisablePProf       bool
	MetricsHttpAddress string
	PProfHttpAddress   string

	// gfsp modular configuration
	Approver                       module.Approver
	Authorizer                     module.Authorizer
	Downloader                     module.Downloader
	TaskExecutor                   module.TaskExecutor
	Gater                          module.Modular
	Manager                        module.Manager
	P2P                            module.P2P
	Receiver                       module.Receiver
	Retriever                      module.Modular
	Signer                         module.Signer
	Uploader                       module.Uploader
	NewStrategyTQueueFunc          coretaskqueue.NewTQueueOnStrategy
	NewStrategyTQueueWithLimitFunc coretaskqueue.NewTQueueOnStrategyWithLimit

	// gfsp approval modular configuration
	AccountBucketNumber          int64
	BucketApprovalTimeoutHeight  uint64
	ObjectApprovalTimeoutHeight  uint64
	CreateBucketApprovalParallel int
	CreateObjectApprovalParallel int

	// gfsp download modular configuration
	DownloadObjectParallelPerNode int
	ChallengePieceParallelPerNode int
	BucketFreeQuota               uint64

	// gfsp task executor modular configuration
	ExecutorMaxExecuteNum                uint64
	ExecutorAskTaskInterval              int
	ExecutorAskReplicateApprovalTimeout  int64
	ExecutorAskReplicateApprovalExFactor float64
	ExecutorListenSealTimeoutHeight      int
	ExecutorListenSealRetryTimeout       int
	ExecutorMaxListenSealRetry           int

	// gfsp task gateway modular configuration
	GatewayDomain      string
	GatewayHttpAddress string
	MaxListReadQuota   int64

	// gfsp manager modular configuration
	GlobalMaxUploadingNumber          int // upload + replicate + seal
	GlobalUploadObjectParallel        int // only upload
	GlobalReplicatePieceParallel      int
	GlobalSealObjectParallel          int
	GlobalReceiveObjectParallel       int
	GlobalGCObjectParallel            int
	GlobalGCZombieParallel            int
	GlobalGCMetaParallel              int
	GlobalDownloadObjectTaskCacheSize int
	GlobalChallengePieceTaskCacheSize int
	GlobalBatchGcObjectTimeInterval   int
	GlobalGcObjectBlockInterval       uint64
	GlobalGcObjectSafeBlockDistance   uint64
	GlobalSyncConsensusInfoInterval   uint64

	// gfsp p2p modular configuration
	P2PPrivateKey                       string
	P2PAddress                          string
	P2PBootstrap                        []string
	P2PPingPeriod                       int
	ReplicatePieceTimeoutHeight         uint64
	AskReplicateApprovalParallelPerNode int

	// gfsp receiver modular configuration
	ReceivePieceParallelPerNode int

	// gfsp retriever modular configuration
	QuerySPParallelPerNode uint64

	// gfsp singer modular configuration
	GasLimit           uint64
	OperatorPrivateKey string
	FundingPrivateKey  string
	SealPrivateKey     string
	ApprovalPrivateKey string
	GcPrivateKey       string

	// gfsp uploader modular configuration
	UploadObjectParallelPerNode int
}
