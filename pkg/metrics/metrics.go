package metrics

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	coremodule "github.com/bnb-chain/greenfield-storage-provider/core/module"
	corercmgr "github.com/bnb-chain/greenfield-storage-provider/core/rcmgr"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var (
	MetricsModularName = strings.ToLower("Metrics")
)

var _ coremodule.Modular = &Metrics{}

// Metrics is used to monitor sp services
type Metrics struct {
	httpAddress string
	registry    *prometheus.Registry
	httpServer  *http.Server
}

func NewMetrics(address string) *Metrics {
	return &Metrics{
		httpAddress: address,
		registry:    prometheus.NewRegistry(),
	}
}

// Name describes metrics service name
func (m *Metrics) Name() string {
	return MetricsModularName
}

// Start HTTP server
func (m *Metrics) Start(ctx context.Context) error {
	m.RegisterMetricItems(MetricsItems...)
	go m.serve()
	return nil
}

// Stop HTTP server
func (m *Metrics) Stop(ctx context.Context) error {
	var errs []error
	if err := m.httpServer.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

// Enabled returns whether starts prometheus metrics
func (m *Metrics) Enabled() bool {
	return false
}

func (m *Metrics) RegisterMetricItems(cs ...prometheus.Collector) {
	m.registry.MustRegister(cs...)
}

func (m *Metrics) serve() {
	router := mux.NewRouter()
	router.Path("/metrics").Handler(promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))
	m.httpServer = &http.Server{
		Addr:    m.httpAddress,
		Handler: router,
	}
	if err := m.httpServer.ListenAndServe(); err != nil {
		log.Errorw("failed to listen and serve", "error", err)
		return
	}
}

func (m *Metrics) ReserveResource(
	ctx context.Context,
	state *corercmgr.ScopeStat) (
	corercmgr.ResourceScopeSpan, error) {
	return &corercmgr.NullScope{}, nil
}
func (m *Metrics) ReleaseResource(
	ctx context.Context,
	scope corercmgr.ResourceScopeSpan) {
	scope.Done()
}

// NilMetrics is a no-op Metrics
type NilMetrics struct{}

// Name is a no-op
func (NilMetrics) Name() string {
	return ""
}

// Start is a no-op
func (NilMetrics) Start(ctx context.Context) error {
	return nil
}

// Stop is a no-op
func (NilMetrics) Stop(ctx context.Context) error {
	return nil
}

// Enabled is a no-op
func (NilMetrics) Enabled() bool {
	return false
}
