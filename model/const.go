package model

// define storage provider support service name.
const (
	GatewayService    = "Gateway"
	UploaderService   = "Uploader"
	DownloaderService = "Downloader"
	ChallengeService  = "Challenge"
	StoneHubService   = "StoneHub"
	StoneNodeService  = "StoneNode"
	SyncerService     = "Syncer"
	SignerService     = "Signer"
)

// define payload data redundancy size.
const (
	InlineSize  = 1 * 1024 * 1024
	SegmentSize = 16 * 1024 * 1024
	EC_M        = 4
	EC_K        = 2
)

type PieceType int32

const (
	SegmentPieceType PieceType = 0
	ECPieceType      PieceType = 1
)

// define piece store constants.
const (
	BufPoolSize  = 32 << 10
	ChecksumAlgo = "Crc32c"
	OctetStream  = "application/octet-stream"
)

// RPC config
const (
	// server and client max send or recv msg size
	MaxCallMsgSize = 25 * 1024 * 1024
)

// Gateway
const (
	// path
	AdminPath          = "/greenfield/admin/v1/"
	GetApprovalSubPath = "get-approval"

	// query key
	TransactionQuery = "transaction"
	PutObjectV2Query = "putobjectv2"
	ActionQuery      = "action"

	// Greenfield header key
	GnfdRequestIDHeader       = "X-Gnfd-Request-ID"
	GnfdContentLengthHeader   = "X-Gnfd-Content-Length"
	GnfdContentTypeHeader     = "X-Gnfd-Content-Type"
	GnfdChecksumHeader        = "X-Gnfd-Checksum"
	GnfdIsPrivateHeader       = "X-Gnfd-Is-Private"
	GnfdTransactionHashHeader = "X-Gnfd-Txn-Hash"
	GnfdResourceHeader        = "X-Gnfd-Resource"
	GnfdPreSignatureHeader    = "X-Gnfd-Pre-Signature"
	// GnfdRedundancyTypeHeader can be EC or Replica, EC is default
	GnfdRedundancyTypeHeader = "X-Gnfd-Redundancy-Type"
	GnfdAuthorizationHeader  = "Authorization"
	GnfdDateHeader           = "X-Gnfd-Date"

	// http header key
	ContentTypeHeader   = "Content-Type"
	ETagHeader          = "ETag"
	ContentLengthHeader = "Content-Length"

	// header value
	ContentTypeXMLHeaderValue        = "application/xml"
	ReplicaRedundancyTypeHeaderValue = "Replica"

	// signature const value
	SignAlgorithm = "ECDSA-secp256k1"
	SignedMsg     = "SignedMsg"
	Signature     = "Signature"
	SignTypeV1    = "authTypeV1"
	SignTypeV2    = "authTypeV2"
)

// define backend store type name.
const (
	MemoryDB string = "memory"
	MySqlDB  string = "mysql"
	LevelDB  string = "leveldb"
)

// environment constants
const (
	// AWS environment constants
	AWSAccessKey    = "AWS_ACCESS_KEY"
	AWSSecretKey    = "AWS_SECRET_KEY"
	AWSSessionToken = "AWS_SESSION_TOKEN"

	// MetaDB environment constants
	MetaDBUser     = "META_DB_USER"
	MetaDBPassword = "META_DB_PASSWORD"

	// JobDB environment constants
	JobDBUser     = "JOB_DB_USER"
	JobDBPassword = "JOB_DB_PASSWORD"
)
