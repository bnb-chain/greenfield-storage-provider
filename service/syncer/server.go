package syncer

import (
	"context"
	"net"

	"google.golang.org/grpc"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

type syncerConfig struct {
	Port                 string
	PieceStoreConfigFile string
}

// SyncerService synchronizes ec data to piece store
type SyncerService struct {
	cfg       syncerConfig
	grpcSever *grpc.Server
	signer    *signerClient
	store     *storeClient
}

// Name describes the name of SyncerService
func (s *SyncerService) Name() string {
	return "SyncerService"
}

// Start running SyncerService
func (s *SyncerService) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", ":"+s.cfg.Port)
	if err != nil {
		log.Errorw("Start SyncerService net.Listen error", "error", err)
		return err
	}
	s.grpcSever = grpc.NewServer()
	service.RegisterSyncerServiceServer(s.grpcSever, &SyncerImpl{syncer: s})
	go func() {
		if err := s.grpcSever.Serve(lis); err != nil {
			log.Errorw("gRPC server Serve error", "error", err)
		}
	}()
	s.signer = newSignerClient()
	if s.store, err = newStoreClient(s.cfg.PieceStoreConfigFile); err != nil {
		log.Errorw("Syncer service starts newStoreClient failed", "error", err)
		return err
	}
	log.Info("Start SyncerService successfully")
	return nil
}

// Stop running SyncerService
func (s *SyncerService) Stop(ctx context.Context) error {
	return nil
}
