package piecestore

// define storage type constants
const (
	S3Store       = "s3"
	MinioStore    = "minio"
	DiskFileStore = "file"
	MemoryStore   = "memory"
)

// piece store storage environment constants
const (
	// BucketURL defines env variable name for bucket url
	BucketURL = "BUCKET_URL"
	// AWSAccessKey defines env variable name for aws access key
	AWSAccessKey = "AWS_ACCESS_KEY"
	// AWSSecretKey defines env variable name for aws secret key
	AWSSecretKey = "AWS_SECRET_KEY"
	// AWSSessionToken defines env variable name for aws session token
	AWSSessionToken = "AWS_SESSION_TOKEN"
	// MinioRegion defines env variable name for minio region
	MinioRegion = "MINIO_REGION"
	// MinioAccessKey defines env variable name for minio access key
	MinioAccessKey = "MINIO_ACCESS_KEY"
	// MinioSecretKey defines env variable name for minio secret key
	MinioSecretKey = "MINIO_SECRET_KEY"
	// MinioSessionToken defines env variable name for minio session token
	MinioSessionToken = "MINIO_SESSION_TOKEN"
)

// define piece store constants.
const (
	// BufPoolSize define buffer pool size
	BufPoolSize = 32 << 10
	// ChecksumAlgo define validation Algorithm Name
	ChecksumAlgo = "Crc32c"
)
