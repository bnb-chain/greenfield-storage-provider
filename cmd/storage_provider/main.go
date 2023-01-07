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
	StoneHubService  = "StoneHub"
	GetaWayService   = "Gateway"
	UploaderService  = "Uploader"
	StoneNodeService = "StoneNode"
	SyncerService    = "Syncer"
)

func main() {
	flag.Parse()
	if *version {
		fmt.Print(DumpLogo() + DumpVersion())
		os.Exit(0)
	}

	var (
		ctx = context.Background()
		cfg = config.LoadConfig(*configFile)
		slc = lifecycle.NewServiceLifecycle()
	)
	for _, serviceName := range cfg.Service {
		switch serviceName {
		case GetaWayService:
			if cfg.GatewayCfg == nil {
				cfg.GatewayCfg = config.DefaultStorageProviderConfig.GatewayCfg
			}
			server, err := gateway.NewGatewayService(cfg.GatewayCfg)
			if err != nil {
				log.Errorw("gateway init failed", "error", err)
				os.Exit(1)
			}
			slc.RegisterServices(server)
		case UploaderService:
			if cfg.UploaderCfg == nil {
				cfg.UploaderCfg = config.DefaultStorageProviderConfig.UploaderCfg
			}
			server, err := uploader.NewUploaderService(cfg.UploaderCfg)
			if err != nil {
				log.Errorw("uploader init failed", "error", err)
				os.Exit(1)
			}
			slc.RegisterServices(server)
		case StoneHubService:
			if cfg.StoneHubCfg == nil {
				cfg.StoneHubCfg = config.DefaultStorageProviderConfig.StoneHubCfg
			}
			server, err := stonehub.NewStoneHubService(cfg.StoneHubCfg)
			if err != nil {
				log.Errorw("stone hub init failed", "error", err)
				os.Exit(1)
			}
			slc.RegisterServices(server)
		case StoneNodeService:
			if cfg.StoneNodeCfg == nil {
				cfg.StoneNodeCfg = config.DefaultStorageProviderConfig.StoneNodeCfg
			}
			server, err := stonenode.NewStoneNodeService(cfg.StoneNodeCfg)
			if err != nil {
				log.Errorw("stone node init failed", "error", err)
				os.Exit(1)
			}
			slc.RegisterServices(server)
		case SyncerService:
			if cfg.SyncerCfg == nil {
				cfg.SyncerCfg = config.DefaultStorageProviderConfig.SyncerCfg
			}
			server, err := syncer.NewSyncerService(cfg.SyncerCfg)
			if err != nil {
				log.Errorw("syncer init failed", "error", err)
				os.Exit(1)
			}
			slc.RegisterServices(server)
		}
		log.Info("init service success ", serviceName)

	}
	// start all services, and listen signals
	slc.Signals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP).StartServices(ctx).Wait(ctx)
}
