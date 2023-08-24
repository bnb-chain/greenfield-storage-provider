package metadata

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/modular/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"
)

const (
	// DefaultQuerySPParallelPerNode defines the max parallel for retrieving request
	DefaultQuerySPParallelPerNode int64 = 10240
	// DefaultBsDBSwitchCheckIntervalSec defines the default db switch check interval in seconds
	DefaultBsDBSwitchCheckIntervalSec = 30
)

var (
	BsModules []string
	BsWorkers uint

	ChainID      string
	ChainAddress []string

	SpOperatorAddress string
	GatewayDomainName string
)

func NewMetadataModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	metadata := &MetadataModular{baseApp: app}
	if err := DefaultMetadataOptions(metadata, cfg); err != nil {
		return nil, err
	}
	// register metadata service to gfsp base app's grpc server
	types.RegisterGfSpMetadataServiceServer(metadata.baseApp.ServerForRegister(), metadata)
	return metadata, nil
}

func DefaultMetadataOptions(metadata *MetadataModular, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Parallel.QuerySPParallelPerNode == 0 {
		cfg.Parallel.QuerySPParallelPerNode = DefaultQuerySPParallelPerNode
	}
	if cfg.Bucket.FreeQuotaPerBucket == 0 {
		cfg.Bucket.FreeQuotaPerBucket = downloader.DefaultBucketFreeQuota
	}

	metadata.freeQuotaPerBucket = cfg.Bucket.FreeQuotaPerBucket
	metadata.maxMetadataRequest = cfg.Parallel.QuerySPParallelPerNode

	metadata.baseApp.SetGfBsDB(metadata.baseApp.GfBsDBMaster())

	BsModules = cfg.BlockSyncer.Modules
	//BsEnableDualDB = cfg.BlockSyncer.EnableDualDB
	BsWorkers = cfg.BlockSyncer.Workers
	ChainID = cfg.Chain.ChainID
	ChainAddress = cfg.Chain.ChainAddress
	SpOperatorAddress = cfg.SpAccount.SpOperatorAddress
	GatewayDomainName = cfg.Gateway.DomainName

	//startDBSwitchListener(time.Second*time.Duration(cfg.Metadata.BsDBSwitchCheckIntervalSec), cfg, metadata)

	return nil
}

// startDBSwitchListener sets up a ticker to periodically check for a database switch signal.
// If a signal is detected, it triggers the switchDB() method to switch to the new database.
// The ticker is stopped when the Metadata gRPC service is stopped, ensuring that
// resources are properly managed and released.
//func startDBSwitchListener(switchInterval time.Duration, cfg *gfspconfig.GfSpConfig, metadata *MetadataModular) {
//	// create a ticker to periodically check for a new database name
//	dbSwitchTicker := time.NewTicker(switchInterval)
//	checkSignal(cfg, metadata)
//	// launch a goroutine to handle the ticker events
//	go func() {
//		// loop until the context is canceled (e.g., when the Metadata service is stopped)
//		for range dbSwitchTicker.C {
//			checkSignal(cfg, metadata)
//		}
//	}()
//}

//func checkSignal(cfg *gfspconfig.GfSpConfig, metadata *MetadataModular) {
//	// check if there is a signal from block syncer database to switch the database
//	signal, err := metadata.baseApp.GfBsDBMaster().GetSwitchDBSignal()
//	if err != nil || signal == nil {
//		log.Errorw("failed to get switch db signal", "err", err)
//	}
//	// if a signal db is not equal to current metadata db, attempt to switch the database
//	//if signal.IsMaster != cfg.Metadata.IsMasterDB {
//	//	switchDB(signal.IsMaster, cfg, metadata)
//	//}
//}

// switchDB is responsible for switching between the primary and backup Block Syncer databases.
// Depending on the current value of the IsMasterDB in the Metadata configuration, it switches
// the active Block Syncer database to the backup or the primary database.
// After switching, it toggles the value of the IsMasterDB to indicate the active database.
//func switchDB(flag bool, cfg *gfspconfig.GfSpConfig, metadata *MetadataModular) {
//	if flag {
//		metadata.baseApp.SetGfBsDB(metadata.baseApp.GfBsDBMaster())
//	} else {
//		metadata.baseApp.SetGfBsDB(metadata.baseApp.GfBsDBBackup())
//	}
//	cfg.Metadata.IsMasterDB = flag
//	log.Info("db switched successfully")
//}
