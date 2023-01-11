package model

const (
	InlineSize  = 1 * 1024 * 1024
	SegmentSize = 16 * 1024 * 1024
	EC_M        = 4
	EC_K        = 2
)

// Piece store constants
const (
	BufPoolSize  = 32 << 10
	ChecksumAlgo = "Crc32c"
	OctetStream  = "application/octet-stream"
)

const (
	MemoryDB = "memory"
	MySqlDB  = "MySQL"
	LevelDB  = "leveldb"
)

// RPC config
const (
	// server and client max send or recv msg size
	MaxCallMsgSize = 10 * 1024 * 1024
)
