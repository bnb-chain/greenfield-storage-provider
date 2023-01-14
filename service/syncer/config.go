package syncer

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/leveldb"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type SyncerConfig struct {
	Address         string
	StorageProvider string
	MetaDBType      string
	MetaDB          *leveldb.MetaLevelDBConfig
	PieceConfig     *storage.PieceStoreConfig
}

var DefaultSyncerConfig = &SyncerConfig{
	Address:         "127.0.0.1:5324",
	StorageProvider: "bnb-sp",
	MetaDBType:      model.LevelDB,
	MetaDB:          leveldb.DefaultMetaLevelDBConfig,
	PieceConfig:     storage.DefaultPieceStoreConfig,
}
