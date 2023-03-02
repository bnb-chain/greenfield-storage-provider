package syncer

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type SyncerConfig struct {
	SpOperatorAddress string
	GRPCAddress       string
	SignerGRPCAddress string
	SpDBConfig        *config.SQLDBConfig
	PieceStoreConfig  *storage.PieceStoreConfig
}

var DefaultSyncerConfig = &SyncerConfig{
	SpOperatorAddress: model.SpOperatorAddress,
	GRPCAddress:       model.SyncerGRPCAddress,
	SignerGRPCAddress: model.SignerGRPCAddress,
	SpDBConfig:        config.DefaultSQLDBConfig,
	PieceStoreConfig:  storage.DefaultPieceStoreConfig,
}
