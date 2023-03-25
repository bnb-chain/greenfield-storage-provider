package main

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	"github.com/bnb-chain/greenfield-storage-provider/config"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/service/blocksyncer"
	"github.com/bnb-chain/greenfield-storage-provider/service/challenge"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/service/gateway"
	"github.com/bnb-chain/greenfield-storage-provider/service/manager"
	metadata "github.com/bnb-chain/greenfield-storage-provider/service/metadata/service"
	"github.com/bnb-chain/greenfield-storage-provider/service/p2p"
	"github.com/bnb-chain/greenfield-storage-provider/service/receiver"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer"
	"github.com/bnb-chain/greenfield-storage-provider/service/tasknode"
	"github.com/bnb-chain/greenfield-storage-provider/service/uploader"
)

// initLog initializes global log level and log path.
func initLog(ctx *cli.Context, cfg *config.StorageProviderConfig) error {
	if cfg.LogCfg == nil {
		cfg.LogCfg = config.DefaultLogConfig
	}
	if ctx.IsSet(utils.LogLevelFlag.Name) {
		cfg.LogCfg.Level = ctx.String(utils.LogLevelFlag.Name)
	}
	if ctx.IsSet(utils.LogPathFlag.Name) {
		cfg.LogCfg.Path = ctx.String(utils.LogPathFlag.Name)
	}
	if ctx.IsSet(utils.LogStdOutputFlag.Name) {
		cfg.LogCfg.Path = ""
	}
	logLevel, err := log.ParseLevel(cfg.LogCfg.Level)
	if err != nil {
		return err
	}
	log.Init(logLevel, cfg.LogCfg.Path)
	return nil
}

// initMetrics initializes global metrics.
func initMetrics(ctx *cli.Context, cfg *config.StorageProviderConfig) error {
	if cfg.MetricsCfg == nil {
		cfg.MetricsCfg = config.DefaultMetricsConfig
	}
	if ctx.IsSet(utils.MetricsEnabledFlag.Name) {
		cfg.MetricsCfg.Enabled = ctx.Bool(utils.MetricsEnabledFlag.Name)
	}
	if ctx.IsSet(utils.MetricsHTTPFlag.Name) {
		cfg.MetricsCfg.HTTPAddress = ctx.String(utils.MetricsHTTPFlag.Name)
	}
	if cfg.MetricsCfg.Enabled {
		slc := lifecycle.NewServiceLifecycle()
		slc.RegisterServices(metrics.NewMetrics(cfg.MetricsCfg))
	}
	return nil
}

// initResourceManager initializes global resource manager.
func initResourceManager(ctx *cli.Context) error {
	if ctx.IsSet(utils.DisableResourceManagerFlag.Name) &&
		ctx.Bool(utils.DisableResourceManagerFlag.Name) {
		return nil
	}
	var (
		limits = rcmgr.DefaultLimitConfig
		err    error
	)
	if ctx.IsSet(utils.ResourceManagerConfigFlag.Name) {
		limits, err = rcmgr.NewLimitConfigFromToml(
			ctx.String(utils.ResourceManagerConfigFlag.Name))
		if err != nil {
			return err
		}
	}
	log.Infow("resource manager", "limits", limits.String())
	if _, err = rcmgr.NewResourceManager(limits); err != nil {
		return err
	}
	return nil
}

// initService initializes service instance by name and config.
func initService(serviceName string, cfg *config.StorageProviderConfig) (server lifecycle.Service, err error) {
	switch serviceName {
	case model.GatewayService:
		gCfg, err := cfg.MakeGatewayConfig()
		if err != nil {
			return nil, err
		}
		server, err = gateway.NewGatewayService(gCfg)
		if err != nil {
			return nil, err
		}
	case model.UploaderService:
		uCfg, err := cfg.MakeUploaderConfig()
		if err != nil {
			return nil, err
		}
		server, err = uploader.NewUploaderService(uCfg)
		if err != nil {
			return nil, err
		}
	case model.DownloaderService:
		dCfg, err := cfg.MakeDownloaderConfig()
		if err != nil {
			return nil, err
		}
		server, err = downloader.NewDownloaderService(dCfg)
		if err != nil {
			return nil, err
		}
	case model.ReceiverService:
		sCfg, err := cfg.MakeReceiverConfig()
		if err != nil {
			return nil, err
		}
		server, err = receiver.NewReceiverService(sCfg)
		if err != nil {
			return nil, err
		}
	case model.ChallengeService:
		cCfg, err := cfg.MakeChallengeConfig()
		if err != nil {
			return nil, err
		}
		server, err = challenge.NewChallengeService(cCfg)
		if err != nil {
			return nil, err
		}
	case model.SignerService:
		sCfg, chainCfg, err := cfg.MakeSignerConfig()
		if err != nil {
			return nil, err
		}
		server, err = signer.NewSignerServer(sCfg, chainCfg)
		if err != nil {
			return nil, err
		}
	case model.TaskNodeService:
		snCfg, err := cfg.MakeTaskNodeConfig()
		if err != nil {
			return nil, err
		}
		server, err = tasknode.NewTaskNodeService(snCfg)
		if err != nil {
			return nil, err
		}
	case model.MetadataService:
		mCfg, err := cfg.MakeMetadataServiceConfig()
		if err != nil {
			return nil, err
		}
		server, err = metadata.NewMetadataService(mCfg)
		if err != nil {
			return nil, err
		}
	case model.BlockSyncerService:
		bsCfg, err := cfg.MakeBlockSyncerConfig()
		if err != nil {
			return nil, err
		}
		server, err = blocksyncer.NewBlockSyncerService(bsCfg)
		if err != nil {
			return nil, err
		}
	case model.ManagerService:
		managerCfg, err := cfg.MakeManagerServiceConfig()
		if err != nil {
			return nil, err
		}
		server, err = manager.NewManagerService(managerCfg)
		if err != nil {
			return nil, err
		}
	case model.P2PService:
		p2pCfg, err := cfg.MakeP2PServiceConfig()
		if err != nil {
			return nil, err
		}
		server, err = p2p.NewP2PServer(p2pCfg)
		if err != nil {
			return nil, err
		}
	default:
		log.Errorw("unknown service", "service", serviceName)
		return nil, fmt.Errorf("unknown service: %s", serviceName)
	}
	return server, nil
}
