package downloader

import (
	"context"
	"fmt"

	lru "github.com/hashicorp/golang-lru"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
)

var _ module.Downloader = &DownloadModular{}

type DownloadModular struct {
	baseApp           *gfspapp.GfSpBaseApp
	scope             rcmgr.ResourceScope
	pieceCache        *lru.Cache
	downloading       int64
	downloadParallel  int64
	challenging       int64
	challengeParallel int64

	// bucketFreeQuota defines the free read quota per bucket, if exceed
	// the quota, the account should buy traffic.
	bucketFreeQuota uint64
}

func (d *DownloadModular) Name() string {
	return module.DownloadModularName
}

func (d *DownloadModular) Start(ctx context.Context) error {
	scope, err := d.baseApp.ResourceManager().OpenService(d.Name())
	if err != nil {
		return err
	}
	d.scope = scope
	return nil
}

func (d *DownloadModular) Stop(ctx context.Context) error {
	d.scope.Release()
	return nil
}

func (d *DownloadModular) ReserveResource(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
	span, err := d.scope.BeginSpan()
	if err != nil {
		return nil, err
	}
	err = span.ReserveResources(state)
	if err != nil {
		return nil, err
	}
	return span, nil
}

func (d *DownloadModular) ReleaseResource(ctx context.Context, span rcmgr.ResourceScopeSpan) {
	span.Done()
}

func cacheKey(pieceKey string, offset, length int64) string {
	return fmt.Sprintf("piece:%s-offset:%d-length:%d", pieceKey, offset, length)
}
