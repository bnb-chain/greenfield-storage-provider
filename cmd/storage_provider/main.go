package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/bnb-chain/greenfield-storage-provider/service/blocksyncer"

	"github.com/bnb-chain/greenfield-storage-provider/config"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/service/challenge"
	"github.com/bnb-chain/greenfield-storage-provider/service/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/service/gateway"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer"
	"github.com/bnb-chain/greenfield-storage-provider/service/stonehub"
	"github.com/bnb-chain/greenfield-storage-provider/service/stonenode"
	"github.com/bnb-chain/greenfield-storage-provider/service/syncer"
	"github.com/bnb-chain/greenfield-storage-provider/service/uploader"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

var (
	version    = flag.Bool("version", false, "print version")
	configFile = flag.String("config", "../../config/config.toml", "config file path")
)

// initService init service instance by name and config.
func initService(serviceName string, cfg *config.StorageProviderConfig) (server lifecycle.Service, err error) {
	switch serviceName {
	case model.GatewayService:
		if cfg.GatewayCfg == nil {
			cfg.GatewayCfg = config.DefaultStorageProviderConfig.GatewayCfg
		}
		server, err = gateway.NewGatewayService(cfg.GatewayCfg)
		if err != nil {
			return nil, err
		}
	case model.UploaderService:
		if cfg.UploaderCfg == nil {
			cfg.UploaderCfg = config.DefaultStorageProviderConfig.UploaderCfg
		}
		server, err = uploader.NewUploaderService(cfg.UploaderCfg)
		if err != nil {
			return nil, err
		}
	case model.DownloaderService:
		if cfg.DownloaderCfg == nil {
			cfg.DownloaderCfg = config.DefaultStorageProviderConfig.DownloaderCfg
		}
		server, err = downloader.NewDownloaderService(cfg.DownloaderCfg)
		if err != nil {
			return nil, err
		}
	case model.StoneHubService:
		if cfg.StoneHubCfg == nil {
			cfg.StoneHubCfg = config.DefaultStorageProviderConfig.StoneHubCfg
		}
		server, err = stonehub.NewStoneHubService(cfg.StoneHubCfg)
		if err != nil {
			return nil, err
		}
	case model.StoneNodeService:
		if cfg.StoneNodeCfg == nil {
			cfg.StoneNodeCfg = config.DefaultStorageProviderConfig.StoneNodeCfg
		}
		server, err = stonenode.NewStoneNodeService(cfg.StoneNodeCfg)
		if err != nil {
			return nil, err
		}
	case model.SyncerService:
		if cfg.SyncerCfg == nil {
			cfg.SyncerCfg = config.DefaultStorageProviderConfig.SyncerCfg
		}
		server, err = syncer.NewSyncerService(cfg.SyncerCfg)
		if err != nil {
			return nil, err
		}
	case model.ChallengeService:
		if cfg.ChallengeCfg == nil {
			cfg.ChallengeCfg = config.DefaultStorageProviderConfig.ChallengeCfg
		}
		server, err = challenge.NewChallengeService(cfg.ChallengeCfg)
		if err != nil {
			return nil, err
		}
	case model.SignerService:
		if cfg.SignerCfg == nil {
			cfg.SignerCfg = config.DefaultStorageProviderConfig.SignerCfg
		}
		server, err = signer.NewSignerServer(cfg.SignerCfg)
		if err != nil {
			return nil, err
		}
	case model.BlockSyncerService:
		server, err = blocksyncer.NewBlockSyncerService(cfg.BlockSyncerCfg)
		if err != nil {
			return nil, err
		}
	default:
		log.Errorw("unknown service", "service", serviceName)
		return nil, fmt.Errorf("unknow service: %s", serviceName)
	}
	return server, nil
}

func main() {
	flag.Parse()
	if *version {
		fmt.Print(DumpLogo() + DumpVersion())
		os.Exit(0)
	}

	cfg := config.LoadConfig(*configFile)
	slc := lifecycle.NewServiceLifecycle()
	for _, serviceName := range cfg.Service {
		// 1. init service instance.
		service, err := initService(serviceName, cfg)
		if err != nil {
			log.Errorw("init service failed", "service_name", serviceName, "error", err)
			os.Exit(1)
		}
		log.Debugw("init service success", "service_name", serviceName)
		// 2. register service to lifecycle.
		slc.RegisterServices(service)
	}
	// 3. start all services and listen os signals.
	ctx := context.Background()
	slc.Signals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP).StartServices(ctx).Wait(ctx)
}
