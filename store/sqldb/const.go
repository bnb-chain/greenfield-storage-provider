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
	// UploadEventTableName defines the event of uploading object
	UploadEventTableName = "upload_event"
)
