package uploader

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	DefaultUploadObjectParallelPerNode = 1024
)

func init() {
	gfspapp.RegisterModularInfo(UploadModularName, UploadModularDescription, NewUploadModular)
}

func NewUploadModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	if cfg.Customize.Uploader != nil {
		app.SetUploader(cfg.Customize.Uploader)
		return cfg.Customize.Uploader, nil
	}
	uploader := &UploadModular{baseApp: app}
	if err := DefaultUploaderOptions(uploader, cfg); err != nil {
		return nil, err
	}
	app.SetUploader(uploader)
	return uploader, nil
}

func DefaultUploaderOptions(uploader *UploadModular, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Parallel.UploadObjectParallelPerNode == 0 {
		cfg.Parallel.UploadObjectParallelPerNode = DefaultUploadObjectParallelPerNode
	}
	uploader.uploadQueue = cfg.Customize.NewStrategyTQueueFunc(
		uploader.Name()+"-upload-object", cfg.Parallel.UploadObjectParallelPerNode)
	return nil
}
