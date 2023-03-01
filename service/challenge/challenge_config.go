package challenge

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type ChallengeConfig struct {
	GrpcAddress      string
	SPDBConfig       *config.SQLDBConfig
	PieceStoreConfig *storage.PieceStoreConfig
}

var DefaultChallengeConfig = &ChallengeConfig{
	GrpcAddress:      model.ChallengeGrpcAddress,
	SPDBConfig:       config.DefaultSQLDBConfig,
	PieceStoreConfig: storage.DefaultPieceStoreConfig,
}
