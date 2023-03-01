package metadata

import (
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/metasql"
)

type MetadataConfig struct {
	Address         string
	MetaSqlDBConfig *config.SqlDBConfig
}

var DefaultMetadataConfig = &MetadataConfig{
	Address:         "127.0.0.1:9733",
	MetaSqlDBConfig: metasql.DefaultMetaSqlDBConfig,
}
