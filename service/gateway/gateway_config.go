package gateway

type GatewayConfig struct {
	Address                  string
	Domain                   string
	UploaderServiceAddress   string
	DownloaderServiceAddress string
	ChainConfig              *chainClientConfig
}

var DefaultGatewayConfig = &GatewayConfig{
	Address:                  "127.0.0.1:5310",
	Domain:                   "bfs.nodereal.com",
	UploaderServiceAddress:   "127.0.0.1:5311",
	DownloaderServiceAddress: "127.0.0.1:5523",
	ChainConfig:              defaultChainClientConfig,
}
