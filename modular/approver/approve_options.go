package approver

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	DefaultAccountBucketNumber                 = 100
	DefaultBucketApprovalTimeoutHeight  uint64 = 10
	DefaultObjectApprovalTimeoutHeight  uint64 = 10
	DefaultCreateBucketApprovalParallel        = 1024
	DefaultCreateObjectApprovalParallel        = 1024
	CreateBucketApprovalQueueSuffix            = "-create-bucket-approval"
	CreateObjectApprovalQueueSuffix            = "-create-object-approval"
)

func NewApprovalModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	approver := &ApprovalModular{baseApp: app}
	if err := DefaultApprovalOptions(cfg, approver); err != nil {
		return nil, err
	}
	return approver, nil
}

func DefaultApprovalOptions(cfg *gfspconfig.GfSpConfig, approver *ApprovalModular) error {
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
		approver.Name()+CreateBucketApprovalQueueSuffix,
		cfg.Parallel.GlobalCreateBucketApprovalParallel)
	approver.objectQueue = cfg.Customize.NewStrategyTQueueFunc(
		approver.Name()+CreateObjectApprovalQueueSuffix,
		cfg.Parallel.GlobalCreateObjectApprovalParallel)
	return nil
}
