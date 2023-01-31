package stonehub

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/urfave/cli"

	stypesv1pb "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	cliCtx "github.com/bnb-chain/greenfield-storage-provider/test/test_tool/context"
)

var QueryStoneCommand = cli.Command{
	Name:   "query_stone",
	Usage:  "Query the stone from stone hub",
	Action: queryStone,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "t,TxHash",
			Value: "",
			Usage: "Transaction hash of create object"},
	},
}

func queryStone(c *cli.Context) {
	ctx := cliCtx.GetContext()
	if ctx.CurrentService != cliCtx.StoneHubService {
		fmt.Println("please cd StoneHubService namespace, try again")
		return
	}
	txHash, err := hex.DecodeString(c.String("t"))
	if err != nil {
		fmt.Println("tx hash param decode error: ", err)
		return
	}
	req := &stypesv1pb.StoneHubServiceQueryStoneRequest{
		TxHash: txHash,
	}
	client, err := GetStoneHubClient()
	if err != nil {
		return
	}
	rpcCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rsp, err := client.QueryStone(rpcCtx, req)
	if err != nil {
		fmt.Println("send create object rpc error:", err)
		return
	}
	if rsp.ObjectInfo != nil {
		fmt.Println("object info: ", rsp.ObjectInfo)
	}
	if rsp.ErrMessage != nil {
		fmt.Println(rsp.ErrMessage)
		return
	}
	fmt.Println("job info: ", rsp.JobInfo)
	if rsp.PendingPrimaryJob != nil && len(rsp.PendingPrimaryJob.TargetIdx) > 0 {
		fmt.Println("pending primary piece job: ", rsp.PendingPrimaryJob.TargetIdx)
	} else {
		fmt.Println("primary piece jobs are completed.")
	}
	if rsp.PendingSecondaryJob != nil && len(rsp.PendingSecondaryJob.TargetIdx) > 0 {
		fmt.Println("pending secondary piece job: ", rsp.PendingSecondaryJob.TargetIdx)
	} else {
		fmt.Println("secondary piece jobs are completed.")
	}
	return
}
