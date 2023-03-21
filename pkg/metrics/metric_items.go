package metrics

import (
	metricshttp "github.com/bnb-chain/greenfield-storage-provider/pkg/metrics/http"
	openmetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/prometheus/client_golang/prometheus"
)

// this file is used to write metric items in sp service
var (
	// DefaultGRPCServerMetrics create default gRPC server metrics
	DefaultGRPCServerMetrics = openmetrics.NewServerMetrics(openmetrics.WithServerHandlingTimeHistogram())
	// DefaultGRPCClientMetrics create default gRPC client metrics
	DefaultGRPCClientMetrics = openmetrics.NewClientMetrics(openmetrics.WithClientHandlingTimeHistogram(),
		openmetrics.WithClientStreamSendHistogram(), openmetrics.WithClientStreamRecvHistogram())
	// DefaultHTTPServerMetrics create default HTTP server metrics
	DefaultHTTPServerMetrics = metricshttp.NewServerMetrics()
	// PanicsTotal record the number of rpc panics
	PanicsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "grpc_req_panics_recovered_total",
		Help: "Total number of gRPC requests recovered from internal panic.",
	}, []string{"grpc_type", "grpc_service", "grpc_method"})
)
