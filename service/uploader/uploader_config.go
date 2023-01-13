package uploader

import (
	"github.com/bnb-chain/inscription-storage-provider/store/metadb/leveldb"
	"github.com/bnb-chain/inscription-storage-provider/store/piecestore/storage"
)

type UploaderConfig struct {
	StorageProvider        string
	Address                string
	StoneHubServiceAddress string
	PieceStoreConfig       *storage.PieceStoreConfig
	MetaDBConfig           *leveldb.MetaLevelDBConfig
}

var DefaultUploaderConfig = &UploaderConfig{
	StorageProvider:        "bnb-sp",
	Address:                "127.0.0.1:5311",
	StoneHubServiceAddress: "127.0.0.1:5323",
	PieceStoreConfig:       storage.DefaultPieceStoreConfig,
	MetaDBConfig:           leveldb.DefaultMetaLevelDBConfig,
}
