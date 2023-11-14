package executor

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	// DefaultExecutorMaxExecuteNum defines the default max parallel execute task number.
	DefaultExecutorMaxExecuteNum int64 = 64
	// DefaultExecutorAskTaskInterval defines the default ask task interval from manager.
	DefaultExecutorAskTaskInterval int = 1
	// DefaultExecutorAskReplicateApprovalTimeout defines the ask replicate piece approval
	// timeout that send the request to the p2p node,
	DefaultExecutorAskReplicateApprovalTimeout int64 = 4
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
	// DefaultExecutorObjectMigrationRetryTimeout defines the sleep time when object migration
	// fail, until retry DefaultExecutorMaxObjectMigrationRetry times, the task is set error.
	DefaultExecutorObjectMigrationRetryTimeout int = 2
	// DefaultExecutorMaxObjectMigrationRetry defines the default max retry number for object migration.
	DefaultExecutorMaxObjectMigrationRetry int = 5
	// DefaultStatisticsOutputInterval defines the default interval for output statistics info,
	// it is used to log and debug.
	DefaultStatisticsOutputInterval int = 60
	// DefaultSleepInterval defines the sleep interval when failed to ask task
	// it is millisecond level
	DefaultSleepInterval = 100
)

const (
	ExecutorSuccessAskTask   = "executor_ask_task_success"
	ExecutorRunTask          = "executor_run_task"
	ExecutorFailureAskTask   = "executor_ask_task_failure"
	ExecutorFailureAskNoTask = "executor_ask_no_task_failure"

	ExecutorSuccessReplicateTask  = "executor_replicate_task_success"
	ExecutorFailureReplicateTask  = "executor_replicate_task_failure"
	ExecutorSuccessSealObjectTask = "executor_seal_object_task_success"
	ExecutorFailureSealObjectTask = "executor_seal_object_task_failure"
	ExecutorSuccessReceiveTask    = "executor_receive_task_success"
	ExecutorFailureReceiveTask    = "executor_receive_task_failure"
	ExecutorSuccessRecoveryTask   = "executor_recovery_task_success"
	ExecutorFailureRecoveryTask   = "executor_recovery_task_failure"

	ExecutorSuccessReportTask = "executor_report_task_to_manager_success"
	ExecutorFailureReportTask = "executor_report_task_to_manager_failure"

	ExecutorSuccessP2P                = "executor_p2p_success"
	ExecutorFailureP2P                = "executor_p2p_failure"
	ExecutorSuccessReplicateAllPiece  = "executor_replicate_all_piece_success"
	ExecutorFailureReplicateAllPiece  = "executor_replicate_all_piece_failure"
	ExecutorSuccessReplicateOnePiece  = "executor_replicate_one_piece_success"
	ExecutorFailureReplicateOnePiece  = "executor_replicate_one_piece_failure"
	ExecutorSuccessDoneReplicatePiece = "executor_done_replicate_piece_success"
	ExecutorFailureDoneReplicatePiece = "executor_done_replicate_piece_failure"
	ExecutorSuccessSealObject         = "executor_seal_object_success"
	ExecutorFailureSealObject         = "executor_seal_object_failure"
)

func NewExecuteModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	executor := &ExecuteModular{baseApp: app}
	defaultExecutorOptions(executor, cfg)
	return executor, nil
}

func defaultExecutorOptions(executor *ExecuteModular, cfg *gfspconfig.GfSpConfig) {
	if cfg.Executor.MaxExecuteNumber == 0 {
		// TODO:: DefaultExecutorMaxExecuteNum should core_num * multiple, the core_num is compatible with docker
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

	if cfg.Executor.ObjectMigrationRetryTimeout == 0 {
		cfg.Executor.ObjectMigrationRetryTimeout = DefaultExecutorObjectMigrationRetryTimeout
	}
	if cfg.Executor.MaxObjectMigrationRetry == 0 {
		cfg.Executor.MaxObjectMigrationRetry = DefaultExecutorMaxObjectMigrationRetry
	}

	executor.maxListenSealRetry = cfg.Executor.MaxListenSealRetry
	executor.statisticsOutputInterval = DefaultStatisticsOutputInterval
	executor.enableSkipFailedToMigrateObject = cfg.Executor.EnableSkipFailedToMigrateObject
	executor.objectMigrationRetryTimeout = cfg.Executor.ObjectMigrationRetryTimeout
	executor.maxObjectMigrationRetry = cfg.Executor.MaxObjectMigrationRetry
}
