package main

import (
	"context"
	"flag"
	"os"
	"syscall"

	"github.com/bnb-chain/inscription-storage-provider/config"
	"github.com/bnb-chain/inscription-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/inscription-storage-provider/service/stonehub"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

var configFile = flag.String("config", "../../config/config.toml", "config file path")

func main() {
	flag.Parse()
	spCfg := config.LoadSPConfig(*configFile)

	lifecycle := lifecycle.NewServiceLifecycle()
	for _, serviceName := range spCfg.Service {
		switch serviceName {
		case "StoneHub":
			server, err := stonehub.NewStoneHubService(spCfg.StoneHubCfg)
			if err != nil {
				log.Error("stone hub init fail", "error", err)
				os.Exit(1)
			}
			log.Info("init service success", serviceName)
			lifecycle.RegisterServices(server)
		}
	}
	ctx := context.Background()
	lifecycle.Signals(syscall.SIGINT, syscall.SIGTERM).StartServices(ctx).Wait(ctx)
}
