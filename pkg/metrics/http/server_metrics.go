package http

import (
	"encoding/xml"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"

	modelgateway "github.com/bnb-chain/greenfield-storage-provider/model/gateway"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// ServerMetrics represents a collection of metrics to be registered on a
// Prometheus metrics registry for an HTTP server.
type ServerMetrics struct {
	serverReqTotalCounter  *prometheus.CounterVec
	serverReqInflightGauge *prometheus.GaugeVec
	serverReqSizeSummary   *prometheus.SummaryVec
	serverRespSizeSummary  *prometheus.SummaryVec
	serverReqDuration      *prometheus.HistogramVec
}

// NewServerMetrics returns an instance of ServerMetrics
func NewServerMetrics(opts ...ServerMetricsOption) *ServerMetrics {
	var config serverMetricsConfig
	config.apply(opts)
	return &ServerMetrics{
		// host and path is not monitored because of virtual path
		serverReqTotalCounter: prometheus.NewCounterVec(
			config.counterOpts.apply(prometheus.CounterOpts{
				Name: "http_server_received_total_requests",
				Help: "Tracks the total number of HTTP requests.",
			}), []string{"handler_name", "method", "code", "error_code"}),
		serverReqInflightGauge: prometheus.NewGaugeVec(
			config.gaugeOpts.apply(prometheus.GaugeOpts{
				Name: "http_server_inflight_requests",
				Help: "Current number of HTTP requests the handler is responding to.",
			}), []string{"handler_name", "method"}),
		serverReqSizeSummary: prometheus.NewSummaryVec(
			config.summaryOpts.apply(prometheus.SummaryOpts{
				Name: "http_request_size_bytes",
				Help: "Tracks the size of HTTP requests.",
			}), []string{"handler_name", "method", "code", "error_code"}),
		serverRespSizeSummary: prometheus.NewSummaryVec(
			config.summaryOpts.apply(prometheus.SummaryOpts{
				Name: "http_response_size_bytes",
				Help: "Tracks the size of HTTP responses.",
			}), []string{"handler_name", "method", "code", "error_code"}),
		serverReqDuration: prometheus.NewHistogramVec(
			config.histogramOpts.apply(prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Tracks the latencies for HTTP requests.",
				Buckets: prometheus.DefBuckets,
			}), []string{"handler_name", "method", "code", "error_code"}),
	}
}

// NewRegisteredServerMetrics returns a custom ServerMetrics object registered with the user's registry
// and registers some common metrics associated with every instance.
func NewRegisteredServerMetrics(registry prometheus.Registerer, opts ...ServerMetricsOption) *ServerMetrics {
	customServerMetrics := NewServerMetrics(opts...)
	customServerMetrics.MustRegister(registry)
	return customServerMetrics
}

// Register registers the metrics with the registry.
func (m *ServerMetrics) Register(registry prometheus.Registerer) error {
	for _, collector := range m.toRegister() {
		if err := registry.Register(collector); err != nil {
			return err
		}
	}
	return nil
}

// MustRegister registers the metrics with the registry
// Panic if any error occurs much like DefaultRegisterer of Prometheus.
func (m *ServerMetrics) MustRegister(registry prometheus.Registerer) {
	registry.MustRegister(m.toRegister()...)
}

func (m *ServerMetrics) toRegister() []prometheus.Collector {
	res := []prometheus.Collector{
		m.serverReqTotalCounter,
		m.serverReqInflightGauge,
		m.serverReqSizeSummary,
		m.serverRespSizeSummary,
		m.serverReqDuration,
	}
	return res
}

// Describe sends the super-set of all possible descriptors of metrics collected by this Collector to the provided
// channel and returns once the last descriptor has been sent.
func (m *ServerMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.serverReqTotalCounter.Describe(ch)
	m.serverReqInflightGauge.Describe(ch)
	m.serverReqSizeSummary.Describe(ch)
	m.serverRespSizeSummary.Describe(ch)
	m.serverReqDuration.Describe(ch)
}

// Collect is called by the Prometheus registry when collecting metrics. The implementation sends each
// collected metric via the provided channel and returns once the last metric has been sent.
func (m *ServerMetrics) Collect(ch chan<- prometheus.Metric) {
	m.serverReqTotalCounter.Collect(ch)
	m.serverReqInflightGauge.Collect(ch)
	m.serverReqSizeSummary.Collect(ch)
	m.serverRespSizeSummary.Collect(ch)
	m.serverReqDuration.Collect(ch)
}

// InstrumentationHandler initializes all metrics, with their appropriate null value, for all HTTP methods registered
// on an HTTP server. This is useful, to ensure that all metrics exist when collecting and querying.
func (m *ServerMetrics) InstrumentationHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		wd := &responseWriterDelegator{w: w}
		next.ServeHTTP(wd, r)

		method := r.Method
		code := wd.Status()
		handlerName := mux.CurrentRoute(r).GetName()

		var (
			errorResp = &modelgateway.ErrorResponse{}
			errorCode string
		)
		if code != strconv.Itoa(http.StatusOK) {
			body := wd.GetBody()
			err := xml.Unmarshal(body, errorResp)
			if err != nil {
				log.Infow("cannot parse gateway error response", "error", err)
				errorCode = "-1" // unknown error code
				goto METRICS
			}
			errorCode = strconv.Itoa(int(errorResp.Code))
			log.Infow("print error response", "error code", errorCode, "error resp", errorResp)
		} else {
			errorCode = "0" // no error
		}

	METRICS:
		m.serverReqTotalCounter.WithLabelValues(handlerName, method, code, errorCode).Inc()
		gauge := m.serverReqInflightGauge.WithLabelValues(handlerName, method)
		gauge.Inc()
		defer gauge.Dec()
		m.serverReqSizeSummary.WithLabelValues(handlerName, method, code, errorCode).Observe(float64(computeApproximateRequestSize(r)))
		m.serverRespSizeSummary.WithLabelValues(handlerName, method, code, errorCode).Observe(float64(wd.size))
		observer := m.serverReqDuration.WithLabelValues(handlerName, method, code, errorCode)
		observer.Observe(time.Since(now).Seconds())

		var traceID string
		span := trace.SpanFromContext(r.Context())
		if span != nil && span.SpanContext().IsSampled() {
			traceID = span.SpanContext().TraceID().String()
		}
		if traceID != "" {
			observer.(prometheus.ExemplarObserver).ObserveWithExemplar(
				time.Since(now).Seconds(),
				prometheus.Labels{
					"traceID": traceID,
				},
			)
		}
	})
}
