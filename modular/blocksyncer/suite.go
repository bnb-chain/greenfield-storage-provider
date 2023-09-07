package blocksyncer

import (
	"context"
	"fmt"
	"runtime"

	"github.com/forbole/juno/v4/log"
	"github.com/stretchr/testify/suite"
	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
)

type BlockSyncerE2eBaseSuite struct {
	suite.Suite
	Context context.Context
}

var App *cli.App

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
)

var (
	appName  = "gnfd-sp"
	appUsage = "the Greenfield Storage Provider command line interface"
)

var VersionCmd = &cli.Command{
	Action:      versionAction,
	Name:        "version",
	Aliases:     []string{"v"},
	Usage:       "Print version information",
	Category:    "MISCELLANEOUS COMMANDS",
	Description: `The output of this command is supposed to be machine-readable.`,
}

const (
	StorageProviderLogo = `Greenfield Storage Provider
    __                                                       _     __
    _____/ /_____  _________ _____ ____     ____  _________ _   __(_)___/ /__  _____
    / ___/ __/ __ \/ ___/ __  / __  / _ \   / __ \/ ___/ __ \ | / / / __  / _ \/ ___/
    (__  ) /_/ /_/ / /  / /_/ / /_/ /  __/  / /_/ / /  / /_/ / |/ / / /_/ /  __/ /
    /____/\__/\____/_/   \__,_/\__, /\___/  / .___/_/   \____/|___/_/\__,_/\___/_/
    /____/       /_/
    `
)

// DumpLogo output greenfield storage provider logo
func DumpLogo() string {
	return StorageProviderLogo
}

func versionAction(ctx *cli.Context) error {
	fmt.Print(DumpLogo() + "\n" + DumpVersion())
	return nil
}

var (
	Version    string
	CommitID   string
	BranchName string
	BuildTime  string
)

// DumpVersion output the storage provider version information
func DumpVersion() string {
	return fmt.Sprintf("Version : %s\n"+
		"Branch  : %s\n"+
		"Commit  : %s\n"+
		"Build   : %s %s %s %s\n",
		Version,
		BranchName,
		CommitID,
		runtime.Version(), runtime.GOOS, runtime.GOARCH, BuildTime)
}

func RegisterModular() {
	gfspapp.RegisterModular(module.BlockSyncerModularName, module.BlockSyncerModularDescription, NewBlockSyncerModular)
}

func StorageProvider(ctx *cli.Context) error {
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

func (s *BlockSyncerE2eBaseSuite) SetupSuite() {
	s.Context = context.Background()

	App = cli.NewApp()
	App.Name = appName
	App.Usage = appUsage
	App.Action = StorageProvider
	App.HideVersion = true
	App.Flags = utils.MergeFlags(
		configFlags,
		rcmgrFlags,
		logFlags,
	)
	App.Commands = []*cli.Command{}
	RegisterModular()
}
