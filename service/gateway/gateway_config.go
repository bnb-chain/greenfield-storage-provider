package gateway

import (
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
)

type GatewayConfig struct {
	SpOperatorAddress        string
	HTTPAddress              string
	Domain                   string
	UploaderServiceAddress   string
	DownloaderServiceAddress string
	SignerServiceAddress     string
	ChallengeServiceAddress  string
	SyncerServiceAddress     string
	ChainConfig              *gnfd.GreenfieldChainConfig
}
