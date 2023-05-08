package retriver

import (
	"context"

	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/modular/retriever/types"
)

const (
	RetrieveModularName        = "retriever"
	RetrieveModularDescription = "retrieve modular supports retrieve sp meta "
)

var _ module.Modular = &RetrieveModular{}

type RetrieveModular struct {
	endpoint string
	baseApp  *gfspapp.GfSpBaseApp
	scope    rcmgr.ResourceScope

	maxRetrieveRequest uint64
	retrievingRequest  uint64
}

func (r *RetrieveModular) Name() string {
	return RetrieveModularName
}

func (r *RetrieveModular) Start(ctx context.Context) error {
	types.RegisterGfSpRetrieverServiceServer(r.baseApp.ServerForRegister(), r)
	reflection.Register(r.baseApp.ServerForRegister())

	return nil
}

func (r *RetrieveModular) Stop(ctx context.Context) error {
	r.scope.Release()
	return nil
}

func (r *RetrieveModular) Description() string {
	return RetrieveModularDescription
}

func (r *RetrieveModular) Endpoint() string {
	return r.endpoint
}

func (r *RetrieveModular) ReserveResource(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
	return &rcmgr.NullScope{}, nil
}

func (r *RetrieveModular) ReleaseResource(ctx context.Context, span rcmgr.ResourceScopeSpan) {
	return
}
