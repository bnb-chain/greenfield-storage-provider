package metadata

import (
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
)

// MetadataConfig is the metadata service config
type MetadataConfig struct {
	GRPCAddress        string
	BsDBConfig         *config.SQLDBConfig
	BsDBSwitchedConfig *config.SQLDBConfig
	// BSDBFlag is used to determine which DB is currently being used
	BSDBFlag bool
}
