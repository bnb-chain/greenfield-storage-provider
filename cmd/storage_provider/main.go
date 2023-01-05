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
	"github.com/bnb-chain/inscription-storage-provider/service/uploader"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

var (
	version    = flag.Bool("version", false, "print version")
	configFile = flag.String("config", "../../config/config.toml", "config file path")
)

func main() {
	flag.Parse()
	if *version {
		fmt.Print(DumpLogo() + DumpVersion())
		os.Exit(0)
	}

	spCfg := config.LoadConfig(*configFile)
	slc := lifecycle.NewServiceLifecycle()
	for _, serviceName := range spCfg.Service {
		switch serviceName {
		case "StoneHub":
			if spCfg.StoneHubCfg == nil {
				spCfg.StoneHubCfg = config.DefaultStorageProviderConfig.StoneHubCfg
			}
			server, err := stonehub.NewStoneHubService(spCfg.StoneHubCfg)
			if err != nil {
				log.Error("stone hub init fail", "error", err)
				os.Exit(1)
			}
			slc.RegisterServices(server)
		case "Gateway":
			if spCfg.GatewayCfg == nil {
				spCfg.GatewayCfg = config.DefaultStorageProviderConfig.GatewayCfg
			}
			server, err := gateway.NewGatewayService(spCfg.GatewayCfg)
			if err != nil {
				log.Error("gateway init fail", "error", err)
				os.Exit(1)
			}
			slc.RegisterServices(server)
		case "Uploader":
			if spCfg.UploaderCfg == nil {
				spCfg.UploaderCfg = config.DefaultStorageProviderConfig.UploaderCfg
			}
			server, err := uploader.NewUploaderService(spCfg.UploaderCfg)
			if err != nil {
				log.Error("uploader init fail", "error", err)
				os.Exit(1)
			}
			slc.RegisterServices(server)
		}
		log.Info("init service success ", serviceName)

	}
	ctx := context.Background()
	slc.Signals(syscall.SIGINT, syscall.SIGTERM).StartServices(ctx).Wait(ctx)
}
