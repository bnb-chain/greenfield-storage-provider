package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bnb-chain/greenfield-storage-provider/cmd/command"
	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/modular/approver"
	"github.com/bnb-chain/greenfield-storage-provider/modular/authorizer"
	"github.com/bnb-chain/greenfield-storage-provider/modular/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/modular/executor"
	"github.com/bnb-chain/greenfield-storage-provider/modular/gater"
	"github.com/bnb-chain/greenfield-storage-provider/modular/manager"
	modularp2p "github.com/bnb-chain/greenfield-storage-provider/modular/p2p"
	"github.com/bnb-chain/greenfield-storage-provider/modular/receiver"
	"github.com/bnb-chain/greenfield-storage-provider/modular/retriever"
	"github.com/bnb-chain/greenfield-storage-provider/modular/singer"
	"github.com/bnb-chain/greenfield-storage-provider/modular/uploader"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/util"
)

var (
	appName  = "gnfd-sp"
	appUsage = "the Greenfield Storage Provider command line interface"
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
		// config category commands
		command.ConfigDumpCmd,
		// query category commands
		command.QueryTaskCmd,
		// p2p category commands
		command.P2PCreateKeysCmd,
		// miscellaneous category commands
		VersionCmd,
		command.ListModularCmd,
	}
	registerModular()
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// makeConfig loads the configuration from local file and replace the fields by flags.
func makeConfig(ctx *cli.Context) (*gfspconfig.GfSpConfig, error) {
	cfg := &gfspconfig.GfSpConfig{}
	if ctx.IsSet(utils.ConfigFileFlag.Name) {
		err := utils.LoadConfig(ctx.String(utils.ConfigFileFlag.Name), cfg)
		if err != nil {
			log.Errorw("failed to load config file", "error", err)
			return nil, err
		}
	}
	if ctx.IsSet(utils.ServerFlag.Name) {
		cfg.Server = util.SplitByComma(ctx.String(utils.ServerFlag.Name))
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

// registerModular registers the module to the module manager, register info include:
// module name, module description and new module func. Module name is an indexer for
// starting, the start module name comes from config file or '--service' command flag.
// Module description uses for 'list' command that shows the SP supports modules info.
// New module func is help module manager to init the module instance.
func registerModular() {
	gfspapp.RegisterModular(module.ApprovalModularName, module.ApprovalModularDescription, approver.NewApprovalModular)
	gfspapp.RegisterModular(module.AuthorizationModularName, module.AuthorizationModularDescription, authorizer.NewAuthorizeModular)
	gfspapp.RegisterModular(module.DownloadModularName, module.DownloadModularDescription, downloader.NewDownloadModular)
	gfspapp.RegisterModular(module.ExecuteModularName, module.ExecuteModularDescription, executor.NewExecuteModular)
	gfspapp.RegisterModular(module.GateModularName, module.GateModularDescription, gater.NewGateModular)
	gfspapp.RegisterModular(module.ManageModularName, module.ManageModularDescription, manager.NewManageModular)
	gfspapp.RegisterModular(module.P2PModularName, module.P2PModularDescription, modularp2p.NewP2PModular)
	gfspapp.RegisterModular(module.ReceiveModularName, module.ReceiveModularDescription, receiver.NewReceiveModular)
	gfspapp.RegisterModular(retriever.RetrieveModularName, retriever.RetrieveModularDescription, retriever.NewRetrieveModular)
	gfspapp.RegisterModular(module.SignerModularName, module.SignerModularDescription, singer.NewSingModular)
	gfspapp.RegisterModular(module.UploadModularName, module.UploadModularDescription, uploader.NewUploadModular)
}

// initLog inits the log configuration from config file and command flags.
func initLog(ctx *cli.Context, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Log.Level == "" {
		// TODO:: change to info
		cfg.Log.Level = "debug"
	}
	if cfg.Log.Path == "" {
		cfg.Log.Path = "./gnfd-sp.log"
	}
	if ctx.IsSet(utils.LogLevelFlag.Name) {
		cfg.Log.Level = ctx.String(utils.LogLevelFlag.Name)
	}
	if ctx.IsSet(utils.LogPathFlag.Name) {
		cfg.Log.Path = ctx.String(utils.LogPathFlag.Name)
	}
	if ctx.IsSet(utils.LogStdOutputFlag.Name) {
		cfg.Log.Path = ""
	}
	level, err := log.ParseLevel(cfg.Log.Level)
	if err != nil {
		return err
	}
	log.Init(level, cfg.Log.Path)
	return nil
}

// makeEnv inits storage provider runtime environment.
func makeEnv(ctx *cli.Context, cfg *gfspconfig.GfSpConfig) error {
	if err := initLog(ctx, cfg); err != nil {
		return err
	}
	return nil
}

// storageProvider is the main entry point into the system if no special subcommand
// is run. It uses default config to  run storage provider services based  on the
// command line arguments and runs it in blocking mode, waiting for it to be shut
// down.
func storageProvider(ctx *cli.Context) error {
	cfg, err := makeConfig(ctx)
	if err != nil {
		log.Errorw("failed to make gf-sp config", "error", err)
		return nil
	}
	err = makeEnv(ctx, cfg)
	if err != nil {
		log.Errorw("failed to make gf-sp env", "error", err)
		return nil
	}
	gfsp, err := gfspapp.NewGfSpBaseApp(cfg)
	if err != nil {
		log.Errorw("failed to init gf-sp app", "error", err)
		return err
	}
	return gfsp.Start(context.Background())
}
