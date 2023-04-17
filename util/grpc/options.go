package grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield-storage-provider/model"
)

// GetDefaultServerOptions returns default gRPC server options
func GetDefaultServerOptions() []grpc.ServerOption {
	options := []grpc.ServerOption{}
	options = append(options, grpc.MaxRecvMsgSize(model.MaxCallMsgSize))
	options = append(options, grpc.MaxSendMsgSize(model.MaxCallMsgSize))
	return options
}

// GetDefaultClientOptions returns default gRPC client options
func GetDefaultClientOptions() []grpc.DialOption {
	options := []grpc.DialOption{}
	options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	options = append(options, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(model.MaxCallMsgSize)))
	options = append(options, grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(model.MaxCallMsgSize)))
	return options
}
