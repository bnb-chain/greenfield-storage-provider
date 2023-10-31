package storage

// define storage type constants
const (
	// S3Store defines storage type for s3
	S3Store = "s3"
	// OSSStore defines storage type for oss
	OSSStore = "oss"
	// B2Store defines storage type for s3
	B2Store = "b2"
	// MinioStore defines storage type for minio
	MinioStore = "minio"
	// DiskFileStore defines storage type for file
	DiskFileStore = "file"
	// LdfsStore defines storage type for ldfs
	LdfsStore = "ldfs"
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

	// OSSRoleARN defines env variable for OSS role arn
	OSSRoleARN = "ALIBABA_CLOUD_ROLE_ARN"
	// OSSWebIdentityTokenFile defines env variable for OSS identity token file
	OSSWebIdentityTokenFile = "ALIBABA_CLOUD_OIDC_TOKEN_FILE"
	// OSSOidcProviderArn defines env variable for OSS oidc provider arn
	OSSOidcProviderArn = "ALIBABA_CLOUD_OIDC_PROVIDER_ARN"
	// OSSAccessKey defines env variable name for OSS access key
	OSSAccessKey = "ALIBABA_CLOUD_ACCESS_KEY"
	// OSSSecretKey defines env variable name for OSS secret key
	OSSSecretKey = "ALIBABA_CLOUD_SECRET_KEY"
	// OSSSessionToken defines env variable name for OSS session token
	OSSSessionToken = "ALIBABA_CLOUD_SESSION_TOKEN"
	// OSSRegion defines env variable name for OSS oss region
	OSSRegion = "ALIBABA_CLOUD_OSS_REGION"

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
	UserAgent = "Greenfield-SP"
	// BufPoolSize define buffer pool size
	BufPoolSize = 32 << 10
	// ChecksumAlgo define validation algorithm name
	ChecksumAlgo = "Crc32c"
)
