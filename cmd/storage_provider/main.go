package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/bnb-chain/greenfield-storage-provider/config"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

var (
	version    = flag.Bool("version", false, "print version")
	configFile = flag.String("config", "./", "config file path")
)

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
