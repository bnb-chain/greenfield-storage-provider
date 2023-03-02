package main

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"gopkg.in/urfave/cli.v1"

	"github.com/bnb-chain/greenfield-storage-provider/config"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var (
	appName  = "gnfd-sp"
	appUsage = "the gnfd-sp command line interface"
)

var app *cli.App

var (
	configFlag = cli.StringFlag{
		Name:  "config",
		Usage: "File path for storage provider configuration",
	}
	versionFlag = cli.BoolFlag{
		Name:  "version",
		Usage: "Show the storage provider version information",
	}

	// flags that configure the storage provider
	spFlags = []cli.Flag{
		configFlag,
		versionFlag,
	}
)

func init() {
	app = cli.NewApp()
	app.Name = appName
	app.Usage = appUsage
	app.Action = storageProvider
	app.HideVersion = true
	app.Flags = append(app.Flags, spFlags...)
	app.Commands = []cli.Command{}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func storageProvider(ctx *cli.Context) error {
	if ctx.GlobalIsSet(versionFlag.Name) {
		fmt.Print(DumpLogo() + DumpVersion())
		return nil
	}
	if !ctx.GlobalIsSet(configFlag.Name) {
		return fmt.Errorf("invalid params")
	}

	cfg := config.LoadConfig(configFlag.Value)
	slc := lifecycle.NewServiceLifecycle()
	for _, serviceName := range cfg.Service {
		// 1. init service instance.
		service, err := initService(serviceName, cfg)
		if err != nil {
			log.Errorw("failed to init service", "service", serviceName, "error", err)
			os.Exit(1)
		}
		log.Debugw("success to init service ", "service", serviceName)
		// 2. register service to lifecycle.
		slc.RegisterServices(service)
	}
	// 3. start all services and listen os signals.
	slcCtx := context.Background()
	slc.Signals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP).StartServices(slcCtx).Wait(slcCtx)
	return nil
}
