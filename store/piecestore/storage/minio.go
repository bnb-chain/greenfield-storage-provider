package storage

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

var _ ObjectStorage = &minioStore{}

var (
	// re-used minio sessions dramatically improve performance
	minioSessionCache = &SessionCache{
		sessions: map[ObjectStorageConfig]*session.Session{},
	}
)

type minioStore struct {
	s3Store
}

func newMinioStore(cfg ObjectStorageConfig) (ObjectStorage, error) {
	minioSession, bucket, err := minioSessionCache.newMinioSession(cfg)
	if err != nil {
		log.Errorw("failed to new minio session", "error", err)
		return nil, err
	}
	log.Infow("new minio store succeeds", "bucket", bucket)
	return &minioStore{s3Store{bucketName: bucket, api: s3.New(minioSession)}}, nil
}

func (sc *SessionCache) newMinioSession(cfg ObjectStorageConfig) (*session.Session, string, error) {
	sc.Lock()
	defer sc.Unlock()

	endpoint, bucketName, disableSSL, err := parseMinioBucketURL(cfg.BucketURL)
	if err != nil {
		log.Errorw("failed to parse minio bucket url", "error", err)
		return nil, "", err
	}
	if sess, ok := sc.sessions[cfg]; ok {
		return sess, bucketName, nil
	}

	// get region
	region, ok := os.LookupEnv(model.MinioRegion)
	if !ok {
		region = endpoints.UsEast1RegionID
	}
	awsConfig := &aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(!disableSSL),
		S3ForcePathStyle: aws.Bool(true),
		HTTPClient:       getHTTPClient(cfg.TLSInsecureSkipVerify),
	}
	key := getSecretKeyFromEnv(model.MinioAccessKey, model.MinioSecretKey, model.MinioSessionToken)
	if key.accessKey != "" && key.secretKey != "" {
		awsConfig.Credentials = credentials.NewStaticCredentials(key.accessKey, key.secretKey, key.sessionToken)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create minio session: %s", err)
	}
	sc.sessions[cfg] = sess
	return sess, bucketName, nil
}

func (m *minioStore) String() string {
	return fmt.Sprintf("minio://%s/", m.s3Store.bucketName)
}

func parseMinioBucketURL(bucketURL string) (string, string, bool, error) {
	// 1. parse bucket url
	if !strings.Contains(bucketURL, "://") {
		bucketURL = fmt.Sprintf("http://%s", bucketURL)
	}
	uri, err := url.ParseRequestURI(bucketURL)
	if err != nil {
		return "", "", false, fmt.Errorf("invalid endpoint %s: %s", bucketURL, err)
	}

	// 2. check if minio uses https
	ssl := strings.ToLower(uri.Scheme) == "https"

	// 3. get bucket name
	if len(uri.Path) < 2 {
		return "", "", false, fmt.Errorf("no bucket name provided in %s", bucketURL)
	}
	bucketName := uri.Path[1:]
	if strings.Contains(bucketName, "/") && strings.HasPrefix(bucketName, "minio/") {
		bucketName = bucketName[len("minio/"):]
	}
	bucketName = strings.Split(bucketName, "/")[0]

	return uri.Host, bucketName, ssl, nil
}
