package syncer

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/inscription-storage-provider/service/client"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

const (
	syncerServiceName string = "Syncer"
)

// SyncerService synchronizes ec data to piece store
type Syncer struct {
	cfg   *SyncerConfig
	name  string
	store *client.StoreClient
}

// NewSyncerService creates a syncer service to upload piece to piece store
func NewSyncerService(config *SyncerConfig) (*Syncer, error) {
	s := &Syncer{
		cfg:  config,
		name: syncerServiceName,
	}
	if err := s.InitClient(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Syncer) InitClient() error {
	store, err := client.NewStoreClient(s.cfg.PieceConfig)
	if err != nil {
		log.Errorw("syncer starts piece store client failed", "error", err)
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
	resCh := make(chan struct{})
	go s.serve(resCh)
	return nil
}

// Stop running SyncerService
func (s *Syncer) Stop(ctx context.Context) error {
	return nil
}

// serve start syncer rpc service
func (s *Syncer) serve(resCh chan struct{}) {
	lis, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		log.Errorw("syncer listen failed", "error", err)
		return
	}
	grpcServer := grpc.NewServer()
	service.RegisterSyncerServiceServer(grpcServer, s)
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Errorw("syncer serve failed", "error", err)
		return
	}
	resCh <- struct{}{}
}
