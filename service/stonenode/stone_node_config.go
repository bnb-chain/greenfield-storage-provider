package stonenode

import "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"

type StoneNodeConfig struct {
	Address                string
	GatewayAddress         []string
	StoneHubServiceAddress string
	SyncerAddress          []string
	StorageProvider        string
	PieceConfig            *storage.PieceStoreConfig
	StoneJobLimit          int64
}

var DefaultStoneNodeConfig = &StoneNodeConfig{
	Address:                "127.0.0.1:9433",
	GatewayAddress:         []string{"127.0.0.1:9034", "127.0.0.1:9035", "127.0.0.1:9036", "127.0.0.1:9037", "127.0.0.1:9038", "127.0.0.1:9039"},
	StoneHubServiceAddress: "127.0.0.1:9333",
	SyncerAddress:          []string{"127.0.0.1:9543", "127.0.0.1:9553", "127.0.0.1:9563", "127.0.0.1:9573", "127.0.0.1:9583", "127.0.0.1:9593"},
	StorageProvider:        "bnb-sp",
	PieceConfig:            storage.DefaultPieceStoreConfig,
	StoneJobLimit:          64,
}
