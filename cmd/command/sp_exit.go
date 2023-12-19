package command

import (
	"fmt"
	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"

	"github.com/urfave/cli/v2"
)

const spExitCommands = "Sp Exit Commands"

var SpExitCmd = &cli.Command{
	Action: SpExitAction,
	Name:   "spExit",
	Usage:  "sp exit",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
	},
	Category:    spExitCommands,
	Description: `Running this command sends exit tx to the chain and cannot be canceled after execution`,
}

func SpExitAction(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		println(err.Error())
		return err
	}

	msg := &virtualgrouptypes.MsgStorageProviderExit{
		StorageProvider: cfg.SpAccount.SpOperatorAddress,
	}

	spClient := utils.MakeGfSpClient(cfg)
	tx, err := spClient.SpExit(ctx.Context, msg)
	if err != nil {
		println(err.Error())
		return err
	}
	fmt.Printf("tx successfully! tx_hash:%s", tx)
	return nil
}

var CompleteSpExitCmd = &cli.Command{
	Action: CompleteSpExitAction,
	Name:   "completeSpExit",
	Usage:  "complete sp exit",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
	},
	Category:    spExitCommands,
	Description: `When Successor has recovered all resources, you can use this CMD to complete the exit`,
}

func CompleteSpExitAction(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		println(err.Error())
		return err
	}

	msg := &virtualgrouptypes.MsgCompleteStorageProviderExit{
		StorageProvider: cfg.SpAccount.SpOperatorAddress,
		Operator:        cfg.SpAccount.SpOperatorAddress,
	}

	spClient := utils.MakeGfSpClient(cfg)
	tx, err := spClient.CompleteSpExit(ctx.Context, msg)
	if err != nil {
		println(err.Error())
		return err
	}
	fmt.Printf("tx successfully! tx_hash:%s", tx)
	return nil
}
