package metadata

import (
	"runtime"
	"time"

	"github.com/zkMeLabs/mechain-storage-provider/base/gfspapp"
	"github.com/zkMeLabs/mechain-storage-provider/base/gfspconfig"
	coremodule "github.com/zkMeLabs/mechain-storage-provider/core/module"
	"github.com/zkMeLabs/mechain-storage-provider/modular/metadata/types"
	"github.com/zkMeLabs/mechain-storage-provider/pkg/metrics"
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

	managerInfo  types.ManagerInfo
	executorInfo types.ExecutorInfo
	gcInfo       types.GCInfo

	MonthlyFreeQuota uint64
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

	metadata.maxMetadataRequest = cfg.Parallel.QuerySPParallelPerNode

	metadata.baseApp.SetGfBsDB(metadata.baseApp.GfBsDBMaster())

	BsModules = cfg.BlockSyncer.Modules
	BsWorkers = cfg.BlockSyncer.Workers
	ChainID = cfg.Chain.ChainID
	ChainAddress = cfg.Chain.ChainAddress
	SpOperatorAddress = cfg.SpAccount.SpOperatorAddress
	GatewayDomainName = cfg.Gateway.DomainName

	managerInfo.EnableLoadTask = cfg.Manager.EnableLoadTask
	managerInfo.EnableHealthChecker = cfg.Manager.EnableHealthyChecker
	managerInfo.EnableTaskRetryScheduler = cfg.Manager.EnableTaskRetryScheduler

	executorInfo.ListenSealRetryTimeout = int64(cfg.Executor.ListenSealRetryTimeout)
	executorInfo.BucketTrafficKeepTimeDay = int64(cfg.Executor.BucketTrafficKeepTimeDay)
	executorInfo.ReadRecordKeepTimeDay = int64(cfg.Executor.ReadRecordKeepTimeDay)

	gcInfo.EnableGcZombie = cfg.GC.EnableGCZombie
	gcInfo.EnableGcMeta = cfg.GC.EnableGCMeta
	gcInfo.GcMetaTimeInterval = int64(cfg.GC.GCMetaTimeInterval)

	if cfg.Quota.MonthlyFreeQuota == 0 {
		MonthlyFreeQuota = gfspapp.DefaultSpMonthlyFreeQuota
	} else {
		MonthlyFreeQuota = cfg.Quota.MonthlyFreeQuota
	}

	startGoRoutineListener()
	return nil
}

// startGoRoutineListener sets up a ticker to periodically check for go routine count of metadata service.
func startGoRoutineListener() {
	go func() {
		dbSwitchTicker := time.NewTicker(500 * time.Millisecond)
		for range dbSwitchTicker.C {
			metrics.GoRoutineCount.Set(float64(runtime.NumGoroutine()))
		}
	}()
}
