package challenge

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	mwgrpc "github.com/bnb-chain/greenfield-storage-provider/pkg/middleware/grpc"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/service/challenge/types"
	psclient "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	utilgrpc "github.com/bnb-chain/greenfield-storage-provider/util/grpc"
)

var _ lifecycle.Service = &Challenge{}

// Challenge implements the gRPC of ChallengeService,
// responsible for handling challenge piece request.
type Challenge struct {
	config     *ChallengeConfig
	spDB       sqldb.SPDB
	rcScope    rcmgr.ResourceScope
	pieceStore *psclient.StoreClient
	grpcServer *grpc.Server
}

// NewChallengeService returns an instance of Challenge that implementation of
// the lifecycle.Service and ChallengeService interface
func NewChallengeService(cfg *ChallengeConfig) (*Challenge, error) {
	var (
		challenge *Challenge
		err       error
	)

	challenge = &Challenge{
		config: cfg,
	}
	if challenge.pieceStore, err = psclient.NewStoreClient(cfg.PieceStoreConfig); err != nil {
		log.Errorw("failed to create piece store client", "error", err)
		return nil, err
	}
	if challenge.spDB, err = sqldb.NewSpDB(cfg.SpDBConfig); err != nil {
		log.Errorw("failed to create sp db client", "error", err)
		return nil, err
	}
	if challenge.rcScope, err = rcmgr.ResrcManager().OpenService(model.ChallengeService); err != nil {
		log.Errorw("failed to open challenge resource scope", "error", err)
		return nil, err
	}

	return challenge, nil
}

// Name return the challenge service name, for the lifecycle management
func (challenge *Challenge) Name() string {
	return model.ChallengeService
}

// Start the challenge gRPC service
func (challenge *Challenge) Start(ctx context.Context) error {
	errCh := make(chan error)
	go challenge.serve(errCh)
	err := <-errCh
	return err
}

// Stop the challenge gRPC service and recycle the resources
func (challenge *Challenge) Stop(ctx context.Context) error {
	challenge.grpcServer.GracefulStop()
	challenge.rcScope.Release()
	return nil
}

func (challenge *Challenge) serve(errCh chan error) {
	lis, err := net.Listen("tcp", challenge.config.GRPCAddress)
	errCh <- err
	if err != nil {
		log.Errorw("failed to listen", "error", err)
		return
	}

	options := utilgrpc.GetDefaultServerOptions()
	if metrics.GetMetrics().Enabled() {
		options = append(options, mwgrpc.GetDefaultServerInterceptor()...)
	}
	challenge.grpcServer = grpc.NewServer(options...)
	types.RegisterChallengeServiceServer(challenge.grpcServer, challenge)
	reflection.Register(challenge.grpcServer)
	if err = challenge.grpcServer.Serve(lis); err != nil {
		log.Errorw("failed to serve", "error", err)
		return
	}
}
