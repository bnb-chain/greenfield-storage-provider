package client

import (
	"context"

	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	mwgrpc "github.com/bnb-chain/greenfield-storage-provider/pkg/middleware/grpc"
	"github.com/bnb-chain/greenfield-storage-provider/service/tasknode/types"
	servicetype "github.com/bnb-chain/greenfield-storage-provider/service/types"
	utilgrpc "github.com/bnb-chain/greenfield-storage-provider/util/grpc"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// TaskNodeClient is a task node gRPC service client wrapper
type TaskNodeClient struct {
	address  string
	conn     *grpc.ClientConn
	taskNode types.TaskNodeServiceClient
}

// NewTaskNodeClient return a TaskNodeClient instance
func NewTaskNodeClient(address string) (*TaskNodeClient, error) {
	options := utilgrpc.GetDefaultClientOptions()
	if metrics.GetMetrics().Enabled() {
		options = append(options, mwgrpc.GetDefaultClientInterceptor()...)
	}
	conn, err := grpc.DialContext(context.Background(), address, options...)
	if err != nil {
		log.Errorw("failed to dial task node", "error", err)
		return nil, err
	}
	client := &TaskNodeClient{
		address:  address,
		conn:     conn,
		taskNode: types.NewTaskNodeServiceClient(conn),
	}
	return client, nil
}

// Close the task node gPRC connection
func (client *TaskNodeClient) Close() error {
	return client.conn.Close()
}

// ReplicateObject async replicate an object payload to other storage provider and seal object
func (client *TaskNodeClient) ReplicateObject(ctx context.Context, object *storagetypes.ObjectInfo, opts ...grpc.CallOption) error {
	_, err := client.taskNode.ReplicateObject(ctx, &types.ReplicateObjectRequest{ObjectInfo: object}, opts...)
	return err
}

// QueryReplicatingObject query a replicating object payload information by object id
func (client *TaskNodeClient) QueryReplicatingObject(ctx context.Context, objectID uint64) (*servicetype.ReplicatePieceInfo, error) {
	resp, err := client.taskNode.QueryReplicatingObject(ctx, &types.QueryReplicatingObjectRequest{ObjectId: objectID})
	if err != nil {
		return nil, err
	}
	return resp.GetReplicatePieceInfo(), err
}
