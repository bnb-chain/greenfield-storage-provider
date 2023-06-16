package metadata

import (
	"context"
	"strings"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
)

var (
	MetadataModularName        = strings.ToLower("Metadata")
	MetadataModularDescription = "Retrieves sp metadata and info."
)

const (
	DefaultMetadataStatisticsInterval = 60
)

var _ module.Modular = &MetadataModular{}

type MetadataModular struct {
	baseApp *gfspapp.GfSpBaseApp
	scope   rcmgr.ResourceScope

	// freeQuotaPerBucket defines the free read quota per bucket
	freeQuotaPerBucket uint64
	// maxMetadataRequest defines the max handling metadata request number
	maxMetadataRequest int64
	// retrievingRequest defines the handling retrieve request number
	retrievingRequest int64
}

func (r *MetadataModular) Name() string {
	return MetadataModularName
}

func (r *MetadataModular) Start(ctx context.Context) error {
	// Default the bsDB to master db at start
	r.baseApp.SetGfBsDB(r.baseApp.GfBsDBMaster())
	scope, err := r.baseApp.ResourceManager().OpenService(r.Name())
	if err != nil {
		return err
	}
	r.scope = scope
	return nil
}

func (r *MetadataModular) Stop(ctx context.Context) error {
	r.scope.Release()
	//r.dbSwitchTicker.Stop()
	return nil
}

func (r *MetadataModular) ReserveResource(
	ctx context.Context,
	state *rcmgr.ScopeStat) (
	rcmgr.ResourceScopeSpan, error) {
	return &rcmgr.NullScope{}, nil
}

func (r *MetadataModular) ReleaseResource(
	ctx context.Context,
	span rcmgr.ResourceScopeSpan) {
	span.Done()
}
