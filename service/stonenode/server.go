package stonenode

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	merrors "github.com/bnb-chain/inscription-storage-provider/model/errors"
	"github.com/bnb-chain/inscription-storage-provider/service/client"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

const (
	ServiceNameStoneNode string = "StoneNode"
	AllocStonePeriod            = time.Second * 1
)

// StoneNodeService manages stone execution units
type StoneNodeService struct {
	cfg  *StoneNodeConfig
	name string
	//syncer   *client.SyncerClient
	//stoneHub *client.StoneHubClient
	//store    *client.StoreClient
	syncer     client.SyncerAPI
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
		name:       ServiceNameStoneNode,
		stopCh:     make(chan struct{}),
		stoneLimit: config.StoneJobLimit,
	}
	if err := node.initClient(); err != nil {
		return nil, err
	}
	return node, nil
}

// InitClient inits store client and rpc client
func (node *StoneNodeService) initClient() error {
	if node.running.Load() == true {
		return merrors.ErrStoneNodeStarted
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
		log.Info("1")
		return merrors.ErrStoneNodeStarted
	}
	go func() {
		log.Info("2")
		var stoneJobCounter int64 // atomic
		allocTimer := time.NewTimer(AllocStonePeriod)
		ctx, cancel := context.WithCancel(context.Background())
		log.Info("3")
		for {
			select {
			case <-allocTimer.C:
				log.Info("4")
				go func() {
					log.Info("5")
					if !node.running.Load() {
						log.Info("6")
						log.Errorw("stone node service stopped, can not alloc stone.")
						return
					}
					log.Info("7")
					atomic.AddInt64(&stoneJobCounter, 1)
					defer atomic.AddInt64(&stoneJobCounter, -1)
					log.Infow("8")
					if atomic.LoadInt64(&stoneJobCounter) > node.stoneLimit {
						log.Info("9")
						log.Errorw("stone job running number exceeded, skip current alloc stone.")
						return
					}
					log.Info("10")
					// TBD::exceed stoneLimit or alloc empty stone,
					// stone node need one backoff strategy.
					node.allocStone(ctx)
				}()
			case <-node.stopCh:
				log.Info("11")
				cancel()
				return
			}
		}
	}()
	log.Info("12")
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
	log.Info("enter into allocStone")
	resp, err := node.stoneHub.AllocStoneJob(ctx)
	ctx = log.Context(ctx, resp)
	log.Info("18")
	if err != nil {
		log.CtxErrorw(ctx, "alloc stone from stone hub failed", "error", err)
		return
	}
	log.Info("19")
	// TBD:: stone node will support more types of stone job,
	// currently only support upload secondary piece job.
	if err := node.syncPieceToSecondarySP(ctx, resp); err != nil {
		log.CtxErrorw(ctx, "upload secondary piece job failed", "error", err)
		return
	}
	log.CtxInfow(ctx, "upload secondary piece job success!")
	return
}
