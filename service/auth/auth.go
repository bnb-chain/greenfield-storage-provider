package auth

/*
import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/core/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	authtypes "github.com/bnb-chain/greenfield-storage-provider/service/auth/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
)

var _ lifecycle.Service = &AuthServer{}

// AuthServer auth service
type AuthServer struct {
	config     *AuthConfig
	spDB       sqldb.SPDB
	grpcServer *grpc.Server
}

// NewAuthServer return an instance of AuthServer
func NewAuthServer(config *AuthConfig) (*AuthServer, error) {
	spDB, err := sqldb.NewSpDB(config.SpDBConfig)
	if err != nil {
		return nil, err
	}
	p := &AuthServer{
		config: config,
		spDB:   spDB,
	}
	return p, nil
}

// Name return the auth server name, for the lifecycle management
func (auth *AuthServer) Name() string {
	return model.AuthService
}

// Start the auth gRPC service
func (auth *AuthServer) Start(ctx context.Context) error {
	errCh := make(chan error)
	go auth.serve(errCh)
	err := <-errCh
	return err
}

// Stop the auth gRPC service and recycle the resources
func (auth *AuthServer) Stop(ctx context.Context) error {
	auth.grpcServer.GracefulStop()
	return nil
}

// Serve starts grpc service.
func (auth *AuthServer) serve(errCh chan error) {
	lis, err := net.Listen("tcp", auth.config.GRPCAddress)
	errCh <- err
	if err != nil {
		log.Errorw("failed to listen", "err", err)
		return
	}

	grpcServer := grpc.NewServer()
	authtypes.RegisterAuthServiceServer(grpcServer, auth)
	auth.grpcServer = grpcServer
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Errorw("failed to start grpc server", "err", err)
		return
	}
}

*/
