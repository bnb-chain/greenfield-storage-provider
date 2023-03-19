package config

import (
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	tomlconfig "github.com/forbole/juno/v4/cmd/migrate/toml"
	databaseconfig "github.com/forbole/juno/v4/database/config"
	loggingconfig "github.com/forbole/juno/v4/log/config"
	"github.com/forbole/juno/v4/node/remote"
	parserconfig "github.com/forbole/juno/v4/parser/config"
	"github.com/forbole/juno/v4/types/config"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/service/challenge"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/service/gateway"
	"github.com/bnb-chain/greenfield-storage-provider/service/manager"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata"
	"github.com/bnb-chain/greenfield-storage-provider/service/receiver"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer"
	"github.com/bnb-chain/greenfield-storage-provider/service/tasknode"
	"github.com/bnb-chain/greenfield-storage-provider/service/uploader"
)

// MakeGatewayConfig make gateway service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeGatewayConfig() (*gateway.GatewayConfig, error) {
	gCfg := &gateway.GatewayConfig{
		SpOperatorAddress: cfg.SpOperatorAddress,
		ChainConfig:       cfg.ChainConfig,
	}
	if _, ok := cfg.ListenAddress[model.GatewayService]; ok {
		gCfg.HTTPAddress = cfg.ListenAddress[model.GatewayService]
	} else {
		return nil, fmt.Errorf("missing gateway HTTP address configuration for gateway service")
	}
	if _, ok := cfg.Endpoint[model.GatewayService]; ok {
		gCfg.Domain = cfg.Endpoint[model.GatewayService]
	} else {
		return nil, fmt.Errorf("missing gateway endpoint configuration for gateway service")
	}
	if _, ok := cfg.Endpoint[model.UploaderService]; ok {
		gCfg.UploaderServiceAddress = cfg.Endpoint[model.UploaderService]
	} else {
		return nil, fmt.Errorf("missing uploader gRPC address configuration for gateway service")
	}
	if _, ok := cfg.Endpoint[model.DownloaderService]; ok {
		gCfg.DownloaderServiceAddress = cfg.Endpoint[model.DownloaderService]
	} else {
		return nil, fmt.Errorf("missing downloader gRPC address configuration for gateway service")
	}
	if _, ok := cfg.Endpoint[model.SignerService]; ok {
		gCfg.SignerServiceAddress = cfg.Endpoint[model.SignerService]
	} else {
		return nil, fmt.Errorf("missing signer gRPC address configuration for gateway service")
	}
	if _, ok := cfg.Endpoint[model.ChallengeService]; ok {
		gCfg.ChallengeServiceAddress = cfg.Endpoint[model.ChallengeService]
	} else {
		return nil, fmt.Errorf("missing challenge gRPC address configuration for gateway service")
	}
	if _, ok := cfg.Endpoint[model.ReceiverService]; ok {
		gCfg.ReceiverServiceAddress = cfg.Endpoint[model.ReceiverService]
	} else {
		return nil, fmt.Errorf("missing receiver gRPC address configuration for gateway service")
	}
	if _, ok := cfg.Endpoint[model.MetadataService]; ok {
		gCfg.MetadataServiceAddress = cfg.Endpoint[model.MetadataService]
	} else {
		return nil, fmt.Errorf("missing metadata gPRC address configuration for gateway service")
	}
	return gCfg, nil
}

// MakeUploaderConfig make uploader service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeUploaderConfig() (*uploader.UploaderConfig, error) {
	uCfg := &uploader.UploaderConfig{
		SpDBConfig:       cfg.SpDBConfig,
		PieceStoreConfig: cfg.PieceStoreConfig,
	}
	if _, ok := cfg.ListenAddress[model.UploaderService]; ok {
		uCfg.GRPCAddress = cfg.ListenAddress[model.UploaderService]
	} else {
		return nil, fmt.Errorf("missing uploader gRPC address configuration for uploader service")
	}
	if _, ok := cfg.Endpoint[model.SignerService]; ok {
		uCfg.SignerGrpcAddress = cfg.Endpoint[model.SignerService]
	} else {
		return nil, fmt.Errorf("missing signer gRPC address configuration for uploader service")
	}
	if _, ok := cfg.Endpoint[model.TaskNodeService]; ok {
		uCfg.TaskNodeGrpcAddress = cfg.Endpoint[model.TaskNodeService]
	} else {
		return nil, fmt.Errorf("missing task node gRPC address configuration for uploader service")
	}
	return uCfg, nil
}

// MakeDownloaderConfig make downloader service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeDownloaderConfig() (*downloader.DownloaderConfig, error) {
	dCfg := &downloader.DownloaderConfig{
		SpDBConfig:       cfg.SpDBConfig,
		PieceStoreConfig: cfg.PieceStoreConfig,
	}
	if _, ok := cfg.ListenAddress[model.DownloaderService]; ok {
		dCfg.GRPCAddress = cfg.ListenAddress[model.DownloaderService]
	} else {
		return nil, fmt.Errorf("missing downloader gRPC address configuration for downloader service")
	}
	return dCfg, nil
}

// MakeReceiverConfig make receiver service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeReceiverConfig() (*receiver.ReceiverConfig, error) {
	sCfg := &receiver.ReceiverConfig{
		SpOperatorAddress: cfg.SpOperatorAddress,
		SpDBConfig:        cfg.SpDBConfig,
		PieceStoreConfig:  cfg.PieceStoreConfig,
	}
	if _, ok := cfg.ListenAddress[model.ReceiverService]; ok {
		sCfg.GRPCAddress = cfg.ListenAddress[model.ReceiverService]
	} else {
		return nil, fmt.Errorf("missing receiver gRPC address configuration for receiver service")
	}
	if _, ok := cfg.Endpoint[model.SignerService]; ok {
		sCfg.SignerGRPCAddress = cfg.Endpoint[model.SignerService]
	} else {
		return nil, fmt.Errorf("missing signer gRPC address configuration for receiver service")
	}
	return sCfg, nil
}

// MakeChallengeConfig make challenge service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeChallengeConfig() (*challenge.ChallengeConfig, error) {
	cCfg := &challenge.ChallengeConfig{
		SpDBConfig:       cfg.SpDBConfig,
		PieceStoreConfig: cfg.PieceStoreConfig,
	}
	if _, ok := cfg.ListenAddress[model.ChallengeService]; ok {
		cCfg.GRPCAddress = cfg.ListenAddress[model.ChallengeService]
	} else {
		return nil, fmt.Errorf("missing challenge gRPC address configuration for challenge service")
	}
	return cCfg, nil
}

// MakeSignerConfig make singer service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeSignerConfig() (*signer.SignerConfig, *gnfd.GreenfieldChainConfig, error) {
	sCfg := cfg.SignerCfg
	signerAddr, ok := cfg.ListenAddress[model.SignerService]
	if !ok {
		return nil, nil, fmt.Errorf("missing signer gRPC address configuration for signer service")
	}
	sCfg.GRPCAddress = signerAddr
	return cfg.SignerCfg, cfg.ChainConfig, nil
}

// MakeTaskNodeConfig make task node service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeTaskNodeConfig() (*tasknode.TaskNodeConfig, error) {
	snCfg := &tasknode.TaskNodeConfig{
		SpOperatorAddress: cfg.SpOperatorAddress,
		SpDBConfig:        cfg.SpDBConfig,
		PieceStoreConfig:  cfg.PieceStoreConfig,
		ChainConfig:       cfg.ChainConfig,
	}
	if _, ok := cfg.ListenAddress[model.TaskNodeService]; ok {
		snCfg.GRPCAddress = cfg.ListenAddress[model.TaskNodeService]
	} else {
		return nil, fmt.Errorf("missing task node gRPC address configuration for task node service")
	}
	if _, ok := cfg.Endpoint[model.SignerService]; ok {
		snCfg.SignerGrpcAddress = cfg.Endpoint[model.SignerService]
	} else {
		return nil, fmt.Errorf("missing signer gRPC address configuration for task node service")
	}
	return snCfg, nil
}

// MakeMetadataServiceConfig make meta data service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeMetadataServiceConfig() (*metadata.MetadataConfig, error) {
	mCfg := &metadata.MetadataConfig{
		SpDBConfig: cfg.SpDBConfig,
	}
	if _, ok := cfg.ListenAddress[model.MetadataService]; ok {
		mCfg.GRPCAddress = cfg.ListenAddress[model.MetadataService]
	} else {
		return nil, fmt.Errorf("missing meta data gRPC address configuration for meta data service")
	}
	return mCfg, nil
}

// MakeManagerServiceConfig make manager service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeManagerServiceConfig() (*manager.ManagerConfig, error) {
	managerConfig := &manager.ManagerConfig{
		SpOperatorAddress: cfg.SpOperatorAddress,
		ChainConfig:       cfg.ChainConfig,
		SpDBConfig:        cfg.SpDBConfig,
	}
	return managerConfig, nil
}

// MakeBlockSyncerConfig make block syncer service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeBlockSyncerConfig() (*tomlconfig.TomlConfig, error) {
	rpcAddress := cfg.ChainConfig.NodeAddr[0].TendermintAddresses[0]
	grpcAddress := cfg.ChainConfig.NodeAddr[0].GreenfieldAddresses[0]

	return &tomlconfig.TomlConfig{
		Chain: config.ChainConfig{
			Bech32Prefix: "cosmos",
			Modules:      cfg.BlockSyncerCfg.Modules,
		},
		Node: tomlconfig.NodeConfig{
			Type: "remote",
			RPC: &remote.RPCConfig{
				ClientName: "juno",
				Address:    rpcAddress,
			},
			GRPC: &remote.GRPCConfig{
				Address:  grpcAddress,
				Insecure: true,
			},
		},
		Parser: parserconfig.Config{
			Workers: 1,
		},
		Database: databaseconfig.Config{
			Type:               "mysql",
			DSN:                cfg.BlockSyncerCfg.Dsn,
			PartitionBatchSize: model.DefaultPartitionSize,
			MaxIdleConnections: 1,
			MaxOpenConnections: 1,
		},
		Logging: loggingconfig.Config{
			Level: "debug",
		},
	}, nil
}

// MakeMetricsMonitorConfig make metrics monitor config from StorageProviderConfig
func (cfg StorageProviderConfig) MakeMetricsMonitorConfig() (*metrics.MetricsMonitorConfig, error) {
	metricsConfig := &metrics.MetricsMonitorConfig{
		Enabled: true,
	}
	if _, ok := cfg.ListenAddress[model.MetricsMonitorService]; ok {
		metricsConfig.HTTPAddress = cfg.ListenAddress[model.MetadataService]
	} else {
		return nil, fmt.Errorf("missing meta data HTTP address configuration for metrics monitor service")
	}
	return metricsConfig, nil
}
