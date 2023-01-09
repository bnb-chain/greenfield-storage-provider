package stonenode

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/service/client"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

const (
	ServiceNameStoneNode string = "StoneNode"
	AllocStonePeriod            = time.Second * 1
)

// StoneNodeService manages stone execution units
type StoneNodeService struct {
	cfg        *StoneNodeConfig
	name       string
	syncer     *client.SyncerClient
	stoneHub   *client.StoneHubClient
	store      *client.StoreClient
	stoneLimit int64

	running atomic.Bool
	stopCh  chan struct{}
}

// NewStoneNodeService returns StoneNodeService instance
func NewStoneNodeService(config *StoneNodeConfig) (*StoneNodeService, error) {
	node := &StoneNodeService{
		cfg:        config,
		name:       ServiceNameStoneNode,
		stopCh:     make(chan struct{}),
		stoneLimit: config.StoneJobLimit,
	}
	if err := node.InitClient(); err != nil {
		return nil, err
	}
	return node, nil
}

// InitClient inits store client and rpc client
func (node *StoneNodeService) InitClient() error {
	if node.running.Load() == true {
		return errors.New("stone node resource is running")
	}
	store, err := client.NewStoreClient(node.cfg.PieceConfig)
	if err != nil {
		log.Errorw("stone node inits piece store client failed", "error", err)
		return err
	}
	stoneHub, err := client.NewStoneHubClient(node.cfg.StoneHubServiceAddress)
	if err != nil {
		log.Errorw("stone node inits stone hub client failed", "error", err)
		return err
	}
	syncer, err := client.NewSyncerClient(node.cfg.SyncerServiceAddress)
	if err != nil {
		log.Errorw("stone node inits syncer client failed", "error", err)
		return err
	}
	node.store = store
	node.stoneHub = stoneHub
	node.syncer = syncer
	return nil
}

// Name returns the name of StoneNodeService, implement lifecycle interface
func (node *StoneNodeService) Name() string {
	return node.name
}

// Start running StoneNodeService, implement lifecycle interface
func (node *StoneNodeService) Start(ctx context.Context) error {
	if node.running.Swap(true) {
		return errors.New("stone node resource is running")
	}
	go func() {
		var stoneJobCounter int64 // atomic
		allocTimer := time.NewTimer(AllocStonePeriod)
		ctx, cancel := context.WithCancel(context.Background())
		for {
			select {
			case <-allocTimer.C:
				go func() {
					if !node.running.Load() {
						log.Errorw("stone node service stopped, can not alloc stone.")
						return
					}
					atomic.AddInt64(&stoneJobCounter, 1)
					defer atomic.AddInt64(&stoneJobCounter, -1)
					if atomic.LoadInt64(&stoneJobCounter) > node.stoneLimit {
						log.Errorw("stone job running number exceeded, skip current alloc stone.")
						return
					}
					// TBD::exceed stoneLimit or alloc empty stone,
					// stone node need one backoff strategy.
					node.allocStone(ctx)
				}()
			case <-node.stopCh:
				cancel()
				return
			}
		}
	}()
	return nil
}

// Stop running StoneNodeService, implement lifecycle interface
func (node *StoneNodeService) Stop(ctx context.Context) error {
	if !node.running.Swap(false) {
		return errors.New("stone node service has stopped")
	}
	close(node.stopCh)
	var errs []error
	if err := node.stoneHub.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := node.syncer.Close(); err != nil {
		errs = append(errs, err)
	}
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

// allocStone sends rpc request to stone hub alloc stone job.
func (node *StoneNodeService) allocStone(ctx context.Context) {
	resp, err := node.stoneHub.AllocStoneJob(ctx)
	ctx = log.Context(ctx, resp)
	if err != nil {
		log.CtxErrorw(ctx, "alloc stone from stone hub failed", "error", err)
		return
	}
	// TBD:: stone node will support more types of stone job,
	// currently only support upload secondary piece job.
	if err := node.syncPieceToSecondarySP(ctx, resp); err != nil {
		log.CtxErrorw(ctx, "upload secondary piece job failed", "error", err)
		return
	}
	log.CtxInfow(ctx, "upload secondary piece job success!")
	return
}
