package storage

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	"github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/viki-org/dnscache"
)

var (
	// Re-used AWS sessions dramatically improve performance
	s3SessionCache = &SessionCache{
		sessions: map[ObjectStorageConfig]*session.Session{},
	}
	disableSSL         bool
	isVirtualHostStyle bool
)

type s3Store struct {
	bucketName string
	api        s3iface.S3API
}

func newS3Store(cfg *ObjectStorageConfig) (ObjectStorage, error) {
	awsSession, bucket, err := s3SessionCache.newSession(*cfg)
	if err != nil {
		log.Errorw("s3 newSession error", "error", err)
		return nil, err
	}
	log.Infow("newS3Store succeeds", "bucket", bucket)

	return &s3Store{bucketName: bucket, api: s3.New(awsSession)}, nil
}

func (s *s3Store) String() string {
	return fmt.Sprintf("s3://%s/", s.bucketName)
}

func (s *s3Store) CreateBucket(ctx context.Context) error {
	_, err := s.api.CreateBucketWithContext(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucketName),
	})
	if err != nil && isErrExists(err) {
		log.Errorw("ObjectStorage S3 CreateBucket error", "error", err)
		err = nil
	}
	return err
}

func isErrExists(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, s3.ErrCodeBucketAlreadyExists) || strings.Contains(msg, s3.ErrCodeBucketAlreadyOwnedByYou)
}

func (s *s3Store) GetObject(ctx context.Context, key string, offset, limit int64) (io.ReadCloser, error) {
	params := &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}
	if offset > 0 || limit > 0 {
		var r string
		if limit > 0 {
			r = fmt.Sprintf("bytes=%d-%d", offset, offset+limit-1)
		} else {
			r = fmt.Sprintf("bytes=%d-", offset)
		}
		params.Range = aws.String(r)
	}
	resp, err := s.api.GetObjectWithContext(ctx, params)
	if err != nil {
		log.Errorw("ObjectStorage S3 GetObject error", "error", err)
		return nil, err
	}
	if offset == 0 && limit == -1 {
		cs := resp.Metadata[model.ChecksumAlgo]
		if cs != nil {
			resp.Body = verifyChecksum(resp.Body, aws.StringValue(cs))
		}
	}
	return resp.Body, nil
}

func (s *s3Store) PutObject(ctx context.Context, key string, reader io.Reader) error {
	var body io.ReadSeeker
	if b, ok := reader.(io.ReadSeeker); ok {
		body = b
	} else {
		data, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}

	checksum := generateChecksum(body)
	params := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(model.OctetStream),
		Metadata:    map[string]*string{model.ChecksumAlgo: aws.String(checksum)},
	}
	_, err := s.api.PutObjectWithContext(ctx, params)
	return err
}

func (s *s3Store) DeleteObject(ctx context.Context, key string) error {
	param := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}
	_, err := s.api.DeleteObjectWithContext(ctx, param)
	if err != nil && strings.Contains(err.Error(), "NoSuckKey") {
		log.Errorw("ObjectStorage S3 DeleteObject error", "error", err)
		err = nil
	}
	return err
}

func (s *s3Store) HeadBucket(ctx context.Context) error {
	if _, err := s.api.HeadBucketWithContext(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucketName),
	}); err != nil {
		log.Errorw("ObjectStorage S3 HeadBucket error", "error", err)
		if reqErr, ok := err.(awserr.RequestFailure); ok {
			if reqErr.StatusCode() == http.StatusNotFound {
				return errors.BucketNotExisted
			}
		}
		return err
	}
	return nil
}

func (s *s3Store) HeadObject(ctx context.Context, key string) (Object, error) {
	param := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}
	r, err := s.api.HeadObjectWithContext(ctx, param)
	if err != nil {
		if e, ok := err.(awserr.RequestFailure); ok && e.StatusCode() == http.StatusNotFound {
			err = os.ErrNotExist
		}
		log.Errorw("ObjectStorage S3 HeadObject error", "error", err)
		return nil, err
	}
	return &object{
		key,
		aws.Int64Value(r.ContentLength),
		aws.TimeValue(r.LastModified),
		strings.HasSuffix(key, "/"),
	}, nil
}

func (s *s3Store) ListObjects(ctx context.Context, prefix, marker, delimiter string, limit int64) ([]Object, error) {
	param := &s3.ListObjectsInput{
		Bucket:    aws.String(s.bucketName),
		Prefix:    aws.String(prefix),
		Marker:    aws.String(marker),
		MaxKeys:   aws.Int64(limit),
		Delimiter: aws.String(delimiter),
	}
	resp, err := s.api.ListObjectsWithContext(ctx, param)
	if err != nil {
		log.Errorw("ObjectStorage S3 ListObjects error", "error", err)
		return nil, err
	}

	n := len(resp.Contents)
	objs := make([]Object, n)
	for i := 0; i < n; i++ {
		o := resp.Contents[i]
		if !strings.HasPrefix(*o.Key, prefix) || *o.Key < marker {
			return nil, fmt.Errorf("found invalid key %s from List, prefix: %s, marker: %s", *o.Key, prefix, marker)
		}
		objs[i] = &object{
			aws.StringValue(o.Key),
			aws.Int64Value(o.Size),
			aws.TimeValue(o.LastModified),
			strings.HasSuffix(aws.StringValue(o.Key), "/"),
		}
	}

	if delimiter != "" {
		for _, p := range resp.CommonPrefixes {
			objs = append(objs, &object{aws.StringValue(p.Prefix), 0, time.Unix(0, 0), true})
		}
		sort.Slice(objs, func(i, j int) bool { return objs[i].Key() < objs[j].Key() })
	}
	return objs, nil
}

func (s *s3Store) ListAllObjects(ctx context.Context, prefix, marker string) (<-chan Object, error) {
	return nil, errors.NotSupportedMethod
}

// SessionCache holds session.Session according to model.ObjectStorage and it synchronizes access/modification
type SessionCache struct {
	sync.Mutex
	sessions map[ObjectStorageConfig]*session.Session
}

// newSession initializes a new AWS session with region fallback and custom config
func (sc *SessionCache) newSession(cfg ObjectStorageConfig) (*session.Session, string, error) {
	sc.Lock()
	defer sc.Unlock()

	if sess, ok := sc.sessions[cfg]; ok {
		return sess, "", nil
	}

	bucketName, region, err := parseEndPoint(cfg.BucketURL)
	if err != nil {
		log.Errorw("s3 parseEndPoint error", "error", err)
		return nil, "", err
	}

	awsConfig := &aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(cfg.BucketURL),
		DisableSSL:       aws.Bool(disableSSL),
		HTTPClient:       getHTTPClient(cfg.TlsInsecureSkipVerify),
		S3ForcePathStyle: aws.Bool(!isVirtualHostStyle),
		Retryer:          newCustomS3Retryer(cfg.MaxRetries, time.Duration(cfg.MinRetryDelay)),
	}
	if cfg.TestMode == false {
		if cfg.NoSignRequest {
			awsConfig.Credentials = credentials.AnonymousCredentials
		} else if cfg.AccessKey != "" && cfg.SecretKey != "" {
			awsConfig.Credentials = credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, cfg.SessionToken)
		}
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, "", fmt.Errorf("Failed to create aws session: %s", err)
	}

	sc.sessions[cfg] = sess
	return sess, bucketName, nil
}

func (sc *SessionCache) clear() {
	sc.Lock()
	defer sc.Unlock()
	sc.sessions = map[ObjectStorageConfig]*session.Session{}
}

func parseEndPoint(endPoint string) (string, string, error) {
	endPoint = strings.Trim(endPoint, "/")
	uri, err := url.ParseRequestURI(endPoint)
	if err != nil {
		log.Errorw("ParseRequestURI error", "endPoint", endPoint, "error", err)
		return "", "", err
	}

	var (
		bucketName string
		region     string
	)
	if uri.Path != "" { // Path style: https://s3.<region>.amazonaws.com(.cn)/<bucket>
		pathParts := strings.Split(uri.Path, "/")
		bucketName = pathParts[1]
		if strings.Contains(uri.Host, ".amazonaws.com") {
			endPoint = uri.Host
			region = parseRegion(endPoint)
		}
		isVirtualHostStyle = false
	} else { // Virtual hosted style: https://<bucketName>.s3.<region>.amazonaws.com(.cn)
		if strings.Contains(uri.Host, ".amazonaws.com") {
			hostParts := strings.SplitN(uri.Host, ".s3", 2)
			bucketName = hostParts[0]
			endPoint = "s3" + hostParts[1]
			region = parseRegion(endPoint)
			isVirtualHostStyle = true
		}
	}

	if region == "" {
		region = endpoints.UsEast1RegionID
	}

	ssl := strings.ToLower(uri.Scheme) == "https"
	if !ssl {
		disableSSL = true
	}

	return bucketName, region, nil
}

func parseRegion(endpoint string) string {
	if strings.HasPrefix(endpoint, "s3-") || strings.HasPrefix(endpoint, "s3.") {
		endpoint = endpoint[3:]
	}
	if strings.HasPrefix(endpoint, "dualstack") {
		endpoint = endpoint[len("dualstack."):]
	}
	if endpoint == "amazonaws.com" {
		endpoint = endpoints.UsEast1RegionID + "." + endpoint
	}
	region := strings.Split(endpoint, ".")[0]
	if region == "external-1" {
		region = endpoints.UsEast1RegionID
	}
	return region
}

func getHTTPClient(tlsInsecureSkipVerify bool) *http.Client {
	resolver := dnscache.New(time.Minute)
	rand.Seed(time.Now().Unix())

	return &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: tlsInsecureSkipVerify},
			TLSHandshakeTimeout:   time.Second * 20,
			ResponseHeaderTimeout: time.Second * 30,
			IdleConnTimeout:       time.Second * 300,
			MaxIdleConnsPerHost:   500,
			DialContext: func(ctx context.Context, network string, address string) (net.Conn, error) {
				separator := strings.LastIndex(address, ":")
				host := address[:separator]
				port := address[separator:]
				ips, err := resolver.Fetch(host)
				if err != nil {
					return nil, err
				}
				if len(ips) == 0 {
					return nil, fmt.Errorf("No such host: %s", host)
				}

				var conn net.Conn
				n := len(ips)
				first := rand.Intn(n)
				dialer := &net.Dialer{Timeout: time.Second * 10}
				for i := 0; i < n; i++ {
					ip := ips[(first+i)%n]
					address = ip.String()
					if port != "" {
						address = net.JoinHostPort(address, port[1:])
					}
					conn, err = dialer.DialContext(ctx, network, address)
					if err == nil {
						return conn, nil
					}
				}
				return nil, err
			},
			DisableCompression: true,
		},
		Timeout: time.Hour,
	}
}
