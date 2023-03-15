package model

import "strings"

// define storage provider include service
var (
	// GatewayService defines the name of gateway service
	GatewayService = strings.ToLower("Gateway")
	// UploaderService defines the name of uploader service
	UploaderService = strings.ToLower("Uploader")
	// DownloaderService defines the name of downloader service
	DownloaderService = strings.ToLower("Downloader")
	// ChallengeService defines the name of challenge service
	ChallengeService = strings.ToLower("Challenge")
	// TaskNodeService defines the name of task node service
	TaskNodeService = strings.ToLower("TaskNode")
	// ReceiverService defines the name of receiver service
	ReceiverService = strings.ToLower("Receiver")
	// SignerService defines the name of signer service
	SignerService = strings.ToLower("Signer")
	// MetadataService defines the name of metadata service
	MetadataService = strings.ToLower("Metadata")
	// BlockSyncerService defines the name of block sync service
	BlockSyncerService = strings.ToLower("BlockSyncer")
	// ManagerService defines the name of manager service
	ManagerService = strings.ToLower("Manager")
)

// SpServiceDesc defines the service description in storage provider
var SpServiceDesc = map[string]string{
	GatewayService:     "Receives the sdk request",
	UploaderService:    "Uploads object payload to greenfield",
	DownloaderService:  "Downloads object from the backend and statistical read traffic",
	ChallengeService:   "Provides the ability to query the integrity hash and piece data",
	TaskNodeService:    "Executes background task",
	ReceiverService:    "Receives data pieces of an object from other storage provider and store",
	SignerService:      "Sign the transaction and broadcast to chain",
	MetadataService:    "Provides the ability to query meta data",
	BlockSyncerService: "Syncs block data to db",
}

// define storage provider service gRPC default address
const (
	// GatewayHTTPAddress default HTTP address of gateway
	GatewayHTTPAddress = "localhost:9033"
	// UploaderGRPCAddress default gRPC address of uploader
	UploaderGRPCAddress = "localhost:9133"
	// DownloaderGRPCAddress default gRPC address of downloader
	DownloaderGRPCAddress = "localhost:9233"
	// ChallengeGRPCAddress default gRPC address of challenge
	ChallengeGRPCAddress = "localhost:9333"
	// TaskNodeGRPCAddress default gRPC address of task node
	TaskNodeGRPCAddress = "localhost:9433"
	// ReceiverGRPCAddress default gRPC address of receiver
	ReceiverGRPCAddress = "localhost:9533"
	// SignerGRPCAddress default gRPC address of signer
	SignerGRPCAddress = "localhost:9633"
	// MetadataGRPCAddress default gRPC address of meta data service
	MetadataGRPCAddress = "localhost:9733"
)

// define greenfield chain default address
const (
	// GreenfieldAddress default greenfield chain address
	GreenfieldAddress = "localhost:9090"
	// TendermintAddress default Tendermint address
	TendermintAddress = "http://localhost:26750"
	// GreenfieldChainID default greenfield chainID
	GreenfieldChainID = "greenfield_9000-1741"
	// WhiteListCIDR default whitelist CIDR
	WhiteListCIDR = "127.0.0.1/32"
)

// environment constants
const (
	// SpDBUser defines env variable name for sp db user name
	SpDBUser = "SP_DB_USER"
	// SpDBPasswd defines env variable name for sp db user passwd
	SpDBPasswd = "SP_DB_PASSWORD"
	// SpDBAddress defines env variable name for sp db address
	SpDBAddress = "SP_DB_ADDRESS"
	// SpDBDataBase defines env variable name for sp db database
	SpDBDataBase = "SP_DB_DATABASE"

	// BsDBUser defines env variable name for block syncer db user name
	BsDBUser = "BS_DB_USER"
	// BsDBPasswd defines env variable name for block syncer db user passwd
	BsDBPasswd = "BS_DB_PASSWORD"
	// BsDBAddress defines env variable name for block syncer db address
	BsDBAddress = "BS_DB_ADDRESS"
	// BsDBDataBase defines env variable name for block syncer db database
	BsDBDataBase = "BS_DB_DATABASE"

	// SpOperatorAddress defines env variable name for sp operator address
	SpOperatorAddress = "greenfield-storage-provider"
	// SpSignerAPIKey defines env variable for signer api key
	SpSignerAPIKey = "SIGNER_API_KEY"
	// SpOperatorPrivKey defines env variable name for sp operator priv key
	SpOperatorPrivKey = "SIGNER_OPERATOR_PRIV_KEY"
	// SpFundingPrivKey defines env variable name for sp funding priv key
	SpFundingPrivKey = "SIGNER_FUNDING_PRIV_KEY"
	// SpApprovalPrivKey defines env variable name for sp approval priv key
	SpApprovalPrivKey = "SIGNER_APPROVAL_PRIV_KEY"
	// SpSealPrivKey defines env variable name for sp seal priv key
	SpSealPrivKey = "SIGNER_SEAL_PRIV_KEY"
	// DsnBlockSyncer defines env variable name for block syncer dsn
	DsnBlockSyncer = "BLOCK_SYNCER_DSN"
)

// http header constants
const (
	// ContentTypeHeader is used to indicate the media type of the resource
	ContentTypeHeader = "Content-Type"
	// ETagHeader is an MD5 digest of the object data
	ETagHeader = "ETag"
	// RangeHeader asks the server to send only a portion of an HTTP message back to a client
	RangeHeader = "Range"
	// ContentRangeHeader response HTTP header indicates where in a full body message a partial message belongs
	ContentRangeHeader = "Content-Range"
	// OctetStream is used to indicate the binary files
	OctetStream = "application/octet-stream"
	// ContentTypeJSONHeaderValue is used to indicate json
	ContentTypeJSONHeaderValue = "application/json"
	// ContentTypeXMLHeaderValue is used to indicate xml
	ContentTypeXMLHeaderValue = "application/xml"

	// SignAlgorithm uses secp256k1 with the ECDSA algorithm
	SignAlgorithm = "ECDSA-secp256k1"
	// SignedMsg is the request hash
	SignedMsg = "SignedMsg"
	// Signature is the request signature
	Signature = "Signature"
	// SignTypeV1 is an authentication algorithm, which is used by dapps
	SignTypeV1 = "authTypeV1"
	// SignTypeV2 is an authentication algorithm, which is used by metamask
	SignTypeV2 = "authTypeV2"

	// GetApprovalPath defines get-approval path style suffix
	GetApprovalPath = "/greenfield/admin/v1/get-approval"
	// ActionQuery defines get-approval's type, currently include create bucket and create object
	ActionQuery = "action"
	// GetBucketReadQuotaQuery defines bucket read quota query, which is used to route request
	GetBucketReadQuotaQuery = "read-quota"
	// GetBucketReadQuotaMonthQuery defines bucket read quota query month
	GetBucketReadQuotaMonthQuery = "year-month"
	// ListBucketReadRecordQuery defines list bucket read record query, which is used to route request
	ListBucketReadRecordQuery = "list-read-record"
	// ListBucketReadRecordMaxRecordsQuery defines list read record max num
	ListBucketReadRecordMaxRecordsQuery = "max-records"
	// StartTimestampUs defines start timestamp in microsecond, which is used by list read record, [start_ts,end_ts)
	StartTimestampUs = "start-timestamp"
	// EndTimestampUs defines end timestamp in microsecond, which is used by list read record, [start_ts,end_ts)
	EndTimestampUs = "end-timestamp"
	// ChallengePath defines challenge path style suffix
	ChallengePath = "/greenfield/admin/v1/challenge"
	// SyncPath defines sync-object path style
	SyncPath = "/greenfield/receiver/v1/sync-piece"
	// GnfdRequestIDHeader defines trace-id, trace request in sp
	GnfdRequestIDHeader = "X-Gnfd-Request-ID"
	// GnfdTransactionHashHeader defines blockchain tx-hash
	GnfdTransactionHashHeader = "X-Gnfd-Txn-Hash"
	// GnfdAuthorizationHeader defines authorization, verify signature and check authorization
	GnfdAuthorizationHeader = "Authorization"
	// GnfdObjectIDHeader defines object id
	GnfdObjectIDHeader = "X-Gnfd-Object-ID"
	// GnfdPieceIndexHeader defines piece idx, which is used by challenge
	GnfdPieceIndexHeader = "X-Gnfd-Piece-Index"
	// GnfdRedundancyIndexHeader defines redundancy idx, which is used by challenge
	GnfdRedundancyIndexHeader = "X-Gnfd-Redundancy-Index"
	// GnfdIntegrityHashHeader defines integrity hash, which is used by challenge and receiver
	GnfdIntegrityHashHeader = "X-Gnfd-Integrity-Hash"
	// GnfdPieceHashHeader defines piece hash list, which is used by challenge
	GnfdPieceHashHeader = "X-Gnfd-Piece-Hash"
	// GnfdUnsignedApprovalMsgHeader defines unsigned msg, which is used by get-approval
	GnfdUnsignedApprovalMsgHeader = "X-Gnfd-Unsigned-Msg"
	// GnfdSignedApprovalMsgHeader defines signed msg, which is used by get-approval
	GnfdSignedApprovalMsgHeader = "X-Gnfd-Signed-Msg"
	// GnfdObjectInfoHeader define object info, which is used by receiver
	GnfdObjectInfoHeader = "X-Gnfd-Object-Info"
	// GnfdReplicaIdxHeader defines replica idx, which is used by receiver
	GnfdReplicaIdxHeader = "X-Gnfd-Replica-Idx"
	// GnfdSegmentSizeHeader defines segment size, which is used by receiver
	GnfdSegmentSizeHeader = "X-Gnfd-Segment-Size"
	// GnfdIntegrityHashSignatureHeader defines integrity hash signature, which is used by receiver
	GnfdIntegrityHashSignatureHeader = "X-Gnfd-Integrity-Hash-Signature"
	// GnfdUserAddressHeader defines the user address
	GnfdUserAddressHeader = "X-Gnfd-User-Address"
)

// define all kinds of size
const (
	// LruCacheLimit defines maximum number of cached items in service trace queue
	LruCacheLimit = 8192
	// MaxCallMsgSize defines gPRC max send or receive msg size
	MaxCallMsgSize = 25 * 1024 * 1024
	// MaxRetryCount defines getting the latest height from the RPC client max retry count
	MaxRetryCount = 50
	// DefaultSpFreeReadQuotaSize defines sp bucket's default free quota size, the SP can modify it by itself
	DefaultSpFreeReadQuotaSize = 10 * 1024 * 1024 * 1024
	// DefaultStreamBufSize defines gateway stream forward payload buf size
	DefaultStreamBufSize = 64 * 1024
	// DefaultTimeoutHeight defines approval timeout height
	DefaultTimeoutHeight = 100
	// DefaultPartitionSize defines partition size
	DefaultPartitionSize = 10_000
	// DefaultMaxListLimit defines maximum number of the list request
	DefaultMaxListLimit = 1000
)
