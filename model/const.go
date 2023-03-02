package model

// define storage provider include service
const (
	// GatewayService defines the name of gateway service
	GatewayService = "Gateway"
	// UploaderService defines the name of uploader service
	UploaderService = "Uploader"
	// DownloaderService defines the name of downloader service
	DownloaderService = "Downloader"
	// ChallengeService defines the name of challenge service
	ChallengeService = "Challenge"
	// StoneNodeService defines the name of stone node service
	StoneNodeService = "StoneNode"
	// SyncerService defines the name of syncer service
	SyncerService = "Syncer"
	// SignerService defines the name of signer service
	SignerService = "Signer"
	// MetadataService defines the name of metadata service
	MetadataService = "Metadata"
	// BlockSyncerService defines the name of block sync service
	BlockSyncerService = "BlockSyncer"
)

// define storage provider service gGRPC default address
const (
	// UploaderGrpcAddress default HTTP address of uploader
	GatewayHttpAddress = "localhost:9033"
	// UploaderGrpcAddress default gGRPC address of uploader
	UploaderGrpcAddress = "localhost:9133"
	// DownloaderGrpcAddress default gGRPC address of downloader
	DownloaderGrpcAddress = "localhost:9233"
	// ChallengeGrpcAddress default gGRPC address of challenge
	ChallengeGrpcAddress = "localhost:9333"
	// StoneNodeGrpcAddress default gGRPC address of stone node
	StoneNodeGrpcAddress = "localhost:9433"
	// SyncerGrpcAddress default gGRPC address of syncer
	SyncerGrpcAddress = "localhost:9533"
	// SignerGrpcAddress default gGRPC address of signer
	SignerGrpcAddress = "localhost:9633"
)

// environment constants
const (
	// BucketURL defines env variable name for bucket url
	BucketURL = "BUCKET_URL"
	// AWSAccessKey defines env variable name for aws assess key
	AWSAccessKey = "AWS_ACCESS_KEY"
	// AWSSecretKey defines env variable name for aws secret key
	AWSSecretKey = "AWS_SECRET_KEY"
	// AWSSessionToken defines env variable name for aws session token
	AWSSessionToken = "AWS_SESSION_TOKEN"

	// SpDBUser defines env variable name for sp db user name
	SpDBUser = "SP_DB_USER"
	// SpDBPasswd defines env variable name for sp db user passwd
	SpDBPasswd = "SP_DB_PASSWORD"

	// SpOperatorAddress defines env variable name for sp operator address
	SpOperatorAddress = "SP_OPERATOR_PUB_KEY"
	// SpOperatorPrivKey defines env variable name for sp operator priv key
	SpOperatorPrivKey = "SIGNER_OPERATOR_PRIV_KEY"
	// SpFundingPrivKey defines env variable name for sp funding priv key
	SpFundingPrivKey = "SIGNER_FUNDING_PRIV_KEY"
	// SpApprovalPrivKey defines env variable name for sp approval priv key
	SpApprovalPrivKey = "SIGNER_APPROVAL_PRIV_KEY"
	// SpSealPrivKey defines env variable name for sp seal priv key
	SpSealPrivKey = "SIGNER_SEAL_PRIV_KEY"
)

// define cache size
const (
	// LruCacheLimit define maximum number of cached items in service trace queue
	LruCacheLimit = 8192
)

// define piece store constants.
const (
	// BufPoolSize define buffer pool size
	BufPoolSize = 32 << 10
	// ChecksumAlgo define validation Algorithm Name
	ChecksumAlgo = "Crc32c"
)

// RPC config
const (
	// MaxCallMsgSize defines gPRCt max send or recv msg size
	MaxCallMsgSize = 25 * 1024 * 1024
)

// http header constants
const (
	// http header key
	OctetStream               = "application/octet-stream"
	ContentTypeHeader         = "Content-Type"
	ETagHeader                = "ETag"
	ContentLengthHeader       = "Content-Length"
	ContentTypeXMLHeaderValue = "application/xml"
	RangeHeader               = "Range"
	ContentRangeHeader        = "Content-Range"

	// path
	AdminPath          = "/greenfield/admin/v1/"
	SyncerPath         = "/greenfield/syncer/v1/sync-piece"
	GetApprovalSubPath = "get-approval"
	ChallengeSubPath   = "challenge"

	// query key
	TransactionQuery = "transaction"
	ActionQuery      = "action"

	// Greenfield header key
	GnfdRequestIDHeader       = "X-Gnfd-Request-ID"
	GnfdChecksumHeader        = "X-Gnfd-Checksum"
	GnfdIsPrivateHeader       = "X-Gnfd-Is-Private"
	GnfdTransactionHashHeader = "X-Gnfd-Txn-Hash"
	GnfdResourceHeader        = "X-Gnfd-Resource"
	GnfdPreSignatureHeader    = "X-Gnfd-Pre-Signature"

	// GnfdRedundancyTypeHeader can be EC or Replica, EC is default
	GnfdRedundancyTypeHeader      = "X-Gnfd-Redundancy-Type"
	GnfdAuthorizationHeader       = "Authorization"
	GnfdDateHeader                = "X-Gnfd-Date"
	GnfdObjectIDHeader            = "X-Gnfd-Object-ID"
	GnfdPieceIndexHeader          = "X-Gnfd-Piece-Index"
	GnfdRedundancyIndexHeader     = "X-Gnfd-Redundancy-Index"
	GnfdIntegrityHashHeader       = "X-Gnfd-Integrity-Hash"
	GnfdPieceHashHeader           = "X-Gnfd-Piece-Hash"
	GnfdUnsignedApprovalMsgHeader = "X-Gnfd-Unsigned-Msg"
	GnfdSignedApprovalMsgHeader   = "X-Gnfd-Signed-Msg"

	// StoneNode to gateway request header
	GnfdSPIDHeader              = "X-Gnfd-SP-ID"
	GnfdPieceCountHeader        = "X-Gnfd-Piece-Count"
	GnfdApprovalSignatureHeader = "X-Gnfd-Approval-Signature"

	// gateway to StoneNode response header
	GnfdPieceChecksumHeader          = "X-Gnfd-Piece-Checksum"
	GnfdIntegrityHashSignatureHeader = "X-Gnfd-Integrity-Hash-Signature"

	// header value
	ReplicaRedundancyTypeHeaderValue = "Replica"
	InlineRedundancyTypeHeaderValue  = "Inline"

	// signature const value
	SignAlgorithm = "ECDSA-secp256k1"
	SignedMsg     = "SignedMsg"
	Signature     = "Signature"
	SignTypeV1    = "authTypeV1"
	SignTypeV2    = "authTypeV2"
)
