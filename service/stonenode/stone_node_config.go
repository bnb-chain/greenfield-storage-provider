package stonenode

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

type StoneNodeConfig struct {
	StorageProvider        string
	Address                string
	GatewayAddress         []string
	StoneHubServiceAddress string
	P2PServiceAddress      string

	PieceStoreConfig *storage.PieceStoreConfig
	StoneJobLimit    int64
}

var DefaultStoneNodeConfig = &StoneNodeConfig{
	StorageProvider:        model.StorageProvider,
	Address:                model.DefaultStoneNodeAddress,
	StoneHubServiceAddress: model.DefaultStoneHubAddress,
	P2PServiceAddress:      model.DefaultP2PServiceAddress,
	PieceStoreConfig:       storage.DefaultPieceStoreConfig,
	GatewayAddress:         DefaultSecondaryGateway,
	StoneJobLimit:          DefaultStoneJobLimit,
}

var DefaultSecondaryGateway = []string{
	"localhost:9034",
	"localhost:9035",
	"localhost:9036",
	"localhost:9037",
	"localhost:9038",
	"localhost:9039",
}

func overrideConfigFromEnv(config *StoneNodeConfig) {
	config.StorageProvider = model.StorageProvider
}
