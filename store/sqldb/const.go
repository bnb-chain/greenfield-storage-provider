package sqldb

// define table name constant.
const (
	// UploadObjectProgressTableName defines the gc object task table name.
	UploadObjectProgressTableName = "upload_object_progress"
	// GCObjectProgressTableName defines the gc object task table name.
	GCObjectProgressTableName = "gc_object_progress"
	// PieceHashTableName defines the piece hash table name.
	PieceHashTableName = "piece_hash"
	// IntegrityMetaTableName defines the integrity meta table name.
	IntegrityMetaTableName = "integrity_meta"
	// SpInfoTableName defines the SP info table name.
	SpInfoTableName = "sp_info"
	// StorageParamsTableName defines the storage params info table name.
	StorageParamsTableName = "storage_params"
	// BucketTrafficTableName defines the bucket traffic table name, which is used for recoding the used quota by bucket.
	BucketTrafficTableName = "bucket_traffic"
	// ReadRecordTableName defines the read record table name.
	ReadRecordTableName = "read_record"
	// ServiceConfigTableName defines the SP configuration table name.
	ServiceConfigTableName = "service_config"
	// OffChainAuthKeyTableName defines the off chain auth key table name.
	OffChainAuthKeyTableName = "off_chain_auth_key"
	// PutObjectSuccessTableName  defines the event of successfully putting object
	PutObjectSuccessTableName = "put_object_success_event_log"
	// PutObjectEventTableName defines the event of putting object
	PutObjectEventTableName = "put_object_event_log"
	// UploadTimeoutTableName defines the event of uploading object to primary sp timeout
	UploadTimeoutTableName = "upload_timeout_event_log"
	// UploadFailedTableName defines the event of unsuccessfully uploading object to primary sp
	UploadFailedTableName = "upload_failed_event_log"
	// ReplicateTimeoutTableName defines the event of replicating object to secondary sp timeout
	ReplicateTimeoutTableName = "replicate_timeout_event_log"
	// ReplicateFailedTableName defines the event of unsuccessfully uploading object to secondary sp
	ReplicateFailedTableName = "replicate_failed_event_log"
	// SealTimeoutTableName defines the event of sealing object timeout
	SealTimeoutTableName = "seal_timeout_event_log"
	// SealFailedTableName defines the event of unsuccessfully sealing object timeout
	SealFailedTableName = "seal_failed_event_log"
	// MigrateSubscribeProgressTableName defines the progress of subscribe migrate event.
	MigrateSubscribeProgressTableName = "migrate_subscribe_progress"
	// SwapOutTableName is the swap out unit table.
	SwapOutTableName = "swap_out_unit"
	// MigrateGVGTableName defines the progress of subscribe migrate event.
	MigrateGVGTableName = "migrate_gvg"
)

// define error name constant.
const (
	TableAlreadyExistsErrorPrefix = "Error 1050 (42S01)"
)
