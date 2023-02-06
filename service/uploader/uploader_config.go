package uploader

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metalevel"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metasql"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type UploaderConfig struct {
	StorageProvider        string
	Address                string
	StoneHubServiceAddress string
	PieceStoreConfig       *storage.PieceStoreConfig
	MetaDBType             string
	MetaLevelDBConfig      *config.LevelDBConfig
	MetaSqlDBConfig        *config.SqlDBConfig
}

var DefaultUploaderConfig = &UploaderConfig{
	StorageProvider:        "bnb-sp",
	Address:                "127.0.0.1:5311",
	StoneHubServiceAddress: "127.0.0.1:5323",
	PieceStoreConfig:       storage.DefaultPieceStoreConfig,
	MetaDBType:             model.LevelDB,
	MetaLevelDBConfig:      metalevel.DefaultMetaLevelDBConfig,
	MetaSqlDBConfig:        metasql.DefaultMetaSqlDBConfig,
}
