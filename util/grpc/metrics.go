package grpc

import (
	"runtime/debug"

	openmetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

var (
	gRPCPanicRecoveryHandler = func(p interface{}) (err error) {
		metrics.PanicsTotal.WithLabelValues().Inc()
		log.Errorw("recovered from panic", "panic", p, "stack", debug.Stack())
		return status.Errorf(codes.Internal, "%s", p)
	}
)

// GetDefaultServerInterceptor returns default gRPC server interceptor
func GetDefaultServerInterceptor() []grpc.ServerOption {
	options := []grpc.ServerOption{}
	options = append(options, grpc.ChainUnaryInterceptor(openmetrics.UnaryServerInterceptor(metrics.DefaultGRPCServerMetrics),
		grpcrecovery.UnaryServerInterceptor(grpcrecovery.WithRecoveryHandler(gRPCPanicRecoveryHandler))))
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
