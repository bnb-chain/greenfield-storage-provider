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

// define backend store type name.
const (
	MemoryDB string = "memory"
	MySqlDB  string = "mysql"
	LevelDB  string = "leveldb"
)

// RPC config
const (
	// server and client max send or recv msg size
	MaxCallMsgSize = 40 * 1024 * 1024
)
