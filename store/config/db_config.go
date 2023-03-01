package config

// SQLDBConfig is sql db config
type SQLDBConfig struct {
	User     string
	Passwd   string
	Address  string
	Database string
}

// LevelDBConfig is leveldb config
type LevelDBConfig struct {
	Path        string
	NameSpace   string
	Cache       int
	FileHandles int
	ReadOnly    bool
}
