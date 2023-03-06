package stonenode

import (
	"github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type StoneNodeConfig struct {
	SpOperatorAddress string
	GRPCAddress       string
	SignerGrpcAddress string
	SpDBConfig        *config.SQLDBConfig
	PieceStoreConfig  *storage.PieceStoreConfig
	ChainConfig       *greenfield.GreenfieldChainConfig
}
