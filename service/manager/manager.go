package manager

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
)

var (
	// RefreshStorageParamsTimer define the period of refresh storage params
	RefreshStorageParamsTimer = 60 * 60
	// RefreshSPInfoTimer define the period of refresh sp info
	RefreshSPInfoTimer = 60 * 60
)

// Manager is responsible for managing the storage provider cluster
type Manager struct {
	cfg     *ManagerConfig
	running atomic.Bool
	stopCh  chan struct{}
	chain   *gnfd.Greenfield
	spDB    sqldb.SPDB
}

// NewManagerService returns an instance of Manager that implementation of the lifecycle.Service
func NewManagerService(cfg *ManagerConfig) (*Manager, error) {
	chain, err := gnfd.NewGreenfield(cfg.ChainConfig)
	if err != nil {
		log.Errorw("failed to create chain client", "error", err)
		return nil, err
	}
	spDB, err := sqldb.NewSpDB(cfg.SpDBConfig)
	if err != nil {
		log.Errorw("failed to create spdb client", "error", err)
		return nil, err
	}
	manager := &Manager{
		cfg:   cfg,
		spDB:  spDB,
		chain: chain,
	}
	return manager, nil
}

// Name return the manager service name, for the lifecycle management
func (m *Manager) Name() string {
	return model.ManagerService
}

// Start stone hub service, implement the lifecycle interface.
func (m *Manager) Start(ctx context.Context) error {
	if m.running.Swap(true) {
		return errors.New("manager has already started")
	}

	// start background task
	go m.eventLoop()
	return nil
}

// eventLoop background goroutine, responsible for refreshing sp info and storage params
func (m *Manager) eventLoop() {
	m.refreshStorageParams()
	m.refreshSPInfo()
	refreshStorageParamsTicker := time.NewTicker(time.Duration(RefreshStorageParamsTimer) * time.Second)
	refreshSPInfoTicker := time.NewTicker(time.Duration(RefreshSPInfoTimer) * time.Second)
	for {
		select {
		case <-refreshStorageParamsTicker.C:
			go m.refreshStorageParams()
		case <-refreshSPInfoTicker.C:
			go m.refreshSPInfo()
		case <-m.stopCh:
			return
		}
	}
}

// refreshStorageParams fetch storage params from chain and update to spdb
func (m *Manager) refreshStorageParams() {
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

// refreshSPInfo fetch sp info from chain and update to spdb
func (m *Manager) refreshSPInfo() {
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
		if spInfo.OperatorAddress == m.cfg.SpOperatorAddress {
			if err = m.spDB.SetOwnSpInfo(spInfo); err != nil {
				log.Errorw("failed to set own sp info", "error", err)
				return
			}
		}
	}
	log.Infow("succeed to refresh sp info", "sp_info", spInfoList)
}

// Stop manager service, implement the lifecycle interface
func (m *Manager) Stop(ctx context.Context) error {
	if !m.running.Swap(false) {
		return errors.New("manager has already stop")
	}
	close(m.stopCh)
	return nil
}
