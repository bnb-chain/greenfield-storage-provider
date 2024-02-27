package downloader

import (
	lru "github.com/hashicorp/golang-lru"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/base/gfspconfig"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
)

const (
	// DefaultDownloadObjectParallelPerNode defines the default max download object parallel
	// per downloader
	DefaultDownloadObjectParallelPerNode = 10240
	// DefaultChallengePieceParallelPerNode defines the default max challenge piece parallel
	// per downloader
	DefaultChallengePieceParallelPerNode = 10240
	// DefaultBucketFreeQuota defines the default free read quota per bucket
	DefaultBucketFreeQuota = 10 * 1024 * 1024 * 1024
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

	cache, err := lru.New(cfg.Parallel.DownloadObjectParallelPerNode)
	if err != nil {
		return err
	}

	downloader.pieceCache = cache
	downloader.downloadParallel = int64(cfg.Parallel.DownloadObjectParallelPerNode)
	downloader.challengeParallel = int64(cfg.Parallel.ChallengePieceParallelPerNode)
	if cfg.Quota.MonthlyFreeQuota == 0 {
		downloader.monthlyFreeQuota = gfspapp.DefaultSpMonthlyFreeQuota
	} else {
		downloader.monthlyFreeQuota = cfg.Quota.MonthlyFreeQuota
	}

	return nil
}
