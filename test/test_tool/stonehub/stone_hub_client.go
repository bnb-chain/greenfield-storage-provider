package stonehub

import (
	"fmt"

	"google.golang.org/grpc"

	cliCtx "github.com/bnb-chain/greenfield-storage-provider/test/test_tool/context"

	service "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
)

func GetStoneHubClient() (service.StoneHubServiceClient, error) {
	ctx := cliCtx.GetContext()
	conn, err := grpc.Dial(ctx.Cfg.StoneHubAddr, grpc.WithInsecure())
	if err != nil {
		fmt.Println("dial stone hub error: ", err)
		return nil, err
	}
	client := service.NewStoneHubServiceClient(conn)
	return client, nil
}
