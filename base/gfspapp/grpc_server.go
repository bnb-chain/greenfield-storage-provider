package gfspapp

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/base/types/gfspserver"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	utilgrpc "github.com/bnb-chain/greenfield-storage-provider/util/grpc"
)

const (
	// MaxServerCallMsgSize defines the max message size for grpc server
	MaxServerCallMsgSize = 3 * 1024 * 1024 * 1024
)

func DefaultGrpcServerOptions() []grpc.ServerOption {
	var options []grpc.ServerOption
	options = append(options, grpc.MaxRecvMsgSize(MaxServerCallMsgSize))
	options = append(options, grpc.MaxSendMsgSize(MaxServerCallMsgSize))
	return options
}

func (g *GfSpBaseApp) newRpcServer(options ...grpc.ServerOption) {
	options = append(options, DefaultGrpcServerOptions()...)
	if g.EnableMetrics() {
		options = append(options, utilgrpc.GetDefaultServerInterceptor()...)
	}
	g.server = grpc.NewServer(options...)
	gfspserver.RegisterGfSpApprovalServiceServer(g.server, g)
	gfspserver.RegisterGfSpAuthenticationServiceServer(g.server, g)
	gfspserver.RegisterGfSpDownloadServiceServer(g.server, g)
	gfspserver.RegisterGfSpManageServiceServer(g.server, g)
	gfspserver.RegisterGfSpP2PServiceServer(g.server, g)
	gfspserver.RegisterGfSpResourceServiceServer(g.server, g)
	gfspserver.RegisterGfSpReceiveServiceServer(g.server, g)
	gfspserver.RegisterGfSpSignServiceServer(g.server, g)
	gfspserver.RegisterGfSpUploadServiceServer(g.server, g)
	gfspserver.RegisterGfSpQueryTaskServiceServer(g.server, g)
	reflection.Register(g.server)
}

func (g *GfSpBaseApp) StartRpcServer(ctx context.Context) error {
	lis, err := net.Listen("tcp", g.grpcAddress)
	if err != nil {
		log.Errorw("failed to listen tcp address", "address", g.grpcAddress, "error", err)
		return err
	}
	go func() {
		if err = g.server.Serve(lis); err != nil {
			log.Errorw("failed to start gf-sp app grpc server", "error", err)
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
