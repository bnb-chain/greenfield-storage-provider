package storage

// PieceStoreConfig contains some parameters which are used to run PieceStore
type PieceStoreConfig struct {
	Shards int                  // store the blocks into N buckets by hash of key
	Store  *ObjectStorageConfig // config of object storage
}

// DefaultPieceStoreConfig if no config in config file, use this default config
var DefaultPieceStoreConfig = &PieceStoreConfig{
	Shards: 0,
	Store: &ObjectStorageConfig{
		Storage:               "s3",
		BucketURL:             "https://s3.ap-northeast-1.amazonaws.com/example",
		NoSignRequest:         false,
		MaxRetries:            5,
		MinRetryDelay:         0,
		TlsInsecureSkipVerify: false,
	},
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
