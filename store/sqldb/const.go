package sqldb

// define table name constant
const (
	// JobTableName defines the job context table name
	JobTableName = "job"
	// ObjectTableName defines the object table name
	ObjectTableName = "object"
	// IntegrityMetaTableName defines the integrity meta table name
	IntegrityMetaTableName = "integrity_meta"
	// SpInfoTableName defines the SP info table name, the data come from greenfield chain
	SpInfoTableName = "sp_info"
	// StorageParamsTableName defines the storage params info table name, the data come from greenfield chain
	StorageParamsTableName = "storage_params"
	// BucketTrafficTableName defines the bucket traffic table name, uses for recoding the used quota by bucket
	BucketTrafficTableName = "bucket_traffic"
	// ReadRecordTableName defines the read record table name
	ReadRecordTableName = "read_record"
	// ServiceConfigTableName defines the SP configuration table name
	ServiceConfigTableName = "service_config"
)
