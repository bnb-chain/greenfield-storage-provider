package sqldb

// define table name constant
const (
	JobTableName           = "job"
	ObjectTableName        = "object"
	IntegrityMetaTableName = "integrity_meta"
	SPInfoTableName        = "sp_info"
	StorageParamsTableName = "storage_params"
	BucketTrafficTableName = "bucket_traffic"
	ReadRecordTableName    = "read_record"
)

// SPDB environment constants
const (
	SPDBUser   = "SP_DB_USER"
	SPDBPasswd = "SP_DB_PASSWORD"
)
