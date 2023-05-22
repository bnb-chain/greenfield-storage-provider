package metadata

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/modular/downloader"
	"github.com/bnb-chain/greenfield-storage-provider/modular/metadata/types"

	servicemeta "github.com/bnb-chain/greenfield-storage-provider/service/metadata"
)

const (
	// DefaultQuerySPParallelPerNode defines the max parallel for retrieving request
	DefaultQuerySPParallelPerNode int64 = 10240
)

func NewMetadataModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	metadata := &MetadataModular{baseApp: app}
	if err := DefaultMetadataOptions(metadata, cfg); err != nil {
		return nil, err
	}
	// register metadata service to gfsp base app's grpc server
	types.RegisterGfSpMetadataServiceServer(metadata.baseApp.ServerForRegister(), metadata)
	return metadata, nil
}

func DefaultMetadataOptions(metadata *MetadataModular, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Parallel.QuerySPParallelPerNode == 0 {
		cfg.Parallel.QuerySPParallelPerNode = DefaultQuerySPParallelPerNode
	}
	if cfg.Bucket.FreeQuotaPerBucket == 0 {
		cfg.Bucket.FreeQuotaPerBucket = downloader.DefaultBucketFreeQuota
	}
	metadata.freeQuotaPerBucket = cfg.Bucket.FreeQuotaPerBucket
	metadata.maxMetadataRequest = cfg.Parallel.QuerySPParallelPerNode

	metadataConfig := &servicemeta.MetadataConfig{
		BsDBConfig:         &cfg.BsDB,
		BsDBSwitchedConfig: &cfg.BsDBBackup,
	}

	metadata.config = metadataConfig
	return nil
}
