package gateway

import gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"

type GatewayConfig struct {
	SpId                     string
	Address                  string
	Domain                   string
	UploaderServiceAddress   string
	DownloaderServiceAddress string
	ChainConfig              *gnfd.GreenfieldChainConfig
}

var DefaultGatewayConfig = &GatewayConfig{
	Address:                  "127.0.0.1:9033",
	Domain:                   "bfs.nodereal.com",
	UploaderServiceAddress:   "127.0.0.1:9133",
	DownloaderServiceAddress: "127.0.0.1:9233",
	ChainConfig:              gnfd.DefaultGreenfieldChainConfig,
}
