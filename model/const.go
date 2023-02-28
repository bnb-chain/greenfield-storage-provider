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
	// GatewayHttpAddress default Http address of gateway
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
	// SpOperatorAddress defines env variable name for sp operator address
	SpOperatorAddress = "SP_OPERATOR_ADDRESS"
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
	// ContentTypeHeader and below are standard http protocols
	ContentTypeHeader         = "Content-Type"
	ETagHeader                = "ETag"
	ContentTypeXMLHeaderValue = "application/xml"
	RangeHeader               = "Range"
	ContentRangeHeader        = "Content-Range"
	OctetStream               = "application/octet-stream"

	// SignAlgorithm and below are the signature-related constants
	SignAlgorithm = "ECDSA-secp256k1"
	SignedMsg     = "SignedMsg"
	Signature     = "Signature"
	SignTypeV1    = "authTypeV1"
	SignTypeV2    = "authTypeV2"

	// AdminPath defines get-approval and challenge path style prefix
	AdminPath = "/greenfield/admin/v1/"
	// GetApprovalSubPath defines get-approval path style suffix
	GetApprovalSubPath = "get-approval"
	// ActionQuery defines get-approval's type, currently include create bucket and create object
	ActionQuery = "action"
	// ChallengeSubPath defines challenge path style suffix
	ChallengeSubPath = "challenge"
	// SyncerPath defines sync-object path style
	SyncerPath = "/greenfield/syncer/v1/sync-piece"
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
	// GnfdIntegrityHashHeader defines integrity hash, which is used by challenge and syncer
	GnfdIntegrityHashHeader = "X-Gnfd-Integrity-Hash"
	// GnfdPieceHashHeader defines piece hash list, which is used by challenge
	GnfdPieceHashHeader = "X-Gnfd-Piece-Hash"
	// GnfdUnsignedApprovalMsgHeader defines unsigned msg, which is used by get-approval
	GnfdUnsignedApprovalMsgHeader = "X-Gnfd-Unsigned-Msg"
	// GnfdSignedApprovalMsgHeader defines signed msg, which is used by get-approval
	GnfdSignedApprovalMsgHeader = "X-Gnfd-Signed-Msg"
	// GnfdObjectInfoHeader define object info, which is used by syncer
	GnfdObjectInfoHeader = "X-Gnfd-Object-Info"
	// GnfdReplicateIdxHeader defines replicate idx, which is used by syncer
	GnfdReplicateIdxHeader = "X-Gnfd-Replicate-Idx"
	// GnfdSegmentSizeHeader defines segment size, which is used by syncer
	GnfdSegmentSizeHeader = "X-Gnfd-Segment-Size"
	// GnfdIntegrityHashSignatureHeader defines integrity hash signature, which is used by syncer
	GnfdIntegrityHashSignatureHeader = "X-Gnfd-Integrity-Hash-Signature"
)
