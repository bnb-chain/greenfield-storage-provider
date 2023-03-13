package gateway

import (
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
)

type GatewayConfig struct {
	SpOperatorAddress        string
	HTTPAddress              string
	Domain                   string
	ChainConfig              *gnfd.GreenfieldChainConfig
	UploaderServiceAddress   string
	DownloaderServiceAddress string
	SignerServiceAddress     string
	ChallengeServiceAddress  string
	SyncerServiceAddress     string
	MetadataServiceAddress   string
}
