package pprof

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strings"

	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/felixge/fgprof"
	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var (
	PProfModularName = strings.ToLower("PProf")
)
var _ coremodule.Modular = &PProf{}

// PProf is used to analyse the performance sp service
type PProf struct {
	httpAddress string
	httpServer  *http.Server
}

// NewPProf returns an instance of pprof
func NewPProf(address string) *PProf {
	return &PProf{httpAddress: address}
}

// Name describes pprof service name
func (p *PProf) Name() string {
	return PProfModularName
}

// Start HTTP server
func (p *PProf) Start(ctx context.Context) error {
	go p.serve()
	return nil
}

// Stop HTTP server
func (p *PProf) Stop(ctx context.Context) error {
	var errs []error
	if err := p.httpServer.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

func (p *PProf) serve() {
	router := mux.NewRouter()
	p.registerProfiler(router)
	p.httpServer = &http.Server{
		Addr:    p.httpAddress,
		Handler: router,
	}
	if err := p.httpServer.ListenAndServe(); err != nil {
		log.Errorw("failed to listen and serve", "error", err)
		return
	}
}

func (p *PProf) ReserveResource(
	ctx context.Context,
	state *corercmgr.ScopeStat) (
	corercmgr.ResourceScopeSpan, error) {
	return &corercmgr.NullScope{}, nil
}
func (p *PProf) ReleaseResource(
	ctx context.Context,
	scope corercmgr.ResourceScopeSpan) {
	scope.Done()
	return
}

func (p *PProf) registerProfiler(r *mux.Router) {
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)
	r.Handle("/debug/fgprof", fgprof.Handler())
}
