package storage

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	credentials_aliyun "github.com/aliyun/credentials-go/credentials"
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

	key := getAliyunSecretKeyFromEnv(AliyunRegion, AliyunAccessKey, AliyunSecretKey, AliyunSessionToken)
	creds := credentials.NewStaticCredentials(key.accessKey, key.secretKey, key.sessionToken)
	awsConfig := &aws.Config{
		Region:           aws.String(key.region),
		Endpoint:         aws.String(endpoint),
		DisableSSL:       aws.Bool(!disableSSL),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      creds,
		HTTPClient:       getHTTPClient(cfg.TLSInsecureSkipVerify),
	}
	log.Debugw("aliyun env", "AliyunRegion", key.region, "AliyunAccessKey", key.accessKey,
		"AliyunSecretKey", key.secretKey, "AliyunSessionToken", key.sessionToken)

	var sess *session.Session
	switch cfg.IAMType {
	case AKSKIAMType:
		key := getAliyunSecretKeyFromEnv(AliyunRegion, AliyunAccessKey, AliyunSecretKey, AliyunSessionToken)
		if key.accessKey == "NoSignRequest" {
			awsConfig.Credentials = credentials.AnonymousCredentials
		} else if key.accessKey != "" && key.secretKey != "" {
			awsConfig.Credentials = credentials.NewStaticCredentials(key.accessKey, key.secretKey, key.sessionToken)
		}
		sess = session.Must(session.NewSession(awsConfig))
		log.Debugw("use aksk to access aliyunfs", "region", *sess.Config.Region, "endpoint", *sess.Config.Endpoint)
	case SAIAMType:
		irsa, roleARN, tokenPath := checkAliyunAvailable()
		log.Debugw("aliyun env", "irsa", irsa, "roleARN", roleARN, "tokenPath", tokenPath)
		if irsa {
			cred, err := credentials_aliyun.NewCredential(nil)

			if err != nil {
				panic(err)
			}

			accessKeyId, _ := cred.GetAccessKeyId()
			accessKeySecret, _ := cred.GetAccessKeySecret()
			securityToken, _ := cred.GetSecurityToken()

			log.Debugw("aliyun env", "accessKeyId", accessKeyId, "accessKeySecret", accessKeySecret, "securityToken", securityToken)

			awsConfig.Credentials = credentials.NewStaticCredentials(*accessKeyId, *accessKeySecret, *securityToken)
		} else {
			return nil, "", fmt.Errorf("failed to use sa to access aliyunfs")
		}
		sess = session.Must(session.NewSession(awsConfig))
	default:
		log.Errorf("unknown IAM type: %s", cfg.IAMType)
		return nil, "", fmt.Errorf("unknown IAM type: %s", cfg.IAMType)
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

func checkAliyunAvailable() (bool, string, string) {
	irsa := true
	roleARN, exists := os.LookupEnv(AliyunRoleARN)
	if !exists {
		irsa = false
		log.Error("failed to read oss role arn")
	}
	tokenPath, exists := os.LookupEnv(AliyunWebIdentityTokenFile)
	if !exists {
		irsa = false
		log.Error("failed to read oss web identity token file")
	}
	return irsa, roleARN, tokenPath
}
