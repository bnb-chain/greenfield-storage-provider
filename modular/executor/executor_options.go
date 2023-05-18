package executor

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	DefaultExecutorMaxExecuteNum                int64   = 1024
	DefaultExecutorAskTaskInterval              int     = 1
	DefaultExecutorAskReplicateApprovalTimeout  int64   = 10
	DefaultExecutorAskReplicateApprovalExFactor float64 = 1.0
	DefaultExecutorListenSealTimeoutHeight      int     = 10
	DefaultExecutorListenSealRetryTimeout       int     = 3 * 2
	DefaultExecutorMaxListenSealRetry           int     = 3
	DefaultStatisticsOutputInterval             int     = 60
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
	if cfg.Executor.AskReplicateApprovalExFactor == 0 {
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
