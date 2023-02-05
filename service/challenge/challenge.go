package challenge

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/service/client"
	stypesv1pb "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/leveldb"
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

// initClient init related client resource.
func (challenge *Challenge) initDB() (err error) {
	switch challenge.config.MetaType {
	case model.LevelDB:
		if challenge.config.MetaDB == nil {
			challenge.config.MetaDB = DefaultChallengeConfig.MetaDB
		}
		challenge.metaDB, err = leveldb.NewMetaDB(challenge.config.MetaDB)
		if err != nil {
			return
		}
	case model.MySqlDB:
		// TODO:: meta support SQL
	default:
		return fmt.Errorf("meta db not support type %s", challenge.config.MetaType)
	}
	challenge.pieceStore, err = client.NewStoreClient(challenge.config.PieceConfig)
	if err != nil {
		log.Errorw("syncer starts piece store client failed", "error", err)
		return
	}
	return
}

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
			log.Errorw("syncer listen failed", "error", err)
			return
		}
		grpcServer := grpc.NewServer()
		stypesv1pb.RegisterChallengeServiceServer(grpcServer, challenge)
		reflection.Register(grpcServer)
		if err = grpcServer.Serve(lis); err != nil {
			log.Errorw("syncer serve failed", "error", err)
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
