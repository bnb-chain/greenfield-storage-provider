package metadata

import (
	"context"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata"
)

var (
	MetadataModularName        = strings.ToLower("Metadata")
	MetadataModularDescription = "Retrieves sp metadata and info."
)

const (
	DefaultMetadataStatisticsInterval = 60
)

var _ module.Modular = &MetadataModular{}

type MetadataModular struct {
	baseApp *gfspapp.GfSpBaseApp
	scope   rcmgr.ResourceScope

	// freeQuotaPerBucket defines the free read quota per bucket
	freeQuotaPerBucket uint64
	// maxMetadataRequest defines the max handling metadata request number
	maxMetadataRequest int64
	// retrievingRequest defines the handling retrieve request number
	retrievingRequest int64
	config            *metadata.MetadataConfig
	dbSwitchTicker    *time.Ticker
}

func (r *MetadataModular) Name() string {
	return MetadataModularName
}

func (r *MetadataModular) Start(ctx context.Context) error {
	// Default the bsDB to master db at start
	r.baseApp.SetGfBsDB(r.baseApp.GfBsDBMaster())
	// Start the timed listener to switch the database
	r.startDBSwitchListener(time.Second * time.Duration(r.config.BsDBSwitchCheckIntervalSec))
	scope, err := r.baseApp.ResourceManager().OpenService(r.Name())
	if err != nil {
		return err
	}
	r.scope = scope
	return nil
}

func (r *MetadataModular) Stop(ctx context.Context) error {
	r.scope.Release()
	r.dbSwitchTicker.Stop()
	return nil
}

func (r *MetadataModular) ReserveResource(
	ctx context.Context,
	state *rcmgr.ScopeStat) (
	rcmgr.ResourceScopeSpan, error) {
	return &rcmgr.NullScope{}, nil
}

func (r *MetadataModular) ReleaseResource(
	ctx context.Context,
	span rcmgr.ResourceScopeSpan) {
	span.Done()
}

// startDBSwitchListener sets up a ticker to periodically check for a database switch signal.
// If a signal is detected, it triggers the switchDB() method to switch to the new database.
// The ticker is stopped when the Metadata gRPC service is stopped, ensuring that
// resources are properly managed and released.
func (r *MetadataModular) startDBSwitchListener(switchInterval time.Duration) {
	// create a ticker to periodically check for a new database name
	r.dbSwitchTicker = time.NewTicker(switchInterval)
	// set the bsdb to be master db at start
	r.config.IsMasterDB = true
	// launch a goroutine to handle the ticker events
	go func() {
		// check once at the start of the system
		r.checkSignal()
		// loop until the context is canceled (e.g., when the Metadata service is stopped)
		for range r.dbSwitchTicker.C {
			r.checkSignal()
		}
	}()
}

func (r *MetadataModular) checkSignal() {
	// check if there is a signal from block syncer database to switch the database
	signal, err := r.baseApp.GfBsDB().GetSwitchDBSignal()
	if err != nil || signal == nil {
		log.Errorw("failed to get switch db signal", "err", err)
	}
	log.Debugf("switchDB check: signal: %t and IsMasterDB: %t", signal.IsMaster, r.config.IsMasterDB)
	// if a signal db is not equal to current metadata db, attempt to switch the database
	if signal.IsMaster != r.config.IsMasterDB {
		r.switchDB(signal.IsMaster)
	}
}

// switchDB is responsible for switching between the primary and backup Block Syncer databases.
// Depending on the current value of the IsMasterDB in the Metadata configuration, it switches
// the active Block Syncer database to the backup or the primary database.
// After switching, it toggles the value of the IsMasterDB to indicate the active database.
func (r *MetadataModular) switchDB(flag bool) {
	if flag {
		r.baseApp.SetGfBsDB(r.baseApp.GfBsDBMaster())
	} else {
		r.baseApp.SetGfBsDB(r.baseApp.GfBsDBBackup())
	}
	r.config.IsMasterDB = flag
	log.Info("db switched successfully")
}
