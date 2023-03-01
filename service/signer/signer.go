package signer

import (
	"context"
	"net"

	"github.com/cloudflare/cfssl/whitelist"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var _ lifecycle.Service = &SignerServer{}

// SignerServer signer service
type SignerServer struct {
	config    *SignerConfig
	whitelist *whitelist.BasicNet
	client    *client.GreenfieldChainSignClient

	server *grpc.Server
}

// NewSignerServer return SignerServer instance
func NewSignerServer(config *SignerConfig) (*SignerServer, error) {
	overrideConfigFromEnv(config)

	whitelist := whitelist.NewBasicNet()
	for _, cidr := range config.WhitelistCIDR {
		_, subnet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		whitelist.Add(subnet)
	}

	client, err := client.NewGreenfieldChainSignClient(
		config.GreenfieldChainConfig.GRPCAddr,
		config.GreenfieldChainConfig.ChainID,
		config.GreenfieldChainConfig.GasLimit,
		config.GreenfieldChainConfig.OperatorPrivateKey,
		config.GreenfieldChainConfig.FundingPrivateKey,
		config.GreenfieldChainConfig.SealPrivateKey,
		config.GreenfieldChainConfig.ApprovalPrivateKey)
	if err != nil {
		return nil, err
	}
	return &SignerServer{
		config:    config,
		client:    client,
		whitelist: whitelist,
	}, nil
}

// Name describe service name
func (signer *SignerServer) Name() string {
	return model.SignerService
}

// Start a service, this method should be used in non-block form
func (signer *SignerServer) Start(ctx context.Context) error {
	// start rpc service
	go signer.serve()
	return nil
}

// Stop a service, this method should be used in non-block form
func (signer *SignerServer) Stop(ctx context.Context) error {
	// stop rpc service
	signer.server.Stop()
	return nil
}

// Serve starts grpc signer service.
func (signer *SignerServer) serve() {
	lis, err := net.Listen("tcp", signer.config.Address)
	if err != nil {
		log.Errorw("failed to listen", "address", signer.config.Address, "error", err)
		return
	}
	signer.server = grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			signer.IPWhitelistInterceptor(),
			signer.AuthInterceptor(),
		)),
	)

	types.RegisterSignerServiceServer(signer.server, signer)
	// register reflection service
	reflection.Register(signer.server)
	if err := signer.server.Serve(lis); err != nil {
		log.Errorf("grpc serve error : %v", err)
		return
	}
}
