package manager

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	gnfd "github.com/bnb-chain/greenfield-storage-provider/pkg/greenfield"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/sqldb"
)

var _ lifecycle.Service = &Manager{}

var (
	// RefreshSPInfoAndStorageParamsTimer define the period of refresh sp info and storage params
	RefreshSPInfoAndStorageParamsTimer = 5 * 60
)

// Manager module is responsible for implementing internal management functions.
// Currently, it supports periodic update of sp info list and storage params information in spdb.
// TODO::support gc and configuration management, etc.
type Manager struct {
	cfg     *ManagerConfig
	running atomic.Value
	stopCh  chan struct{}
	chain   *gnfd.Greenfield
	spDB    sqldb.SPDB
}

// NewManagerService returns an instance of manager
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
	return nil
}

// eventLoop background goroutine, responsible for refreshing sp info and storage params
func (m *Manager) eventLoop() {
	m.refreshSPInfoAndStorageParams()
	refreshSPInfoAndStorageParamsTicker := time.NewTicker(time.Duration(RefreshSPInfoAndStorageParamsTimer) * time.Second)
	for {
		select {
		case <-refreshSPInfoAndStorageParamsTicker.C:
			go m.refreshSPInfoAndStorageParams()
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
		if spInfo.OperatorAddress == m.cfg.SpOperatorAddress {
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
