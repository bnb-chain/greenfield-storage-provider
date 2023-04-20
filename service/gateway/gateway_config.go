package gateway

import (
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	localhttp "github.com/bnb-chain/greenfield-storage-provider/pkg/middleware/http"
)

// GatewayConfig defines gateway service config
type GatewayConfig struct {
	SpOperatorAddress        string
	HTTPAddress              string
	Domain                   string
	ChainConfig              *gnfd.GreenfieldChainConfig
	UploaderServiceAddress   string
	DownloaderServiceAddress string
	SignerServiceAddress     string
	ChallengeServiceAddress  string
	ReceiverServiceAddress   string
	MetadataServiceAddress   string
	AuthServiceAddress       string
	APILimiterCfg            *localhttp.APILimiterConfig
}
