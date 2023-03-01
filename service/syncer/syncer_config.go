package syncer

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type SyncerConfig struct {
	SpOperatorAddress string
	GrpcAddress       string
	SignerGrpcAddress string
	SPDBConfig        *config.SQLDBConfig
	PieceStoreConfig  *storage.PieceStoreConfig
}

var DefaultSyncerConfig = &SyncerConfig{
	SpOperatorAddress: model.SpOperatorAddress,
	GrpcAddress:       model.SyncerGrpcAddress,
	SignerGrpcAddress: model.SignerGrpcAddress,
	SPDBConfig:        config.DefaultSQLDBConfig,
	PieceStoreConfig:  storage.DefaultPieceStoreConfig,
}
