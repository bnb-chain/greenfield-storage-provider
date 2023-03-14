package sqldb

// define table name constant
const (
	// JobTableName defines the job context table name
	JobTableName = "job"
	// ObjectTableName defines the object table name
	ObjectTableName = "object"
	// IntegrityMetaTableName defines the integrity meta table name
	IntegrityMetaTableName = "integrity_meta"
	// SpInfoTableName defines the SP info table name
	SpInfoTableName = "sp_info"
	// StorageParamsTableName defines the storage params info table name
	StorageParamsTableName = "storage_params"
	// BucketTrafficTableName defines the bucket traffic table name, which is used for recoding the used quota by bucket
	BucketTrafficTableName = "bucket_traffic"
	// ReadRecordTableName defines the read record table name
	ReadRecordTableName = "read_record"
	// ServiceConfigTableName defines the SP configuration table name
	ServiceConfigTableName = "service_config"
)

// define metadata query statement
const (
	// DeletedObjectsDefaultSize defines the default size of ListDeletedObjectsByBlockNumberRange response
	DeletedObjectsDefaultSize = 1000
)
