package command

import (
	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"

	"github.com/urfave/cli/v2"
)

const spExitCommands = "Sp Exit Commands"

var SpExitCmd = &cli.Command{
	Action: SpExitAction,
	Name:   "spExit ",
	Usage:  "sp exit",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
	},
	Category:    spExitCommands,
	Description: ``,
}

func SpExitAction(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		return err
	}

	msg := &virtualgrouptypes.MsgStorageProviderExit{
		StorageProvider: cfg.SpAccount.SpOperatorAddress,
	}

	spClient := utils.MakeGfSpClient(cfg)
	_, err = spClient.SpExit(ctx.Context, msg)
	return err
}

var CompleteSpExitCmd = &cli.Command{
	Action: CompleteSpExitAction,
	Name:   "completeSpExit ",
	Usage:  "complete sp exit",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
	},
	Category:    spExitCommands,
	Description: ``,
}

func CompleteSpExitAction(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		return err
	}

	msg := &virtualgrouptypes.MsgCompleteStorageProviderExit{
		StorageProvider: cfg.SpAccount.SpOperatorAddress,
	}

	spClient := utils.MakeGfSpClient(cfg)
	_, err = spClient.CompleteSpExit(ctx.Context, msg)
	return err
}
