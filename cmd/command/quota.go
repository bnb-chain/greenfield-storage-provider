package command

import (
	"context"
	"fmt"

	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	"github.com/urfave/cli/v2"
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
	oldQuota, err := chain.QuerySPFreeQuota(localCtx, cfg.SpAccount.SpOperatorAddress)
	if err != nil {
		return err
	}

	if oldQuota == freeQuota {
		return fmt.Errorf("quota value: %d is same as chain meta, no need to update \n", freeQuota)
	}

	txnHash, err := chain.UpdateSPQuota(localCtx, cfg.SpAccount.SpOperatorAddress, freeQuota)
	if err != nil {
		return err
	}
	fmt.Printf("update sp free quota from %d to %d successfully, txn hash is %s \n", oldQuota, freeQuota, txnHash)

	return nil
}
