package manager

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"github.com/bnb-chain/greenfield-storage-provider/service/manager/types"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
	utilgrpc "github.com/bnb-chain/greenfield-storage-provider/util/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var _ lifecycle.Service = &Manager{}

var (
	// RefreshSPInfoAndStorageParamsTimer defines the period of refresh sp info and storage params.
	RefreshSPInfoAndStorageParamsTimer = 5 * 60
	// GCManagerPriorityQueueTimer defines the period of gc manager priority queue.
	GCManagerPriorityQueueTimer = 60
)

// Manager module is responsible for implementing internal management functions.
// Currently, it supports periodic update of sp info list and storage params information in sp db.
// TODO::support gc and configuration management, etc.
type Manager struct {
	config     *ManagerConfig
	pqueue     *MPQueue
	running    atomic.Value
	stopCh     chan struct{}
	chain      *gnfd.Greenfield
	spDB       sqldb.SPDB
	grpcServer *grpc.Server
}

// NewManagerService returns an instance of manager
func NewManagerService(cfg *ManagerConfig) (*Manager, error) {
	var (
		manager *Manager
		err     error
	)

	manager = &Manager{
		config: cfg,
		stopCh: make(chan struct{}),
	}
	if manager.chain, err = gnfd.NewGreenfield(cfg.ChainConfig); err != nil {
		log.Errorw("failed to create chain client", "error", err)
		return nil, err
	}
	if manager.spDB, err = sqldb.NewSpDB(cfg.SpDBConfig); err != nil {
		log.Errorw("failed to create spdb client", "error", err)
		return nil, err
	}
	manager.pqueue = NewMPQueue(manager.chain, cfg.UploadQueueCap, cfg.ReplicateQueueCap,
		cfg.SealQueueCap, cfg.GCObjectQueueCap)

	return manager, nil
}

// Name return the manager service name
func (m *Manager) Name() string {
	return model.ManagerService
}

// Start function start background goroutine to execute refresh sp meta
func (m *Manager) Start(ctx context.Context) error {
	if m.running.Swap(true) == true {
		return errors.New("manager has already started")
	}

	// start background task
	go m.eventLoop()
	errCh := make(chan error)
	go m.serve(errCh)
	err := <-errCh
	return err
}

// serve start the manager gRPC service
func (m *Manager) serve(errCh chan error) {
	lis, err := net.Listen("tcp", m.config.GRPCAddress)
	errCh <- err
	if err != nil {
		log.Errorw("failed to listen", "error", err)
		return
	}

	options := utilgrpc.GetDefaultServerOptions()
	if metrics.GetMetrics().Enabled() {
		options = append(options, utilgrpc.GetDefaultServerInterceptor()...)
	}
	m.grpcServer = grpc.NewServer(options...)
	types.RegisterManagerServiceServer(m.grpcServer, m)
	reflection.Register(m.grpcServer)
	if err := m.grpcServer.Serve(lis); err != nil {
		log.Errorw("failed to start grpc server", "error", err)
		return
	}
}

// eventLoop background goroutine, responsible for refreshing sp info and storage params
func (m *Manager) eventLoop() {
	m.refreshSPInfoAndStorageParams()
	refreshSPInfoAndStorageParamsTicker := time.NewTicker(time.Duration(RefreshSPInfoAndStorageParamsTimer) * time.Second)
	gcMPQueueTicker := time.NewTicker(time.Duration(GCManagerPriorityQueueTimer) * time.Second)
	for {
		select {
		case <-refreshSPInfoAndStorageParamsTicker.C:
			go m.refreshSPInfoAndStorageParams()
		case <-gcMPQueueTicker.C:
			go m.pqueue.GCMPQueueTask()
		case <-m.stopCh:
			return
		}
	}
}

// refreshSPInfoAndStorageParams fetch sp info and storage params from chain and update to spdb
func (m *Manager) refreshSPInfoAndStorageParams() {
	spInfoList, err := m.chain.QuerySPInfo(context.Background())
	if err != nil {
		log.Errorw("failed to query sp info", "error", err)
		return
	}
	if err = m.spDB.UpdateAllSp(spInfoList); err != nil {
		log.Errorw("failed to update sp info", "error", err)
		return
	}
	for _, spInfo := range spInfoList {
		if spInfo.OperatorAddress == m.config.SpOperatorAddress {
			if err = m.spDB.SetOwnSpInfo(spInfo); err != nil {
				log.Errorw("failed to set own sp info", "error", err)
				return
			}
		}
	}
	log.Infow("succeed to refresh sp info", "sp_info", spInfoList)
	storageParams, err := m.chain.QueryStorageParams(context.Background())
	if err != nil {
		log.Errorw("failed to query storage params", "error", err)
		return
	}
	if err = m.spDB.SetStorageParams(storageParams); err != nil {
		log.Errorw("failed to update storage params", "error", err)
		return
	}
	log.Infow("succeed to refresh storage params", "params", storageParams)
}

// Stop manager background goroutine
func (m *Manager) Stop(ctx context.Context) error {
	if m.running.Swap(false) == false {
		return errors.New("manager has already stop")
	}
	close(m.stopCh)
	return nil
}
