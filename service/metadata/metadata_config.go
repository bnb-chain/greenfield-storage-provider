package metadata

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata/store"
)

type MetadataConfig struct {
	Address    string
	MetaDBType string
	DBConfig   store.DBConfig
}

var DefaultMetadataConfig = &MetadataConfig{
	Address:    "127.0.0.1:9733",
	MetaDBType: model.MySqlDB,
	DBConfig:   store.DefaultDBConfig,
}
