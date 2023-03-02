package challenge

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type ChallengeConfig struct {
	GRPCAddress      string
	SpDBConfig       *config.SQLDBConfig
	PieceStoreConfig *storage.PieceStoreConfig
}

var DefaultChallengeConfig = &ChallengeConfig{
	GRPCAddress:      model.ChallengeGRPCAddress,
	SpDBConfig:       config.DefaultSQLDBConfig,
	PieceStoreConfig: storage.DefaultPieceStoreConfig,
}
