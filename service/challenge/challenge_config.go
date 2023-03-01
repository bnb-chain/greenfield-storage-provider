package challenge

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type ChallengeConfig struct {
	GrpcAddress      string
	SpDBConfig       *store.SqlDBConfig
	PieceStoreConfig *storage.PieceStoreConfig
}

var DefaultChallengeConfig = &ChallengeConfig{
	GrpcAddress:      model.ChallengeGrpcAddress,
	SpDBConfig:       store.DefaultSqlDBConfig,
	PieceStoreConfig: storage.DefaultPieceStoreConfig,
}
