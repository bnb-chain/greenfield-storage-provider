package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var (
	mMonitor MetricsMonitor
	once     sync.Once
)

// MetricsMonitor defines abstract method
type MetricsMonitor interface {
	lifecycle.Service
	Enabled() bool
}

// Metrics is used to monitor sp services
type Metrics struct {
	config     *MetricsConfig
	registry   *prometheus.Registry
	httpServer *http.Server
}

// NewMetrics returns a singleton instance of Metrics.
// Note: if you want to use metrics service in storage provider, you must call NewMetrics in initMetricsConfig func.
// If you use GetMetrics method straightly without calling NewMetrics firstly, you won't start metrics service to collect
// stats data about sp.
func NewMetrics(cfg *MetricsConfig) MetricsMonitor {
	return initMetrics(cfg)
}

// GetMetrics gets an instance of MetricsMonitor, you can use this in the service logic of sp
func GetMetrics() MetricsMonitor {
	return initMetrics(nil)
}

// initMetrics is used to init metrics according to MetricsConfig
func initMetrics(cfg *MetricsConfig) MetricsMonitor {
	once.Do(func() {
		if cfg == nil || !cfg.Enabled {
			mMonitor = NilMetrics{}
		} else {
			mMonitor = &Metrics{
				config:   cfg,
				registry: prometheus.NewRegistry(),
			}
		}
	})
	return mMonitor
}

// Name describes metrics service name
func (m *Metrics) Name() string {
	return model.MetricsService
}

// Start HTTP server
func (m *Metrics) Start(ctx context.Context) error {
	m.registerMetricItems()
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
	if m.config != nil {
		return m.config.Enabled
	} else {
		return false
	}
}

func (m *Metrics) registerMetricItems() {
	m.registry.MustRegister(DefaultGRPCServerMetrics, DefaultGRPCClientMetrics, DefaultHTTPServerMetrics,
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}), PanicsTotal)
}

func (m *Metrics) serve() {
	router := mux.NewRouter()
	router.Path("/metrics").Handler(promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))
	m.httpServer = &http.Server{
		Addr:    m.config.HTTPAddress,
		Handler: router,
	}
	if err := m.httpServer.ListenAndServe(); err != nil {
		log.Errorw("failed to listen and serve", "error", err)
		return
	}
}

// NilMetrics is a no-op Metrics
type NilMetrics struct{}

// Name is a no-op
func (NilMetrics) Name() string {
	return ""
}

// Start is a no-op
func (NilMetrics) Start(ctx context.Context) error {
	return merrors.ErrUnsupportedMethod
}

// Stop is a no-op
func (NilMetrics) Stop(ctx context.Context) error {
	return merrors.ErrUnsupportedMethod
}

// Enabled is a no-op
func (NilMetrics) Enabled() bool {
	return false
}
