package stonehub

import (
	"fmt"

	"google.golang.org/grpc"

	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	cliCtx "github.com/bnb-chain/greenfield-storage-provider/test/test_tool/context"
)

func GetStoneHubClient() (stypes.StoneHubServiceClient, error) {
	ctx := cliCtx.GetContext()
	conn, err := grpc.Dial(ctx.Cfg.StoneHubAddr)
	if err != nil {
		fmt.Println("dial stone hub error: ", err)
		return nil, err
	}
	client := stypes.NewStoneHubServiceClient(conn)
	return client, nil
}
