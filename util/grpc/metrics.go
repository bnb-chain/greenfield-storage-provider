package grpc

import (
	openmetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

// GetDefaultServerInterceptor returns default gRPC server interceptor
func GetDefaultServerInterceptor() []grpc.ServerOption {
	options := []grpc.ServerOption{}
	options = append(options, grpc.ChainUnaryInterceptor(openmetrics.UnaryServerInterceptor(metrics.DefaultGRPCServerMetrics)))
	options = append(options, grpc.ChainStreamInterceptor(openmetrics.StreamServerInterceptor(metrics.DefaultGRPCServerMetrics)))
	return options
}

// GetDefaultClientInterceptor returns default gRPC client interceptor
func GetDefaultClientInterceptor() []grpc.DialOption {
	options := []grpc.DialOption{}
	options = append(options, grpc.WithChainUnaryInterceptor(openmetrics.UnaryClientInterceptor(metrics.DefaultGRPCClientMetrics)))
	options = append(options, grpc.WithChainStreamInterceptor(openmetrics.StreamClientInterceptor(metrics.DefaultGRPCClientMetrics)))
	return options
}
