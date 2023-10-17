package probe

import (
	"context"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var (
	ProbeModularName = strings.ToLower("Probe")
)

var _ coremodule.Modular = &Probe{}

// PProf is used to analyse the performance sp service
type Probe struct {
	httpAddress string
	httpServer  *http.Server
	httpProbe   *HTTPProbe
}

// NewProbe returns an instance of probe
func NewProbe(address string, probe *HTTPProbe) *Probe {
	return &Probe{
		httpAddress: address,
		httpProbe:   probe,
	}
}

// Name describes probe service name
func (p *Probe) Name() string {
	return ProbeModularName
}

// Start HTTP server
func (p *Probe) Start(ctx context.Context) error {
	go p.serve()
	return nil
}

// Stop HTTP server
func (p *Probe) Stop(ctx context.Context) error {
	if err := p.httpServer.Shutdown(ctx); err != nil {
		log.Errorw("failed to shutdown http server", "error", err)
		return err
	}
	return nil
}

func (p *Probe) serve() {
	router := mux.NewRouter()
	p.registerProbes(router, p.httpProbe)
	p.httpServer = &http.Server{
		Addr:    p.httpAddress,
		Handler: router,
	}
	if err := p.httpServer.ListenAndServe(); err != nil {
		log.Errorw("failed to listen and serve", "error", err)
		return
	}
}

func (p *Probe) ReserveResource(ctx context.Context, state *corercmgr.ScopeStat) (corercmgr.ResourceScopeSpan, error) {
	return &corercmgr.NullScope{}, nil
}

func (p *Probe) ReleaseResource(ctx context.Context, scope corercmgr.ResourceScopeSpan) {
	scope.Done()
}

func (p *Probe) registerProbes(r *mux.Router, h *HTTPProbe) {
	r.HandleFunc("/-/healthy", h.HealthyHandler())
	r.HandleFunc("/-/ready", h.ReadyHandler())
}
