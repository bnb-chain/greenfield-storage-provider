package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/cmd/conf"
	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
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
	// flags that configure the storage provider
	spFlags = []cli.Flag{
		utils.VersionFlag,
		utils.ConfigFileFlag,
		utils.ConfigRemoteFlag,
		utils.ServerFlag,
		utils.DBUserFlag,
		utils.DBPasswordFlag,
		utils.DBAddressFlag,
		utils.LogLevelFlag,
		utils.LogPathFlag,
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
	cfg := &config.StorageProviderConfig{}
	var services []string
	if ctx.IsSet(utils.ServerFlag.Name) {
		services = utils.SplitTagsFlag(ctx.String(utils.ServerFlag.Name))
		for _, service := range services {
			if _, ok := model.SpServiceDesc[strings.ToLower(service)]; !ok {
				return nil, fmt.Errorf("invalid service %s", service)
			}
		}
	}
	if ctx.IsSet(utils.ConfigRemoteFlag.Name) {
		//&storeconf.SQLDBConfig{
		//	User:     ctx.String(utils.DBUserFlag.Name),
		//	Passwd:   ctx.String(utils.DBPasswordFlag.Name),
		//	Address:  ctx.String(utils.DBAddressFlag.Name),
		//	Database: "job_db",
		//}
		// TODO:: new sp db load config by services
	} else if ctx.IsSet(utils.ConfigFileFlag.Name) {
		cfg = config.LoadConfig(ctx.String(utils.ConfigFileFlag.Name))
	} else {
		cfg = config.DefaultStorageProviderConfig
	}
	if ctx.IsSet(utils.ServerFlag.Name) {
		cfg.Service = services
	}
	return cfg, nil
}

// storageProvider is the main entry point into the system if no special subcommand is ran.
// It uses default config to run storage provider services based on the command line arguments
// and runs it in blocking mode, waiting for it to be shut down.
func storageProvider(ctx *cli.Context) error {
	if ctx.IsSet(utils.VersionFlag.Name) {
		fmt.Print(DumpLogo() + "\n" + DumpVersion())
		return nil
	}
	logLevel, err := log.ParseLevel(ctx.String(utils.LogLevelFlag.Name))
	if err != nil {
		return err
	}
	log.Init(logLevel, ctx.String(utils.LogPathFlag.Name))

	cfg, err := makeConfig(ctx)
	if err != nil {
		return err
	}
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
