package retriver

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspmdmgr"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	DefaultQuerySPParallelPerNode uint64 = 10240
)

func init() {
	gfspmdmgr.RegisterModularInfo(RetrieveModularName, RetrieveModularDescription, NewRetrieveModular)
}

func NewRetrieveModular(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig,
	opts ...gfspconfig.Option) (
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
	return nil
}
