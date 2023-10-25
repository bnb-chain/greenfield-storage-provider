package command

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/util"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

const migrateCommands = "MIGRATE COMMANDS"

var spOperatorAddressFlag = &cli.StringFlag{
	Name:     "operatorAddress",
	Usage:    "The operator account address of the storage provider who want to exit from the greenfield storage network",
	Required: true,
}

var SPExitCmd = &cli.Command{
	Name:  "sp.exit",
	Usage: "Used for sp exits from the Greenfield storage network",
	Description: `Using this command, it will send an transaction to Greenfield blockchain to tell this SP is prepared ` +
		`to exit from Greenfield storage network.`,
	Category: migrateCommands,
	Action:   CW.spExit,
	Flags: []cli.Flag{
		spOperatorAddressFlag,
	},
}

func (w *CMDWrapper) spExit(ctx *cli.Context) error {
	err := w.init(ctx)
	if err != nil {
		return err
	}
	operatorAddress := ctx.String(spOperatorAddressFlag.Name)
	if operatorAddress != w.config.SpAccount.SpOperatorAddress {
		fmt.Printf("failed to check operator address, actual=%v, expected=%v\n", operatorAddress, w.config.SpAccount.SpOperatorAddress)
		return fmt.Errorf("invalid operator address")
	}
	txHash, err := w.grpcAPI.SPExit(ctx.Context, &virtualgrouptypes.MsgStorageProviderExit{StorageProvider: operatorAddress})
	if err != nil {
		fmt.Printf("failed to send sp exit tx, operatorAddress: %s, error:%s\n", operatorAddress, err)
		return err
	}
	fmt.Printf("send sp exit txn successfully, txn hash: %s\n", txHash)
	return nil
}

/*
The following commands are only used in debug scenarios.
*/
var swapOutFamilyID = &cli.StringFlag{
	Name:  "familyID",
	Usage: "The familyID who intends to swap out",
}

var swapOutGVGIDList = &cli.StringFlag{
	Name:  "gvgIDList",
	Usage: "The gvgIDList who intends to swap out, eg: '1,2,3'",
}

var CompleteSPExitCmd = &cli.Command{
	Name:  "sp.complete.exit",
	Usage: "Only used in debugging scenarios, online use not allowed. Used for sp complete exits from the Greenfield storage network.",
	Description: `Using this command, it will send an transaction to Greenfield blockchain to tell this SP is prepared ` +
		`to complete exit from Greenfield storage network.`,
	Category: migrateCommands,
	Action:   CW.completeSPExit,
	Flags: []cli.Flag{
		spOperatorAddressFlag,
	},
}

func (w *CMDWrapper) completeSPExit(ctx *cli.Context) error {
	err := w.init(ctx)
	if err != nil {
		return err
	}
	operatorAddress := ctx.String(spOperatorAddressFlag.Name)
	if operatorAddress != w.config.SpAccount.SpOperatorAddress {
		fmt.Printf("failed to check operator address, actual=%v, expected=%v\n", operatorAddress, w.config.SpAccount.SpOperatorAddress)
		return fmt.Errorf("invalid operator address")
	}
	txHash, err := w.grpcAPI.CompleteSPExit(ctx.Context, &virtualgrouptypes.MsgCompleteStorageProviderExit{StorageProvider: operatorAddress})
	if err != nil {
		fmt.Printf("failed to send complete sp exit tx, operatorAddress: %s, error:%s\n", operatorAddress, err)
		return err
	}
	fmt.Printf("send complete sp exit txn successfully, txn hash: %s\n", txHash)
	return nil
}

var CompleteSwapOutCmd = &cli.Command{
	Name:  "sp.complete.swapout",
	Usage: "Only used in debugging scenarios, online use not allowed. Used for swap out from the Greenfield storage network.",
	Description: `Using this command, it will send an transaction to Greenfield blockchain to tell this SP is prepared ` +
		`to swap out from Greenfield storage network.`,
	Category: migrateCommands,
	Action:   CW.completeSwapOut,
	Flags: []cli.Flag{
		spOperatorAddressFlag,
		swapOutFamilyID,
		swapOutGVGIDList,
	},
}

func (w *CMDWrapper) completeSwapOut(ctx *cli.Context) error {
	err := w.init(ctx)
	if err != nil {
		return err
	}
	operatorAddress := ctx.String(spOperatorAddressFlag.Name)
	if operatorAddress != w.config.SpAccount.SpOperatorAddress {
		return fmt.Errorf("invalid operator address")
	}
	familyID := uint32(ctx.Uint64(ctx.String(swapOutFamilyID.Name)))
	gvgIDList, err := util.StringToUint32Slice(ctx.String(swapOutGVGIDList.Name))
	if err != nil {
		fmt.Printf("failed to send complete swap out tx, operatorAddress: %s, gvgIDList: %s\n", operatorAddress, ctx.String(swapOutGVGIDList.Name))
		return err
	}
	txHash, err := w.grpcAPI.CompleteSwapOut(ctx.Context, &virtualgrouptypes.MsgCompleteSwapOut{
		StorageProvider:            operatorAddress,
		GlobalVirtualGroupFamilyId: familyID,
		GlobalVirtualGroupIds:      gvgIDList})
	if err != nil {
		fmt.Printf("failed to send complete swap out txn, operatorAddress: %s\n", operatorAddress)
		return err
	}
	fmt.Printf("send complete swap out txn successfully, txn hash: %s\n", txHash)
	return nil
}
