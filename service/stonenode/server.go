package stonenode

import (
	"context"

	"google.golang.org/grpc"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"

	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

const (
	stoneNodeServiceName string = "StoneNodeService"
)

type stoneNodeConfig struct {
	addr                 string
	PieceStoreConfigFile string
}

type StoneNodeService struct {
	sn    *StoneNode
	cfg   *stoneNodeConfig
	store *storeClient
}

// Name describes the name of StoneNodeService
func (s *StoneNodeService) Name() string {
	return stoneNodeServiceName
}

// Start running StoneNodeService
func (s *StoneNodeService) Start(ctx context.Context) error {
	conn, err := grpc.Dial(s.cfg.addr)
	if err != nil {
		log.Errorw("Start StoneNodeService failed", "error", err)
		return err
	}
	defer conn.Close()

	s.sn.syncerClient = service.NewSyncerServiceClient(conn)
	s.sn.stoneHubClient = service.NewStoneHubServiceClient(conn)
	log.Info("Start StoneNodeService successfully")
	return nil
}

// Stop running StoneNodeService
func (s *StoneNodeService) Stop(ctx context.Context) error {
	return nil
}
