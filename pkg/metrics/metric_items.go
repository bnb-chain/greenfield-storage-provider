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
	// PieceStoreTimeHistogram records piece store request time
	PieceStoreTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "piece_store_handling_seconds",
		Help:    "Track the latency for piece store requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"method_name"})
	// PieceStoreRequestTotal records piece store total request
	PieceStoreRequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "piece_store_total_requests",
		Help: "Track piece store handles total request",
	}, []string{"method_name"})
	SPDBTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "sp_db_handling_seconds",
		Help:    "Track the latency for spdb requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"method_name"})
)
