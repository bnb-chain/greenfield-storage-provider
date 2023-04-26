package tasknode

import (
	"github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

// TaskNodeConfig defines TaskNode service config
type TaskNodeConfig struct {
	SpOperatorAddress  string
	GRPCAddress        string
	SignerGrpcAddress  string
	P2PGrpcAddress     string
	ManagerGrpcAddress string
	SpDBConfig         *config.SQLDBConfig
	PieceStoreConfig   *storage.PieceStoreConfig
	ChainConfig        *greenfield.GreenfieldChainConfig
}
