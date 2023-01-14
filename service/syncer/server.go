package syncer

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/store/metadb/leveldb"

	"github.com/bnb-chain/greenfield-storage-provider/store/metadb"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/service/client"
	service "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

const (
	ServiceNameSyncer string = "Syncer"
)

// SyncerService synchronizes ec data to piece store
type Syncer struct {
	cfg     *SyncerConfig
	name    string
	store   client.PieceStoreAPI
	metaDB  metadb.MetaDB // storage provider meta db
	running atomic.Bool
}

// NewSyncerService creates a syncer service to upload piece to piece store
func NewSyncerService(config *SyncerConfig) (*Syncer, error) {
	s := &Syncer{
		cfg:  config,
		name: ServiceNameSyncer,
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

func (s *Syncer) initClient() error {
	store, err := client.NewStoreClient(s.cfg.PieceConfig)
	if err != nil {
		log.Errorw("syncer starts piece store client failed", "error", err)
		return err
	}
	s.store = store
	return nil
}

func (s *Syncer) initDB() error {
	var err error
	switch s.cfg.MetaDBType {
	case model.LevelDB:
		if s.cfg.MetaDB == nil {
			s.cfg.MetaDB = DefaultSyncerConfig.MetaDB
		}
		s.metaDB, err = leveldb.NewMetaDB(s.cfg.MetaDB)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported meta db type %s", s.cfg.MetaDBType)
	}
	return nil
}

// Name describes the name of SyncerService
func (s *Syncer) Name() string {
	return s.name
}

// Start running SyncerService
func (s *Syncer) Start(ctx context.Context) error {
	if s.running.Swap(true) {
		return errors.New("syncer service is running")
	}
	errCh := make(chan error)
	go s.serve(errCh)
	err := <-errCh
	return err
}

// Stop running SyncerService
func (s *Syncer) Stop(ctx context.Context) error {
	if !s.running.Swap(false) {
		return errors.New("syncer service has already stopped")
	}
	return nil
}

// serve start syncer rpc service
func (s *Syncer) serve(errCh chan error) {
	lis, err := net.Listen("tcp", s.cfg.Address)
	errCh <- err
	if err != nil {
		log.Errorw("syncer listen failed", "error", err)
		return
	}
	grpcServer := grpc.NewServer(grpc.MaxSendMsgSize(model.MaxCallMsgSize), grpc.MaxRecvMsgSize(model.MaxCallMsgSize))
	service.RegisterSyncerServiceServer(grpcServer, s)
	reflection.Register(grpcServer)
	if err = grpcServer.Serve(lis); err != nil {
		log.Errorw("syncer serve failed", "error", err)
		return
	}
	return
}
