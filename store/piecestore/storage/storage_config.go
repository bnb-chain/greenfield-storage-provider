package storage

// PieceStoreConfig contains some parameters which are used to run PieceStore
type PieceStoreConfig struct {
	Shards int                  // store the blocks into N buckets by hash of key
	Store  *ObjectStorageConfig // config of object storage
}

var DefaultPieceStoreConfig = &PieceStoreConfig{
	Shards: 0,
	Store: &ObjectStorageConfig{
		Storage:               "s3",
		BucketURL:             "https://s3.us-east-1.amazonaws.com/tf-nodereal-dev-bsc-storage",
		AccessKey:             "",
		SecretKey:             "",
		SessionToken:          "",
		NoSignRequest:         false,
		MaxRetries:            3,
		MinRetryDelay:         0,
		TlsInsecureSkipVerify: false,
	},
}

type ObjectStorageConfig struct {
	Storage               string // backend storage type (e.g. s3, file, memory)
	BucketURL             string // the bucket URL of object storage to store data
	AccessKey             string // access key for object storage
	SecretKey             string // secret key for object storage
	SessionToken          string // temporary credential used to access backend storage
	NoSignRequest         bool   // whether access public bucket
	MaxRetries            int    // the number of max retries that will be performed
	MinRetryDelay         int64  // the minimum retry delay after which retry will be performed
	TlsInsecureSkipVerify bool   // whether skip the certificate verification of HTTPS requests
}
