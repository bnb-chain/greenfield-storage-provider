package executor

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	// DefaultExecutorMaxExecuteNum defines the default max parallel execute task number.
	DefaultExecutorMaxExecuteNum int64 = 1024
	// DefaultExecutorAskTaskInterval defines the default ask task interval from manager.
	DefaultExecutorAskTaskInterval int = 1
	// DefaultExecutorAskReplicateApprovalTimeout defines the ask replicate piece approval
	// timeout that send the request to the p2p node,
	DefaultExecutorAskReplicateApprovalTimeout int64 = 10
	// DefaultExecutorAskReplicateApprovalExFactor defines the expanded factor for asking
	// secondary SP.
	// Example: need data chunk + data parity chunk numbers SPs as secondary, consider fault
	// 	tolerance, should collect (data chunk + data parity chunk) * factor numbers SPs as
	//	backup secondary, if some of these are failed to replicate can pick up again from
	//  backups. So it is always bigger than 1.0.
	DefaultExecutorAskReplicateApprovalExFactor float64 = 1.0
	// DefaultExecutorListenSealTimeoutHeight defines the default listen seal object on
	// greenfield timeout height, if after current block height + timeout height, the object
	// is not sealed, it is judged failed to seal object on greenfield.
	DefaultExecutorListenSealTimeoutHeight int = 10
	// DefaultExecutorListenSealRetryTimeout defines the sleep time when listen seal object
	// fail, until retry ExecutorMaxListenSealRetry times, the task is set error.
	DefaultExecutorListenSealRetryTimeout int = 2
	// DefaultExecutorMaxListenSealRetry defines the default max retry number for listening
	// object.
	DefaultExecutorMaxListenSealRetry int = 3
	// DefaultStatisticsOutputInterval defines the default interval for output statistics info,
	// it is used to log and debug.
	DefaultStatisticsOutputInterval int = 60
)

func NewExecuteModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	executor := &ExecuteModular{baseApp: app}
	if err := DefaultExecutorOptions(executor, cfg); err != nil {
		return nil, err
	}
	return executor, nil
}

func DefaultExecutorOptions(executor *ExecuteModular, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Executor.MaxExecuteNumber == 0 {
		cfg.Executor.MaxExecuteNumber = DefaultExecutorMaxExecuteNum
	}
	executor.maxExecuteNum = cfg.Executor.MaxExecuteNumber
	if cfg.Executor.AskTaskInterval == 0 {
		cfg.Executor.AskTaskInterval = DefaultExecutorAskTaskInterval
	}
	executor.askTaskInterval = cfg.Executor.AskTaskInterval
	if cfg.Executor.AskReplicateApprovalTimeout == 0 {
		cfg.Executor.AskReplicateApprovalTimeout = DefaultExecutorAskReplicateApprovalTimeout
	}
	executor.askReplicateApprovalTimeout = cfg.Executor.AskReplicateApprovalTimeout
	if cfg.Executor.AskReplicateApprovalExFactor < 1.0 {
		cfg.Executor.AskReplicateApprovalExFactor = DefaultExecutorAskReplicateApprovalExFactor
	}
	executor.askReplicateApprovalExFactor = cfg.Executor.AskReplicateApprovalExFactor
	if cfg.Executor.ListenSealTimeoutHeight == 0 {
		cfg.Executor.ListenSealTimeoutHeight = DefaultExecutorListenSealTimeoutHeight
	}
	executor.listenSealTimeoutHeight = cfg.Executor.ListenSealTimeoutHeight
	if cfg.Executor.ListenSealRetryTimeout == 0 {
		cfg.Executor.ListenSealRetryTimeout = DefaultExecutorListenSealRetryTimeout
	}
	executor.listenSealRetryTimeout = cfg.Executor.ListenSealRetryTimeout
	if cfg.Executor.MaxListenSealRetry == 0 {
		cfg.Executor.MaxListenSealRetry = DefaultExecutorMaxListenSealRetry
	}
	executor.maxListenSealRetry = cfg.Executor.MaxListenSealRetry
	executor.statisticsOutputInterval = DefaultStatisticsOutputInterval
	return nil
}
