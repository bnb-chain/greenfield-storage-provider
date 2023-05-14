package executor

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspmdmgr"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	DefaultExecutorMaxExecuteNum                uint64  = 1024
	DefaultExecutorAskTaskInterval              int     = 1
	DefaultExecutorAskReplicateApprovalTimeout  int64   = 10
	DefaultExecutorAskReplicateApprovalExFactor float64 = 1.0
	DefaultExecutorListenSealTimeoutHeight      int     = 10
	DefaultExecutorListenSealRetryTimeout       int     = 3 * 2
	DefaultExecutorMaxListenSealRetry           int     = 3
)

func init() {
	gfspmdmgr.RegisterModularInfo(ExecuteModularName, ExecuteModularDescription, NewExecuteModular)
}

func NewExecuteModular(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig,
	opts ...gfspconfig.Option) (
	coremodule.Modular, error) {
	if cfg.TaskExecutor != nil {
		app.SetTaskExecutor(cfg.TaskExecutor)
		return cfg.TaskExecutor, nil
	}
	executor := &ExecuteModular{baseApp: app}
	opts = append(opts, executor.DefaultExecutorOptions)
	for _, opt := range opts {
		if err := opt(app, cfg); err != nil {
			return nil, err
		}
	}
	app.SetTaskExecutor(executor)
	return executor, nil
}

func (e *ExecuteModular) DefaultExecutorOptions(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig) error {
	if cfg.ExecutorMaxExecuteNum == 0 {
		cfg.ExecutorMaxExecuteNum = DefaultExecutorMaxExecuteNum
	}
	if cfg.ExecutorAskTaskInterval == 0 {
		cfg.ExecutorAskTaskInterval = DefaultExecutorAskTaskInterval
	}
	if cfg.ExecutorAskReplicateApprovalTimeout == 0 {
		cfg.ExecutorAskReplicateApprovalTimeout = DefaultExecutorAskReplicateApprovalTimeout
	}
	if cfg.ExecutorAskReplicateApprovalExFactor == 0 {
		cfg.ExecutorAskReplicateApprovalExFactor = DefaultExecutorAskReplicateApprovalExFactor
	}
	if cfg.ExecutorListenSealTimeoutHeight == 0 {
		cfg.ExecutorListenSealTimeoutHeight = DefaultExecutorListenSealTimeoutHeight
	}
	if cfg.ExecutorListenSealRetryTimeout == 0 {
		cfg.ExecutorListenSealRetryTimeout = DefaultExecutorListenSealRetryTimeout
	}
	if cfg.ExecutorMaxListenSealRetry == 0 {
		cfg.ExecutorMaxListenSealRetry = DefaultExecutorMaxListenSealRetry
	}
	return nil
}
