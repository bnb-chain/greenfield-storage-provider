package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aliyun/credentials-go/credentials"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

type ossStore struct {
	client *oss.Client
	bucket *oss.Bucket
}

func (o *ossStore) String() string {
	return fmt.Sprintf("oss://%s/", o.bucket.BucketName)
}

func (o *ossStore) CreateBucket(ctx context.Context) error {
	err := o.bucket.Client.CreateBucket(o.bucket.BucketName)
	if err != nil && isErrExists(err) {
		err = nil
	}
	return err
}

func (o *ossStore) GetObject(ctx context.Context, key string, off, limit int64) (resp io.ReadCloser, err error) {
	var respHeader http.Header
	if off > 0 || limit > 0 {
		var r string
		if limit > 0 {
			r = fmt.Sprintf("%d-%d", off, off+limit-1)
		} else {
			r = fmt.Sprintf("%d-", off)
		}
		resp, err = o.bucket.GetObject(key, oss.NormalizedRange(r), oss.RangeBehavior("standard"), oss.GetResponseHeader(&respHeader))
	} else {
		resp, err = o.bucket.GetObject(key, oss.GetResponseHeader(&respHeader))
		if err == nil {
			resp = verifyChecksum(resp,
				resp.(*oss.Response).Headers.Get(oss.HTTPHeaderOssMetaPrefix+ChecksumAlgo))
		}
	}
	return resp, err
}

func (o *ossStore) PutObject(ctx context.Context, key string, in io.Reader) error {
	var (
		option     []oss.Option
		respHeader http.Header
	)
	if rs, ok := in.(io.ReadSeeker); ok {
		option = append(option, oss.Meta(ChecksumAlgo, generateChecksum(rs)))
	}
	option = append(option, oss.GetResponseHeader(&respHeader))
	err := o.bucket.PutObject(key, in, option...)
	return err
}

func (o *ossStore) DeleteObject(ctx context.Context, key string) error {
	var respHeader http.Header
	return o.bucket.DeleteObject(key, oss.GetResponseHeader(&respHeader))
}

func (o *ossStore) DeleteObjectsByPrefix(ctx context.Context, key string) (uint64, error) {
	var (
		objectKeys           []string
		objectKeySizeMap     = make(map[string]uint64)
		continueDeleteObject = true
		batchSize            = int64(1000)
		size                 uint64
	)

	for continueDeleteObject {
		objs, err := o.ListObjects(ctx, key, "", "", batchSize)
		if err != nil {
			log.Errorw("DeleteObjectsByPrefix list objects error", "error", err)
			return size, err
		}

		if len(objs) == 0 {
			log.CtxDebugw(ctx, "No object is listed in oss by prefix", "prefix", key)
			return 0, nil
		}

		// if the object listed here is less than required batch size, meaning it is the last page
		if int64(len(objs)) < batchSize {
			continueDeleteObject = false
		}

		for _, obj := range objs {
			objectKeys = append(objectKeys, obj.Key())
			objectKeySizeMap[obj.Key()] = uint64(obj.Size())
		}
		deletedObjResults, err := o.bucket.DeleteObjects(objectKeys)
		if err != nil {
			log.Errorw("DeleteObjectsByPrefix delete objects error", "error", err)
		}
		for _, deletedObjKey := range deletedObjResults.DeletedObjects {
			size += objectKeySizeMap[deletedObjKey]
		}
	}
	return size, nil
}

func (o *ossStore) HeadBucket(ctx context.Context) error {
	ok, err := o.client.IsBucketExist(o.bucket.BucketName)
	if !ok {
		log.Errorw("OSS bucket does not exist", "bucket name", o.bucket.BucketName, "error", err)
		return fmt.Errorf("bucket %s does not exist", o.bucket.BucketName)
	}
	return err
}

func (o *ossStore) HeadObject(ctx context.Context, key string) (Object, error) {
	r, err := o.bucket.GetObjectMeta(key)
	if err != nil {
		log.Errorw("failed to head object", "error", err)
		if e, ok := err.(oss.ServiceError); ok && e.StatusCode == http.StatusNotFound {
			err = os.ErrNotExist
		}
		return nil, err
	}

	lastModified := r.Get("Last-Modified")
	if lastModified == "" {
		return nil, fmt.Errorf("cannot get last modified time")
	}
	contentLength := r.Get("Content-Length")
	mtime, _ := time.Parse(time.RFC1123, lastModified)
	size, _ := strconv.ParseInt(contentLength, 10, 64)
	return &object{
		key,
		size,
		mtime,
		strings.HasSuffix(key, "/"),
	}, nil
}

func (o *ossStore) ListObjects(ctx context.Context, prefix, marker, delimiter string, limit int64) ([]Object, error) {
	if limit > 1000 {
		limit = 1000
	}
	result, err := o.bucket.ListObjects(oss.Prefix(prefix), oss.Marker(marker), oss.Delimiter(delimiter),
		oss.MaxKeys(int(limit)))
	if err != nil {
		log.Errorw("failed to list objects", "error", err)
		return nil, err
	}
	n := len(result.Objects)
	objs := make([]Object, n)
	for i := 0; i < n; i++ {
		obj := result.Objects[i]
		objs[i] = &object{obj.Key, obj.Size, obj.LastModified, strings.HasSuffix(obj.Key, "/")}
	}
	if delimiter != "" {
		for _, obj := range result.CommonPrefixes {
			objs = append(objs, &object{obj, 0, time.Unix(0, 0), true})
		}
		sort.Slice(objs, func(i, j int) bool { return objs[i].Key() < objs[j].Key() })
	}
	return objs, nil
}

func (o *ossStore) ListAllObjects(ctx context.Context, prefix, marker string) (<-chan Object, error) {
	return nil, ErrUnsupportedMethod
}

func newOSSStore(cfg ObjectStorageConfig) (ObjectStorage, error) {
	var (
		cli          *oss.Client
		credProvider oss.CredentialsProvider
		err          error
	)
	endpoint, bucketName, region, err := parseOSS(cfg.BucketURL)
	if err != nil {
		log.Errorw("failed to parse oss bucket url", "error", err)
		return nil, err
	}

	switch cfg.IAMType {
	case AKSKIAMType:
		key := getOSSSecretKeyFromEnv(OSSRegion, OSSAccessKey, OSSSecretKey, OSSSessionToken)
		if key.accessKey != "" && key.secretKey != "" {
			cli, err = oss.New(endpoint, key.accessKey, key.secretKey, oss.SecurityToken(key.sessionToken),
				oss.Region(region), oss.HTTPClient(getHTTPClient(cfg.TLSInsecureSkipVerify)))
			if err != nil {
				log.Errorw("failed to use ak sk iam type to new oss", "error", err)
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("cannot get access key and secret key in os env vars")
		}
		log.Debug("use aksk to access oss")
	case SAIAMType:
		credProvider, err = newOIDCCredentialProvider()
		if err != nil {
			log.Errorw("failed to new oidc credential provider", "error", err)
		}
		cli, err = oss.New(endpoint, "", "", oss.SetCredentialsProvider(credProvider))
		if err != nil {
			log.Errorw("failed to use sa iam type to new oss", "error", err)
			return nil, err
		}
		log.Debug("use sa to access oss")
	default:
		log.Errorf("unknown IAM type: %s", cfg.IAMType)
		return nil, fmt.Errorf("unknown IAM type: %s", cfg.IAMType)
	}

	cli.Config.Timeout = 10
	cli.Config.RetryTimes = uint(cfg.MaxRetries)
	cli.Config.HTTPTimeout.ConnectTimeout = time.Second * 30
	cli.Config.HTTPTimeout.ReadWriteTimeout = time.Second * 60
	cli.Config.HTTPTimeout.HeaderTimeout = time.Second * 60
	cli.Config.HTTPTimeout.LongTimeout = time.Second * 300
	cli.Config.IsEnableCRC = false
	cli.Config.UserAgent = UserAgent

	bucket, err := cli.Bucket(bucketName)
	if err != nil {
		return nil, fmt.Errorf("cannot get bucket instance %s: %s", bucketName, err)
	}
	return &ossStore{client: cli, bucket: bucket}, nil
}

func newOIDCCredentialProvider() (oss.CredentialsProvider, error) {
	ok, roleArn, oidcProviderArn, tokenPath := checkOIDCAvailable()
	if !ok {
		log.Error("failed to check oss oidc")
		return nil, fmt.Errorf("no oidc env vars")
	}
	config := new(credentials.Config).
		SetType("oidc_role_arn").
		SetRoleArn(roleArn).
		SetOIDCProviderArn(oidcProviderArn).
		SetOIDCTokenFilePath(tokenPath).SetRoleSessionName("sp-oss")
	oidcCredential, err := credentials.NewCredential(config)
	if err != nil {
		log.Errorw("failed to new oidc credentials", "error", err)
	}
	provider := newOSSCredentialsProvider(oidcCredential)
	return &provider, nil
}

func checkOIDCAvailable() (bool, string, string, string) {
	oidc := true
	roleArn, exists := os.LookupEnv(OSSRoleARN)
	if !exists {
		oidc = false
		log.Error("failed to read oss role arn")
	}
	oidcProviderArn, exists := os.LookupEnv(OSSOidcProviderArn)
	if !exists {
		oidc = false
		log.Error("failed to read oss oidc provider arn")
	}
	tokenPath, exists := os.LookupEnv(OSSWebIdentityTokenFile)
	if !exists {
		oidc = false
		log.Error("failed to read oss web identity token file")
	}
	return oidc, roleArn, oidcProviderArn, tokenPath
}

func parseOSS(bucketURL string) (string, string, string, error) {
	if !strings.Contains(bucketURL, "://") {
		bucketURL = fmt.Sprintf("https://%s", bucketURL)
	}
	uri, err := url.ParseRequestURI(bucketURL)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid bucket: %s, error: %v", bucketURL, err)
	}

	hostParts := strings.SplitN(uri.Host, ".", 2)
	var endpoint string
	if len(hostParts) > 1 {
		endpoint = uri.Scheme + "://" + hostParts[1]
	} else {
		return "", "", "", fmt.Errorf("cannot get oss domain name: %s", bucketURL)
	}
	regionParts := strings.SplitN(hostParts[1], ".", 2)
	if len(regionParts) != 2 {
		return "", "", "", fmt.Errorf("cannot get oss region: %s", bucketURL)
	}
	region := regionParts[0]
	bucketName := hostParts[0]

	return endpoint, bucketName, region, nil
}

type ossCredentials struct {
	AccessKeyID     string
	AccessKeySecret string
	SecurityToken   string
}

type defaultCredentialsProvider struct {
	cred credentials.Credential
}

func (o *ossCredentials) GetAccessKeyID() string {
	return o.AccessKeyID
}

func (o *ossCredentials) GetAccessKeySecret() string {
	return o.AccessKeySecret
}

func (o *ossCredentials) GetSecurityToken() string {
	return o.SecurityToken
}

func (d *defaultCredentialsProvider) GetCredentials() oss.Credentials {
	accessKey, _ := d.cred.GetAccessKeyId()
	secretKey, _ := d.cred.GetAccessKeySecret()
	securityToken, _ := d.cred.GetSecurityToken()
	return &ossCredentials{
		AccessKeyID:     *accessKey,
		AccessKeySecret: *secretKey,
		SecurityToken:   *securityToken,
	}
}

func newOSSCredentialsProvider(credential credentials.Credential) defaultCredentialsProvider {
	return defaultCredentialsProvider{cred: credential}
}
