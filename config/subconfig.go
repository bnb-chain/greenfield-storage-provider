package config

import (
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/service/challenge"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/service/gateway"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer"
	"github.com/bnb-chain/greenfield-storage-provider/service/stonenode"
	"github.com/bnb-chain/greenfield-storage-provider/service/syncer"
	"github.com/bnb-chain/greenfield-storage-provider/service/uploader"
)

// MakeGatewayConfig make gateway service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeGatewayConfig() (*gateway.GatewayConfig, error) {
	gCfg := &gateway.GatewayConfig{
		SpOperatorAddress: cfg.SpOperatorAddress,
		Domain:            cfg.Domain,
		ChainConfig:       cfg.ChainConfig,
	}
	if _, ok := cfg.HTTPAddress[model.GatewayHTTPAddress]; ok {
		gCfg.HTTPAddress = cfg.HTTPAddress[model.GatewayHTTPAddress]
	} else {
		return nil, fmt.Errorf("missing gateway HTTP address configuration")
	}
	if _, ok := cfg.GRPCAddress[model.UploaderGRPCAddress]; ok {
		gCfg.UploaderServiceAddress = cfg.GRPCAddress[model.UploaderGRPCAddress]
	} else {
		return nil, fmt.Errorf("missing uploader gPRC address configuration for gateway service")
	}
	if _, ok := cfg.GRPCAddress[model.DownloaderGRPCAddress]; ok {
		gCfg.DownloaderServiceAddress = cfg.GRPCAddress[model.DownloaderGRPCAddress]
	} else {
		return nil, fmt.Errorf("missing downloader gPRC address configuration for gateway service")
	}
	if _, ok := cfg.GRPCAddress[model.SignerGRPCAddress]; ok {
		gCfg.SignerServiceAddress = cfg.GRPCAddress[model.SignerGRPCAddress]
	} else {
		return nil, fmt.Errorf("missing signer gPRC address configuration for gateway service")
	}
	if _, ok := cfg.GRPCAddress[model.ChallengeGRPCAddress]; ok {
		gCfg.ChallengeServiceAddress = cfg.GRPCAddress[model.ChallengeGRPCAddress]
	} else {
		return nil, fmt.Errorf("missing challenge gPRC address configuration for gateway service")
	}
	if _, ok := cfg.GRPCAddress[model.SyncerGRPCAddress]; ok {
		gCfg.SyncerServiceAddress = cfg.GRPCAddress[model.SyncerGRPCAddress]
	} else {
		return nil, fmt.Errorf("missing syncer gPRC address configuration for gateway service")
	}
	return gCfg, nil
}

// MakeUploaderConfig make uploader service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeUploaderConfig() (*uploader.UploaderConfig, error) {
	uCfg := &uploader.UploaderConfig{
		SpDBConfig:       cfg.SpDBConfig,
		PieceStoreConfig: cfg.PieceStoreConfig,
	}
	if _, ok := cfg.GRPCAddress[model.UploaderService]; ok {
		uCfg.GRPCAddress = cfg.GRPCAddress[model.UploaderService]
	} else {
		return nil, fmt.Errorf("missing uploader gPRC address configuration for uploader service")
	}
	if _, ok := cfg.GRPCAddress[model.SignerService]; ok {
		uCfg.SignerGrpcAddress = cfg.GRPCAddress[model.SignerService]
	} else {
		return nil, fmt.Errorf("missing signer gPRC address configuration for uploader service")
	}
	if _, ok := cfg.GRPCAddress[model.StoneNodeGRPCAddress]; ok {
		uCfg.StoneNodeGrpcAddress = cfg.GRPCAddress[model.StoneNodeGRPCAddress]
	} else {
		return nil, fmt.Errorf("missing stone node gPRC address configuration for uploader service")
	}
	return uCfg, nil
}

// MakeDownloaderConfig make downloader service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeDownloaderConfig() (*downloader.DownloaderConfig, error) {
	dCfg := &downloader.DownloaderConfig{
		SpDBConfig:       cfg.SpDBConfig,
		ChainConfig:      cfg.ChainConfig,
		PieceStoreConfig: cfg.PieceStoreConfig,
	}
	if _, ok := cfg.GRPCAddress[model.DownloaderService]; ok {
		dCfg.GRPCAddress = cfg.GRPCAddress[model.DownloaderService]
	} else {
		return nil, fmt.Errorf("missing downloader gPRC address configuration for downloader service")
	}
	return dCfg, nil
}

// MakeSyncerConfig make syncer service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeSyncerConfig() (*syncer.SyncerConfig, error) {
	sCfg := &syncer.SyncerConfig{
		SpOperatorAddress: cfg.SpOperatorAddress,
		SpDBConfig:        cfg.SpDBConfig,
		PieceStoreConfig:  cfg.PieceStoreConfig,
	}
	if _, ok := cfg.GRPCAddress[model.SyncerService]; ok {
		sCfg.GRPCAddress = cfg.GRPCAddress[model.SyncerService]
	} else {
		return nil, fmt.Errorf("missing syncer gPRC address configuration for syncer service")
	}
	if _, ok := cfg.GRPCAddress[model.SignerGRPCAddress]; ok {
		sCfg.SignerGRPCAddress = cfg.GRPCAddress[model.SignerGRPCAddress]
	} else {
		return nil, fmt.Errorf("missing signer gPRC address configuration for syncer service")
	}
	return sCfg, nil
}

// MakeChallengeConfig make challenge service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeChallengeConfig() (*challenge.ChallengeConfig, error) {
	cCfg := &challenge.ChallengeConfig{
		SpDBConfig:       cfg.SpDBConfig,
		PieceStoreConfig: cfg.PieceStoreConfig,
	}
	if _, ok := cfg.GRPCAddress[model.ChallengeService]; ok {
		cCfg.GRPCAddress = cfg.GRPCAddress[model.ChallengeService]
	} else {
		return nil, fmt.Errorf("missing challenge gPRC address configuration for challenge service")
	}
	return cCfg, nil
}

// MakeSignerConfig make singer service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeSignerConfig() (*signer.SignerConfig, error) {
	return cfg.SignerCfg, nil
}

// MakeStoneNodeConfig make stone node service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeStoneNodeConfig() (*stonenode.StoneNodeConfig, error) {
	snCfg := &stonenode.StoneNodeConfig{
		SpOperatorAddress: cfg.SpOperatorAddress,
		SpDBConfig:        cfg.SpDBConfig,
		PieceStoreConfig:  cfg.PieceStoreConfig,
		ChainConfig:       cfg.ChainConfig,
	}
	if _, ok := cfg.GRPCAddress[model.StoneNodeService]; ok {
		snCfg.GRPCAddress = cfg.GRPCAddress[model.StoneNodeService]
	} else {
		return nil, fmt.Errorf("missing stone node gPRC address configuration for stone node service")
	}
	if _, ok := cfg.GRPCAddress[model.SignerGRPCAddress]; ok {
		snCfg.SignerGrpcAddress = cfg.GRPCAddress[model.SignerGRPCAddress]
	} else {
		return nil, fmt.Errorf("missing signer gPRC address configuration for stone node service")
	}
	return snCfg, nil
}

// MakeMetadataServiceConfig make meta data service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeMetadataServiceConfig() (*metadata.MetadataConfig, error) {
	mCfg := &metadata.MetadataConfig{
		SpDBConfig: cfg.SpDBConfig,
	}
	if _, ok := cfg.HTTPAddress[model.MetadataService]; ok {
		mCfg.Address = cfg.HTTPAddress[model.MetadataService]
	} else {
		return nil, fmt.Errorf("missing meta data HTTP address configuration for mate data service")
	}
	return mCfg, nil
}
