package client

import (
	"context"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	servicetypes "github.com/bnb-chain/greenfield-storage-provider/service/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/receiver/types"
)

// ReceiverClient is a receiver gRPC service client wrapper
type ReceiverClient struct {
	address  string
	conn     *grpc.ClientConn
	receiver types.ReceiverServiceClient
}

// NewReceiverClient return a ReceiverClient instance
func NewReceiverClient(address string) (*ReceiverClient, error) {
	conn, err := grpc.DialContext(context.Background(), address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(model.MaxCallMsgSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(model.MaxCallMsgSize)))
	if err != nil {
		log.Errorw("failed to dial receiver", "error", err)
		return nil, err
	}
	client := &ReceiverClient{
		address:  address,
		conn:     conn,
		receiver: types.NewReceiverServiceClient(conn),
	}
	return client, nil
}

// Close the receiver gPRC connection
func (client *ReceiverClient) Close() error {
	return client.conn.Close()
}

// SyncObject an object payload with object info
func (client *ReceiverClient) SyncObject(
	ctx context.Context,
	opts ...grpc.CallOption) (types.ReceiverService_SyncObjectClient, error) {
	return client.receiver.SyncObject(ctx, opts...)
}

// QuerySyncingObject a syncing object info by object id
func (client *ReceiverClient) QuerySyncingObject(ctx context.Context, objectID sdkmath.Uint) (*servicetypes.SegmentInfo, error) {
	req := &types.QuerySyncingObjectRequest{
		ObjectId: objectID,
	}
	resp, err := client.receiver.QuerySyncingObject(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetSegmentInfo(), nil
}
