package syncer

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type SyncerConfig struct {
	SPOperatorAddress string
	GRPCAddress       string
	SignerGRPCAddress string
	SPDBConfig        *config.SQLDBConfig
	PieceStoreConfig  *storage.PieceStoreConfig
}

var DefaultSyncerConfig = &SyncerConfig{
	SPOperatorAddress: model.SPOperatorAddress,
	GRPCAddress:       model.SyncerGRPCAddress,
	SignerGRPCAddress: model.SignerGRPCAddress,
	SPDBConfig:        config.DefaultSQLDBConfig,
	PieceStoreConfig:  storage.DefaultPieceStoreConfig,
}
