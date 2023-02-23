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
	StorageProvider:        model.StorageProvider,
	Address:                model.DefaultUploaderAddress,
	StoneHubServiceAddress: model.DefaultStoneHubAddress,
	PieceStoreConfig:       storage.DefaultPieceStoreConfig,
	MetaDBType:             model.LevelDB,
	MetaLevelDBConfig:      metalevel.DefaultMetaLevelDBConfig,
	MetaSqlDBConfig:        metasql.DefaultMetaSqlDBConfig,
}

func overrideConfigFromEnv(config *UploaderConfig) {
	config.StorageProvider = model.StorageProvider
}
