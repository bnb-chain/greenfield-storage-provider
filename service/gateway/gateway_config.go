package gateway

type GatewayConfig struct {
	Address          string
	Domain           string
	UploaderConfig   *uploadProcessorConfig
	ChainConfig      *chainClientConfig
	DownloaderConfig *downloadProcessorConfig
}

var DefaultGatewayConfig = &GatewayConfig{
	Address:          "127.0.0.1:5310",
	Domain:           "bfs.nodereal.com",
	UploaderConfig:   defaultUploadProcessorConfig,
	ChainConfig:      defaultChainClientConfig,
	DownloaderConfig: defaultDownloadProcessorConfig,
}
