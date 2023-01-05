package stonenode

import (
	"context"
	"time"

	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

const (
	stoneNodeServiceName string = "StoneNode"
	grpcTimeout                 = time.Second * 5
)

type StoneNodeService struct {
	cfg      *StoneNodeConfig
	name     string
	syncer   service.SyncerServiceClient
	stoneHub service.StoneHubServiceClient
	store    *storeClient
}

// NewStoneNodeService creates a stone node service
func NewStoneNodeService(config *StoneNodeConfig) (*StoneNodeService, error) {
	s := &StoneNodeService{
		cfg:  config,
		name: stoneNodeServiceName,
	}
	if err := s.InitClient(); err != nil {
		log.Errorw("stone node init client failed", "error", err)
		return nil, err
	}
	return s, nil
}

// InitClient inits store client and rpc client
func (s *StoneNodeService) InitClient() error {
	store, err := newStoreClient(s.cfg.PieceConfig)
	if err != nil {
		log.Errorw("stone node inits newStoreClient failed", "error", err)
		return err
	}
	s.store = store

	stoneHub, err := newStoneHubClient(s.cfg.StoneHubServiceAddress)
	if err != nil {
		log.Errorw("stone node inits newStoneHubClient failed", "error", err)
	}
	s.stoneHub = stoneHub

	syncer, err := newSyncerClient(s.cfg.SyncerServiceAddress)
	if err != nil {
		log.Errorw("stone node inits newSyncerClient failed", "error", err)
	}
	s.syncer = syncer
	return nil
}

// Name describes the name of StoneNodeService
func (s *StoneNodeService) Name() string {
	return s.name
}

// Start running StoneNodeService
func (s *StoneNodeService) Start(ctx context.Context) error {
	go func() {
		for {
			if err := s.doEC(ctx, s.cfg.StorageProvider); err != nil {
				log.Errorw("do ec failed", "error", err)
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()
	log.Info("Start StoneNodeService successfully")
	return nil
}

// Stop running StoneNodeService
func (s *StoneNodeService) Stop(ctx context.Context) error {
	log.Info("Stop stone node!")
	return nil
}
