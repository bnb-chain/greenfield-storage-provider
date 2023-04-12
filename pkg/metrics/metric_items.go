package metrics

import (
	openmetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/prometheus/client_golang/prometheus"

	metricshttp "github.com/bnb-chain/greenfield-storage-provider/pkg/metrics/http"
)

const serviceLabelName = "service"

// this file is used to write metric items in sp service
var (
	// DefaultGRPCServerMetrics create default gRPC server metrics
	DefaultGRPCServerMetrics = openmetrics.NewServerMetrics(openmetrics.WithServerHandlingTimeHistogram())
	// DefaultGRPCClientMetrics create default gRPC client metrics
	DefaultGRPCClientMetrics = openmetrics.NewClientMetrics(openmetrics.WithClientHandlingTimeHistogram(),
		openmetrics.WithClientStreamSendHistogram(), openmetrics.WithClientStreamRecvHistogram())
	// DefaultHTTPServerMetrics create default HTTP server metrics
	DefaultHTTPServerMetrics = metricshttp.NewServerMetrics()

	// PanicsTotal records the number of rpc panics
	PanicsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "grpc_req_panics_recovered_total",
		Help: "Total number of gRPC requests recovered from internal panic.",
	}, []string{"grpc_type", "grpc_service", "grpc_method"})
	// BlockHeightLagGauge records the current block height of block syncer service
	BlockHeightLagGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "block_syncer_height",
		Help: "Current block number of block syncer progress.",
	}, []string{serviceLabelName})
	// SealObjectTimeHistogram records sealing object time of task node service
	SealObjectTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "task_node_seal_object_time",
		Help:    "Track task node service the time of sealing object on chain.",
		Buckets: prometheus.DefBuckets,
	}, []string{serviceLabelName})
	// SealObjectTotalCounter records total seal object number
	SealObjectTotalCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "task_node_seal_object_total",
		Help: "Track task node service handles total seal object number",
	}, []string{"success_or_failure"})
	// ReplicateObjectTaskGauge records total replicate object number
	ReplicateObjectTaskGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "task_node_replicate_object_task_number",
		Help: "Track task node service replicate object task",
	}, []string{serviceLabelName})
)
