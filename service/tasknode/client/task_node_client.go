package client

import (
	"context"

	servicetype "github.com/bnb-chain/greenfield-storage-provider/service/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/tasknode/types"
)

// TaskNodeClient is a task node gRPC service client wrapper
type TaskNodeClient struct {
	address  string
	conn     *grpc.ClientConn
	taskNode types.TaskNodeServiceClient
}

// NewTaskNodeClient return a TaskNodeClient instance
func NewTaskNodeClient(address string) (*TaskNodeClient, error) {
	conn, err := grpc.DialContext(context.Background(), address,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
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
func (client *TaskNodeClient) QueryReplicatingObject(ctx context.Context, objectID uint64) (*servicetype.ReplicateSegmentInfo, error) {
	resp, err := client.taskNode.QueryReplicatingObject(ctx, &types.QueryReplicatingObjectRequest{ObjectId: objectID})
	if err != nil {
		return nil, err
	}
	return resp.GetReplicateSegmentInfo(), err
}
