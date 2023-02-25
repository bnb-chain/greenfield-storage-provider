package conf

import (
	"context"
	"fmt"

	gtypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/urfave/cli"

	"github.com/bnb-chain/greenfield-storage-provider/config"
	"github.com/bnb-chain/greenfield-storage-provider/service/p2p/client"
)

var DumpConfigCommand = cli.Command{
	Name:   "dump.config",
	Usage:  "dump default config",
	Action: DumpDefaultConfig,
}

func DumpDefaultConfig(c *cli.Context) {
	err := config.SaveConfig("./", config.DefaultStorageProviderConfig)
	if err != nil {
		fmt.Println(err)
	}
}

var AskCommand = cli.Command{
	Name:   "ask_approval",
	Usage:  "dump default config",
	Action: AskCommandAction,
}

func AskCommandAction(c *cli.Context) {
	client, err := client.NewP2PServiceRpcClient("localhost:9933")
	if err != nil {
		fmt.Println(err)
	}
	approvals, err := client.AskSecondaryApproval(context.Background(), &gtypes.MsgCreateObject{}, 6, 9)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(approvals)
}
