package gater

import (
	"context"
	"net/http"
	"sync"
	"time"

	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

var _ module.Modular = &GateModular{}

type GateModular struct {
	domain      string
	httpAddress string
	baseApp     *gfspapp.GfSpBaseApp
	scope       rcmgr.ResourceScope
	httpServer  *http.Server

	maxListReadQuota int64
	maxPayloadSize   uint64

	param *storagetypes.Params
	mux   sync.RWMutex
}

func (g *GateModular) Name() string {
	return module.GateModularName
}

func (g *GateModular) Start(ctx context.Context) error {
	scope, err := g.baseApp.ResourceManager().OpenService(g.Name())
	if err != nil {
		return err
	}
	g.scope = scope
	go g.server(ctx)
	return nil
}

func (g *GateModular) server(ctx context.Context) {
	router := mux.NewRouter().SkipClean(true)
	if g.baseApp.EnableMetrics() {
		router.Use(metrics.DefaultHTTPServerMetrics.InstrumentationHandler)
	}
	g.RegisterHandler(router)
	server := &http.Server{
		Addr:    g.httpAddress,
		Handler: router,
	}
	g.httpServer = server
	if err := server.ListenAndServe(); err != nil {
		log.Errorw("failed to listen", "error", err)
		return
	}
	go g.eventLoop(ctx)
}

func (g *GateModular) Stop(ctx context.Context) error {
	g.scope.Release()
	g.httpServer.Shutdown(ctx)
	return nil
}

func (g *GateModular) ReserveResource(
	ctx context.Context,
	state *rcmgr.ScopeStat) (
	rcmgr.ResourceScopeSpan, error) {
	span, err := g.scope.BeginSpan()
	if err != nil {
		return nil, err
	}
	err = span.ReserveResources(state)
	if err != nil {
		return nil, err
	}
	return span, nil
}

func (g *GateModular) ReleaseResource(
	ctx context.Context,
	span rcmgr.ResourceScopeSpan) {
	span.Done()
}

func (g *GateModular) eventLoop(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			params, err := g.baseApp.Consensus().QueryStorageParams(ctx)
			if err != nil {
				log.CtxErrorw(ctx, "failed to query storage params from chain", "error", err)
				continue
			}
			g.mux.Lock()
			g.param = params
			g.mux.Unlock()
		}
	}
}

func (g *GateModular) GetStorageParams() (*storagetypes.Params, error) {
	g.mux.RLock()
	defer g.mux.RUnlock()
	if g.param == nil {
		return nil, ErrConsensus
	}
	return g.param, nil
}
