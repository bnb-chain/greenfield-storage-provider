package gateway

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
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

var DefaultGatewayConfig = &GatewayConfig{
	SpOperatorAddress:        model.SpOperatorAddress,
	HTTPAddress:              model.GatewayHTTPAddress,
	Domain:                   "gnfd.nodereal.com",
	UploaderServiceAddress:   model.UploaderGRPCAddress,
	DownloaderServiceAddress: model.DownloaderGRPCAddress,
	SyncerServiceAddress:     model.SyncerGRPCAddress,
	SignerServiceAddress:     model.SignerGRPCAddress,
	ChallengeServiceAddress:  model.ChallengeGRPCAddress,
	ChainConfig:              gnfd.DefaultGreenfieldChainConfig,
}
