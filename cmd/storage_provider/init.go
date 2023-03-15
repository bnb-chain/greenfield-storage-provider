package main

import (
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	"github.com/bnb-chain/greenfield-storage-provider/service/blocksyncer"
	"github.com/bnb-chain/greenfield-storage-provider/service/challenge"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/service/manager"
	"github.com/bnb-chain/greenfield-storage-provider/service/receiver"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer"
	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/config"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/gateway"
	metadata "github.com/bnb-chain/greenfield-storage-provider/service/metadata/service"
	"github.com/bnb-chain/greenfield-storage-provider/service/tasknode"
	"github.com/bnb-chain/greenfield-storage-provider/service/uploader"
)

// initLog init global log level and log path.
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

// initService init service instance by name and config.
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
		sCfg, _ := cfg.MakeSignerConfig()
		server, err = signer.NewSignerServer(sCfg)
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
		server, err = blocksyncer.NewBlockSyncerService(cfg.BlockSyncerCfg)
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
	default:
		log.Errorw("unknown service", "service", serviceName)
		return nil, fmt.Errorf("unknown service: %s", serviceName)
	}
	return server, nil
}
