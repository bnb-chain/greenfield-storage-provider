package storage

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

var _ ObjectStorage = &ldfsStore{}

var (
	// re-used ldfs sessions dramatically improve performance
	ldfsSessionCache = &SessionCache{
		sessions: map[ObjectStorageConfig]*session.Session{},
	}
)

type ldfsStore struct {
	s3Store
}

func newLdfsStore(cfg ObjectStorageConfig) (ObjectStorage, error) {
	ldfsSession, bucket, err := ldfsSessionCache.newLdfsSession(cfg)
	if err != nil {
		log.Errorw("failed to new ldfs session", "error", err)
		return nil, err
	}
	log.Infow("new ldfs store succeeds", "bucket", bucket)
	return &ldfsStore{s3Store{bucketName: bucket, api: s3.New(ldfsSession)}}, nil
}

func (sc *SessionCache) newLdfsSession(cfg ObjectStorageConfig) (*session.Session, string, error) {
	sc.Lock()
	defer sc.Unlock()

	endpoint, bucketName, disableSSL, err := parseLdfsBucketURL(cfg.BucketURL)
	if err != nil {
		log.Errorw("failed to parse ldfs bucket url", "error", err)
		return &session.Session{}, "", err
	}
	if sess, ok := sc.sessions[cfg]; ok {
		return sess, bucketName, nil
	}

	// There is no concept of `region` in LDFS
	awsConfig := &aws.Config{
		Region:           aws.String("ldfs"),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(!disableSSL),
		S3ForcePathStyle: aws.Bool(true),
		HTTPClient:       getHTTPClient(cfg.TLSInsecureSkipVerify),
	}
	// We don't need additional authentication.
	// Because we use a whitelist to restrict the IPs that can access LDFS.
	awsConfig.Credentials = credentials.NewStaticCredentials("ldfs", "ldfs", "")

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return &session.Session{}, "", fmt.Errorf("failed to create ldfs session: %s", err)
	}
	sc.sessions[cfg] = sess
	return sess, bucketName, nil
}

func (m *ldfsStore) String() string {
	return fmt.Sprintf("ldfs://%s/", m.s3Store.bucketName)
}

func parseLdfsBucketURL(bucketURL string) (string, string, bool, error) {
	// 1. parse bucket url
	if !strings.Contains(bucketURL, "://") {
		bucketURL = fmt.Sprintf("http://%s", bucketURL)
	}
	uri, err := url.ParseRequestURI(bucketURL)
	if err != nil {
		return "", "", false, fmt.Errorf("invalid endpoint %s: %s", bucketURL, err)
	}

	// 2. check if ldfs uses https
	ssl := strings.ToLower(uri.Scheme) == "https"

	// 3. get bucket name
	if len(uri.Path) < 2 {
		return "", "", false, fmt.Errorf("no bucket name provided in %s", bucketURL)
	}
	bucketName := strings.Split(uri.Path, "/")[1]
	return uri.Host, bucketName, ssl, nil
}
