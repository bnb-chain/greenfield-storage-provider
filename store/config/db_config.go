package config

var (
	// MinCache is the minimum amount of memory in megabytes to allocate to leveldb
	// read and write caching, split half and half.
	MinCache = 16

	// MinHandles is the minimum number of files handles to allocate to the open
	// database files.
	MinHandles = 16
)

// SqlDBConfig is sql-db config
type SqlDBConfig struct {
	User     string
	Passwd   string
	Address  string
	Database string
}

// LevelDBConfig is level-db config
type LevelDBConfig struct {
	Path        string
	NameSpace   string
	Cache       int
	FileHandles int
	ReadOnly    bool
}
