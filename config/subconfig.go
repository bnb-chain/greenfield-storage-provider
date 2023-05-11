package config

import (
	"fmt"

	tomlconfig "github.com/forbole/juno/v4/cmd/migrate/toml"
	databaseconfig "github.com/forbole/juno/v4/database/config"
	loggingconfig "github.com/forbole/juno/v4/log/config"
	"github.com/forbole/juno/v4/node/remote"
	parserconfig "github.com/forbole/juno/v4/parser/config"
	"github.com/forbole/juno/v4/types/config"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	localhttp "github.com/bnb-chain/greenfield-storage-provider/pkg/middleware/http"
	"github.com/bnb-chain/greenfield-storage-provider/service/auth"
	"github.com/bnb-chain/greenfield-storage-provider/service/challenge"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/service/gateway"
	"github.com/bnb-chain/greenfield-storage-provider/service/manager"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata"
	"github.com/bnb-chain/greenfield-storage-provider/service/p2p"
	"github.com/bnb-chain/greenfield-storage-provider/service/receiver"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer"
	"github.com/bnb-chain/greenfield-storage-provider/service/stopserving"
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
	if _, ok := cfg.Endpoint[model.AuthService]; ok {
		gCfg.AuthServiceAddress = cfg.Endpoint[model.AuthService]
	} else {
		return nil, fmt.Errorf("missing auth gPRC address configuration for gateway service")
	}
	if cfg.RateLimiter != nil {
		defaultMap := make(map[string]localhttp.MemoryLimiterConfig)
		for _, c := range cfg.RateLimiter.PathPattern {
			defaultMap[c.Key] = localhttp.MemoryLimiterConfig{
				RateLimit:  c.RateLimit,
				RatePeriod: c.RatePeriod,
			}
		}
		patternMap := make(map[string]localhttp.MemoryLimiterConfig)
		for _, c := range cfg.RateLimiter.HostPattern {
			patternMap[c.Key] = localhttp.MemoryLimiterConfig{
				RateLimit:  c.RateLimit,
				RatePeriod: c.RatePeriod,
			}
		}
		apiLimitsMap := make(map[string]localhttp.MemoryLimiterConfig)
		for _, c := range cfg.RateLimiter.APILimits {
			apiLimitsMap[c.Key] = localhttp.MemoryLimiterConfig{
				RateLimit:  c.RateLimit,
				RatePeriod: c.RatePeriod,
			}
		}
		gCfg.APILimiterCfg = &localhttp.APILimiterConfig{
			PathPattern:  defaultMap,
			HostPattern:  patternMap,
			APILimits:    apiLimitsMap,
			HTTPLimitCfg: cfg.RateLimiter.HTTPLimitCfg,
		}
		gCfg.BandwidthLimitCfg = &localhttp.BandwidthLimiterConfig{
			Enable: cfg.BandwidthLimiter.Enable,
			R:      10,
			B:      1000,
		}
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
	if _, ok := cfg.Endpoint[model.P2PService]; ok {
		snCfg.P2PGrpcAddress = cfg.Endpoint[model.P2PService]
	} else {
		return nil, fmt.Errorf("missing p2p server gRPC address configuration for task node service")
	}
	return snCfg, nil
}

// MakeMetadataServiceConfig make meta data service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeMetadataServiceConfig() (*metadata.MetadataConfig, error) {
	mCfg := &metadata.MetadataConfig{
		BsDBConfig:                 cfg.BsDBConfig,
		BsDBSwitchedConfig:         cfg.BsDBSwitchedConfig,
		BsDBSwitchCheckIntervalSec: cfg.MetadataCfg.BsDBSwitchCheckIntervalSec,
		IsMasterDB:                 cfg.MetadataCfg.IsMasterDB,
	}
	if _, ok := cfg.ListenAddress[model.MetadataService]; ok {
		mCfg.GRPCAddress = cfg.ListenAddress[model.MetadataService]
	} else {
		return nil, fmt.Errorf("missing metadata gRPC address configuration for meta data service")
	}
	return mCfg, nil
}

// MakeManagerServiceConfig make manager service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeManagerServiceConfig() (*manager.ManagerConfig, error) {
	managerConfig := &manager.ManagerConfig{
		SpOperatorAddress: cfg.SpOperatorAddress,
		ChainConfig:       cfg.ChainConfig,
		SpDBConfig:        cfg.SpDBConfig,
		PieceStoreConfig:  cfg.PieceStoreConfig,
	}
	if _, ok := cfg.Endpoint[model.MetadataService]; ok {
		managerConfig.MetadataGrpcAddress = cfg.Endpoint[model.MetadataService]
	} else {
		return nil, fmt.Errorf("missing metadata server gRPC address configuration for manager service")
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
			Workers: int64(cfg.BlockSyncerCfg.Workers),
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
		RecreateTables: cfg.BlockSyncerCfg.RecreateTables,
	}, nil
}

// MakeP2PServiceConfig make p2p service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeP2PServiceConfig() (*p2p.P2PConfig, error) {
	pCfg := &p2p.P2PConfig{
		SpOperatorAddress: cfg.SpOperatorAddress,
		SpDBConfig:        cfg.SpDBConfig,
		P2PConfig:         cfg.P2PCfg,
	}
	if _, ok := cfg.ListenAddress[model.P2PService]; ok {
		pCfg.GRPCAddress = cfg.ListenAddress[model.P2PService]
	} else {
		return nil, fmt.Errorf("missing p2p service gRPC address configuration for p2p service")
	}
	if _, ok := cfg.Endpoint[model.SignerService]; ok {
		pCfg.SignerGrpcAddress = cfg.Endpoint[model.SignerService]
	} else {
		return nil, fmt.Errorf("missing signer gRPC address configuration for p2p service")
	}
	return pCfg, nil
}

// MakeAuthServiceConfig make auth service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeAuthServiceConfig() (*auth.AuthConfig, error) {
	aCfg := &auth.AuthConfig{
		SpDBConfig:        cfg.SpDBConfig,
		SpOperatorAddress: cfg.SpOperatorAddress,
	}
	if _, ok := cfg.ListenAddress[model.AuthService]; ok {
		aCfg.GRPCAddress = cfg.ListenAddress[model.AuthService]
	} else {
		return nil, fmt.Errorf("missing auth gRPC address configuration for auth service")
	}
	return aCfg, nil
}

// MakeStopServingServiceConfig make stop serving service config from StorageProviderConfig
func (cfg *StorageProviderConfig) MakeStopServingServiceConfig() (*stopserving.StopServingConfig, error) {
	ssCfg := &stopserving.StopServingConfig{
		SpOperatorAddress: cfg.SpOperatorAddress,
		DiscontinueConfig: cfg.DiscontinueCfg,
	}
	if _, ok := cfg.ListenAddress[model.SignerService]; ok {
		ssCfg.SignerGrpcAddress = cfg.ListenAddress[model.SignerService]
	} else {
		return nil, fmt.Errorf("missing signer gRPC address configuration for stop serving service")
	}
	if _, ok := cfg.Endpoint[model.MetadataService]; ok {
		ssCfg.MetadataGrpcAddress = cfg.Endpoint[model.MetadataService]
	} else {
		return nil, fmt.Errorf("missing metadata gRPC address configuration for stop serving service")
	}
	return ssCfg, nil
}
