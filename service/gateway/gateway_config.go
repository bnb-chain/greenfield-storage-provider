package gateway

import (
	"github.com/bnb-chain/greenfield-storage-provider/model"
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
)

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
}

var DefaultGatewayConfig = &GatewayConfig{
	StorageProvider:          model.SpOperatorAddress,
	Address:                  model.GatewayHttpAddress,
	Domain:                   "gnfd.nodereal.com",
	UploaderServiceAddress:   model.UploaderService,
	DownloaderServiceAddress: model.DownloaderGrpcAddress,
	SyncerServiceAddress:     model.SyncerGrpcAddress,
	SignerServiceAddress:     model.SignerGrpcAddress,
	ChallengeServiceAddress:  model.ChallengeGrpcAddress,
	ChainConfig:              gnfd.DefaultGreenfieldChainConfig,
}
