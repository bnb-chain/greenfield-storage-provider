package metrics

import (
	"regexp"

	openmetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	metricshttp "github.com/bnb-chain/greenfield-storage-provider/pkg/metrics/http"
)

var MetricsItems = []prometheus.Collector{
	// Grpc metrics category
	DefaultGRPCServerMetrics,
	DefaultGRPCClientMetrics,
	DefaultHTTPServerMetrics,

	// task queue metrics category
	QueueSizeGauge,
	QueueCapGauge,
	QueueTime,
	TaskInQueueTime,

	// piece store metrics category
	PieceStoreTime,
	PieceStoreCounter,
	PieceStoreUsageAmountGauge,

	// db metrics category
	SPDBTime,
	SPDBCounter,

	// chain metrics category
	GnfdChainTime,
	GnfdChainCounter,
	BlockHeightLagGauge,

	// common module metrics items
	ReqCounter,
	ReqTime,
	ReqPieceSize,

	// task executor module metrics category
	ExecutorCounter,
	ExecutorTime,
	GCObjectCounter,
	MaxTaskNumberGauge,
	RunningTaskNumberGauge,
	RemainingMemoryGauge,
	RemainingTaskGauge,
	RemainingHighPriorityTaskGauge,
	RemainingMediumPriorityTaskGauge,
	RemainingLowTaskGauge,

	// manager metrics module category
	ManagerCounter,
	ManagerTime,
	GCBlockNumberGauge,

	// workflow metrics category
	PerfApprovalTime,
	PerfPutObjectTime,

	// Perf workflow category
	PerfAuthTimeHistogram,
	PerfReceivePieceTimeHistogram,
	PerfGetObjectTimeHistogram,
	PerfChallengeTimeHistogram,

	// blocksyncer metrics category
	BlocksyncerCatchTime,
	BlockEventCount,
	BlocksyncerLogicTime,
	BlocksyncerWriteDBTime,
	ChainLatestHeight,
	ChainRPCTime,

	// metadata metrics category
	MetadataReqTime,

	// general runtime metrics category
	GoRoutineCount,

	// sp exit and bucket migration category
	MigrateGVGTimeHistogram,
	MigrateGVGCounter,
	MigrateObjectTimeHistogram,
	MigrateObjectCounter,

	// golang runtime and process metrics
	collectors.NewGoCollector(collectors.WithGoCollectorRuntimeMetrics(collectors.GoRuntimeMetricsRule{
		Matcher: regexp.MustCompile("/.*")})),
	collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
}

// basic metrics items
var (
	// DefaultGRPCServerMetrics defines default gRPC server metrics
	DefaultGRPCServerMetrics = openmetrics.NewServerMetrics(openmetrics.WithServerHandlingTimeHistogram())
	// DefaultGRPCClientMetrics defines default gRPC client metrics
	DefaultGRPCClientMetrics = openmetrics.NewClientMetrics(openmetrics.WithClientHandlingTimeHistogram(),
		openmetrics.WithClientStreamSendHistogram(), openmetrics.WithClientStreamRecvHistogram())
	// DefaultHTTPServerMetrics defines default HTTP server metrics
	DefaultHTTPServerMetrics = metricshttp.NewServerMetrics()

	// task queue metrics
	QueueSizeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "queue_size",
		Help: "Track the task queue used size.",
	}, []string{"queue_size"})
	QueueCapGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "queue_capacity",
		Help: "Track the task queue capacity.",
	}, []string{"queue_capacity"})
	QueueTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "queue_time",
		Help:    "Track the task of queue operator time.",
		Buckets: prometheus.DefBuckets,
	}, []string{"queue_time"})
	TaskInQueueTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "task_in_queue_time",
		Help:    "Track the task of alive time duration in queue from task is pushed.",
		Buckets: prometheus.DefBuckets,
	}, []string{"task_in_queue_time"})

	// piece store metrics
	PieceStoreTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "piece_store_time",
		Help:    "Track the time of operating piece store.",
		Buckets: prometheus.DefBuckets,
	}, []string{"piece_store_time"})
	PieceStoreCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "piece_store_counter",
		Help: "Track total counter of operating piece store.",
	}, []string{"piece_store_counter"})
	PieceStoreUsageAmountGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "usage_amount_piece_store",
		Help: "Track usage amount of piece store.",
	}, []string{"usage_amount_piece_store"})

	// spdb metrics
	SPDBTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "sp_db_time",
		Help:    "Track the time of operating spdb",
		Buckets: prometheus.DefBuckets,
	}, []string{"sp_db_time"})
	SPDBCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "sp_db_counter",
		Help: "Track total counter of operating spdb.",
	}, []string{"sp_db_counter"})

	GnfdChainTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gnfd_chain_time",
		Help:    "Track the time of greenfield chain api costs.",
		Buckets: prometheus.DefBuckets,
	}, []string{"gnfd_chain_time"})
	GnfdChainCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gnfd_chain_counter",
		Help: "Track the counter of greenfield chain api.",
	}, []string{"gnfd_chain_counter"})
	BlockHeightLagGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "block_syncer_height",
		Help: "Current block number of block syncer progress.",
	}, []string{"block_syncer_height"})
)

// module metrics items, include gateway, approver, uploader, manager, task executor,
// receiver, challenge, downloader
var (
	// common module metrics items
	ReqCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "request_qps",
		Help: "Track total request counter.",
	}, []string{"request_qps"})
	ReqTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "request_time",
		Help:    "Track the request time.",
		Buckets: prometheus.DefBuckets,
	}, []string{"request_time"})
	ReqPieceSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "request_piece_size",
		Help:    "Track the request object piece payload size.",
		Buckets: prometheus.DefBuckets,
	}, []string{"request_piece_size"})

	// task executor mertics items
	ExecutorCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "executor_counter",
		Help: "Track total request counter.",
	}, []string{"request_qps"})
	ExecutorTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "executor_time",
		Help:    "Track the executor time.",
		Buckets: prometheus.DefBuckets,
	}, []string{"request_time"})
	MaxTaskNumberGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "max_task_num",
		Help: "Track the max task number of task executor.",
	}, []string{"max_task_num"})
	RunningTaskNumberGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "running_task_num",
		Help: "Track the running task number of task executor.",
	}, []string{"running_task_num"})
	RemainingMemoryGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "remaining_memory_resource",
		Help: "Track remaining memory size of task executor.",
	}, []string{"remaining_memory_resource"})
	RemainingTaskGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "remaining_task_resource",
		Help: "Track remaining resource of total task number.",
	}, []string{"remaining_task_resource"})
	RemainingHighPriorityTaskGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "remaining_high_task_resource",
		Help: "Track remaining resource of high priority task number.",
	}, []string{"remaining_high_task_resource"})
	RemainingMediumPriorityTaskGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "remaining_medium_task_resource",
		Help: "Track remaining resource of medium task number.",
	}, []string{"remaining_medium_task_resource"})
	RemainingLowTaskGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "remaining_low_task_resource",
		Help: "Track remaining resource of low task number.",
	}, []string{"remaining_task_resource"})
	GCObjectCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "delete_object_number",
		Help: "Track deleted object number.",
	}, []string{"delete_object_number"})

	// manager mertics items
	ManagerCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "manager_counter",
		Help: "Track total request counter.",
	}, []string{"request_qps"})
	ManagerTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "manager_time",
		Help:    "Track the manager time.",
		Buckets: prometheus.DefBuckets,
	}, []string{"request_time"})
	GCBlockNumberGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gc_block_number",
		Help: "Track the next gc block number.",
	}, []string{"gc_block_number"})
)

// workflow metrics items
var (
	PerfApprovalTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "perf_approval_time",
		Help:    "Track approval workflow costs.",
		Buckets: prometheus.DefBuckets,
	}, []string{"perf_approval_time"})
	PerfPutObjectTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "perf_put_object_time",
		Help:    "Track put object workflow costs.",
		Buckets: prometheus.DefBuckets,
	}, []string{"perf_put_object_time"})
)

var (
	PerfAuthTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "perf_auth_time",
		Help:    "Track auth workflow costs.",
		Buckets: prometheus.DefBuckets,
	}, []string{"perf_auth_time"})
	PerfReceivePieceTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "perf_receive_time",
		Help:    "Track receive piece workflow costs.",
		Buckets: prometheus.DefBuckets,
	}, []string{"perf_receive_time"})
	PerfGetObjectTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "perf_get_object_time",
		Help:    "Track get object workflow costs.",
		Buckets: prometheus.DefBuckets,
	}, []string{"perf_get_object_time"})
	PerfChallengeTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "perf_challenge_piece_time",
		Help:    "Track challenge piece workflow costs.",
		Buckets: prometheus.DefBuckets,
	}, []string{"perf_challenge_piece_time"})
)

var (
	BlocksyncerCatchTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "blocksyncer_catch_time",
		Help: "Track the time of catch block time. ",
	})
	BlocksyncerLogicTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "blocksyncer_logic_time",
		Help: "Track the time of catch block time. ",
	})
	BlocksyncerWriteDBTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "blocksyncer_write_db_time",
		Help: "Track the time of catch block time. ",
	})
	ChainRPCTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "chain_rpc_time",
		Help: "Track the time of chain rpc. ",
	})
	BlockEventCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "block_event_count",
		Help: "Track the sql count of block. ",
	})
	ChainLatestHeight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "chain_latest_height",
		Help: "Track the height of chain. ",
	})
)

var (
	MetadataReqTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "metadata_request_time",
		Help:    "Track the metadata request time.",
		Buckets: prometheus.DefBuckets,
	}, []string{"status", "level", "method_name", "code_or_msg"})
)

var (
	GoRoutineCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_routine_count",
		Help: "Track the current go routine count. ",
	})
)

// SP exit and bucket migration metrics
var (
	MigrateGVGTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "migrate_gvg_time",
		Help:    "Track migrate gvg workflow costs",
		Buckets: prometheus.DefBuckets,
	}, []string{"migrate_gvg_time"})
	MigrateGVGCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "migrate_gvg_counter",
		Help: "Track migrate gvg number",
	}, []string{"migrate_gvg_counter"})
	MigrateObjectTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "migrate_object_time",
		Help:    "Track migrate object workflow costs",
		Buckets: prometheus.DefBuckets,
	}, []string{"migrate_object_time"})
	MigrateObjectCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "migrate_object_counter",
		Help: "Track migrate object number",
	}, []string{"migrate_object_counter"})
)
