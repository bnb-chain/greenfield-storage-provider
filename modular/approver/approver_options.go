package approver

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/modular/manager"
)

const (
	// DefaultAccountBucketNumber defines the default value of bucket number is
	// owned by the same account
	DefaultAccountBucketNumber = 100
	// DefaultBucketApprovalTimeoutHeight defines the default value of timeout
	// height for creating bucket approval
	DefaultBucketApprovalTimeoutHeight uint64 = 100
	// DefaultObjectApprovalTimeoutHeight defines the default value of timeout
	//	// height for creating object approval
	DefaultObjectApprovalTimeoutHeight uint64 = 100
	// DefaultCreateBucketApprovalParallel defines the default value of parallel
	// for approved create bucket per approver
	DefaultCreateBucketApprovalParallel = 10240
	// DefaultCreateObjectApprovalParallel defines the default value of parallel
	// for approved create object per approver
	DefaultCreateObjectApprovalParallel = 10240
)

func NewApprovalModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	approver := &ApprovalModular{baseApp: app}
	defaultApprovalOptions(cfg, approver)
	return approver, nil
}

func defaultApprovalOptions(cfg *gfspconfig.GfSpConfig, approver *ApprovalModular) {
	if cfg.Bucket.AccountBucketNumber == 0 {
		cfg.Bucket.AccountBucketNumber = DefaultAccountBucketNumber
	}
	approver.accountBucketNumber = cfg.Bucket.AccountBucketNumber
	if cfg.Approval.BucketApprovalTimeoutHeight == uint64(0) {
		cfg.Approval.BucketApprovalTimeoutHeight = DefaultBucketApprovalTimeoutHeight
	}
	approver.bucketApprovalTimeoutHeight = cfg.Approval.BucketApprovalTimeoutHeight
	if cfg.Approval.ObjectApprovalTimeoutHeight == uint64(0) {
		cfg.Approval.ObjectApprovalTimeoutHeight = DefaultObjectApprovalTimeoutHeight
	}
	approver.objectApprovalTimeoutHeight = cfg.Approval.ObjectApprovalTimeoutHeight
	if cfg.Parallel.GlobalCreateBucketApprovalParallel == 0 {
		cfg.Parallel.GlobalCreateBucketApprovalParallel = DefaultCreateBucketApprovalParallel
	}
	if cfg.Parallel.GlobalCreateObjectApprovalParallel == 0 {
		cfg.Parallel.GlobalCreateObjectApprovalParallel = DefaultCreateObjectApprovalParallel
	}
	approver.bucketQueue = cfg.Customize.NewStrategyTQueueFunc(
		approver.Name()+"-create-bucket-approval", cfg.Parallel.GlobalCreateBucketApprovalParallel)
	approver.objectQueue = cfg.Customize.NewStrategyTQueueFunc(
		approver.Name()+"-create-object-approval", cfg.Parallel.GlobalCreateObjectApprovalParallel)
	if cfg.Parallel.GlobalMigrateGVGParallel == 0 {
		cfg.Parallel.GlobalMigrateGVGParallel = manager.DefaultGlobalMigrateGVGParallel
	}
	approver.migrateGVGLimit = cfg.Parallel.GlobalMigrateGVGParallel
}
