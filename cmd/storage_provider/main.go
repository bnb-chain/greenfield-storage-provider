package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/cmd/p2p"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/pelletier/go-toml/v2"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/urfave/cli/v2"
)

var (
	appName  = "gnfd-sp"
	appUsage = "the gnfd-sp command line interface"
)

var app *cli.App

// flags that configure the storage provider
var (
	configFlags = []cli.Flag{
		utils.ConfigFileFlag,
		utils.ServerFlag,
	}

	rcmgrFlags = []cli.Flag{
		utils.DisableResourceManagerFlag,
	}

	logFlags = []cli.Flag{
		utils.LogLevelFlag,
		utils.LogPathFlag,
		utils.LogStdOutputFlag,
	}

	metricsFlags = []cli.Flag{
		utils.MetricsDisableFlag,
		utils.MetricsHTTPFlag,
	}

	pprofFlags = []cli.Flag{
		utils.PProfDisableFlag,
		utils.PProfHTTPFlag,
	}
)

func init() {
	app = cli.NewApp()
	app.Name = appName
	app.Usage = appUsage
	app.Action = storageProvider
	app.HideVersion = true
	app.Flags = utils.MergeFlags(
		configFlags,
		rcmgrFlags,
		logFlags,
		metricsFlags,
		pprofFlags,
	)
	app.Commands = []*cli.Command{
		// p2p category commands
		p2p.P2PCreateKeysCmd,
		// miscellaneous category commands
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

func loadConfig(file string, cfg *gfspconfig.GfSpConfig) error {
	if cfg == nil {
		return errors.New("failed to load config file, the config param invalid")
	}
	bz, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return toml.Unmarshal(bz, cfg)
}

// makeConfig loads the configuration from local file and replace the fields by flags.
func makeConfig(ctx *cli.Context) (*gfspconfig.GfSpConfig, error) {
	cfg := &gfspconfig.GfSpConfig{}
	if ctx.IsSet(utils.ConfigFileFlag.Name) {
		err := loadConfig(ctx.String(utils.ConfigFileFlag.Name), cfg)
		if err != nil {
			log.Errorw("failed to load config file", "error", err)
			return nil, err
		}
	}
	if ctx.IsSet(utils.ServerFlag.Name) {
		servers := util.SplitByComma(ctx.String(utils.ServerFlag.Name))
		cfg.Server = servers
	}
	if ctx.IsSet(utils.MetricsDisableFlag.Name) {
		cfg.Monitor.DisableMetrics = ctx.Bool(utils.MetricsDisableFlag.Name)
	}
	if ctx.IsSet(utils.MetricsHTTPFlag.Name) {
		cfg.Monitor.MetricsHttpAddress = ctx.String(utils.MetricsHTTPFlag.Name)
	}
	if ctx.IsSet(utils.PProfDisableFlag.Name) {
		cfg.Monitor.DisablePProf = ctx.Bool(utils.PProfDisableFlag.Name)
	}
	if ctx.IsSet(utils.PProfHTTPFlag.Name) {
		cfg.Monitor.PProfHttpAddress = ctx.String(utils.PProfHTTPFlag.Name)
	}
	if ctx.IsSet(utils.DisableResourceManagerFlag.Name) {
		cfg.Rcmgr.DisableRcmgr = ctx.Bool(utils.DisableResourceManagerFlag.Name)
	}
	return cfg, nil
}

// makeEnv init storage provider runtime environment
func makeEnv(ctx *cli.Context, cfg *gfspconfig.GfSpConfig) error {
	var (
		logLevel = "debug"
		logPath  = "./log"
	)
	if ctx.IsSet(utils.LogLevelFlag.Name) {
		logLevel = ctx.String(utils.LogLevelFlag.Name)
	}
	if ctx.IsSet(utils.LogPathFlag.Name) {
		logPath = ctx.String(utils.LogPathFlag.Name)
	}
	if ctx.IsSet(utils.LogStdOutputFlag.Name) {
		logPath = ""
	}
	loglvl, err := log.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	log.Init(loglvl, logPath)
	return nil
}

// storageProvider is the main entry point into the system if no special subcommand
// is ran. It uses default config to  run storage provider services based  on the
// command line arguments and runs it in blocking mode, waiting for it to be shut
// down.
func storageProvider(ctx *cli.Context) error {
	cfg, err := makeConfig(ctx)
	if err != nil {
		log.Errorw("failed to make gfsp config", "error", err)
		return nil
	}
	err = makeEnv(ctx, cfg)
	if err != nil {
		log.Errorw("failed to make gfsp env", "error", err)
		return nil
	}
	gfsp, err := gfspapp.NewGfSpBaseApp(cfg)
	if err != nil {
		log.Errorw("failed to init gfsp app", "error", err)
		return err
	}
	return gfsp.Start(context.Background())
}
