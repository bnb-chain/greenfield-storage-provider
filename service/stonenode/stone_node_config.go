package stonenode

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/store"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type StoneNodeConfig struct {
	SpOperatorAddress string
	GrpcAddress       string
	SignerGrpcAddress string
	SpDBConfig        *store.SqlDBConfig
	PieceStoreConfig  *storage.PieceStoreConfig
	ChainConfig       *greenfield.GreenfieldChainConfig
}

var DefaultStoneNodeConfig = &StoneNodeConfig{
	SpOperatorAddress: "bnb-sp",
	GrpcAddress:       model.StoneNodeGrpcAddress,
	SignerGrpcAddress: model.SignerGrpcAddress,
	SpDBConfig:        store.DefaultSqlDBConfig,
	PieceStoreConfig:  storage.DefaultPieceStoreConfig,
	ChainConfig:       greenfield.DefaultGreenfieldChainConfig,
}
