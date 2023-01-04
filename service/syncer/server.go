package syncer

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

const (
	syncerServiceName string = "Syncer"
)

type syncerConfig struct {
	Port                 string
	PieceStoreConfigFile string
}

// SyncerService synchronizes ec data to piece store
type Syncer struct {
	cfg   *SyncerConfig
	name  string
	store *storeClient
}

func NewSyncerService(config *SyncerConfig) (*Syncer, error) {
	return &Syncer{
		cfg:  config,
		name: syncerServiceName,
	}, nil
}

func (s *Syncer) Init() error {
	store, err := newStoreClient(s.cfg.PieceConfig)
	if err != nil {
		log.Errorw("Syncer starts newStoreClient failed", "error", err)
		return err
	}
	s.store = store
	return nil
}

// Name describes the name of SyncerService
func (s *Syncer) Name() string {
	return s.name
}

// Start running SyncerService
func (s *Syncer) Start(ctx context.Context) error {
	go s.Serve()
	return nil
}

// Stop running SyncerService
func (s *Syncer) Stop(ctx context.Context) error {
	log.Info("Stop syncer service")
	return nil
}

func (s *Syncer) Serve() {
	lis, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		log.Errorw("syncer net.Listen failed", "error", err)
		return
	}
	grpcServer := grpc.NewServer()
	service.RegisterSyncerServiceServer(grpcServer, s)
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Errorw("syncer serve failed", "error", err)
		return
	}
	log.Info("Start syncer successfully")
}
