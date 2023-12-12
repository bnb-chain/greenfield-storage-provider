package command

//
//import (
//	"errors"
//
//	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
//	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
//	"github.com/urfave/cli/v2"
//)
//
//const swapInCommands = "SwapIn Commands"
//
//var gvgIDFlag = &cli.Uint64Flag{
//	Name:     "gvg",
//	Usage:    "gvg id",
//	Aliases:  []string{"gvg"},
//	Required: true,
//}
//
//var vgfIDFlag = &cli.Uint64Flag{
//	Name:     "vgf",
//	Usage:    "",
//	Aliases:  []string{"vgf"},
//	Required: true,
//}
//
//var targetSPIDFlag = &cli.Uint64Flag{
//	Name:     "targetSP",
//	Usage:    "target sp",
//	Aliases:  []string{"sp"},
//	Required: true,
//}
//
//var SwapInCmd = &cli.Command{
//	Action: SwapInAction,
//	Name:   "swap in",
//	Usage:  "Successor swap in GVG/VGF",
//	Flags: []cli.Flag{
//		utils.ConfigFileFlag,
//		gvgIDFlag,
//		vgfIDFlag,
//		targetSPIDFlag,
//	},
//	Category:    swapInCommands,
//	Description: ``,
//}
//
//var RecoverGVGCmd = &cli.Command{
//	Action: RecoverGVGAction,
//	Name:   "swap in",
//	Usage:  "Successor swap in GVG/VGF",
//	Flags: []cli.Flag{
//		utils.ConfigFileFlag,
//		gvgIDFlag,
//	},
//	Category:    swapInCommands,
//	Description: ``,
//}
//
//var RecoverVGFCmd = &cli.Command{
//	Action: RecoverVGFAction,
//	Name:   "recover-vgf",
//	Usage:  "recover objects in vgf",
//	Flags: []cli.Flag{
//		utils.ConfigFileFlag,
//		vgfIDFlag,
//	},
//	Category:    swapInCommands,
//	Description: ``,
//}
//
////var CompleteSwapInCmd = &cli.Command{
////	Action: CompleteSwapInAction,
////	Name:   "complete swap in",
////	Usage:  "complete swap in",
////	Flags: []cli.Flag{
////		utils.ConfigFileFlag,
////		gvgIDFlag,
////		vgfIDFlag,
////	},
////	Category:    swapInCommands,
////	Description: ``,
////}
//
//func SwapInAction(ctx *cli.Context) error {
//	cfg, err := utils.MakeConfig(ctx)
//	if err != nil {
//		return err
//	}
//
//	targetSpID := ctx.Uint64(targetSPIDFlag.Name)
//	gvgID := ctx.Uint64(gvgIDFlag.Name)
//	gvgfID := ctx.Uint64(gvgIDFlag.Name)
//
//	reserveSwapIn := &virtualgrouptypes.MsgReserveSwapIn{
//		TargetSpId:                 uint32(targetSpID),
//		GlobalVirtualGroupFamilyId: uint32(gvgfID),
//		GlobalVirtualGroupId:       uint32(gvgID),
//		StorageProvider:            cfg.SpAccount.SpOperatorAddress,
//	}
//
//	spClient := utils.MakeGfSpClient(cfg)
//	_, err = spClient.ReserveSwapIn(ctx.Context, reserveSwapIn)
//	return err
//}
//
////func CompleteSwapInAction(ctx *cli.Context) error {
////	cfg, err := utils.MakeConfig(ctx)
////	if err != nil {
////		return err
////	}
////
////	gvgID := ctx.Uint64(gvgIDFlag.Name)
////	gvgfID := ctx.Uint64(gvgIDFlag.Name)
////
////	completeSwapIn := &virtualgrouptypes.MsgCompleteSwapIn{
////		GlobalVirtualGroupFamilyId: uint32(gvgfID),
////		GlobalVirtualGroupId:       uint32(gvgID),
////		StorageProvider:            cfg.SpAccount.SpOperatorAddress,
////	}
////
////	spClient := utils.MakeGfSpClient(cfg)
////	_, err = spClient.(ctx.Context, completeSwapIn)
////	return err
////}
//
//func RecoverGVGAction(ctx *cli.Context) error {
//	cfg, err := utils.MakeConfig(ctx)
//	if err != nil {
//		return err
//	}
//
//	// get client
//	chainClient, err := utils.MakeGnfd(cfg)
//	if err != nil {
//		return err
//	}
//	spClient := utils.MakeGfSpClient(cfg)
//
//	// check swapIn info
//	sp, err := chainClient.QuerySP(ctx.Context, cfg.SpAccount.SpOperatorAddress)
//	if err != nil {
//		return err
//	}
//	gvgID := ctx.Uint64(gvgIDFlag.Name)
//	swapInInfo, err := chainClient.QuerySwapInInfo(ctx.Context, 0, uint32(gvgID))
//	if err != nil {
//		return err
//	}
//	if swapInInfo.GetSuccessorSpId() != sp.GetId() {
//		return errors.New("sp is not successor sp")
//	}
//
//	// trigger
//	//return spClient.TriggerRecoverForSuccessorSP(ctx.Context, 0, uint32(gvgID), swapInInfo.)
//}
//
//func RecoverVGFAction(ctx *cli.Context) error {
//	cfg, err := utils.MakeConfig(ctx)
//	if err != nil {
//		return err
//	}
//
//	// get client
//	chainClient, err := utils.MakeGnfd(cfg)
//	if err != nil {
//		return err
//	}
//	spClient := utils.MakeGfSpClient(cfg)
//
//	// check swapIn info
//	sp, err := chainClient.QuerySP(ctx.Context, cfg.SpAccount.SpOperatorAddress)
//	if err != nil {
//		return err
//	}
//	vgfID := ctx.Uint64(vgfIDFlag.Name)
//	swapInInfo, err := chainClient.QuerySwapInInfo(ctx.Context, uint32(vgfID), 0)
//	if err != nil {
//		return err
//	}
//	if swapInInfo.GetSuccessorSpId() != sp.GetId() {
//		return errors.New("sp is not successor sp")
//	}
//
//	// trigger
//	//return spClient.TriggerRecoverForSuccessorSP(ctx.Context, uint32(vgfID), 0, swapInInfo.TargetSpId)
//}
