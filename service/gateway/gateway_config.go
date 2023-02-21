package gateway

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
	StorageProvider:          "bnb-sp",
	Address:                  "127.0.0.1:9033",
	Domain:                   "gnfd.nodereal.com",
	UploaderServiceAddress:   "127.0.0.1:9133",
	DownloaderServiceAddress: "127.0.0.1:9233",
	SyncerServiceAddress:     "127.0.0.1:9533",
	ChallengeServiceAddress:  "127.0.0.1:9633",
	ChainConfig:              DefaultChainClientConfig,
}
