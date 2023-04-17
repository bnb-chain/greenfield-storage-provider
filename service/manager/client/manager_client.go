package client

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	tqueuetypes "github.com/bnb-chain/greenfield-storage-provider/pkg/taskqueue/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/service/manager/types"
	utilgrpc "github.com/bnb-chain/greenfield-storage-provider/util/grpc"
)

// ManagerClient is a challenge gRPC service client wrapper
type ManagerClient struct {
	address string
	manager types.ManagerServiceClient
	conn    *grpc.ClientConn
}

const managerRPCServiceName = "service.manager.types.ManagerService"

// NewManagerClient return a manager client instance
func NewManagerClient(address string) (*ManagerClient, error) {
	options := utilgrpc.GetDefaultClientOptions()
	retryOption, err := utilgrpc.GetDefaultGRPCRetryPolicy(managerRPCServiceName)
	if err != nil {
		log.Errorw("failed to get manager client retry option", "error", err)
		return nil, err
	}
	options = append(options, retryOption)
	if metrics.GetMetrics().Enabled() {
		options = append(options, utilgrpc.GetDefaultClientInterceptor()...)
	}
	conn, err := grpc.DialContext(context.Background(), address, options...)
	if err != nil {
		log.Errorw("failed to dial manager", "error", err)
		return nil, err
	}
	client := &ManagerClient{
		address: address,
		conn:    conn,
		manager: types.NewManagerServiceClient(conn),
	}
	return client, nil
}

// Close the manager gPRC connection
func (client *ManagerClient) Close() error {
	return client.conn.Close()
}

// AskUploadObject asks to create object to SP manager.
func (client *ManagerClient) AskUploadObject(ctx context.Context,
	object *storagetypes.ObjectInfo, opts ...grpc.CallOption) (bool, error) {
	task, err := tqueuetypes.NewUploadObjectTask(object)
	if err != nil {
		return false, err
	}
	req := &types.AskUploadObjectRequest{
		UploadObjectTask: task,
	}
	resp, err := client.manager.AskUploadObject(ctx, req, opts...)
	if err != nil {
		return false, err
	}
	return resp.GetAllow(), err
}

// CreateUploadObjectTask asks to upload object to SP manager.
func (client *ManagerClient) CreateUploadObjectTask(ctx context.Context,
	object *storagetypes.ObjectInfo, opts ...grpc.CallOption) error {
	task, err := tqueuetypes.NewUploadObjectTask(object)
	if err != nil {
		return err
	}
	req := &types.CreateUploadObjectTaskRequest{
		UploadObjectTask: task,
	}
	_, err = client.manager.CreateUploadObjectTask(ctx, req, opts...)
	return err
}

// DoneUploadObjectTask notifies the manager the upload object task has been done.
func (client *ManagerClient) DoneUploadObjectTask(ctx context.Context,
	object *storagetypes.ObjectInfo, opts ...grpc.CallOption) error {
	task, err := tqueuetypes.NewUploadObjectTask(object)
	if err != nil {
		return err
	}
	req := &types.DoneUploadObjectTaskRequest{
		UploadObjectTask: task,
	}
	_, err = client.manager.DoneUploadObjectTask(ctx, req, opts...)
	return err
}

// DoneReplicatePieceTask notifies the manager the replicate piece task has been done.
func (client *ManagerClient) DoneReplicatePieceTask(ctx context.Context,
	task *tqueuetypes.ReplicatePieceTask, opts ...grpc.CallOption) error {
	req := &types.DoneReplicatePieceTaskRequest{
		ReplicatePieceTask: task,
	}
	_, err := client.manager.DoneReplicatePieceTask(ctx, req, opts...)
	return err
}

// DoneSealObjectTask notifies the manager the seal object task has been done.
func (client *ManagerClient) DoneSealObjectTask(ctx context.Context,
	task *tqueuetypes.SealObjectTask, opts ...grpc.CallOption) error {
	req := &types.DoneSealObjectTaskRequest{
		SealObjectTask: task,
	}
	_, err := client.manager.DoneSealObjectTask(ctx, req, opts...)
	return err
}

// AskTask asks the task to execute
func (client *ManagerClient) AskTask(ctx context.Context,
	rclimit rcmgr.Limit, opts ...grpc.CallOption) (*types.AskTaskResponse, error) {
	req := &types.AskTaskRequest{
		Limit: types.NewLimits(rclimit),
	}
	resp, err := client.manager.AskTask(ctx, req, opts...)
	return resp, err
}

// DoneGCObjectTask notifies the manager the gc object task has been done.
func (client *ManagerClient) DoneGCObjectTask(ctx context.Context,
	task *tqueuetypes.GCObjectTask, opts ...grpc.CallOption) error {
	req := &types.DoneGCObjectTaskRequest{
		GcObjectTask: task,
	}
	_, err := client.manager.DoneGCObjectTask(ctx, req, opts...)
	return err
}

// ReportGCObjectProcess notifies the manager the gc object task process.
func (client *ManagerClient) ReportGCObjectProcess(ctx context.Context,
	task *tqueuetypes.GCObjectTask, opts ...grpc.CallOption) (bool, error) {
	req := &types.ReportGCObjectProcessRequest{
		GcObjectTask: task,
	}
	resp, err := client.manager.ReportGCObjectProcess(ctx, req, opts...)
	if err != nil {
		return false, err
	}
	return resp.GetCancel(), nil
}
