package model

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
	// BsDBSwitchedUser defines env variable name for switched block syncer db user name
	BsDBSwitchedUser = "BS_DB_SWITCHED_USER"
	// BsDBSwitchedPasswd defines env variable name for switched block syncer db user passwd
	BsDBSwitchedPasswd = "BS_DB_SWITCHED_PASSWORD"
	// BsDBSwitchedAddress defines env variable name for switched block syncer db address
	BsDBSwitchedAddress = "BS_DB_SWITCHED_ADDRESS"
	// BsDBSwitchedDataBase defines env variable name for switched block syncer db database
	BsDBSwitchedDataBase = "BS_DB_SWITCHED_DATABASE"

	// SpOperatorAddress defines env variable name for sp operator address
	SpOperatorAddress = "greenfield-storage-provider"
	// DsnBlockSyncer defines env variable name for block syncer dsn
	DsnBlockSyncer = "BLOCK_SYNCER_DSN"
	// DsnBlockSyncerSwitched defines env variable name for block syncer backup dsn
	DsnBlockSyncerSwitched = "BLOCK_SYNCER_DSN_SWITCHED"
)

// SQLDB default configuration parmas
const (
	// DefaultConnMaxLifetime defines the default max liveness time of connection
	DefaultConnMaxLifetime = 60
	// DefaultConnMaxIdleTime defines the default max idle time of connection
	DefaultConnMaxIdleTime = 30
	// DefaultMaxIdleConns defines the default max number of idle connections
	DefaultMaxIdleConns = 16
	// DefaultMaxOpenConns defines the default max number of open connections
	DefaultMaxOpenConns = 32
)

// define all kinds of http constants
const (
	// ContentTypeHeader is used to indicate the media type of the resource
	ContentTypeHeader = "Content-Type"
	// ContentLengthHeader indicates the size of the message body, in bytes
	ContentLengthHeader = "Content-Length"
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
	// ContentDispositionHeader is used to indicate the media disposition of the resource
	ContentDispositionHeader = "Content-Disposition"
	// ContentDispositionAttachmentValue is used to indicate attachment
	ContentDispositionAttachmentValue = "attachment"
	// ContentDispositionInlineValue is used to indicate inline
	ContentDispositionInlineValue = "inline"

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

	SignTypeOffChain   = "OffChainAuth" // sign type - off-chain-auth
	SignTypePersonal   = "PersonalSign" // sign type -  PersonalSign
	SignAlgorithmEddsa = "EDDSA"

	// GetApprovalPath defines get-approval path style suffix
	GetApprovalPath = "/greenfield/admin/v1/get-approval"
	// ActionQuery defines get-approval's type, currently include create bucket and create object
	ActionQuery = "action"
	// UploadProgressQuery defines upload progress query, which is used to route request
	UploadProgressQuery = "upload-progress"
	// GetBucketReadQuotaQuery defines bucket read quota query, which is used to route request
	GetBucketReadQuotaQuery = "read-quota"
	// GetBucketReadQuotaMonthQuery defines bucket read quota query month
	GetBucketReadQuotaMonthQuery = "year-month"
	// ListBucketReadRecordQuery defines list bucket read record query, which is used to route request
	ListBucketReadRecordQuery = "list-read-record"
	// ListBucketReadRecordMaxRecordsQuery defines list read record max num
	ListBucketReadRecordMaxRecordsQuery = "max-records"
	// ListObjectsMaxKeysQuery defines the maximum number of keys returned to the response
	ListObjectsMaxKeysQuery = "max-keys"
	// ListObjectsStartAfterQuery defines where you want to start listing from
	ListObjectsStartAfterQuery = "start-after"
	// ListObjectsContinuationTokenQuery indicates that the list is being continued on this bucket with a token
	ListObjectsContinuationTokenQuery = "continuation-token"
	// ListObjectsDelimiterQuery defines a character you use to group keys
	ListObjectsDelimiterQuery = "delimiter"
	// ListObjectsPrefixQuery defines limits the response to keys that begin with the specified prefix
	ListObjectsPrefixQuery = "prefix"
	// GetBucketMetaQuery defines get bucket metadata query, which is used to route request
	GetBucketMetaQuery = "bucket-meta"
	// GetObjectMetaQuery defines get object metadata query, which is used to route request
	GetObjectMetaQuery = "object-meta"
	// StartTimestampUs defines start timestamp in microsecond, which is used by list read record, [start_ts,end_ts)
	StartTimestampUs = "start-timestamp"
	// EndTimestampUs defines end timestamp in microsecond, which is used by list read record, [start_ts,end_ts)
	EndTimestampUs = "end-timestamp"
	// ChallengePath defines challenge path style suffix
	ChallengePath = "/greenfield/admin/v1/challenge"
	// ReplicateObjectPiecePath defines replicate-object path style
	ReplicateObjectPiecePath = "/greenfield/receiver/v1/replicate-piece"
	// AuthRequestNoncePath defines path to request auth nonce
	AuthRequestNoncePath = "/auth/request_nonce"
	// AuthUpdateKeyPath defines path to update user public key
	AuthUpdateKeyPath = "/auth/update_key"
	// GnfdRequestIDHeader defines trace-id, trace request in sp
	GnfdRequestIDHeader = "X-Gnfd-Request-ID"
	// GnfdAuthorizationHeader defines authorization, verify signature and check authorization
	GnfdAuthorizationHeader = "Authorization"
	// GnfdReceiveMsgHeader defines receive piece data meta
	GnfdReceiveMsgHeader = "X-Gnfd-Receive-Msg"
	// GnfdReplicatePieceApprovalHeader defines secondary approved msg for replicating piece
	GnfdReplicatePieceApprovalHeader = "X-Gnfd-Replicate-Piece-Approval-Msg"
	// GnfdObjectIDHeader defines object id
	GnfdObjectIDHeader = "X-Gnfd-Object-ID"
	// GnfdPieceIndexHeader defines piece idx, which is used by challenge
	GnfdPieceIndexHeader = "X-Gnfd-Piece-Index"
	// GnfdRedundancyIndexHeader defines redundancy idx, which is used by challenge and receiver
	GnfdRedundancyIndexHeader = "X-Gnfd-Redundancy-Index"
	// GnfdIntegrityHashHeader defines integrity hash, which is used by challenge and receiver
	GnfdIntegrityHashHeader = "X-Gnfd-Integrity-Hash"
	// GnfdPieceHashHeader defines piece hash list, which is used by challenge
	GnfdPieceHashHeader = "X-Gnfd-Piece-Hash"
	// GnfdUnsignedApprovalMsgHeader defines unsigned msg, which is used by get-approval
	GnfdUnsignedApprovalMsgHeader = "X-Gnfd-Unsigned-Msg"
	// GnfdSignedApprovalMsgHeader defines signed msg, which is used by get-approval
	GnfdSignedApprovalMsgHeader = "X-Gnfd-Signed-Msg"
	// GnfdPieceSizeHeader defines piece size, which is used to split by receiver
	GnfdPieceSizeHeader = "X-Gnfd-Piece-Size"
	// GnfdReplicateApproval defines SP approval that allow to replicate piece data, which is used by receiver
	GnfdReplicateApproval = "X-Gnfd-Replicate-Approval"
	// GnfdIntegrityHashSignatureHeader defines integrity hash signature, which is used by receiver
	GnfdIntegrityHashSignatureHeader = "X-Gnfd-Integrity-Hash-Signature"
	// GnfdUserAddressHeader defines the user address
	GnfdUserAddressHeader = "X-Gnfd-User-Address"
	// GnfdResponseXMLVersion defines the response xml version
	GnfdResponseXMLVersion = "1.0"

	// off-chain-auth headers

	// GnfdOffChainAuthAppDomainHeader defines the app domain from where user is trying to do the EDDSA auth interactions
	GnfdOffChainAuthAppDomainHeader = "X-Gnfd-App-Domain"
	// GnfdOffChainAuthAppRegNonceHeader defines nonce for which user is trying to register his/her EDDSA public key
	GnfdOffChainAuthAppRegNonceHeader = "X-Gnfd-App-Reg-Nonce"
	// GnfdOffChainAuthAppRegPublicKeyHeader defines the EDDSA public key for which user is trying to register
	GnfdOffChainAuthAppRegPublicKeyHeader = "X-Gnfd-App-Reg-Public-Key"
	// GnfdOffChainAuthAppRegExpiryDateHeader defines the Expiry-Date is the ISO 8601 datetime string (e.g. 2021-09-30T16:25:24Z), used to register the EDDSA public key
	GnfdOffChainAuthAppRegExpiryDateHeader = "X-Gnfd-App-Reg-Expiry-Date"
)

// // define all kinds of size
const (
	// MaxCallMsgSize defines gPRC max send or receive msg size
	MaxCallMsgSize = 25 * 1024 * 1024
	// DefaultStreamBufSize defines gateway stream forward payload buf size
	DefaultStreamBufSize = 16 * 1024
	// MaxRetryCount defines getting the latest height from the RPC client max retry count
	MaxRetryCount = 50
	// DefaultBlockHeightDiff defines default block height diff of main and backup service
	DefaultBlockHeightDiff = 100
	// DefaultCheckDiffPeriod defines check interval of block height diff
	DefaultCheckDiffPeriod = 1
)
