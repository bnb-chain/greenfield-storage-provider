package gfspapp

import (
	"context"
	"net"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

const (
	MaxCallMsgSize = 32 * 1024 * 1024
)

func DefaultGrpcServerOptions() []grpc.ServerOption {
	var options []grpc.ServerOption
	options = append(options, grpc.MaxRecvMsgSize(MaxCallMsgSize))
	options = append(options, grpc.MaxSendMsgSize(MaxCallMsgSize))
	return options
}

func (g *GfSpBaseApp) newRpcServer(options []grpc.ServerOption) {
	options = append(options, DefaultGrpcServerOptions()...)
	g.server = grpc.NewServer(options...)
}

func (g *GfSpBaseApp) StartRpcServer(ctx context.Context) error {
	lis, err := net.Listen("tcp", g.grpcAddress)
	if err != nil {
		return err
	}
	go func() {
		if err = g.server.Serve(lis); err != nil {
			log.Errorw("failed to start gfsp app grpc server", "error", err)
		}
	}()
	return nil
}

func (g *GfSpBaseApp) StopRpcServer(ctx context.Context) error {
	g.server.GracefulStop()
	return nil
}

func RpcRemoteAddress(ctx context.Context) string {
	var addr string
	if pr, ok := peer.FromContext(ctx); ok {
		if tcpAddr, ok := pr.Addr.(*net.TCPAddr); ok {
			addr = tcpAddr.IP.String()
		} else {
			addr = pr.Addr.String()
		}
	}
	return addr
}
