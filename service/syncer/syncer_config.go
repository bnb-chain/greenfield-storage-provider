package syncer

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metalevel"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metasql"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type SyncerConfig struct {
	Address           string
	StorageProvider   string
	MetaDBType        string
	MetaLevelDBConfig *config.LevelDBConfig
	MetaSqlDBConfig   *config.SqlDBConfig
	PieceConfig       *storage.PieceStoreConfig
}

var DefaultSyncerConfig = &SyncerConfig{
	Address:           "127.0.0.1:5324",
	StorageProvider:   "bnb-sp",
	MetaDBType:        model.LevelDB,
	MetaLevelDBConfig: metalevel.DefaultMetaLevelDBConfig,
	MetaSqlDBConfig:   metasql.DefaultMetaSqlDBConfig,
	PieceConfig:       storage.DefaultPieceStoreConfig,
}
