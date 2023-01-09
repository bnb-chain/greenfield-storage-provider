package stonehub

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/urfave/cli"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	cliCtx "github.com/bnb-chain/inscription-storage-provider/test/test_tool/context"
)

var BeginUploadPayloadCommand = cli.Command{
	Name:   "begin_stone",
	Usage:  "Begin upload payload data",
	Action: beginStone,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "t,TxHash",
			Value: "",
			Usage: "Transaction hash of create object"},
	},
}

func beginStone(c *cli.Context) {
	ctx := cliCtx.GetContext()
	if ctx.CurrentService != cliCtx.StoneHubService {
		fmt.Println("please cd StoneHubService namespace, try again")
		return
	}
	txHash, err := hex.DecodeString(c.String("t"))
	if err != nil {
		fmt.Println("tx hash param decode error: ", err)
	}
	req := &service.StoneHubServiceBeginUploadPayloadRequest{
		TxHash: txHash,
	}
	client, err := GetStoneHubClient()
	if err != nil {
		return
	}
	rpcCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rsp, err := client.BeginUploadPayload(rpcCtx, req)
	if err != nil {
		fmt.Println("send create object rpc error:", err)
		return
	}
	if rsp.ErrMessage != nil {
		fmt.Println(rsp.ErrMessage)
		return
	}
	fmt.Println("begin upload payload data success")
	return
}
