package metrics

import (
	openmetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/prometheus/client_golang/prometheus"

	metricshttp "github.com/bnb-chain/greenfield-storage-provider/pkg/metrics/http"
)

var MetricsItems = []prometheus.Collector{
	// Grpc metrics category
	DefaultGRPCServerMetrics,
	DefaultGRPCClientMetrics,
	// Http metrics category
	DefaultHTTPServerMetrics,
	// Perf workflow category
	PerfUploadTimeHistogram,
	PerfGetApprovalTimeHistogram,
	PerfAuthTimeHistogram,
	PerfReceivePieceTimeHistogram,
	PerfGetObjectTimeHistogram,
	PerfChallengeTimeHistogram,
	// TaskQueue metrics category
	QueueSizeGauge,
	QueueCapGauge,
	TaskInQueueTimeHistogram,
	// PieceStore metrics category
	PutPieceTimeHistogram,
	PutPieceTotalNumberCounter,
	GetPieceTimeHistogram,
	GetPieceTotalNumberCounter,
	DeletePieceTimeHistogram,
	DeletePieceTotalNumberCounter,
	PieceUsageAmountGauge,
	// Front module metrics category
	UploadObjectSizeHistogram,
	DownloadObjectSizeHistogram,
	ChallengePieceSizeHistogram,
	ReceivePieceSizeHistogram,
	// TaskExecutor metrics category
	MaxTaskNumberGauge,
	RunningTaskNumberGauge,
	RemainingMemoryGauge,
	RemainingTaskGauge,
	RemainingHighPriorityTaskGauge,
	RemainingMediumPriorityTaskGauge,
	RemainingLowTaskGauge,
	GCObjectCounter,
	ReplicatePieceSizeCounter,
	ReplicateSucceedCounter,
	ReplicateFailedCounter,
	ReplicatePieceTimeHistogram,
	ExecutorReplicatePieceTaskCounter,
	ExecutorSealObjectTaskCounter,
	ExecutorReceiveTaskCounter,
	ExecutorGCObjectTaskCounter,
	ExecutorGCZombieTaskCounter,
	ExecutorGCMetaTaskCounter,
	// Manager metrics category
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
	// Signer metrics category
	SealObjectTimeHistogram,
	SealObjectSucceedCounter,
	SealObjectFailedCounter,
	RejectUnSealObjectTimeHistogram,
	RejectUnSealObjectSucceedCounter,
	RejectUnSealObjectFailedCounter,
	DiscontinueBucketTimeHistogram,
	DiscontinueBucketSucceedCounter,
	DiscontinueBucketFailedCounter,
	// SPDB metrics category
	SPDBTimeHistogram,
	// BlockSyncer metrics category
	BlockHeightLagGauge,
	// the greenfield chain metrics.
	GnfdChainHistogram,
}

var (
	// DefaultGRPCServerMetrics defines default gRPC server metrics
	DefaultGRPCServerMetrics = openmetrics.NewServerMetrics(openmetrics.WithServerHandlingTimeHistogram())
	// DefaultGRPCClientMetrics defines default gRPC client metrics
	DefaultGRPCClientMetrics = openmetrics.NewClientMetrics(openmetrics.WithClientHandlingTimeHistogram(),
		openmetrics.WithClientStreamSendHistogram(), openmetrics.WithClientStreamRecvHistogram())
	// DefaultHTTPServerMetrics defines default HTTP server metrics
	DefaultHTTPServerMetrics = metricshttp.NewServerMetrics()

	// PerfUploadTimeHistogram is used to perf upload workflow.
	PerfUploadTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "perf_upload_time",
		Help:    "Track upload workflow costs.",
		Buckets: prometheus.DefBuckets,
	}, []string{"perf_upload_time"})
	// PerfGetApprovalTimeHistogram is used to perf get approval workflow
	PerfGetApprovalTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "perf_get_approval_time",
		Help:    "Track get approval workflow costs.",
		Buckets: prometheus.DefBuckets,
	}, []string{"perf_get_approval_time"})
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

	// task queue metrics
	QueueSizeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "queue_size",
		Help: "Track the task queue using size.",
	}, []string{"queue_size"})
	QueueCapGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "queue_capacity",
		Help: "Track the task queue capacity.",
	}, []string{"queue_capacity"})
	TaskInQueueTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "task_in_queue_time",
		Help:    "Track the task of alive time duration in queue from task is created.",
		Buckets: prometheus.DefBuckets,
	}, []string{"task_in_queue_time"})

	// piece store metrics
	PutPieceTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "put_piece_store_time",
		Help:    "Track the time of putting piece data to piece store.",
		Buckets: prometheus.DefBuckets,
	}, []string{"put_piece_store_time"})
	PutPieceTotalNumberCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "put_piece_store_number",
		Help: "Track the total number of putting piece data to piece store.",
	}, []string{"put_piece_store_number"})
	GetPieceTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "get_piece_store_time",
		Help:    "Track the time of getting piece data to piece store.",
		Buckets: prometheus.DefBuckets,
	}, []string{"get_piece_store_time"})
	GetPieceTotalNumberCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "get_piece_store_number",
		Help: "Track the total number of getting piece data to piece store.",
	}, []string{"get_piece_store_number"})
	DeletePieceTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "delete_piece_store_time",
		Help:    "Track the time of deleting piece data to piece store.",
		Buckets: prometheus.DefBuckets,
	}, []string{"delete_piece_store_time"})
	DeletePieceTotalNumberCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "delete_piece_store_number",
		Help: "Track the total number of deleting piece data to piece store.",
	}, []string{"delete_piece_store_number"})
	PieceUsageAmountGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "usage_amount_piece_store",
		Help: "Track usage amount of piece store.",
	}, []string{"usage_amount_piece_store"})

	// front module metrics
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
	DownloadPieceSizeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "download_piece_size",
		Help:    "Track the object piece payload size of downloading.",
		Buckets: prometheus.DefBuckets,
	}, []string{"download_piece_size"})
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
	RecoverPieceTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "recovery_piece_time",
		Help:    "Track the time of recovery piece",
		Buckets: prometheus.DefBuckets,
	}, []string{"recovery_piece_time"})
	ExecutorReplicatePieceTaskCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "replicate_task_count",
		Help: "Track replicate task number.",
	}, []string{"replicate_task_count"})
	ExecutorSealObjectTaskCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "seal_task_count",
		Help: "Track seal task number.",
	}, []string{"seal_task_count"})
	ExecutorReceiveTaskCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "receive_task_count",
		Help: "Track receive task number.",
	}, []string{"receive_task_count"})
	ExecutorGCObjectTaskCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gc_object_task_count",
		Help: "Track gc object task number.",
	}, []string{"gc_object_task_count"})
	ExecutorGCZombieTaskCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gc_zombie_task_count",
		Help: "Track gc zombie task number.",
	}, []string{"gc_zombie_task_count"})
	ExecutorGCMetaTaskCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gc_meta_task_count",
		Help: "Track gc meta task number.",
	}, []string{"gc_meta_task_count"})
	ExecutorRecoveryTaskCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "recover_piece_task_count",
		Help: "Track recovery task number.",
	}, []string{"recovery__task_count"})

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
		Name:    "seal_object_task_time",
		Help:    "Track the time of seal object time on chain.",
		Buckets: prometheus.DefBuckets,
	}, []string{"seal_object_task_time"})
	GCBlockNumberGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gc_block_number",
		Help: "Track the next gc block number.",
	}, []string{"gc_block_number"})
	UploadObjectTaskFailedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "upload_object_task_failure",
		Help: "Track upload object task failure total number",
	}, []string{"upload_object_task_failure"})
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
	DispatchRecoverPieceTaskCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dispatch_recovery_piece_task",
		Help: "Track recovery task total number",
	}, []string{"dispatch_recovery_piece_task"})

	// signer metrics
	SealObjectTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "seal_object_time",
		Help:    "Track the time of seal object time to chain.",
		Buckets: prometheus.DefBuckets,
	}, []string{"seal_object_time"})
	SealObjectSucceedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "seal_object_success",
		Help: "Track seal object success total number",
	}, []string{"seal_object_success"})
	SealObjectFailedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "seal_object_failure",
		Help: "Track seal object failure total number",
	}, []string{"seal_object_failure"})
	RejectUnSealObjectTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "reject_unseal_object_time",
		Help:    "Track the time of reject unseal object time to chain.",
		Buckets: prometheus.DefBuckets,
	}, []string{"reject_unseal_object_time"})
	RejectUnSealObjectSucceedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "reject_unseal_object_success",
		Help: "Track reject unseal object success total number",
	}, []string{"reject_unseal_object_success"})
	RejectUnSealObjectFailedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "reject_unseal_object_failure",
		Help: "Track reject unseal object failure total number",
	}, []string{"reject_unseal_object_failure"})
	DiscontinueBucketTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "discontinue_bucket_time",
		Help:    "Track the time of discontinue bucket time to chain.",
		Buckets: prometheus.DefBuckets,
	}, []string{"discontinue_bucket_time"})
	DiscontinueBucketSucceedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "discontinue_bucket_success",
		Help: "Track discontinue bucket success total number.",
	}, []string{"discontinue_bucket_success"})
	DiscontinueBucketFailedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "discontinue_bucket_failure",
		Help: "Track discontinue bucket failure total number.",
	}, []string{"discontinue_bucket_failure"})

	// spdb metrics
	SPDBTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "sp_db_handling_seconds",
		Help:    "Track the latency for spdb requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"sp_db_handling_seconds"})

	// BlockHeightLagGauge records the current block height of block syncer service
	BlockHeightLagGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "block_syncer_height",
		Help: "Current block number of block syncer progress.",
	}, []string{"block_syncer_height"})

	// GnfdChainHistogram is used to record greenfield chain cost.
	GnfdChainHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gnfd_chain_time",
		Help:    "Track the greenfield chain api costs.",
		Buckets: prometheus.DefBuckets,
	}, []string{"gnfd_chain_time"})
)
