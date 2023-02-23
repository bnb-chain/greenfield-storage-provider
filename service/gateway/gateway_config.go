package gateway

import "github.com/bnb-chain/greenfield-storage-provider/model"

type GatewayConfig struct {
	StorageProvider          string
	Address                  string
	Domain                   string
	UploaderServiceAddress   string
	DownloaderServiceAddress string
	ChallengeServiceAddress  string
	SyncerServiceAddress     string
	ChainConfig              *ChainClientConfig
}

var DefaultGatewayConfig = &GatewayConfig{
	StorageProvider:          model.StorageProvider,
	Address:                  model.DefaultGateAddress,
	Domain:                   "gnfd.nodereal.com",
	UploaderServiceAddress:   model.DefaultUploaderAddress,
	DownloaderServiceAddress: model.DefaultDownloaderAddress,
	SyncerServiceAddress:     model.DefaultSyncerAddress,
	ChallengeServiceAddress:  model.DefaultChallengeAddress,
	ChainConfig:              DefaultChainClientConfig,
}

func overrideConfigFromEnv(config *GatewayConfig) {
	config.StorageProvider = model.StorageProvider
}
