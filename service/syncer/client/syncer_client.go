package client

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/syncer/types"
)

// SyncerClient is a grpc client wrapper.
type SyncerClient struct {
	address string
	syncer  types.SyncerServiceClient
	conn    *grpc.ClientConn
}

// NewSyncerClient return a SyncerClient instance.
func NewSyncerClient(address string) (*SyncerClient, error) {
	conn, err := grpc.DialContext(context.Background(), address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(model.MaxCallMsgSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(model.MaxCallMsgSize)))
	if err != nil {
		log.Errorw("failed to dail syncer", "error", err)
		return nil, err
	}
	client := &SyncerClient{
		address: address,
		conn:    conn,
		syncer:  types.NewSyncerServiceClient(conn),
	}
	return client, nil
}

// SyncObject an object payload with object info.
func (client *SyncerClient) SyncObject(
	ctx context.Context,
	opts ...grpc.CallOption) (types.SyncerService_SyncObjectClient, error) {
	return client.syncer.SyncObject(ctx, opts...)
}

// QuerySyncingObject a syncing object info by object id.
func (client *SyncerClient) QuerySyncingObject(
	ctx context.Context, objectId uint64) (*servicetypes.SegmentInfo, error) {
	req := &types.QuerySyncingObjectRequest{
		ObjectId: objectId,
	}
	resp, err := client.syncer.QuerySyncingObject(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetSegmentInfo(), nil
}
