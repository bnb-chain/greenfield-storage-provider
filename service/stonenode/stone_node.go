package stonenode

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/service/client"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

const (
	allocStonePeriod = time.Second * 1
)

// StoneNodeService manages stone execution units
type StoneNodeService struct {
	cfg        *StoneNodeConfig
	name       string
	stoneHub   client.StoneHubAPI
	store      client.PieceStoreAPI
	stoneLimit int64

	running atomic.Bool
	stopCh  chan struct{}
}

// NewStoneNodeService returns StoneNodeService instance
func NewStoneNodeService(config *StoneNodeConfig) (*StoneNodeService, error) {
	node := &StoneNodeService{
		cfg:        config,
		name:       model.StoneNodeService,
		stopCh:     make(chan struct{}),
		stoneLimit: config.StoneJobLimit,
	}
	if err := node.initClient(); err != nil {
		return nil, err
	}
	return node, nil
}

// initClient inits store client and rpc client
func (node *StoneNodeService) initClient() error {
	if node.running.Load() {
		return merrors.ErrStoneNodeStarted
	}
	store, err := client.NewStoreClient(node.cfg.PieceStoreConfig)
	if err != nil {
		log.Errorw("stone node inits piece store client failed", "error", err)
		return err
	}
	stoneHub, err := client.NewStoneHubClient(node.cfg.StoneHubServiceAddress)
	if err != nil {
		log.Errorw("stone node inits stone hub client failed", "error", err)
		return err
	}
	node.store = store
	node.stoneHub = stoneHub
	return nil
}

// Name returns the name of StoneNodeService, implement lifecycle interface
func (node *StoneNodeService) Name() string {
	return node.name
}

// Start running StoneNodeService, implement lifecycle interface
func (node *StoneNodeService) Start(startCtx context.Context) error {
	if node.running.Swap(true) {
		return merrors.ErrStoneNodeStarted
	}
	go func() {
		var stoneJobCounter int64 // atomic
		allocTicker := time.NewTicker(allocStonePeriod)
		ctx, cancel := context.WithCancel(startCtx)
		for {
			select {
			case <-allocTicker.C:
				go func() {
					if !node.running.Load() {
						log.Errorw("stone node service stopped, can not alloc stone")
						return
					}
					if node.stoneLimit <= 0 {
						log.Errorw("stone node stone limit is zero, forbid pull stone job")
						return
					}
					atomic.AddInt64(&stoneJobCounter, 1)
					defer atomic.AddInt64(&stoneJobCounter, -1)
					if atomic.LoadInt64(&stoneJobCounter) > node.stoneLimit {
						log.Errorw("stone job running number exceeded, skip current alloc stone")
						return
					}
					// TBD::exceed stoneLimit or alloc empty stone,
					// stone node need one backoff strategy.
					node.allocStoneJob(ctx)
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
		return merrors.ErrStoneNodeStopped
	}
	close(node.stopCh)
	var errs []error
	if err := node.stoneHub.Close(); err != nil {
		errs = append(errs, err)
	}
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	return nil
}
