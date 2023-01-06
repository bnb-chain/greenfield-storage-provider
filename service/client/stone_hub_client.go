package client

import (
	"context"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/inscription-storage-provider/util"

	"github.com/bnb-chain/inscription-storage-provider/util/log"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
)

var ClientRpcTimeout = time.Second * 5

var _ io.Closer = &StoneHubClient{}

type StoneHubClient struct {
	address  string
	stoneHub service.StoneHubServiceClient
	conn     *grpc.ClientConn
}

func NewStoneHubClient(address string) (*StoneHubClient, error) {
	ctx, _ := context.WithTimeout(context.Background(), ClientRpcTimeout)
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorw("invoke stoneHub service grpc.DialContext failed", "error", err)
		return nil, err
	}
	client := &StoneHubClient{
		address:  address,
		conn:     conn,
		stoneHub: service.NewStoneHubServiceClient(conn),
	}
	return client, nil
}

func (client *StoneHubClient) AllocStoneJob(ctx context.Context, opts ...grpc.CallOption) (*service.StoneHubServiceAllocStoneJobResponse, error) {
	req := &service.StoneHubServiceAllocStoneJobRequest{TraceId: util.GenerateRequestID()}
	return client.stoneHub.AllocStoneJob(ctx, req, opts...)
}

func (client *StoneHubClient) DoneSecondaryPieceJob(ctx context.Context, in *service.StoneHubServiceDoneSecondaryPieceJobRequest, opts ...grpc.CallOption) (*service.StoneHubServiceDoneSecondaryPieceJobResponse, error) {
	return client.stoneHub.DoneSecondaryPieceJob(ctx, in, opts...)
}

func (client *StoneHubClient) Close() error {
	return client.conn.Close()
}
