package challenge

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/challenge/types"
	psclient "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
)

// Challenge implements the gRPC of ChallengeService,
// responsible for handling challenge piece request.
type Challenge struct {
	config     *ChallengeConfig
	spDB       sqldb.SPDB
	pieceStore *psclient.StoreClient
}

// NewChallengeService returns an instance of Challenge that implementation of
// the lifecycle.Service and ChallengeService interface
func NewChallengeService(config *ChallengeConfig) (challenge *Challenge, err error) {
	pieceStore, err := psclient.NewStoreClient(config.PieceStoreConfig)
	if err != nil {
		return nil, err
	}
	spDB, err := sqldb.NewSpDB(config.SpDBConfig)
	if err != nil {
		return nil, err
	}
	challenge = &Challenge{
		config:     config,
		spDB:       spDB,
		pieceStore: pieceStore,
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

	go func(errCh chan error) {
		lis, err := net.Listen("tcp", challenge.config.GRPCAddress)
		errCh <- err
		if err != nil {
			log.Errorw("challenge listen failed", "error", err)
			return
		}
		grpcServer := grpc.NewServer()
		types.RegisterChallengeServiceServer(grpcServer, challenge)
		reflection.Register(grpcServer)
		if err = grpcServer.Serve(lis); err != nil {
			log.Errorw("challenge serve failed", "error", err)
			return
		}
	}(errCh)

	err := <-errCh
	return err
}

// Stop the challenge gRPC service and recycle the resources
func (challenge *Challenge) Stop(ctx context.Context) error {
	return nil
}
