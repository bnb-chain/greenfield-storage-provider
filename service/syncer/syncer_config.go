package syncer

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type SyncerConfig struct {
	SpOperatorAddress string
	GrpcAddress       string
	SignerGrpcAddress string
	SpDBConfig        *store.SqlDBConfig
	PieceStoreConfig  *storage.PieceStoreConfig
}

var DefaultSyncerConfig = &SyncerConfig{
	SpOperatorAddress: "bnb-sp",
	GrpcAddress:       model.SyncerGrpcAddress,
	SignerGrpcAddress: model.SingerGrpcAddress,
	SpDBConfig:        store.DefaultSqlDBConfig,
	PieceStoreConfig:  storage.DefaultPieceStoreConfig,
}
