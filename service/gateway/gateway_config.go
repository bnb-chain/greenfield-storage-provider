package gateway

import gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"

type GatewayConfig struct {
	StorageProvider          string
	Address                  string
	Domain                   string
	UploaderServiceAddress   string
	DownloaderServiceAddress string
	SignerServiceAddress     string
	ChallengeServiceAddress  string
	SyncerServiceAddress     string
	ChainConfig              *gnfd.GreenfieldChainConfig
	MetadataServiceAddress   string
}

var DefaultGatewayConfig = &GatewayConfig{
	StorageProvider:          "bnb-sp",
	Address:                  "127.0.0.1:9033",
	Domain:                   "gnfd.nodereal.com",
	UploaderServiceAddress:   "127.0.0.1:9133",
	DownloaderServiceAddress: "127.0.0.1:9233",
	SyncerServiceAddress:     "127.0.0.1:9533",
	SignerServiceAddress:     "127.0.0.1:9633",
	ChallengeServiceAddress:  "127.0.0.1:9733",
	MetadataServiceAddress:   "127.0.0.1:9833",
	ChainConfig:              gnfd.DefaultGreenfieldChainConfig,
}
