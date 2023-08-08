package command

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

var freeQuotaFlag = &cli.Uint64Flag{
	Name:     "quota",
	Usage:    "The sp free quota",
	Required: true,
}

var SetQuotaCmd = &cli.Command{
	Action: updateFreeQuotaAction,
	Name:   "update.quota",
	Usage:  "Update the free quota of the SP",

	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		freeQuotaFlag,
	},

	Category: "QUOTA COMMANDS",
	Description: `The update.quota command is used to update the free quota of the SP on greenfield chain, 
				it will send a txn to the chain to finish the updating ", 
  `,
}

func updateFreeQuotaAction(ctx *cli.Context) error {
	cfg, err := utils.MakeConfig(ctx)
	if err != nil {
		return err
	}

	freeQuota := ctx.Uint64(freeQuotaFlag.Name)
	chain, err := utils.MakeGnfd(cfg)
	if err != nil {
		return err
	}

	localCtx := context.Background()
	operatorAddr := cfg.SpAccount.SpOperatorAddress
	oldQuota, err := chain.QuerySPFreeQuota(localCtx, operatorAddr)
	if err != nil {
		return err
	}

	if oldQuota == freeQuota {
		return fmt.Errorf("quota value: %d is same as chain meta, no need to update \n", freeQuota)
	}

	priceInfo, err := chain.QuerySPPrice(localCtx, operatorAddr)
	if err != nil {
		fmt.Printf("failed to get the current quota info %s\n", err.Error())
		return err
	}
	msgUpdateStoragePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     operatorAddr,
		ReadPrice:     priceInfo.ReadPrice,
		StorePrice:    priceInfo.StorePrice,
		FreeReadQuota: freeQuota,
	}

	spClient := utils.MakeGfSpClient(cfg)
	txnHash, err := spClient.UpdateSPPrice(localCtx, msgUpdateStoragePrice)
	if err != nil {
		return err
	}
	fmt.Printf("update sp free quota from %d to %d successfully, txn hash is %s \n", oldQuota, freeQuota, txnHash)

	return nil
}
