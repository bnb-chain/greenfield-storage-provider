package retriever

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/modular/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/modular/retriever/types"
	"github.com/bnb-chain/greenfield-storage-provider/service/metadata"
)

const (
	// DefaultQuerySPParallelPerNode defines the max parallel for retrieving request
	DefaultQuerySPParallelPerNode int64 = 10240
)

func NewRetrieveModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	retriever := &RetrieveModular{baseApp: app}
	if err := DefaultRetrieverOptions(retriever, cfg); err != nil {
		return nil, err
	}
	// register retrieve service to gfsp base app's grpc server
	types.RegisterGfSpRetrieverServiceServer(retriever.baseApp.ServerForRegister(), retriever)
	return retriever, nil
}

func DefaultRetrieverOptions(retriever *RetrieveModular, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Parallel.QuerySPParallelPerNode == 0 {
		cfg.Parallel.QuerySPParallelPerNode = DefaultQuerySPParallelPerNode
	}
	if cfg.Bucket.FreeQuotaPerBucket == 0 {
		cfg.Bucket.FreeQuotaPerBucket = downloader.DefaultBucketFreeQuota
	}
	retriever.freeQuotaPerBucket = cfg.Bucket.FreeQuotaPerBucket
	retriever.maxRetrieveRequest = cfg.Parallel.QuerySPParallelPerNode

	metadataConfig := &metadata.MetadataConfig{
		BsDBConfig:         &cfg.BsDB,
		BsDBSwitchedConfig: &cfg.BsDBBackup,
	}

	retriever.config = metadataConfig
	return nil
}
