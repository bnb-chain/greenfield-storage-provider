package metadata

import (
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
)

type MetadataConfig struct {
	Address    string
	SpDBConfig *config.SQLDBConfig
}

var DefaultMetadataConfig = &MetadataConfig{
	Address:    "127.0.0.1:9733",
	SpDBConfig: config.DefaultSQLDBConfig,
}
