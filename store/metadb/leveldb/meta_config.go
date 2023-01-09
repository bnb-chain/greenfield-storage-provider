package leveldb

type MetaLevelDBConfig struct {
	Path        string
	NameSpace   string
	Cache       int
	FileHandles int
	ReadOnly    bool
}

var DefaultMetaLevelDBConfig = &MetaLevelDBConfig{
	Path:        "./data",
	NameSpace:   "bnb-sp",
	Cache:       4096,
	FileHandles: 100000,
	ReadOnly:    false,
}
