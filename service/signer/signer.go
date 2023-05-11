package signer

import (
	"context"
	"errors"
	"net"

	"github.com/cloudflare/cfssl/whitelist"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer/client"
	"github.com/bnb-chain/greenfield-storage-provider/service/signer/types"
	utilgrpc "github.com/bnb-chain/greenfield-storage-provider/util/grpc"
)

var _ lifecycle.Service = &SignerServer{}

// SignerServer signer service
type SignerServer struct {
	config       *SignerConfig
	chainConfig  *gnfd.GreenfieldChainConfig
	svcWhitelist *whitelist.BasicNet
	client       *client.GreenfieldChainSignClient

	server *grpc.Server
}

// NewSignerServer return SignerServer instance
func NewSignerServer(config *SignerConfig, chainConfig *gnfd.GreenfieldChainConfig) (*SignerServer, error) {
	overrideConfigFromEnv(config)

	svcWhitelist := whitelist.NewBasicNet()
	for _, cidr := range config.WhitelistCIDR {
		_, subnet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		svcWhitelist.Add(subnet)
	}
	if len(chainConfig.NodeAddr) == 0 {
		return nil, errors.New("greenfield nodes missing")
	}

	if len(chainConfig.NodeAddr[0].GreenfieldAddresses) == 0 {
		return nil, errors.New("greenfield endpoints missing")
	}

	log.Infow("print signer config", "config", config)
	client, err := client.NewGreenfieldChainSignClient(
		// TODO: greenfield SDK may support multiple endpoints.
		chainConfig.NodeAddr[0].TendermintAddresses[0],
		chainConfig.ChainID,
		config.GasLimit,
		config.OperatorPrivateKey,
		config.FundingPrivateKey,
		config.SealPrivateKey,
		config.ApprovalPrivateKey,
		config.GcPrivateKey)
	if err != nil {
		return nil, err
	}
	return &SignerServer{
		config:       config,
		chainConfig:  chainConfig,
		client:       client,
		svcWhitelist: svcWhitelist,
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
	lis, err := net.Listen("tcp", signer.config.GRPCAddress)
	if err != nil {
		log.Errorw("failed to listen", "address", signer.config.GRPCAddress, "error", err)
		return
	}
	options := utilgrpc.GetDefaultServerOptions()
	if metrics.GetMetrics().Enabled() {
		options = append(options, utilgrpc.GetDefaultServerInterceptor()...)
	}
	options = append(options, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		signer.IPWhitelistInterceptor(),
		signer.AuthInterceptor(),
	)))
	signer.server = grpc.NewServer(options...)

	types.RegisterSignerServiceServer(signer.server, signer)
	// register reflection service
	reflection.Register(signer.server)
	if err := signer.server.Serve(lis); err != nil {
		log.Errorf("grpc serve error : %v", err)
		return
	}
}
