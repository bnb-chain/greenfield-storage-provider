package pprof

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/felixge/fgprof"
	"github.com/gorilla/mux"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// Pprof is used to analyse the performance sp service
type Pprof struct {
	config     *PprofConfig
	httpServer *http.Server
}

// NewPprof returns an instance of pprof
func NewPprof(cfg *PprofConfig) *Pprof {
	return &Pprof{config: cfg}
}

// Name describes pprof service name
func (p *Pprof) Name() string {
	return model.PprofService
}

// Start HTTP server
func (p *Pprof) Start(ctx context.Context) error {
	if p.config.Enabled {
		go p.serve()
	}
	return nil
}

// Stop HTTP server
func (p *Pprof) Stop(ctx context.Context) error {
	var errs []error
	if err := p.httpServer.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

func (p *Pprof) serve() {
	router := mux.NewRouter()
	p.registerProfiler(router)
	p.httpServer = &http.Server{
		Addr:    p.config.HTTPAddress,
		Handler: router,
	}
	if err := p.httpServer.ListenAndServe(); err != nil {
		log.Errorw("failed to listen and serve", "error", err)
		return
	}
}

func (p *Pprof) registerProfiler(r *mux.Router) {
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)
	r.Handle("/debug/fgprof", fgprof.Handler())
}
