package retriver

import (
	"google.golang.org/grpc/reflection"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/modular/retriever/types"
)

const (
	DefaultQuerySPParallelPerNode uint64 = 10240
)

func init() {
	gfspapp.RegisterModularInfo(RetrieveModularName, RetrieveModularDescription, NewRetrieveModular)
}

func NewRetrieveModular(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig,
	opts ...gfspapp.Option) (
	coremodule.Modular, error) {
	if cfg.Retriever != nil {
		return cfg.Retriever, nil
	}
	receiver := &RetrieveModular{baseApp: app}
	opts = append(opts, receiver.DefaultRetrieverOptions)
	for _, opt := range opts {
		if err := opt(app, cfg); err != nil {
			return nil, err
		}
	}
	return receiver, nil
}

func (r *RetrieveModular) DefaultRetrieverOptions(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig) error {
	if cfg.QuerySPParallelPerNode == 0 {
		cfg.QuerySPParallelPerNode = DefaultQuerySPParallelPerNode
	}
	r.retrievingRequest = cfg.QuerySPParallelPerNode

	types.RegisterGfSpRetrieverServiceServer(r.baseApp.ServerForRegister(), r)
	reflection.Register(r.baseApp.ServerForRegister())
	return nil
}
