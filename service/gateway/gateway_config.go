package gateway

type GatewayConfig struct {
	Address                  string
	Domain                   string
	UploaderServiceAddress   string
	DownloaderServiceAddress string
	ChainConfig              *chainClientConfig
}

var DefaultGatewayConfig = &GatewayConfig{
	Address:                  "127.0.0.1:9033",
	Domain:                   "bfs.nodereal.com",
	UploaderServiceAddress:   "127.0.0.1:9133",
	DownloaderServiceAddress: "127.0.0.1:9233",
	ChainConfig:              defaultChainClientConfig,
}
