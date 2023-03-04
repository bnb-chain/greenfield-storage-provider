package storage

// PieceStoreConfig contains some parameters which are used to run PieceStore
type PieceStoreConfig struct {
	Shards int                  // store the blocks into N buckets by hash of key
	Store  *ObjectStorageConfig // config of object storage
}

// ObjectStorageConfig object storage config
type ObjectStorageConfig struct {
	Storage               string // backend storage type (e.g. s3, file, memory)
	BucketURL             string // the bucket URL of object storage to store data
	NoSignRequest         bool   // whether access public bucket
	MaxRetries            int    // the number of max retries that will be performed
	MinRetryDelay         int64  // the minimum retry delay after which retry will be performed
	TlsInsecureSkipVerify bool   // whether skip the certificate verification of HTTPS requests
	TestMode              bool   // if test mode is true, should provide s3 credentials
}
