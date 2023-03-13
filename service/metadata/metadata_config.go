package metadata

import (
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
)

// MetadataConfig is the metadata service config
type MetadataConfig struct {
	GRPCAddress string
	SpDBConfig  *config.SQLDBConfig
}
