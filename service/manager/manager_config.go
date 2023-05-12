package manager

import (
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

// ManagerConfig defines manager service config
type ManagerConfig struct {
	SpOperatorAddress   string
	ChainConfig         *gnfd.GreenfieldChainConfig
	SpDBConfig          *config.SQLDBConfig
	MetadataGrpcAddress string
	PieceStoreConfig    *storage.PieceStoreConfig
}
