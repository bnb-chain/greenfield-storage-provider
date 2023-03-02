package gateway

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
)

type GatewayConfig struct {
	SPOperatorAddress        string
	Address                  string
	Domain                   string
	UploaderServiceAddress   string
	DownloaderServiceAddress string
	SignerServiceAddress     string
	ChallengeServiceAddress  string
	SyncerServiceAddress     string
	ChainConfig              *gnfd.GreenfieldChainConfig
}

var DefaultGatewayConfig = &GatewayConfig{
	SPOperatorAddress:        model.SPOperatorAddress,
	Address:                  model.GatewayHTTPAddress,
	Domain:                   "gnfd.nodereal.com",
	UploaderServiceAddress:   model.UploaderGRPCAddress,
	DownloaderServiceAddress: model.DownloaderGRPCAddress,
	SyncerServiceAddress:     model.SyncerGRPCAddress,
	SignerServiceAddress:     model.SignerGRPCAddress,
	ChallengeServiceAddress:  model.ChallengeGRPCAddress,
	ChainConfig:              gnfd.DefaultGreenfieldChainConfig,
}
