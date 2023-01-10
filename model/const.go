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

var (
	MemoryDB string = "memory"
	MySqlDB  string = "MySql"
	LevelDB  string = "leveldb"
)
