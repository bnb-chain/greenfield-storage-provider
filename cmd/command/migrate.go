package command

import (
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"github.com/urfave/cli/v2"
)

var spOperatorAddressFlag = &cli.StringFlag{
	Name:  "operatorAddress",
	Usage: "The operator account address of the storage provider who want to exit from the greenfield storage network",
}

var SPExitCmd = &cli.Command{
	Name:  "exit",
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
	txHash, err := client.SPExit(ctx.Context, &virtualgrouptypes.MsgStorageProviderExit{StorageProvider: operatorAddress})
	if err != nil {
		fmt.Printf("failed to send sp exit txn, operatorAddress: %s\n", operatorAddress)
		return err
	}
	fmt.Printf("send sp exit txn successfully, txn hash: %s", txHash)
	return nil
}
