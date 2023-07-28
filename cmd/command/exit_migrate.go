package command

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var spOperatorAddressFlag = &cli.StringFlag{
	Name:  "operatorAddress",
	Usage: "The operator account address of the storage provider who want to exit from the greenfield storage network",
}

var SPExitCmd = &cli.Command{
	Name:  "sp.exit",
	Usage: "Used for sp exits from the Greenfield storage network",
	Description: `Using this command, it will send an transaction to Greenfield blockchain to tell this SP is prepared
		to exit from Greenfield storage network`,
	Category: "MIGRATE COMMANDS",
	Action:   spExit,
	Flags: []cli.Flag{
		spOperatorAddressFlag,
	},
}

func spExit(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		return err
	}
	client := utils.MakeGfSpClient(cfg)
	operatorAddress := ctx.String(spOperatorAddressFlag.Name)
	// TODO: add more verification for cli args
	if operatorAddress != cfg.SpAccount.SpOperatorAddress {
		return fmt.Errorf("invalid operator address")
	}
	txHash, err := client.SPExit(ctx.Context, &virtualgrouptypes.MsgStorageProviderExit{StorageProvider: operatorAddress})
	if err != nil {
		fmt.Printf("failed to send sp exit txn, operatorAddress: %s\n", operatorAddress)
		return err
	}
	fmt.Printf("send sp exit txn successfully, txn hash: %s", txHash)
	return nil
}
