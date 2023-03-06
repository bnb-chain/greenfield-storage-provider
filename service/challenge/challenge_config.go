package challenge

import (
	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type ChallengeConfig struct {
	GRPCAddress      string
	SpDBConfig       *config.SQLDBConfig
	PieceStoreConfig *storage.PieceStoreConfig
}
