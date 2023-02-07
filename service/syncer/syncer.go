package syncer

import (
	"context"
	"net"
	"sync/atomic"

	"github.com/bnb-chain/greenfield-storage-provider/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/service/client"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// Syncer synchronizes ec data to piece store
type Syncer struct {
	config  *SyncerConfig
	name    string
	store   client.PieceStoreAPI
	metaDB  metadb.MetaDB // storage provider meta db
	running atomic.Bool
}

// NewSyncerService creates a syncer service to upload piece to piece store
func NewSyncerService(config *SyncerConfig) (*Syncer, error) {
	s := &Syncer{
		config: config,
		name:   model.SyncerService,
	}
	if err := s.initClient(); err != nil {
		return nil, err
	}
	// init meta db
	if err := s.initDB(); err != nil {
		return nil, err
	}
	return s, nil
}

// initClient
func (s *Syncer) initClient() error {
	store, err := client.NewStoreClient(s.config.PieceConfig)
	if err != nil {
		log.Errorw("syncer starts piece store client failed", "error", err)
		return err
	}
	s.store = store
	return nil
}

// initDB init a meta-db instance
func (s *Syncer) initDB() error {
	var (
		metaDB metadb.MetaDB
		err    error
	)

	metaDB, err = store.NewMetaDB(s.config.MetaDBType,
		s.config.MetaLevelDBConfig, s.config.MetaSqlDBConfig)
	if err != nil {
		log.Errorw("failed to init metaDB", "err", err)
		return err
	}
	s.metaDB = metaDB
	return nil
}

// Name describes the name of SyncerService
func (s *Syncer) Name() string {
	return s.name
}

// Start running SyncerService
func (s *Syncer) Start(ctx context.Context) error {
	if s.running.Swap(true) {
		return merrors.ErrSyncerStarted
	}
	errCh := make(chan error)
	go s.serve(errCh)
	err := <-errCh
	return err
}

// Stop running SyncerService
func (s *Syncer) Stop(ctx context.Context) error {
	if !s.running.Swap(false) {
		return merrors.ErrSyncerStopped
	}
	return nil
}

// serve start syncer rpc service
func (s *Syncer) serve(errCh chan error) {
	lis, err := net.Listen("tcp", s.config.Address)
	errCh <- err
	if err != nil {
		log.Errorw("syncer listen failed", "error", err)
		return
	}
	grpcServer := grpc.NewServer(grpc.MaxSendMsgSize(model.MaxCallMsgSize), grpc.MaxRecvMsgSize(model.MaxCallMsgSize))
	stypes.RegisterSyncerServiceServer(grpcServer, s)
	reflection.Register(grpcServer)
	if err = grpcServer.Serve(lis); err != nil {
		log.Errorw("syncer serve failed", "error", err)
		return
	}
}
