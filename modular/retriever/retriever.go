package retriver

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/model"
)

var (
	RetrieveModularName        = model.RetrieveModular
	RetrieveModularDescription = model.SpServiceDesc[model.RetrieveModular]
)

var _ module.Modular = &RetrieveModular{}

type RetrieveModular struct {
	baseApp *gfspapp.GfSpBaseApp
	scope   rcmgr.ResourceScope

	maxRetrieveRequest uint64
	retrievingRequest  uint64
}

func (r *RetrieveModular) Name() string {
	return RetrieveModularName
}

func (r *RetrieveModular) Start(ctx context.Context) error {
	return nil
}

func (r *RetrieveModular) Stop(ctx context.Context) error {
	r.scope.Release()
	return nil
}

func (r *RetrieveModular) ReserveResource(
	ctx context.Context,
	state *rcmgr.ScopeStat) (
	rcmgr.ResourceScopeSpan, error) {
	return &rcmgr.NullScope{}, nil
}

func (r *RetrieveModular) ReleaseResource(
	ctx context.Context,
	span rcmgr.ResourceScopeSpan) {
	return
}
