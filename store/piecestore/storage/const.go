package storage

// define storage type constants
const (
	// S3Store defines storage type for s3
	S3Store = "s3"
	// B2Store defines storage type for s3
	B2Store = "b2"
	// MinioStore defines storage type for minio
	MinioStore = "minio"
	// DiskFileStore defines storage type for file
	DiskFileStore = "file"
	// LdfsStore defines storage type for ldfs
	LdfsStore = "ldfs"
	// OssfsStore defines storage type for ossfs
	AliyunfsStore = "aliyunfs"
	// MemoryStore defines storage type for memory
	MemoryStore = "memory"
)

// piece store storage config and environment constants
const (
	// AKSKIAMType defines IAM type config which uses access key and secret key to access aws s3
	AKSKIAMType = "AKSK"
	// SAIAMType defines IAM type config which uses service account to access aws s3
	SAIAMType = "SA"

	// AWSRoleARN defines env variable for aws role arn
	AWSRoleARN = "AWS_ROLE_ARN"
	// AWSWebIdentityTokenFile defines env variable for aws identity token file
	AWSWebIdentityTokenFile = "AWS_WEB_IDENTITY_TOKEN_FILE"

	// BucketURL defines env variable name for bucket url
	BucketURL = "BUCKET_URL"
	// AWSAccessKey defines env variable name for aws access key
	AWSAccessKey = "AWS_ACCESS_KEY"
	// AWSSecretKey defines env variable name for aws secret key
	AWSSecretKey = "AWS_SECRET_KEY"
	// AWSSessionToken defines env variable name for aws session token
	AWSSessionToken = "AWS_SESSION_TOKEN"

	// AliyunRoleARN defines env variable for aws role arn
	AliyunRoleARN = "ALIYUN_ROLE_ARN"
	// AliyunWebIdentityTokenFile defines env variable for aws identity token file
	AliyunWebIdentityTokenFile = "ALIYUN_WEB_IDENTITY_TOKEN_FILE"
	// AliyunAccessKey defines env variable name for aws access key
	AliyunAccessKey = "ALIYUN_ACCESS_KEY"
	// AliyunSecretKey defines env variable name for aws secret key
	AliyunSecretKey = "ALIYUN_SECRET_KEY"
	// AliyunSessionToken defines env variable name for aws session token
	AliyunSessionToken = "ALIYUN_SESSION_TOKEN"
	// AliyunSessionToken defines env variable name for aws session token
	AliyunRegion = "ALIYUN_SESSION_TOKEN"

	// MinioRegion defines env variable name for minio region
	MinioRegion = "MINIO_REGION"
	// MinioAccessKey defines env variable name for minio access key
	MinioAccessKey = "MINIO_ACCESS_KEY"
	// MinioSecretKey defines env variable name for minio secret key
	MinioSecretKey = "MINIO_SECRET_KEY"
	// MinioSessionToken defines env variable name for minio session token
	MinioSessionToken = "MINIO_SESSION_TOKEN"

	// B2AccessKey defines env variable name for minio access key
	B2AccessKey = "B2_ACCESS_KEY"
	// B2SecretKey defines env variable name for minio secret key
	B2SecretKey = "B2_SECRET_KEY"
	// B2SessionToken defines env variable name for minio session token
	B2SessionToken = "B2_SESSION_TOKEN"

	// OctetStream is used to indicate the binary files
	OctetStream = "application/octet-stream"
)

// define piece store constants.
const (
	// BufPoolSize define buffer pool size
	BufPoolSize = 32 << 10
	// ChecksumAlgo define validation algorithm name
	ChecksumAlgo = "Crc32c"
)
