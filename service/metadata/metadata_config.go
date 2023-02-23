package metadata

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metasql"
)

type MetadataConfig struct {
	Address         string
	MetaDBType      string
	MetaSqlDBConfig *config.SqlDBConfig
}

var DefaultMetadataConfig = &MetadataConfig{
	Address:         model.DefaultMetaServiceAddress,
	MetaDBType:      model.MySqlDB,
	MetaSqlDBConfig: metasql.DefaultMetaSqlDBConfig,
}
