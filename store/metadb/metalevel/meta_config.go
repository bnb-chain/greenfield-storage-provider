package metalevel

import "github.com/bnb-chain/greenfield-storage-provider/store/config"

var DefaultMetaLevelDBConfig = &config.LevelDBConfig{
	Path:        "./data",
	NameSpace:   "bnb-sp",
	Cache:       4096,
	FileHandles: 100000,
	ReadOnly:    false,
}
