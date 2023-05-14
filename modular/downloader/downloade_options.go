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

func init() {
	gfspapp.RegisterModularInfo(DownloadModularName, DownloadModularDescription, NewDownloadModular)
}

func NewDownloadModular(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig,
	opts ...gfspapp.Option) (
	coremodule.Modular, error) {
	if cfg.Downloader != nil {
		app.SetDownloader(cfg.Downloader)
		return cfg.Downloader, nil
	}
	downloader := &DownloadModular{baseApp: app}
	opts = append(opts, downloader.DefaultDownloaderOptions)
	for _, opt := range opts {
		if err := opt(app, cfg); err != nil {
			return nil, err
		}
	}
	app.SetDownloader(downloader)
	return downloader, nil
}

func (d *DownloadModular) DefaultDownloaderOptions(
	app *gfspapp.GfSpBaseApp,
	cfg *gfspconfig.GfSpConfig) error {
	if cfg.DownloadObjectParallelPerNode == 0 {
		cfg.DownloadObjectParallelPerNode = DefaultDownloadObjectParallelPerNode
	}
	if cfg.ChallengePieceParallelPerNode == 0 {
		cfg.ChallengePieceParallelPerNode = DefaultChallengePieceParallelPerNode
	}
	if cfg.BucketFreeQuota == 0 {
		cfg.BucketFreeQuota = DefaultBucketFreeQuota
	}
	d.downloadQueue = cfg.NewStrategyTQueueFunc(
		d.Name()+DefaultDownloadObjectQueueSuffix,
		cfg.DownloadObjectParallelPerNode)
	d.challengeQueue = cfg.NewStrategyTQueueFunc(
		d.Name()+DefaultChallengePieceQueueSuffix,
		cfg.ChallengePieceParallelPerNode)
	d.bucketFreeQuota = cfg.BucketFreeQuota
	return nil
}
