package approver

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspmdmgr"
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

func init() {
	gfspmdmgr.RegisterModularInfo(ApprovalModularName, ApprovalModularDescription, NewApprovalModular)
}

func NewApprovalModular(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig,
	opts ...gfspconfig.Option) (
	coremodule.Modular, error) {
	if cfg.Approver != nil {
		app.SetApprover(cfg.Approver)
		return cfg.Approver, nil
	}
	approver := &ApprovalModular{baseApp: app}
	opts = append(opts, approver.DefaultApprovalOptions)
	for _, opt := range opts {
		if err := opt(app, cfg); err != nil {
			return nil, err
		}
	}
	app.SetApprover(approver)
	return approver, nil
}

func (a *ApprovalModular) DefaultApprovalOptions(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig) error {
	if cfg.AccountBucketNumber == 0 {
		cfg.AccountBucketNumber = DefaultAccountBucketNumber
	}
	a.accountBucketNumber = cfg.AccountBucketNumber
	if cfg.BucketApprovalTimeoutHeight == uint64(0) {
		cfg.BucketApprovalTimeoutHeight = DefaultBucketApprovalTimeoutHeight
	}
	a.bucketApprovalTimeoutHeight = cfg.BucketApprovalTimeoutHeight
	if cfg.ObjectApprovalTimeoutHeight == uint64(0) {
		cfg.ObjectApprovalTimeoutHeight = DefaultObjectApprovalTimeoutHeight
	}
	a.objectApprovalTimeoutHeight = cfg.ObjectApprovalTimeoutHeight
	if cfg.CreateBucketApprovalParallel == 0 {
		cfg.CreateBucketApprovalParallel = DefaultCreateBucketApprovalParallel
	}
	if cfg.CreateObjectApprovalParallel == 0 {
		cfg.CreateObjectApprovalParallel = DefaultCreateObjectApprovalParallel
	}
	a.bucketQueue = cfg.NewStrategyTQueueFunc(a.Name()+CreateBucketApprovalQueueSuffix,
		cfg.CreateBucketApprovalParallel)
	a.objectQueue = cfg.NewStrategyTQueueFunc(a.Name()+CreateObjectApprovalQueueSuffix,
		cfg.CreateObjectApprovalParallel)
	return nil
}
