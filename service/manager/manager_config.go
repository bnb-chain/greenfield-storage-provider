package manager

import (
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
)

// ManagerConfig defines manager service config
type ManagerConfig struct {
	GRPCAddress         string
	SpOperatorAddress   string
	ChainConfig         *gnfd.GreenfieldChainConfig
	SpDBConfig          *config.SQLDBConfig
	MaxUploadConcurrent int // include upload, replicate and seal.
	UploadQueueCap      int
	ReplicateQueueCap   int
	SealQueueCap        int
	GCObjectQueueCap    int
}

// DefaultManagerConfig is the default config.
var DefaultManagerConfig = &ManagerConfig{
	MaxUploadConcurrent: 10000,
	UploadQueueCap:      5000,
	ReplicateQueueCap:   5000,
	SealQueueCap:        5000,
	GCObjectQueueCap:    5000,
}
