package metrics

import (
	metricshttp "github.com/bnb-chain/greenfield-storage-provider/pkg/metrics/http"
	openmetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/prometheus/client_golang/prometheus"
)

const serviceLabelName = "service"

var MetricsItems = []prometheus.Collector{
	DefaultGRPCServerMetrics,
	DefaultGRPCClientMetrics,
	DefaultHTTPServerMetrics,

	QueueSizeGauge,
	QueueCapGauge,
	TaskInQueueTimeHistogram,

	MaxTaskNumberGauge,
	RunningTaskNumberGauge,
	RemainingMemoryGauge,
	RemainingTaskGauge,
	RemainingHighPriorityTaskGauge,
	RemainingMediumPriorityTaskGauge,
	RemainingLowTaskGauge,
	SealObjectSucceedCounter,
	SealObjectFailedCounter,
	GCObjectCounter,
	ReplicatePieceSizeCounter,
	ReplicateSucceedCounter,
	ReplicateFailedCounter,
	ReplicatePieceTimeHistogram,

	UploadObjectTaskTimeHistogram,
	ReplicateAndSealTaskTimeHistogram,
	ReceiveTaskTimeHistogram,
	SealObjectTaskTimeHistogram,
	GCBlockNumberGauge,
	UploadObjectTaskFailedCounter,
	ReplicatePieceTaskFailedCounter,
	ReceivePieceTaskFailedCounter,
	SealObjectTaskFailedCounter,
	ReplicateCombineSealTaskFailedCounter,
	DispatchReplicatePieceTaskCounter,
	DispatchSealObjectTaskCounter,
	DispatchReceivePieceTaskCounter,
	DispatchGcObjectTaskCounter,
	SealObjectTimeHistogram,
	BlockHeightLagGauge,
}

// this file is used to write metric items in sp service
var (
	// DefaultGRPCServerMetrics create default gRPC server metrics
	DefaultGRPCServerMetrics = openmetrics.NewServerMetrics(openmetrics.WithServerHandlingTimeHistogram())
	// DefaultGRPCClientMetrics create default gRPC client metrics
	DefaultGRPCClientMetrics = openmetrics.NewClientMetrics(openmetrics.WithClientHandlingTimeHistogram(),
		openmetrics.WithClientStreamSendHistogram(), openmetrics.WithClientStreamRecvHistogram())
	// DefaultHTTPServerMetrics create default HTTP server metrics
	DefaultHTTPServerMetrics = metricshttp.NewServerMetrics()

	// queue mertices
	QueueSizeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "queue_size",
		Help: "Current task queue size.",
	}, []string{"queue_size"})
	QueueCapGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "queue_capacity",
		Help: "Task queue capacity.",
	}, []string{"queue_capacity"})
	TaskInQueueTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "task_active_time",
		Help:    "Track the task of time in queue.",
		Buckets: prometheus.DefBuckets,
	}, []string{"task_in_queue_time"})

	// piecestore metrics
	PutPieceTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "put_piece_time",
		Help:    "Track put piece data to store time.",
		Buckets: prometheus.DefBuckets,
	}, []string{"put_piece_time"})
	PutPieceTimeCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "put_piece_number",
		Help: "Track put piece data to store total number.",
	}, []string{"put_piece_number"})
	GetPieceTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "get_piece_time",
		Help:    "Track get piece data to store time.",
		Buckets: prometheus.DefBuckets,
	}, []string{"get_piece_time"})
	GetPieceTimeCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "get_piece_number",
		Help: "Track get piece data to store total number.",
	}, []string{"get_piece_number"})
	DeletePieceTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "delete_piece_time",
		Help:    "Track delete piece data to store time..",
		Buckets: prometheus.DefBuckets,
	}, []string{"delete_piece_time"})
	DeletePieceTimeCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "delete_piece_number",
		Help: "Track get piece data to store total number.",
	}, []string{"delete_piece_number"})
	PieceWriteSizeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "write_size_to_piece_store",
		Help: "Track piece store to use capacity.",
	}, []string{"write_size_to_piece_store"})

	// front modular mertices
	UploadObjectSizeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "upload_object_size",
		Help:    "Track the object payload size of uploading.",
		Buckets: prometheus.DefBuckets,
	}, []string{"upload_object_size"})
	DownloadObjectSizeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "download_object_size",
		Help:    "Track the object payload size of downloading.",
		Buckets: prometheus.DefBuckets,
	}, []string{"download_object_size"})
	ChallengePieceSizeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "challenge_piece_size",
		Help:    "Track the piece size of challenging.",
		Buckets: prometheus.DefBuckets,
	}, []string{"challenge_piece_size"})
	ReceivePieceSizeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "receive_piece_size",
		Help:    "Track the piece size of receiving from primary.",
		Buckets: prometheus.DefBuckets,
	}, []string{"receive_piece_size"})

	// task executor mertics
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
	SealObjectSucceedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "seal_object_success",
		Help: "Track seal object success total number",
	}, []string{"seal_object_success"})
	SealObjectFailedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "seal_object_failure",
		Help: "Track seal object failure total number",
	}, []string{"seal_object_failure"})
	GCObjectCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "delete_object_number",
		Help: "Track deleted object number.",
	}, []string{"delete_object_number"})
	ReplicatePieceSizeCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "replicate_piece_size",
		Help: "Track replicate piece data size.",
	}, []string{"replicate_piece_size"})
	ReplicateSucceedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "replicate_secondary_success",
		Help: "Track replicate secondary success number.",
	}, []string{"replicate_secondary_success"})
	ReplicateFailedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "replicate_secondary_failure",
		Help: "Track replicate secondary failure number.",
	}, []string{"replicate_secondary_failure"})
	ReplicatePieceTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "replicate_piece_time",
		Help:    "Track the time of replicate piece to secondary.",
		Buckets: prometheus.DefBuckets,
	}, []string{"replicate_piece_time"})

	// manager mertics
	UploadObjectTaskTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "upload_primary_time",
		Help:    "Track the time of upload payload to primary.",
		Buckets: prometheus.DefBuckets,
	}, []string{"upload_primary_time"})
	ReplicateAndSealTaskTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "replicate_and_seal_time",
		Help:    "Track the time of replicate to secondary and seal on chain.",
		Buckets: prometheus.DefBuckets,
	}, []string{"replicate_and_seal_time"})
	ReceiveTaskTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "confirm_secondary_piece_seal_on_chain_time",
		Help:    "Track the time of confirm secondary piece seal on chain.",
		Buckets: prometheus.DefBuckets,
	}, []string{"confirm_secondary_piece_on_chain_time"})
	SealObjectTaskTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "seal_object_time",
		Help:    "Track the time of confirm secondary piece seal on chain.",
		Buckets: prometheus.DefBuckets,
	}, []string{"confirm_secondary_piece_on_chain_time"})
	GCBlockNumberGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gc_block_number",
		Help: "Track the next gc block number.",
	}, []string{"gc_block_number"})
	UploadObjectTaskFailedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "upload_object_task_failure",
		Help: "Track upload object task failure total number",
	}, []string{"seal_object_task_failure"})
	ReplicatePieceTaskFailedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "replicate_object_task_failure",
		Help: "Track replicate object task failure total number",
	}, []string{"replicate_object_task_failure"})
	ReceivePieceTaskFailedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "receive_piece_task_failure",
		Help: "Track receive piece task failure total number",
	}, []string{"receive_object_task_failure"})
	SealObjectTaskFailedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "seal_object_task_failure",
		Help: "Track seal object task failure total number",
	}, []string{"seal_object_task_failure"})
	ReplicateCombineSealTaskFailedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "replicate_combine_seal_task_failure",
		Help: "Track combine replicate and seal object failure total number",
	}, []string{"replicate_combine_seal_task_failure"})
	DispatchReplicatePieceTaskCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dispatch_replicate_task",
		Help: "Track replicate object task total number",
	}, []string{"dispatch_replicate_task"})
	DispatchSealObjectTaskCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dispatch_seal_task",
		Help: "Track seal object task total number",
	}, []string{"dispatch_seal_task"})
	DispatchReceivePieceTaskCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dispatch_confirm_receive_task",
		Help: "Track confirm receive task total number",
	}, []string{"dispatch_confirm_receive_task"})
	DispatchGcObjectTaskCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dispatch_gc_object_task",
		Help: "Track gc object task total number",
	}, []string{"dispatch_gc_object_task"})

	// singer metrics
	SealObjectTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "seal_object_time",
		Help:    "Track the time of seal object time to chain.",
		Buckets: prometheus.DefBuckets,
	}, []string{"seal_object_time"})

	// BlockHeightLagGauge records the current block height of block syncer service
	BlockHeightLagGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "block_syncer_height",
		Help: "Current block number of block syncer progress.",
	}, []string{"service"})
	SPDBTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "sp_db_handling_seconds",
		Help:    "Track the latency for spdb requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"method_name"})
)
