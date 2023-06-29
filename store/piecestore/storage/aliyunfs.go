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

var _ ObjectStorage = &aliyunfsStore{}

var (
	aliyunfsSessionCache = &SessionCache{
		sessions: map[ObjectStorageConfig]*session.Session{},
	}
)

type aliyunfsStore struct {
	s3Store
}

func newAliyunfsStore(cfg ObjectStorageConfig) (ObjectStorage, error) {
	aliyunfsSession, bucket, err := aliyunfsSessionCache.newAliyunfsSession(cfg)
	if err != nil {
		log.Errorw("failed to new aliyunfs session", "error", err)
		return nil, err
	}
	log.Infow("new aliyunfs store succeeds", "bucket", bucket)
	return &aliyunfsStore{s3Store{bucketName: bucket, api: s3.New(aliyunfsSession)}}, nil
}

func (sc *SessionCache) newAliyunfsSession(cfg ObjectStorageConfig) (*session.Session, string, error) {
	sc.Lock()
	defer sc.Unlock()

	endpoint, bucketName, disableSSL, err := parseAliyunfsBucketURL(cfg.BucketURL)
	if err != nil {
		log.Errorw("failed to parse aliyunfs bucket url", "error", err)
		return nil, "", err
	}
	if sess, ok := sc.sessions[cfg]; ok {
		return sess, bucketName, nil
	}
	// todo to be changed prod region
	region := "aliyun-cn-hangzhou"
	awsConfig := &aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(!disableSSL),
		S3ForcePathStyle: aws.Bool(true),
		HTTPClient:       getHTTPClient(cfg.TLSInsecureSkipVerify),
	}
	// todo to be changed prod OSS_AccessKeyId and OSS_AccessKeySecret
	awsConfig.Credentials = credentials.NewStaticCredentials("OSS_AccessKeyId", "OSS_AccessKeySecret", "")

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create aliyunfs session: %s", err)
	}
	sc.sessions[cfg] = sess
	return sess, bucketName, nil
}

func (m *aliyunfsStore) String() string {
	return fmt.Sprintf("aliyunfs://%s/", m.s3Store.bucketName)
}

func parseAliyunfsBucketURL(bucketURL string) (string, string, bool, error) {
	// 1. parse bucket url
	if !strings.Contains(bucketURL, "://") {
		bucketURL = fmt.Sprintf("http://%s", bucketURL)
	}
	uri, err := url.ParseRequestURI(bucketURL)
	if err != nil {
		return "", "", false, fmt.Errorf("invalid endpoint %s: %s", bucketURL, err)
	}

	// 2. check if aliyunfs uses https
	ssl := strings.ToLower(uri.Scheme) == "https"

	// 3. get bucket name
	if len(uri.Path) < 2 {
		return "", "", false, fmt.Errorf("no bucket name provided in %s", bucketURL)
	}
	bucketName := strings.Split(uri.Path, "/")[1]
	return uri.Host, bucketName, ssl, nil
}
