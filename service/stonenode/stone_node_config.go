package stonenode

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type StoneNodeConfig struct {
	SpOperatorAddress string
	GrpcAddress       string
	SignerGrpcAddress string
	SpDBConfig        *config.SQLDBConfig
	PieceStoreConfig  *storage.PieceStoreConfig
	ChainConfig       *greenfield.GreenfieldChainConfig
}

var DefaultStoneNodeConfig = &StoneNodeConfig{
	SpOperatorAddress: model.SpOperatorAddress,
	GrpcAddress:       model.StoneNodeGrpcAddress,
	SignerGrpcAddress: model.SignerGrpcAddress,
	SpDBConfig:        config.DefaultSQLDBConfig,
	PieceStoreConfig:  storage.DefaultPieceStoreConfig,
	ChainConfig:       greenfield.DefaultGreenfieldChainConfig,
}
