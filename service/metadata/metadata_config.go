package metadata

import (
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
)

// MetadataConfig is the metadata service config
type MetadataConfig struct {
	GRPCAddress        string
	BsDBConfig         *config.SQLDBConfig
	BsDBSwitchedConfig *config.SQLDBConfig
	// IsMasterDB is used to determine if the master database (BsDBConfig) is currently being used.
	IsMasterDB                 bool
	BsDBSwitchCheckIntervalSec int64
}
