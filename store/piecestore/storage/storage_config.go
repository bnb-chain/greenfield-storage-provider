package storage

// PieceStoreConfig contains some parameters which are used to run PieceStore
type PieceStoreConfig struct {
	// Shards store the blocks into N buckets by hash of key
	Shards int `comment:"required"`
	// Store config of object storage
	Store ObjectStorageConfig
}

// ObjectStorageConfig object storage config
type ObjectStorageConfig struct {
	// Storage backend storage type (e.g. s3, file, memory)
	Storage string `comment:"required"`
	// BucketURL the bucket URL of object storage to store data
	BucketURL string `comment:"optional"`
	// MaxRetries the number of max retries that will be performed
	MaxRetries int `comment:"optional"`
	// MinRetryDelay the minimum retry delay after which retry will be performed
	MinRetryDelay int64 `comment:"optional"`
	// TLSInsecureSkipVerify whether skip the certificate verification of HTTPS requests
	TLSInsecureSkipVerify bool `comment:"optional"`
	// IAMType is identity and access management type which contains two types: AKSKIAMType/SAIAMType
	IAMType string `comment:"required"`
	//for input from console
	AccessKeyID     string `comment:"optional"`
	AccessSecretKey string `comment:"optional"`	
}
