package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/cmd/command"
	"github.com/bnb-chain/greenfield-storage-provider/cmd/command/bs_data_migration"
	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/modular/approver"
	"github.com/bnb-chain/greenfield-storage-provider/modular/authenticator"
	"github.com/bnb-chain/greenfield-storage-provider/modular/blocksyncer"
	"github.com/bnb-chain/greenfield-storage-provider/modular/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/modular/executor"
	"github.com/bnb-chain/greenfield-storage-provider/modular/gater"
	"github.com/bnb-chain/greenfield-storage-provider/modular/manager"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata"
	"github.com/bnb-chain/greenfield-storage-provider/modular/p2p"
	"github.com/bnb-chain/greenfield-storage-provider/modular/receiver"
	"github.com/bnb-chain/greenfield-storage-provider/modular/signer"
	"github.com/bnb-chain/greenfield-storage-provider/modular/uploader"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// registerModular registers the module to the module manager, register info include:
// module name, module description and new module func. Module name is an indexer for
// starting, the start module name comes from config file or '--service' command flag.
// Module description uses for 'list' command that shows the SP supports modules info.
// New module func is help module manager to init the module instance.
func registerModular() {
	gfspapp.RegisterModular(module.ApprovalModularName, module.ApprovalModularDescription, approver.NewApprovalModular)
	gfspapp.RegisterModular(module.AuthenticationModularName, module.AuthenticationModularDescription, authenticator.NewAuthenticationModular)
	gfspapp.RegisterModular(module.DownloadModularName, module.DownloadModularDescription, downloader.NewDownloadModular)
	gfspapp.RegisterModular(module.ExecuteModularName, module.ExecuteModularDescription, executor.NewExecuteModular)
	gfspapp.RegisterModular(module.GateModularName, module.GateModularDescription, gater.NewGateModular)
	gfspapp.RegisterModular(module.ManageModularName, module.ManageModularDescription, manager.NewManageModular)
	gfspapp.RegisterModular(module.P2PModularName, module.P2PModularDescription, p2p.NewP2PModular)
	gfspapp.RegisterModular(module.ReceiveModularName, module.ReceiveModularDescription, receiver.NewReceiveModular)
	gfspapp.RegisterModular(module.SignModularName, module.SignModularDescription, signer.NewSignModular)
	gfspapp.RegisterModular(module.MetadataModularName, module.MetadataModularDescription, metadata.NewMetadataModular)
	gfspapp.RegisterModular(module.UploadModularName, module.UploadModularDescription, uploader.NewUploadModular)
	gfspapp.RegisterModular(module.BlockSyncerModularName, module.BlockSyncerModularDescription, blocksyncer.NewBlockSyncerModular)
}

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
		VersionCmd,
		// config category commands
		command.ConfigDumpCmd,
		// query category commands
		command.ListModulesCmd,
		command.ListErrorsCmd,
		command.QueryTaskCmd,
		command.GetObjectCmd,
		command.ChallengePieceCmd,
		command.GetSegmentIntegrityCmd,
		// query sp exit and bucket migrate status
		command.QueryBucketMigrateCmd,
		command.QuerySPExitCmd,

		// query primary and secondary SP income details
		command.QueryPrimarySPIncomeCmd,
		command.QuerySecondarySPIncomeCmd,
		// p2p category commands
		command.P2PCreateKeysCmd,
		// debug commands
		command.DebugCreateBucketApprovalCmd,
		command.DebugCreateObjectApprovalCmd,
		command.DebugReplicateApprovalCmd,
		command.DebugPutObjectCmd,
		// recovery commands
		command.RecoverObjectCmd,
		command.RecoverPieceCmd,
		// sp exit
		command.SPExitCmd,
		command.CompleteSPExitCmd,  // only for debugging
		command.CompleteSwapOutCmd, // only for debugging
		// update quota
		command.SetQuotaCmd,
		// block syncer
		bs_data_migration.BsDataMigrationCmd,
		// be related to sp exit
		command.SpExitCmd,
		command.CompleteSpExitCmd,
		command.SwapInCmd,
		command.CompleteSwapInCmd,
		command.CancelSwapInCmd,
		command.RecoverGVGCmd,
		command.RecoverVGFCmd,
		command.QueryRecoverProcessCmd,
		command.ListGlobalVirtualGroupsBySecondarySPCmd,
		command.ListVirtualGroupFamiliesBySpIDCmd,
	}
	registerModular()
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// storageProvider is the main entry point into the system if no special subcommand
// is run. It uses default config to  run storage provider services based  on the
// command line arguments and runs it in blocking mode, waiting for it to be shutdown.
func storageProvider(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		log.Errorw("failed to make gf-sp config", "error", err)
		return nil
	}
	err = utils.MakeEnv(ctx, cfg)
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
