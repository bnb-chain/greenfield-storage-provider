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
	//segChan  chan []byte
	errChan chan error
}

func NewStoneNodeService(config *StoneNodeConfig) *StoneNodeService {
	return &StoneNodeService{
		cfg:     config,
		name:    stoneNodeServiceName,
		errChan: make(chan error),
	}
}

func (s *StoneNodeService) Init() error {
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
		if err := s.Init(); err != nil {
			log.Errorw("stone node init failed", "error", err)
			return
		}
	}()

	go func() {
		for {
			if err := s.doEC(ctx); err != nil {
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
	close(s.errChan)
	log.Info("Stop stone node!")
	return nil
}
