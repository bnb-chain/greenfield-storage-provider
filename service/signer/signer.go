package signer

import (
	"context"
	"net"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// SignerServer signer service
type SignerServer struct {
	config          *SignerConfig
	greenfieldChain *GreenfieldChain
}

// NewSignerServer return SignerServer instance
func NewSignerServer(config *SignerConfig) (*SignerServer, error) {
	return &SignerServer{
		greenfieldChain: NewGreenfieldChain(config.GreenfieldChainConfig),
	}, nil
}

// Name describe service name
func (signer *SignerServer) Name() string {
	return model.SignerService
}

// Start a service, this method should be used in non-block form
func (signer *SignerServer) Start(ctx context.Context) error {
	// start background task
	signer.greenfieldChain.wg.Add(1)
	go signer.greenfieldChain.updateClientLoop()

	// start rpc service
	go signer.serve()
	return nil
}

// Stop a service, this method should be used in non-block form
func (signer *SignerServer) Stop(ctx context.Context) error {
	close(signer.greenfieldChain.stopCh)
	signer.greenfieldChain.wg.Wait()
	return nil
}

// Serve starts grpc signer service.
func (signer *SignerServer) serve() {
	lis, err := net.Listen("tcp", signer.config.Address)
	if err != nil {
		log.Errorw("failed to listen", "address", signer.config.Address, "error", err)
		return
	}
	grpcServer := grpc.NewServer()
	stypes.RegisterSignerServiceServer(grpcServer, signer)
	// register reflection service
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Errorf("grpc serve error : %v", err)
		return
	}
}
