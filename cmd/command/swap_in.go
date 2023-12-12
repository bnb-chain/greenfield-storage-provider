package command

import (
	"errors"

	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"

	"github.com/urfave/cli/v2"
)

const swapInCommands = "SwapIn Commands"

var ngvgIDFlag = &cli.Uint64Flag{
	Name:     "ngvgId",
	Usage:    "gvg id",
	Aliases:  []string{"gid"},
	Required: true,
}

var vgfIDFlag = &cli.Uint64Flag{
	Name:     "vgf",
	Usage:    "",
	Aliases:  []string{"f"},
	Required: true,
}

var targetSPIDFlag = &cli.Uint64Flag{
	Name:     "targetSP",
	Usage:    "target sp",
	Aliases:  []string{"sp"},
	Required: true,
}

var SwapInCmd = &cli.Command{
	Action: SwapInAction,
	Name:   "swapIn",
	Usage:  "Successor swap in GVG/VGF",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		ngvgIDFlag,
		vgfIDFlag,
		targetSPIDFlag,
	},
	Category:    swapInCommands,
	Description: ``,
}

var RecoverGVGCmd = &cli.Command{
	Action: RecoverGVGAction,
	Name:   "recoverGVG",
	Usage:  "Successor swap in GVG/VGF",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		ngvgIDFlag,
	},
	Category:    swapInCommands,
	Description: ``,
}

var RecoverVGFCmd = &cli.Command{
	Action: RecoverVGFAction,
	Name:   "recover-vgf",
	Usage:  "recover objects in vgf",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		vgfIDFlag,
	},
	Category:    swapInCommands,
	Description: ``,
}

var CompleteSwapInCmd = &cli.Command{
	Action: CompleteSwapInAction,
	Name:   "completeSwapIn",
	Usage:  "complete swap in",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		ngvgIDFlag,
		vgfIDFlag,
	},
	Category:    swapInCommands,
	Description: ``,
}

func SwapInAction(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		return err
	}

	targetSpID := ctx.Uint64(targetSPIDFlag.Name)
	gvgID := ctx.Uint64(ngvgIDFlag.Name)
	gvgfID := ctx.Uint64(vgfIDFlag.Name)

	reserveSwapIn := &virtualgrouptypes.MsgReserveSwapIn{
		TargetSpId:                 uint32(targetSpID),
		GlobalVirtualGroupFamilyId: uint32(gvgfID),
		GlobalVirtualGroupId:       uint32(gvgID),
		StorageProvider:            cfg.SpAccount.SpOperatorAddress,
	}

	spClient := utils.MakeGfSpClient(cfg)
	_, err = spClient.ReserveSwapIn(ctx.Context, reserveSwapIn)
	return err
}

func CompleteSwapInAction(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		return err
	}

	gvgID := ctx.Uint64(ngvgIDFlag.Name)
	gvgfID := ctx.Uint64(vgfIDFlag.Name)

	completeSwapIn := &virtualgrouptypes.MsgCompleteSwapIn{
		GlobalVirtualGroupFamilyId: uint32(gvgfID),
		GlobalVirtualGroupId:       uint32(gvgID),
		StorageProvider:            cfg.SpAccount.SpOperatorAddress,
	}

	spClient := utils.MakeGfSpClient(cfg)
	_, err = spClient.CompleteSwapIn(ctx.Context, completeSwapIn)
	return err
}

func RecoverGVGAction(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		return err
	}

	// get client
	chainClient, err := utils.MakeGnfd(cfg)
	if err != nil {
		return err
	}
	spClient := utils.MakeGfSpClient(cfg)

	// check swapIn info
	sp, err := chainClient.QuerySP(ctx.Context, cfg.SpAccount.SpOperatorAddress)
	if err != nil {
		return err
	}
	gvgID := ctx.Uint64(ngvgIDFlag.Name)
	swapInInfo, err := chainClient.QuerySwapInInfo(ctx.Context, 0, uint32(gvgID))
	if err != nil {
		return err
	}
	if swapInInfo.GetSuccessorSpId() != sp.GetId() {
		return errors.New("sp is not successor sp")
	}

	//get replicateIndex
	gvgInfo, err := spClient.GetGlobalVirtualGroupByGvgID(ctx.Context, uint32(gvgID))
	if err != nil {
		return err
	}
	vgfId := gvgInfo.GetFamilyId()
	vgfInfo, err := spClient.GetVirtualGroupFamily(ctx.Context, vgfId)
	if err != nil {
		return err
	}
	var replicateIndex int32
	for idx, gvg := range vgfInfo.GetGlobalVirtualGroupIds() {
		if uint64(gvg) == gvgID {
			replicateIndex = int32(idx - 1)
			break
		}
	}
	// trigger
	return spClient.TriggerRecoverForSuccessorSP(ctx.Context, 0, uint32(gvgID), replicateIndex)
}

func RecoverVGFAction(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		return err
	}

	// get client
	chainClient, err := utils.MakeGnfd(cfg)
	if err != nil {
		return err
	}
	spClient := utils.MakeGfSpClient(cfg)

	// check swapIn info
	sp, err := chainClient.QuerySP(ctx.Context, cfg.SpAccount.SpOperatorAddress)
	if err != nil {
		return err
	}
	vgfID := ctx.Uint64(vgfIDFlag.Name)
	swapInInfo, err := chainClient.QuerySwapInInfo(ctx.Context, uint32(vgfID), 0)
	if err != nil {
		return err
	}
	if swapInInfo.GetSuccessorSpId() != sp.GetId() {
		return errors.New("sp is not successor sp")
	}

	// trigger
	return spClient.TriggerRecoverForSuccessorSP(ctx.Context, uint32(vgfID), 0, -1)
}
