package main

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/cmd/conf"
	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	"github.com/bnb-chain/greenfield-storage-provider/config"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

var (
	appName  = "gnfd-sp"
	appUsage = "the gnfd-sp command line interface"
)

var app *cli.App

var (
	// flags that configure the storage provider
	spFlags = []cli.Flag{
		utils.ConfigFileFlag,
		utils.ConfigRemoteFlag,
		utils.ServerFlag,
		utils.DBUserFlag,
		utils.DBPasswordFlag,
		utils.DBAddressFlag,
		utils.DBDataBaseFlag,
		utils.LogLevelFlag,
		utils.LogPathFlag,
		utils.LogStdOutputFlag,
	}
)

func init() {
	app = cli.NewApp()
	app.Name = appName
	app.Usage = appUsage
	app.Action = storageProvider
	app.HideVersion = true
	app.Flags = append(app.Flags, spFlags...)
	app.Commands = []*cli.Command{
		// config category commands
		conf.ConfigDumpCmd,
		conf.ConfigUploadCmd,
		// miscellaneous commands
		VersionCmd,
		utils.ListServiceCmd,
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// makeConfig loads the configuration and creates the storage provider backend.
func makeConfig(ctx *cli.Context) (*config.StorageProviderConfig, error) {
	// load config from remote db or local config file
	cfg := config.DefaultStorageProviderConfig
	if ctx.IsSet(utils.ConfigRemoteFlag.Name) {
		spDB, err := utils.MakeSPDB(ctx)
		if err != nil {
			return nil, err
		}
		_, cfgBytes, err := spDB.GetAllServiceConfigs()
		if err != nil {
			return nil, err
		}
		if err := cfg.JSONUnMarshal([]byte(cfgBytes)); err != nil {
			return nil, err
		}
	} else if ctx.IsSet(utils.ConfigFileFlag.Name) {
		cfg = config.LoadConfig(ctx.String(utils.ConfigFileFlag.Name))
	}
	// override the services to be started by flag
	if ctx.IsSet(utils.ServerFlag.Name) {
		services := util.SplitByComma(ctx.String(utils.ServerFlag.Name))
		cfg.Service = services
	}
	// init log
	if err := initLog(ctx, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// storageProvider is the main entry point into the system if no special subcommand is ran.
// It uses default config to run storage provider services based on the command line arguments
// and runs it in blocking mode, waiting for it to be shut down.
func storageProvider(ctx *cli.Context) error {
	cfg, err := makeConfig(ctx)
	if err != nil {
		return err
	}
	slc := lifecycle.NewServiceLifecycle()
	for _, serviceName := range cfg.Service {
		// init service instance.
		service, err := initService(serviceName, cfg)
		if err != nil {
			log.Errorw("failed to init service", "service", serviceName, "error", err)
			os.Exit(1)
		}
		log.Debugw("success to init service ", "service", serviceName)
		// register service to lifecycle.
		slc.RegisterServices(service)
	}
	// start all services and listen os signals.
	slcCtx := context.Background()
	slc.Signals(syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP).StartServices(slcCtx).Wait(slcCtx)
	return nil
}
