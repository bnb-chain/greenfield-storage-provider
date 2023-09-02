package uploader

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	// DefaultUploadObjectParallelPerNode defines the default max parallel of uploading
	// object per uploader.
	DefaultUploadObjectParallelPerNode = 10240
)

func NewUploadModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	uploader := &UploadModular{baseApp: app}
	defaultUploaderOptions(uploader, cfg)
	return uploader, nil
}

func defaultUploaderOptions(uploader *UploadModular, cfg *gfspconfig.GfSpConfig) {
	if cfg.Parallel.UploadObjectParallelPerNode == 0 {
		cfg.Parallel.UploadObjectParallelPerNode = DefaultUploadObjectParallelPerNode
	}
	uploader.uploadQueue = cfg.Customize.NewStrategyTQueueFunc(
		uploader.Name()+"-upload-object", cfg.Parallel.UploadObjectParallelPerNode)
	uploader.resumeableUploadQueue = cfg.Customize.NewStrategyTQueueFunc(
		uploader.Name()+"-upload-resumable-object", cfg.Parallel.UploadObjectParallelPerNode)
}
