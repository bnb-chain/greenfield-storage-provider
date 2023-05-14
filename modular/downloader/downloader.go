package downloader

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/core/taskqueue"
)

const (
	DownloadModularName        = "downloader"
	DownloadModularDescription = "download modular supports download object and get challenge info"
)

var _ module.Downloader = &DownloadModular{}

type DownloadModular struct {
	baseApp        *gfspapp.GfSpBaseApp
	scope          rcmgr.ResourceScope
	downloadQueue  taskqueue.TQueue
	challengeQueue taskqueue.TQueue

	bucketFreeQuota uint64
}

func (d *DownloadModular) Name() string {
	return DownloadModularName
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
