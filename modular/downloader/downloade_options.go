package downloader

import (
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	DefaultDownloadObjectParallelPerNode = 10240
	DefaultChallengePieceParallelPerNode = 10240
	DefaultBucketFreeQuota               = 10 * 1024 * 1024 * 1024
	DefaultDownloadObjectQueueSuffix     = "-download-queue"
	DefaultChallengePieceQueueSuffix     = "-challenge-piece"
)

func NewDownloadModular(app *gfspapp.GfSpBaseApp, cfg *gfspconfig.GfSpConfig) (coremodule.Modular, error) {
	downloader := &DownloadModular{baseApp: app}
	if err := DefaultDownloaderOptions(downloader, cfg); err != nil {
		return nil, nil
	}
	return downloader, nil
}

func DefaultDownloaderOptions(downloader *DownloadModular, cfg *gfspconfig.GfSpConfig) error {
	if cfg.Parallel.DownloadObjectParallelPerNode == 0 {
		cfg.Parallel.DownloadObjectParallelPerNode = DefaultDownloadObjectParallelPerNode
	}
	if cfg.Parallel.ChallengePieceParallelPerNode == 0 {
		cfg.Parallel.ChallengePieceParallelPerNode = DefaultChallengePieceParallelPerNode
	}
	if cfg.Bucket.FreeQuotaPerBucket == 0 {
		cfg.Bucket.FreeQuotaPerBucket = DefaultBucketFreeQuota
	}
	downloader.downloadQueue = cfg.Customize.NewStrategyTQueueFunc(
		downloader.Name()+DefaultDownloadObjectQueueSuffix,
		cfg.Parallel.DownloadObjectParallelPerNode)
	downloader.challengeQueue = cfg.Customize.NewStrategyTQueueFunc(
		downloader.Name()+DefaultChallengePieceQueueSuffix,
		cfg.Parallel.ChallengePieceParallelPerNode)
	downloader.bucketFreeQuota = cfg.Bucket.FreeQuotaPerBucket
	return nil
}
