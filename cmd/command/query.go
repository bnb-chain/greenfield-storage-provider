package command

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspclient"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/cmd/utils"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var endpointFlag = &cli.StringFlag{
	Name:  "n",
	Usage: "The address of machine that to query tasks",
	Value: "",
}

var keyFlag = &cli.StringFlag{
	Name:  "k",
	Usage: "The sub key of task",
	Value: "",
}

var QueryTaskCmd = &cli.Command{
	Action:   queryTasksAction,
	Name:     "query.task",
	Usage:    "Query running tasks in modules by task sub key",
	Category: "Query COMMANDS",
	Flags: []cli.Flag{
		utils.ConfigFileFlag,
		endpointFlag,
		keyFlag,
	},
	Description: `Query running tasks in modules by task sub key, 
show the tasks that task key contains the inout key detail info`,
}

func queryTasksAction(ctx *cli.Context) error {
	endpoint := gfspapp.DefaultGRPCAddress
	if ctx.IsSet(utils.ConfigFileFlag.Name) {
		cfg := &gfspconfig.GfSpConfig{}
		err := utils.LoadConfig(ctx.String(utils.ConfigFileFlag.Name), cfg)
		if err != nil {
			log.Errorw("failed to load config file", "error", err)
			return err
		}
		endpoint = cfg.GRPCAddress
	}
	if ctx.IsSet(endpointFlag.Name) {
		endpoint = ctx.String(endpointFlag.Name)
	}
	if !ctx.IsSet(keyFlag.Name) {
		return fmt.Errorf("query key should be set")
	}
	key := ctx.String(keyFlag.Name)
	if len(key) == 0 {
		return fmt.Errorf("query key can not empty")
	}
	client := &gfspclient.GfSpClient{}
	infos, err := client.QueryTasks(context.Background(), endpoint, key)
	if err != nil {
		return err
	}
	if len(infos) == 0 {
		return fmt.Errorf("no task match the query key")
	}
	for _, info := range infos {
		fmt.Printf(info + "\n")
	}
	return nil
}
