package retriever

import (
	"context"
	"strings"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
)

var (
	RetrieveModularName        = strings.ToLower("Retriever")
	RetrieveModularDescription = "Retrieves sp metadata and info."
)

var _ module.Modular = &RetrieveModular{}

type RetrieveModular struct {
	baseApp *gfspapp.GfSpBaseApp
	scope   rcmgr.ResourceScope

	freeQuotaPerBucket uint64
	maxRetrieveRequest int64
	retrievingRequest  int64
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
