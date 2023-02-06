package challenge

import (
	"context"
	"fmt"
	"net"

	"github.com/bnb-chain/greenfield-storage-provider/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/service/client"
	stypesv1pb "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// Challenge manage the integrity hash check
type Challenge struct {
	config     *ChallengeConfig
	name       string
	metaDB     metadb.MetaDB // storage provider meta db
	pieceStore *client.StoreClient
}

// NewChallengeService return a Challenge instance.
func NewChallengeService(config *ChallengeConfig) (challenge *Challenge, err error) {
	challenge = &Challenge{
		config: config,
		name:   model.ChallengeService,
	}
	err = challenge.initDB()
	return
}

// initDB init related client resource.
func (challenge *Challenge) initDB() error {
	var (
		metaDB metadb.MetaDB
		err    error
	)

	metaDB, err = store.NewMetaDB(challenge.config.MetaDBType,
		challenge.config.MetaLevelDBConfig, challenge.config.MetaSqlDBConfig)
	if err != nil {
		log.Errorw("failed to init metaDB", "err", err)
		return err
	}
	challenge.metaDB = metaDB

	challenge.pieceStore, err = client.NewStoreClient(challenge.config.PieceConfig)
	if err != nil {
		log.Errorw("challenge starts piece store client failed", "error", err)
		return err
	}
	return nil
}

// Name describes the name of Challenge
func (challenge *Challenge) Name() string {
	return challenge.name
}

// Start implement the lifecycle interface
func (challenge *Challenge) Start(ctx context.Context) error {
	errCh := make(chan error)

	go func(errCh chan error) {
		lis, err := net.Listen("tcp", challenge.config.Address)
		errCh <- err
		if err != nil {
			log.Errorw("challenge listen failed", "error", err)
			return
		}
		grpcServer := grpc.NewServer()
		stypesv1pb.RegisterChallengeServiceServer(grpcServer, challenge)
		reflection.Register(grpcServer)
		if err = grpcServer.Serve(lis); err != nil {
			log.Errorw("challenge serve failed", "error", err)
			return
		}
		return
	}(errCh)

	err := <-errCh
	return err
}

// Stop implement the lifecycle interface
func (challenge *Challenge) Stop(ctx context.Context) error {
	var errs []error
	if err := challenge.metaDB.Close(); err != nil {
		errs = append(errs, err)
	}
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	return nil
}
