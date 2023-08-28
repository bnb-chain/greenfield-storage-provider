package metadata

import (
	"context"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
)

const (
	DefaultMetadataStatisticsInterval = 60
)

var _ coremodule.Modular = &MetadataModular{}

type MetadataModular struct {
	baseApp *gfspapp.GfSpBaseApp
	scope   rcmgr.ResourceScope

	// maxMetadataRequest defines the max handling metadata request number
	maxMetadataRequest int64
	// retrievingRequest defines the handling retrieve request number
	retrievingRequest int64
}

func (r *MetadataModular) Name() string {
	return coremodule.MetadataModularName
}

func (r *MetadataModular) Start(ctx context.Context) error {
	scope, err := r.baseApp.ResourceManager().OpenService(r.Name())
	if err != nil {
		return err
	}
	r.scope = scope
	return nil
}

func (r *MetadataModular) Stop(ctx context.Context) error {
	r.scope.Release()
	// r.dbSwitchTicker.Stop()
	return nil
}

func (r *MetadataModular) ReserveResource(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
	return &rcmgr.NullScope{}, nil
}

func (r *MetadataModular) ReleaseResource(ctx context.Context, span rcmgr.ResourceScopeSpan) {
	span.Done()
}
