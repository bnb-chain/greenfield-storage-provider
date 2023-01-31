package stonehub

import (
	"fmt"

	"google.golang.org/grpc"

	stypesv1pb "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	cliCtx "github.com/bnb-chain/greenfield-storage-provider/test/test_tool/context"
)

func GetStoneHubClient() (stypesv1pb.StoneHubServiceClient, error) {
	ctx := cliCtx.GetContext()
	conn, err := grpc.Dial(ctx.Cfg.StoneHubAddr, grpc.WithInsecure())
	if err != nil {
		fmt.Println("dial stone hub error: ", err)
		return nil, err
	}
	client := stypesv1pb.NewStoneHubServiceClient(conn)
	return client, nil
}
