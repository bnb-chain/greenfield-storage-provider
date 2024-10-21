package uploader

import (
	"github.com/zkMeLabs/mechain-storage-provider/base/gfspapp"
	"github.com/zkMeLabs/mechain-storage-provider/base/gfspconfig"
	coremodule "github.com/zkMeLabs/mechain-storage-provider/core/module"
)

const (
	// DefaultUploadObjectParallelPerNode defines the default max parallel of uploading
	// object per uploader.
	DefaultUploadObjectParallelPerNode = 10240
	// RejectUnSealObjectRetry defines the retry number of sending reject unseal object tx.
	RejectUnSealObjectRetry = 3
	// RejectUnSealObjectTimeout defines the timeout of sending reject unseal object tx.
	RejectUnSealObjectTimeout = 3
	// is not rejected, it is judged failed to reject unseal object on mechain.
	DefaultListenRejectUnSealTimeoutHeight int = 10
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
