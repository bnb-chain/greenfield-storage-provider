package retriever

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var (
	RetrieveModularName        = strings.ToLower("Retriever")
	RetrieveModularDescription = "Retrieves sp metadata and info."
)

const (
	DefaultRetrieverStatisticsInterval = 60
)

var _ module.Modular = &RetrieveModular{}

type RetrieveModular struct {
	baseApp *gfspapp.GfSpBaseApp
	scope   rcmgr.ResourceScope

	// freeQuotaPerBucket defines the free read quota per bucket
	freeQuotaPerBucket uint64
	// maxRetrieveRequest defines the max handling retrieve request number
	maxRetrieveRequest int64
	// retrievingRequest defines the handling retrieve request number
	retrievingRequest int64
}

func (r *RetrieveModular) Name() string {
	return RetrieveModularName
}

func (r *RetrieveModular) Start(ctx context.Context) error {
	scope, err := r.baseApp.ResourceManager().OpenService(r.Name())
	if err != nil {
		return err
	}
	r.scope = scope
	return nil
}

func (r *RetrieveModular) eventLoop(ctx context.Context) {
	statisticsTicker := time.NewTicker(time.Duration(DefaultRetrieverStatisticsInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-statisticsTicker.C:
			log.Infof("retrieveMax[%d], retrieving[%d]",
				atomic.LoadInt64(&r.maxRetrieveRequest), atomic.LoadInt64(&r.retrievingRequest))
		}
	}
}

func (r *RetrieveModular) Stop(ctx context.Context) error {
	r.scope.Release()
	return nil
}

func (r *RetrieveModular) ReserveResource(
	ctx context.Context,
	state *rcmgr.ScopeStat) (
	rcmgr.ResourceScopeSpan, error) {
	return &rcmgr.NullScope{}, nil
}

func (r *RetrieveModular) ReleaseResource(
	ctx context.Context,
	span rcmgr.ResourceScopeSpan) {
	span.Done()
	return
}
