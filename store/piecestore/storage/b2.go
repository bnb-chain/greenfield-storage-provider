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

var _ ObjectStorage = &b2Store{}

var (
	// re-used b2 sessions dramatically improve performance
	b2SessionCache = &SessionCache{
		sessions: map[ObjectStorageConfig]*session.Session{},
	}
)

type b2Store struct {
	s3Store
}

func newB2Store(cfg ObjectStorageConfig) (ObjectStorage, error) {
	b2Session, bucket, err := b2SessionCache.newB2Session(cfg)
	if err != nil {
		log.Errorw("failed to new b2 session", "error", err)
		return nil, err
	}
	log.Infow("new b2 store succeeds", "bucket", bucket)
	return &b2Store{s3Store{bucketName: bucket, api: s3.New(b2Session)}}, nil
}

func (sc *SessionCache) newB2Session(cfg ObjectStorageConfig) (*session.Session, string, error) {
	sc.Lock()
	defer sc.Unlock()

	endpoint, region, bucketName, err := parseB2BucketURL(cfg.BucketURL)
	if err != nil {
		log.Errorw("failed to parse b2 bucket url", "error", err)
		return nil, "", err
	}

	if sess, ok := sc.sessions[cfg]; ok {
		return sess, bucketName, nil
	}

	key := getSecretKeyFromEnv(B2AccessKey, B2SecretKey, B2SessionToken)
	awsConfig := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(key.accessKey, key.secretKey, key.sessionToken),
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(disableSSL),
		S3ForcePathStyle: aws.Bool(true),
		HTTPClient:       getHTTPClient(cfg.TLSInsecureSkipVerify),
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create b2 session: %s", err)
	}

	sc.sessions[cfg] = sess
	return sess, bucketName, nil
}

func (m *b2Store) String() string {
	return fmt.Sprintf("b2://%s/", m.bucketName)
}

func parseB2BucketURL(bucketURL string) (endpoint, region, bucketName string, err error) {
	bucketURL = strings.Trim(bucketURL, "/")
	uri, err := url.ParseRequestURI(bucketURL)
	if err != nil {
		err = fmt.Errorf("failed to parse b2 bucket url: %s", err)
		return
	}

	ssl := strings.ToLower(uri.Scheme) == "https"
	if !ssl {
		disableSSL = true
	}

	endpoint = uri.Host

	if uri.Path != "" {
		// Path style: https://s3.<region>.backblazeb2.com(.cn)/<bucketName>
		bucketName = strings.Split(uri.Path, "/")[1]
		isVirtualHostStyle = false
	} else {
		// Virtual hosted style: https://<bucketName>.s3.<region>.backblazeb2.com(.cn)
		bucketName = strings.SplitN(endpoint, ".s3", 2)[0]
		endpoint = endpoint[len(bucketName)+1:]
		isVirtualHostStyle = true
	}

	region = parseRegion(endpoint)
	return
}
