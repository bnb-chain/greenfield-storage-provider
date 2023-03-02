package stonenode

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type StoneNodeConfig struct {
	SPOperatorAddress string
	GRPCAddress       string
	SignerGrpcAddress string
	SPDBConfig        *config.SQLDBConfig
	PieceStoreConfig  *storage.PieceStoreConfig
	ChainConfig       *greenfield.GreenfieldChainConfig
}

var DefaultStoneNodeConfig = &StoneNodeConfig{
	SPOperatorAddress: model.SPOperatorAddress,
	GRPCAddress:       model.StoneNodeGRPCAddress,
	SignerGrpcAddress: model.SignerGRPCAddress,
	SPDBConfig:        config.DefaultSQLDBConfig,
	PieceStoreConfig:  storage.DefaultPieceStoreConfig,
	ChainConfig:       greenfield.DefaultGreenfieldChainConfig,
}
