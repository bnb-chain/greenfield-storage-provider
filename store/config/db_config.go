package config

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
