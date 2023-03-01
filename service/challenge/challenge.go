package challenge

import (
	"context"
	"net"

	"github.com/bnb-chain/greenfield-storage-provider/service/challenge/types"
	"github.com/bnb-chain/greenfield-storage-provider/store"
	pscli "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// Challenge manage the integrity hash check
type Challenge struct {
	config     *ChallengeConfig
	spDB       store.SPDB
	pieceStore *pscli.StoreClient
}

// NewChallengeService return a Challenge instance.
func NewChallengeService(config *ChallengeConfig) (challenge *Challenge, err error) {
	pieceStore, err := pscli.NewStoreClient(config.PieceStoreConfig)
	if err != nil {
		return nil, err
	}
	// TODO:: new sp db
	challenge = &Challenge{
		config:     config,
		pieceStore: pieceStore,
	}
	return challenge, nil
}

// Name describes the name of Challenge
func (challenge *Challenge) Name() string {
	return model.ChallengeService
}

// Start implement the lifecycle interface
func (challenge *Challenge) Start(ctx context.Context) error {
	errCh := make(chan error)

	go func(errCh chan error) {
		lis, err := net.Listen("tcp", challenge.config.GrpcAddress)
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

// Stop implement the lifecycle interface
func (challenge *Challenge) Stop(ctx context.Context) error {
	return nil
}
