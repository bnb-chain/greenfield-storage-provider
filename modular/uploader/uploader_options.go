package uploader

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspmdmgr"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	DefaultUploadObjectParallelPerNode = 1024
)

func init() {
	gfspmdmgr.RegisterModularInfo(UploadModularName, UploadModularDescription, NewUploadModular)
}

func NewUploadModular(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig,
	opts ...gfspconfig.Option) (
	coremodule.Modular, error) {
	if cfg.Uploader != nil {
		app.SetUploader(cfg.Uploader)
		return cfg.Uploader, nil
	}
	uploader := &UploadModular{baseApp: app}
	opts = append(opts, uploader.DefaultUploaderOptions)
	for _, opt := range opts {
		if err := opt(app, cfg); err != nil {
			return nil, err
		}
	}
	app.SetUploader(uploader)
	return uploader, nil
}

func (u *UploadModular) DefaultUploaderOptions(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig) error {
	if cfg.UploadObjectParallelPerNode == 0 {
		cfg.UploadObjectParallelPerNode = DefaultUploadObjectParallelPerNode
	}
	u.uploadQueue = cfg.NewStrategyTQueueFunc(u.Name()+"-upload-object",
		cfg.UploadObjectParallelPerNode)
	return nil
}
