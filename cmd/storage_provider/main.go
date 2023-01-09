package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/bnb-chain/inscription-storage-provider/config"
	"github.com/bnb-chain/inscription-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/inscription-storage-provider/service/gateway"
	"github.com/bnb-chain/inscription-storage-provider/service/stonehub"
	"github.com/bnb-chain/inscription-storage-provider/service/stonenode"
	"github.com/bnb-chain/inscription-storage-provider/service/syncer"
	"github.com/bnb-chain/inscription-storage-provider/service/uploader"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

var (
	version    = flag.Bool("version", false, "print version")
	configFile = flag.String("config", "../../config/config.toml", "config file path")
)

// define the storage provider supports service names
var (
	StoneHubService  string = "StoneHub"
	GetaWayService   string = "Gateway"
	UploaderService  string = "Uploader"
	StoneNodeService string = "StoneNode"
	SyncerService    string = "Syncer"
)

// initService init service instance by name and config.
func initService(serviceName string, cfg *config.StorageProviderConfig) (server lifecycle.Service, err error) {
	switch serviceName {
	case GetaWayService:
		if cfg.GatewayCfg == nil {
			cfg.GatewayCfg = config.DefaultStorageProviderConfig.GatewayCfg
		}
		server, err = gateway.NewGatewayService(cfg.GatewayCfg)
		if err != nil {
			return nil, err
		}
	case UploaderService:
		if cfg.UploaderCfg == nil {
			cfg.UploaderCfg = config.DefaultStorageProviderConfig.UploaderCfg
		}
		server, err = uploader.NewUploaderService(cfg.UploaderCfg)
		if err != nil {
			return nil, err
		}
	case StoneHubService:
		if cfg.StoneHubCfg == nil {
			cfg.StoneHubCfg = config.DefaultStorageProviderConfig.StoneHubCfg
		}
		server, err = stonehub.NewStoneHubService(cfg.StoneHubCfg)
		if err != nil {
			return nil, err
		}
	case StoneNodeService:
		if cfg.StoneNodeCfg == nil {
			cfg.StoneNodeCfg = config.DefaultStorageProviderConfig.StoneNodeCfg
		}
		server, err = stonenode.NewStoneNodeService(cfg.StoneNodeCfg)
		if err != nil {
			return nil, err
		}
	case SyncerService:
		if cfg.SyncerCfg == nil {
			cfg.SyncerCfg = config.DefaultStorageProviderConfig.SyncerCfg
		}
		server, err = syncer.NewSyncerService(cfg.SyncerCfg)
		if err != nil {
			return nil, err
		}
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
			log.Errorw("init service failed ", "error", err)
		}
		log.Infow("init service success", "service", serviceName)
		// 2. register service to lifecycle.
		slc.RegisterServices(service)
	}
	// 3. start all services and listen os signals.
	ctx := context.Background()
	slc.Signals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP).StartServices(ctx).Wait(ctx)
}
