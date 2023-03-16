package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	openmetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/lifecycle"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	metricshttp "github.com/bnb-chain/greenfield-storage-provider/pkg/metrics/http"
)

var _ lifecycle.Service = &MetricsMonitor{}

var (
	reg = prometheus.NewRegistry()
	// DefaultGRPCServerMetrics create default gRPC server metrics
	DefaultGRPCServerMetrics = openmetrics.NewServerMetrics(openmetrics.WithServerHandlingTimeHistogram())
	// DefaultGRPCClientMetrics create default gRPC client metrics
	DefaultGRPCClientMetrics = openmetrics.NewClientMetrics(openmetrics.WithClientHandlingTimeHistogram())
	// DefaultHTTPServerMetrics create default HTTP server metrics
	DefaultHTTPServerMetrics = metricshttp.NewServerMetrics()
)

func init() {
	reg.MustRegister(DefaultGRPCServerMetrics, DefaultGRPCClientMetrics, DefaultHTTPServerMetrics,
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
}

// Metrics is used to monitor sp services
type MetricsMonitor struct {
	config     *MetricsMonitorConfig
	httpServer *http.Server
}

// NewMetricsMonitor returns an instance of Metrics
func NewMetricsMonitor(cfg *MetricsMonitorConfig) (*MetricsMonitor, error) {
	return &MetricsMonitor{config: cfg}, nil
}

// Name describes service name
func (m *MetricsMonitor) Name() string {
	return model.MetricsMonitorService
}

// Start HTTP server
func (m *MetricsMonitor) Start(ctx context.Context) error {
	go m.serve()
	return nil
}

// Stop HTTP server
func (m *MetricsMonitor) Stop(ctx context.Context) error {
	var errs []error
	if err := m.httpServer.Shutdown(ctx); err != nil {
		errs = append(errs, err)
	}
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

func (m *MetricsMonitor) serve() {
	router := mux.NewRouter()
	router.Path("/metrics").Handler(promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	m.httpServer = &http.Server{
		Addr:    m.config.HTTPAddress,
		Handler: router,
	}
	if err := m.httpServer.ListenAndServe(); err != nil {
		log.Errorw("failed to listen and serve", "error", err)
		return
	}
}
