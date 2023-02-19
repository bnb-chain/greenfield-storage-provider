package gateway

type GatewayConfig struct {
	Address                  string
	Domain                   string
	UploaderServiceAddress   string
	DownloaderServiceAddress string
	ChallengeServiceAddress  string
	ChainConfig              *chainClientConfig
}

var DefaultGatewayConfig = &GatewayConfig{
	Address:                  "127.0.0.1:9033",
	Domain:                   "gnfd.nodereal.com",
	UploaderServiceAddress:   "127.0.0.1:9133",
	DownloaderServiceAddress: "127.0.0.1:9233",
	ChallengeServiceAddress:  "127.0.0.1:9633",
	ChainConfig:              defaultChainClientConfig,
}
