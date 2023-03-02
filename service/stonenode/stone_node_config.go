package stonenode

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type StoneNodeConfig struct {
	SpOperatorAddress string
	GRPCAddress       string
	SignerGrpcAddress string
	SPDBConfig        *config.SQLDBConfig
	PieceStoreConfig  *storage.PieceStoreConfig
	ChainConfig       *greenfield.GreenfieldChainConfig
}

var DefaultStoneNodeConfig = &StoneNodeConfig{
	SpOperatorAddress: model.SpOperatorAddress,
	GRPCAddress:       model.StoneNodeGRPCAddress,
	SignerGrpcAddress: model.SignerGRPCAddress,
	SPDBConfig:        config.DefaultSQLDBConfig,
	PieceStoreConfig:  storage.DefaultPieceStoreConfig,
	ChainConfig:       greenfield.DefaultGreenfieldChainConfig,
}
