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

func NewRetrieveModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	if cfg.Customize.Retriever != nil {
		return cfg.Customize.Retriever, nil
	}
	receiver := &RetrieveModular{baseApp: app}
	if err := DefaultRetrieverOptions(receiver, cfg); err != nil {
		return nil, err
	}
	types.RegisterGfSpRetrieverServiceServer(receiver.baseApp.ServerForRegister(), receiver)
	reflection.Register(receiver.baseApp.ServerForRegister())
	return receiver, nil
}

func DefaultRetrieverOptions(receiver *RetrieveModular, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Parallel.QuerySPParallelPerNode == 0 {
		cfg.Parallel.QuerySPParallelPerNode = DefaultQuerySPParallelPerNode
	}
	receiver.retrievingRequest = cfg.Parallel.QuerySPParallelPerNode
	return nil
}
