package gater

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/base/gfspapp"
	"github.com/bnb-chain/greenfield-storage-provider/core/module"
	"github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

const ReadHeaderTimeout = 20 * time.Minute

var _ module.Modular = &GateModular{}

type GateModular struct {
	env         string
	domain      string
	httpAddress string
	baseApp     *gfspapp.GfSpBaseApp
	scope       rcmgr.ResourceScope
	httpServer  *http.Server

	maxListReadQuota int64
	maxPayloadSize   uint64

	spID uint32
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
	g.httpServer = &http.Server{
		Addr:              g.httpAddress,
		Handler:           router,
		ReadHeaderTimeout: ReadHeaderTimeout,
	}
	if err := g.httpServer.ListenAndServe(); err != nil {
		log.Errorw("failed to listen", "error", err)
		return
	}
}

func (g *GateModular) Stop(ctx context.Context) error {
	g.scope.Release()
	_ = g.httpServer.Shutdown(ctx)
	return nil
}

func (g *GateModular) ReserveResource(ctx context.Context, state *rcmgr.ScopeStat) (rcmgr.ResourceScopeSpan, error) {
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

func (g *GateModular) ReleaseResource(ctx context.Context, span rcmgr.ResourceScopeSpan) {
	span.Done()
}

func (g *GateModular) getSPID() (uint32, error) {
	if g.spID != 0 {
		return g.spID, nil
	}
	spInfo, err := g.baseApp.Consensus().QuerySP(context.Background(), g.baseApp.OperatorAddress())
	if err != nil {
		return 0, err
	}
	g.spID = spInfo.GetId()
	return g.spID, nil
}
