package retriever

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/modular/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/modular/retriever/types"
)

const (
	// DefaultQuerySPParallelPerNode defines the max parallel for retrieving request
	DefaultQuerySPParallelPerNode int64 = 10240
)

func NewRetrieveModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	receiver := &RetrieveModular{baseApp: app}
	if err := DefaultRetrieverOptions(receiver, cfg); err != nil {
		return nil, err
	}
	// register retrieve service to gfsp base app's grpc server
	types.RegisterGfSpRetrieverServiceServer(receiver.baseApp.ServerForRegister(), receiver)
	return receiver, nil
}

func DefaultRetrieverOptions(receiver *RetrieveModular, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Parallel.QuerySPParallelPerNode == 0 {
		cfg.Parallel.QuerySPParallelPerNode = DefaultQuerySPParallelPerNode
	}
	if cfg.Bucket.FreeQuotaPerBucket == 0 {
		cfg.Bucket.FreeQuotaPerBucket = downloader.DefaultBucketFreeQuota
	}
	receiver.freeQuotaPerBucket = cfg.Bucket.FreeQuotaPerBucket
	receiver.maxRetrieveRequest = cfg.Parallel.QuerySPParallelPerNode
	return nil
}
