package metrics

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var (
	// dfl: draft full load
	dflBuckets = []float64{0.3, 1.0, 2.5, 5.0}
)

const (
	requestName = "http_requests_total"
	latencyName = "http_request_duration_seconds"
)

// Opts specifies options how to create new PrometheusMiddleware.
type Opts struct {
	// Buckets specifies a custom buckets to be used in request histogram.
	Buckets []float64
}

// PrometheusMiddleware specifies the metrics that is going to be generated
type PrometheusMiddleware struct {
	request *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

// NewPrometheusMiddleware creates a new PrometheusMiddleware instance
func NewPrometheusMiddleware(opts Opts) *PrometheusMiddleware {
	var prometheusMiddleware PrometheusMiddleware
	prometheusMiddleware.request = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: model.GatewayService,
			Name:      requestName,
			Help:      "How many HTTP requests processed, partitioned by host, http path, handler name, method and status code",
		}, []string{"host", "path", "handler_name", "method", "code"},
	)
	if err := prometheus.Register(prometheusMiddleware.request); err != nil {
		log.Errorw("prometheusMiddleware.request was not registered", "error", err)
	}

	buckets := opts.Buckets
	if len(buckets) == 0 {
		buckets = dflBuckets
	}
	prometheusMiddleware.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    latencyName,
		Help:    "How long it took to process the request, partitioned by host, http path, handler name, method and status code",
		Buckets: buckets,
	}, []string{"host", "path", "handler_name", "method", "code"},
	)
	if err := prometheus.Register(prometheusMiddleware.latency); err != nil {
		log.Errorw("prometheusMiddleware.latency was not registered", "error", err)
	}

	return &prometheusMiddleware
}

// InstrumentHandlerDuration is a middleware that wraps the http.Handler and it records
// how long the handler took to run, which path was called, and the status code.
func (p *PrometheusMiddleware) InstrumentHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		inner := &innerResponseWriter{ResponseWriter: w}
		rw := inner
		next.ServeHTTP(rw, r) // call original

		route := mux.CurrentRoute(r)
		host, _ := route.GetHostTemplate()
		path, _ := route.GetPathTemplate()
		handlerName := route.GetName()
		method := sanitizeMethod(r.Method)
		code := sanitizeCode(inner.status)
		go p.request.WithLabelValues(host, path, handlerName, method, code).Inc()
		go p.latency.WithLabelValues(host, path, handlerName, method, code).Observe(float64(time.Since(begin)) / float64(time.Second))
	})
}

type innerResponseWriter struct {
	http.ResponseWriter
	status      int
	written     int64
	writeHeader bool
}

func (r *innerResponseWriter) WriteHeader(code int) {
	r.status = code
	r.writeHeader = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *innerResponseWriter) Write(b []byte) (int, error) {
	if !r.writeHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.written += int64(n)
	return n, err
}

func sanitizeMethod(m string) string {
	return strings.ToLower(m)
}

func sanitizeCode(s int) string {
	return strconv.Itoa(s)
}
