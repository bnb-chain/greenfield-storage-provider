package storage

import (
	"bytes"
	"compress/flate"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/assert"
)

const (
	mockEndpoint     = "https://s3.mock-region-1.amazonaws.com/mockBucket"
	mockAccessKey    = "mockAccessKey"
	mockSecretKey    = "mockSecretKey"
	mockSessionToken = "mockSessionToken"
	mockS3Bucket     = "s3Bucket"
	mockKey          = "mock"
	mockSize         = 10
	emptyString      = ""
)

var mockModifiedTime = time.Date(2022, time.July, 1, 10, 0, 0, 0, time.UTC)

type mockS3Client struct {
	s3iface.S3API
	createBucketReq  s3.CreateBucketInput
	createBucketResp s3.CreateBucketOutput
	getObjectResp    s3.GetObjectOutput
	putObjectResp    s3.PutObjectOutput
	deleteObjectReq  s3.DeleteObjectInput
	deleteObjectResp s3.DeleteObjectOutput
	headBucketResp   s3.HeadBucketOutput
	headObjectResp   s3.HeadObjectOutput
	listObjectsResp  s3.ListObjectsOutput
}

func (m mockS3Client) CreateBucketWithContext(aws.Context, *s3.CreateBucketInput, ...request.Option) (
	*s3.CreateBucketOutput, error) {
	if *m.createBucketReq.Bucket == "alreadyCreatedBucket" {
		return nil, errors.New("BucketAlreadyExists")
	} else if *m.createBucketReq.Bucket == "alreadyCreatedBucketAndOwnedByYou" {
		return nil, errors.New("BucketAlreadyOwnedByYou")
	}
	return &m.createBucketResp, nil
}

func (m mockS3Client) GetObjectWithContext(aws.Context, *s3.GetObjectInput, ...request.Option) (
	*s3.GetObjectOutput, error) {
	return &m.getObjectResp, nil
}

func (m mockS3Client) PutObjectWithContext(aws.Context, *s3.PutObjectInput, ...request.Option) (
	*s3.PutObjectOutput, error) {
	return &m.putObjectResp, nil
}

func (m mockS3Client) DeleteObjectWithContext(aws.Context, *s3.DeleteObjectInput, ...request.Option) (
	*s3.DeleteObjectOutput, error) {
	if *m.deleteObjectReq.Key == "non_existed_object" {
		return nil, errors.New("NoSuckKey")
	}
	return &m.deleteObjectResp, nil
}

func (m mockS3Client) HeadBucketWithContext(aws.Context, *s3.HeadBucketInput, ...request.Option) (
	*s3.HeadBucketOutput, error) {
	return &m.headBucketResp, nil
}

func (m mockS3Client) HeadObjectWithContext(aws.Context, *s3.HeadObjectInput, ...request.Option) (
	*s3.HeadObjectOutput, error) {
	return &m.headObjectResp, nil
}

func (m mockS3Client) ListObjectsWithContext(aws.Context, *s3.ListObjectsInput, ...request.Option) (
	*s3.ListObjectsOutput, error) {
	return &m.listObjectsResp, nil
}

func setupS3Test(t *testing.T) *s3Store {
	return &s3Store{bucketName: mockS3Bucket}
}

func TestS3Store_String(t *testing.T) {
	store := setupS3Test(t)
	store.api = mockS3Client{}
	result := store.String()
	assert.Equal(t, "s3://s3Bucket/", result)
}

func TestS3Store_newS3Store(t *testing.T) {
	cases := []struct {
		name         string
		cfg          ObjectStorageConfig
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name: "new s3 store successfully",
			cfg: ObjectStorageConfig{
				Storage:   S3Store,
				BucketURL: "http://mockBucket.s3.us-east-1.amazonaws.com",
				IAMType:   AKSKIAMType,
			},
			wantedIsErr:  false,
			wantedErrStr: emptyString,
		},
		{
			name: "failed to new s3 store",
			cfg: ObjectStorageConfig{
				Storage:   S3Store,
				BucketURL: "http://mockBucket.s3.us-east-1.amazonaws.com\r\n",
				IAMType:   AKSKIAMType,
			},
			wantedIsErr:  true,
			wantedErrStr: "net/url: invalid control character in URL",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := newS3Store(tt.cfg)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErrStr)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestS3Store_CreateBucketSuccess(t *testing.T) {
	store := setupS3Test(t)
	cases := []struct {
		name      string
		req       s3.CreateBucketInput
		resp      s3.CreateBucketOutput
		wantedErr error
	}{
		{
			name:      "1",
			req:       s3.CreateBucketInput{Bucket: aws.String(mockS3Bucket)},
			resp:      s3.CreateBucketOutput{Location: aws.String("zh")},
			wantedErr: nil,
		},
		{
			name:      "2",
			req:       s3.CreateBucketInput{Bucket: aws.String("alreadyCreatedBucket")},
			resp:      s3.CreateBucketOutput{},
			wantedErr: nil,
		},
		{
			name:      "3",
			req:       s3.CreateBucketInput{Bucket: aws.String("alreadyCreatedBucketAndOwnedByYou")},
			resp:      s3.CreateBucketOutput{},
			wantedErr: nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store.api = mockS3Client{createBucketReq: tt.req, createBucketResp: tt.resp}
			err := store.CreateBucket(context.TODO())
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestS3Store_GetObjectSuccess(t *testing.T) {
	store := setupS3Test(t)
	cases := []struct {
		name      string
		key       string
		offset    int64
		limit     int64
		resp      s3.GetObjectOutput
		wantedErr error
	}{
		{
			name:      "1",
			key:       mockKey,
			limit:     1,
			resp:      s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader("s3 get"))},
			wantedErr: nil,
		},
		{
			name:      "2",
			key:       mockKey,
			offset:    1,
			limit:     -1,
			resp:      s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader("s3 get"))},
			wantedErr: nil,
		},
		{
			name:   "3",
			key:    mockKey,
			offset: 0,
			limit:  -1,
			resp: s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader("s3 get")),
				Metadata: map[string]*string{
					ChecksumAlgo: aws.String("445758184"),
				}},
			wantedErr: nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store.api = mockS3Client{getObjectResp: tt.resp}
			data, err := store.GetObject(context.TODO(), tt.key, tt.offset, tt.limit)
			assert.Equal(t, tt.wantedErr, err)
			data1, err := io.ReadAll(data)
			if err != nil {
				t.Fatalf("io ReadAll error: %s", err)
			}
			assert.Equal(t, "s3 get", string(data1))
		})
	}
}

func TestS3Store_PutObjectSuccess(t *testing.T) {
	fr := flate.NewReader(strings.NewReader("test"))
	defer fr.Close()
	store := setupS3Test(t)
	cases := []struct {
		name      string
		key       string
		reader    io.Reader
		resp      s3.PutObjectOutput
		wantedErr error
	}{
		{
			name:      "io read seeker",
			key:       mockKey,
			reader:    strings.NewReader("test"),
			resp:      s3.PutObjectOutput{},
			wantedErr: nil,
		},
		{
			name:      "non io read seeker failure",
			key:       mockKey,
			reader:    bytes.NewBufferString("test"),
			resp:      s3.PutObjectOutput{},
			wantedErr: nil,
		},
		{
			name:      "non io read seeker failure",
			key:       mockKey,
			reader:    fr,
			resp:      s3.PutObjectOutput{},
			wantedErr: io.ErrUnexpectedEOF,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store.api = mockS3Client{putObjectResp: tt.resp}
			err := store.PutObject(context.TODO(), tt.key, tt.reader)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestS3Store_DeleteObjectSuccess(t *testing.T) {
	store := setupS3Test(t)
	cases := []struct {
		name      string
		key       string
		req       s3.DeleteObjectInput
		resp      s3.DeleteObjectOutput
		wantedErr error
	}{
		{
			name:      "1",
			key:       mockKey,
			req:       s3.DeleteObjectInput{Key: aws.String(mockKey)},
			resp:      s3.DeleteObjectOutput{},
			wantedErr: nil,
		},
		{
			name:      "2",
			key:       mockKey,
			req:       s3.DeleteObjectInput{Key: aws.String("non_existed_object")},
			resp:      s3.DeleteObjectOutput{},
			wantedErr: nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store.api = mockS3Client{deleteObjectReq: tt.req, deleteObjectResp: tt.resp}
			err := store.DeleteObject(context.TODO(), tt.key)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestS3Store_HeadBucketSuccess(t *testing.T) {
	store := setupS3Test(t)
	store.api = mockS3Client{headObjectResp: s3.HeadObjectOutput{
		ContentType: aws.String("100"),
	}}
	err := store.HeadBucket(context.TODO())
	assert.Nil(t, err)
}

func TestS3Store_HeadObjectSuccess(t *testing.T) {
	store := setupS3Test(t)
	cases := []struct {
		name      string
		key       string
		resp      s3.HeadObjectOutput
		wantedErr error
	}{
		{
			name: "1",
			key:  mockKey,
			resp: s3.HeadObjectOutput{
				ContentLength: aws.Int64(mockSize),
				LastModified:  aws.Time(mockModifiedTime),
			},
			wantedErr: nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store.api = mockS3Client{headObjectResp: tt.resp}
			obj, err := store.HeadObject(context.TODO(), tt.key)
			assert.Equal(t, int64(mockSize), obj.Size())
			assert.Equal(t, mockModifiedTime, obj.ModTime())
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestS3Store_ListObjectsSuccess(t *testing.T) {
	store := setupS3Test(t)
	s3Object := &s3.Object{
		Key:          aws.String(mockKey),
		LastModified: aws.Time(mockModifiedTime),
		Size:         aws.Int64(mockSize),
	}
	objectsList := make([]*s3.Object, 0)
	objectsList = append(objectsList, s3Object)

	object1 := &object{
		key:     mockKey,
		size:    mockSize,
		modTime: mockModifiedTime,
		isDir:   false,
	}
	objList1 := make([]Object, 1)
	objList1[0] = object1

	object2 := &object{
		key:     "mo",
		size:    0,
		modTime: time.Unix(0, 0),
		isDir:   true,
	}
	objList2 := make([]Object, 0)
	objList2 = append(objList2, object2, object1)
	cases := []struct {
		name         string
		prefix       string
		marker       string
		delimiter    string
		resp         s3.ListObjectsOutput
		wantedResult []Object
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name:         "1",
			prefix:       emptyString,
			delimiter:    emptyString,
			marker:       emptyString,
			resp:         s3.ListObjectsOutput{Contents: objectsList},
			wantedResult: objList1,
			wantedIsErr:  false,
			wantedErr:    nil,
		},
		{
			name:      "2",
			prefix:    emptyString,
			marker:    emptyString,
			delimiter: "mo",
			resp: s3.ListObjectsOutput{Contents: objectsList,
				CommonPrefixes: []*s3.CommonPrefix{{Prefix: aws.String("mo")}}},
			wantedResult: objList2,
			wantedIsErr:  false,
			wantedErr:    nil,
		},
		{
			name:      "3",
			prefix:    "l",
			marker:    emptyString,
			delimiter: "mo",
			resp: s3.ListObjectsOutput{Contents: objectsList,
				CommonPrefixes: []*s3.CommonPrefix{{Prefix: aws.String("mo")}}},
			wantedIsErr: true,
			wantedErr:   fmt.Errorf("found invalid key mock from List, prefix: l, marker: "),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store.api = mockS3Client{listObjectsResp: tt.resp}
			objs, err := store.ListObjects(context.TODO(), tt.prefix, emptyString, tt.delimiter, 0)
			assert.Equal(t, tt.wantedErr, err)
			if tt.wantedIsErr {
				assert.Nil(t, objs)
			} else {
				for i, value := range objs {
					assert.Equal(t, tt.wantedResult[i].Key(), value.Key())
					assert.Equal(t, tt.wantedResult[i].ModTime(), value.ModTime())
					assert.Equal(t, tt.wantedResult[i].Size(), value.Size())
				}
			}
		})
	}
}

func TestS3Store_ListAllObjectsSuccess(t *testing.T) {
	store := setupS3Test(t)
	_, err := store.ListAllObjects(context.TODO(), emptyString, emptyString)
	assert.Equal(t, ErrUnsupportedMethod, err)
}

type mockS3ClientError struct {
	s3iface.S3API
	headBucketReq s3.HeadBucketInput
	headObjectReq s3.HeadObjectInput
	// createBucketResp s3.CreateBucketOutput
	// headObjectResp s3.HeadObjectOutput
	// getObjectResp    s3.GetObjectOutput
	// putObjectResp    s3.PutObjectOutput
	// deleteObjectResp s3.DeleteObjectOutput
	// listObjectsResp  s3.ListObjectsOutput
}

func (m mockS3ClientError) CreateBucketWithContext(aws.Context, *s3.CreateBucketInput, ...request.Option) (
	*s3.CreateBucketOutput, error) {
	return nil, errors.New("create bucket error")
}

func (m mockS3ClientError) GetObjectWithContext(aws.Context, *s3.GetObjectInput, ...request.Option) (
	*s3.GetObjectOutput, error) {
	return nil, errors.New("get object error")
}

func (m mockS3ClientError) PutObjectWithContext(aws.Context, *s3.PutObjectInput, ...request.Option) (
	*s3.PutObjectOutput, error) {
	return nil, errors.New("put object error")
}

func (m mockS3ClientError) DeleteObjectWithContext(aws.Context, *s3.DeleteObjectInput, ...request.Option) (
	*s3.DeleteObjectOutput, error) {
	return nil, errors.New("delete object error")
}

func (m mockS3ClientError) HeadBucketWithContext(aws.Context, *s3.HeadBucketInput, ...request.Option) (
	*s3.HeadBucketOutput, error) {
	if *m.headBucketReq.Bucket == mockKey {
		return nil, awserr.NewRequestFailure(awserr.New("NotFound", "mock error", errors.New("head bucket original error")),
			http.StatusNotFound, "mockReqID")
	}
	return nil, errors.New("head bucket error")
}

func (m mockS3ClientError) HeadObjectWithContext(aws.Context, *s3.HeadObjectInput, ...request.Option) (
	*s3.HeadObjectOutput, error) {
	if *m.headObjectReq.Key == mockKey {
		return nil, awserr.NewRequestFailure(awserr.New("NotFound", "mock error", errors.New("head object original error")),
			http.StatusNotFound, "mockReqID")
	}
	return nil, errors.New("head object error")
}

func (m mockS3ClientError) ListObjectsWithContext(aws.Context, *s3.ListObjectsInput, ...request.Option) (
	*s3.ListObjectsOutput, error) {
	return nil, errors.New("list objects error")
}

func TestS3Store_CreateBucketError(t *testing.T) {
	store := setupS3Test(t)
	cases := []struct {
		name      string
		resp      s3.CreateBucketOutput
		wantedErr error
	}{
		{
			name:      "1",
			resp:      s3.CreateBucketOutput{},
			wantedErr: errors.New("create bucket error"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store.api = mockS3ClientError{}
			err := store.CreateBucket(context.TODO())
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestS3Store_GetObjectError(t *testing.T) {
	store := setupS3Test(t)
	cases := []struct {
		name      string
		key       string
		offset    int64
		limit     int64
		resp      s3.GetObjectOutput
		wantedErr error
	}{
		{
			name:      "1",
			key:       mockKey,
			limit:     1,
			resp:      s3.GetObjectOutput{},
			wantedErr: errors.New("get object error"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store.api = mockS3ClientError{}
			data, err := store.GetObject(context.TODO(), mockKey, 0, 0)
			assert.Equal(t, nil, data)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestS3Store_PutObjectError(t *testing.T) {
	store := setupS3Test(t)
	cases := []struct {
		name      string
		key       string
		reader    io.Reader
		resp      s3.PutObjectOutput
		wantedErr error
	}{
		{
			name:      "1",
			key:       mockKey,
			reader:    strings.NewReader("test"),
			resp:      s3.PutObjectOutput{},
			wantedErr: errors.New("put object error"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store.api = mockS3ClientError{}
			err := store.PutObject(context.TODO(), tt.key, tt.reader)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestS3Store_DeleteObjectError(t *testing.T) {
	store := setupS3Test(t)
	cases := []struct {
		name      string
		key       string
		resp      s3.DeleteObjectOutput
		wantedErr error
	}{
		{
			name:      "1",
			resp:      s3.DeleteObjectOutput{},
			wantedErr: errors.New("delete object error"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store.api = mockS3ClientError{}
			err := store.DeleteObject(context.TODO(), tt.key)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestS3Store_HeadBucket(t *testing.T) {
	store := setupS3Test(t)
	cases := []struct {
		name      string
		key       string
		req       s3.HeadBucketInput
		resp      s3.HeadBucketOutput
		wantedErr error
	}{
		{
			name:      "no such bucket",
			req:       s3.HeadBucketInput{Bucket: aws.String(mockKey)},
			resp:      s3.HeadBucketOutput{},
			wantedErr: ErrNoSuchBucket,
		},
		{
			name:      "head bucket error",
			req:       s3.HeadBucketInput{Bucket: aws.String(store.bucketName)},
			resp:      s3.HeadBucketOutput{},
			wantedErr: errors.New("head bucket error"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store.api = mockS3ClientError{headBucketReq: tt.req}
			err := store.HeadBucket(context.TODO())
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestS3Store_HeadObjectError(t *testing.T) {
	store := setupS3Test(t)
	cases := []struct {
		name      string
		key       string
		req       s3.HeadObjectInput
		resp      s3.HeadObjectOutput
		wantedErr error
	}{
		{
			name:      "1",
			key:       mockKey,
			req:       s3.HeadObjectInput{Key: aws.String(mockSecretKey)},
			resp:      s3.HeadObjectOutput{},
			wantedErr: errors.New("head object error"),
		},
		{
			name:      "2",
			key:       mockKey,
			req:       s3.HeadObjectInput{Key: aws.String(mockKey)},
			resp:      s3.HeadObjectOutput{},
			wantedErr: os.ErrNotExist,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store.api = mockS3ClientError{headObjectReq: tt.req}
			obj, err := store.HeadObject(context.TODO(), tt.key)
			assert.Equal(t, nil, obj)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestS3Store_ListObjectsError(t *testing.T) {
	store := setupS3Test(t)
	cases := []struct {
		name      string
		resp      s3.ListObjectsOutput
		wantedErr error
	}{
		{
			name:      "1",
			resp:      s3.ListObjectsOutput{},
			wantedErr: errors.New("list objects error"),
		},
		{
			name:      "2",
			resp:      s3.ListObjectsOutput{},
			wantedErr: errors.New("list objects error"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store.api = mockS3ClientError{}
			objs, err := store.ListObjects(context.TODO(), emptyString, emptyString, emptyString, 0)
			assert.Equal(t, 0, len(objs))
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestS3Store_newSessionAKSKAndCache(t *testing.T) {
	cfg := ObjectStorageConfig{
		Storage:   S3Store,
		BucketURL: mockEndpoint,
		IAMType:   AKSKIAMType,
	}
	_ = os.Setenv(AWSAccessKey, mockAccessKey)
	_ = os.Setenv(AWSSecretKey, mockSecretKey)
	_ = os.Setenv(AWSSessionToken, mockSessionToken)
	defer os.Unsetenv(AWSAccessKey)
	defer os.Unsetenv(AWSSecretKey)
	defer os.Unsetenv(AWSSessionToken)
	_, bucketName, err := s3SessionCache.newSession(cfg)
	defer s3SessionCache.clear()
	if err != nil {
		t.Fatal(err)
	}
	if bucketName != "mockBucket" {
		t.Fatalf("expected mockBucket, got %v", bucketName)
	}

	// get result from map cache
	_, bucketName1, err1 := s3SessionCache.newSession(cfg)
	if err1 != nil {
		t.Fatal(err)
	}
	if bucketName1 != "mockBucket" {
		t.Fatalf("expected mockBucket, got %v", bucketName1)
	}
}

func TestS3Store_newSessionWithSATypeSuccess(t *testing.T) {
	defer s3SessionCache.clear()
	cfg := ObjectStorageConfig{
		Storage:   S3Store,
		BucketURL: mockEndpoint,
		IAMType:   SAIAMType,
	}
	_ = os.Setenv(AWSRoleARN, "mockAWSRoleARN")
	_ = os.Setenv(AWSWebIdentityTokenFile, "mockAWSWebIdentityTokenFile")
	defer os.Unsetenv(AWSRoleARN)
	defer os.Unsetenv(AWSWebIdentityTokenFile)
	_, bucketName, err := s3SessionCache.newSession(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if bucketName != "mockBucket" {
		t.Fatalf("expected mockBucket, got %v", bucketName)
	}

	// get result from map cache
	_, bucketName1, err1 := s3SessionCache.newSession(cfg)
	if err1 != nil {
		t.Fatal(err)
	}
	if bucketName1 != "mockBucket" {
		t.Fatalf("expected mockBucket, got %v", bucketName1)
	}
}

func TestS3Store_newSessionWithNoSignRequest(t *testing.T) {
	_ = os.Setenv(AWSAccessKey, "NoSignRequest")
	defer os.Unsetenv(AWSAccessKey)

	sess, _, err := s3SessionCache.newSession(ObjectStorageConfig{
		Storage:   S3Store,
		BucketURL: mockEndpoint,
		IAMType:   AKSKIAMType,
	})
	defer s3SessionCache.clear()
	if err != nil {
		t.Fatal(err)
	}

	got := sess.Config.Credentials
	expected := credentials.AnonymousCredentials
	if expected != got {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestS3Store_newSessionSomeWrongCases(t *testing.T) {
	cases := []struct {
		name         string
		cfg          ObjectStorageConfig
		wantedResult string
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name: "wrong bucket url",
			cfg: ObjectStorageConfig{
				Storage:   S3Store,
				BucketURL: "http://mockBucket.s3.us-east-1.amazonaws.com\r\n",
				IAMType:   AKSKIAMType,
			},
			wantedResult: "",
			wantedIsErr:  true,
			wantedErrStr: "net/url: invalid control character in URL",
		},
		{
			name: "invalid iam type",
			cfg: ObjectStorageConfig{
				Storage:   S3Store,
				BucketURL: "http://mockBucket.s3.us-east-1.amazonaws.com",
				IAMType:   "unknown",
			},
			wantedResult: "",
			wantedIsErr:  true,
			wantedErrStr: "unknown IAM type: unknown",
		},
		{
			name: "cannot use sa to access s3",
			cfg: ObjectStorageConfig{
				Storage:   S3Store,
				BucketURL: "http://mockBucket.s3.us-east-1.amazonaws.com",
				IAMType:   SAIAMType,
			},
			wantedResult: "",
			wantedIsErr:  true,
			wantedErrStr: "failed to use sa to access s3",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result1, result2, err := s3SessionCache.newSession(tt.cfg)
			defer s3SessionCache.clear()
			assert.Equal(t, &session.Session{}, result1)
			assert.Equal(t, tt.wantedResult, result2)
			assert.Contains(t, err.Error(), tt.wantedErrStr)
		})
	}
}

func TestS3Store_checkIRSAAvailableWithAWSEnv(t *testing.T) {
	_ = os.Setenv(AWSRoleARN, "mockAWSRoleARN")
	_ = os.Setenv(AWSWebIdentityTokenFile, "mockAWSWebIdentityTokenFile")
	defer os.Unsetenv(AWSRoleARN)
	defer os.Unsetenv(AWSWebIdentityTokenFile)
	result1, result2, result3 := checkIRSAAvailable()
	assert.Equal(t, true, result1)
	assert.Equal(t, "mockAWSRoleARN", result2)
	assert.Equal(t, "mockAWSWebIdentityTokenFile", result3)
}

func TestS3Store_checkIRSAAvailableWithoutAWSEnv(t *testing.T) {
	result1, result2, result3 := checkIRSAAvailable()
	assert.Equal(t, false, result1)
	assert.Equal(t, "", result2)
	assert.Equal(t, "", result3)
}

func TestS3Store_parseEndpoint(t *testing.T) {
	cases := []struct {
		name          string
		endpoint      string
		wantedResult1 string
		wantedResult2 string
		wantedResult3 string
		wantedIsErr   bool
		wantedErrStr  string
	}{
		{
			name:          "path style endpoint",
			endpoint:      "http://s3.us-east-1.amazonaws.com/mockBucket",
			wantedResult1: "s3.us-east-1.amazonaws.com",
			wantedResult2: "mockBucket",
			wantedResult3: "us-east-1",
			wantedIsErr:   false,
			wantedErrStr:  "",
		},
		{
			name:          "virtual style endpoint",
			endpoint:      "http://mockBucket.s3.us-east-1.amazonaws.com",
			wantedResult1: "s3.us-east-1.amazonaws.com",
			wantedResult2: "mockBucket",
			wantedResult3: "us-east-1",
			wantedIsErr:   false,
			wantedErrStr:  "",
		},
		{
			name:          "add us east 1 region",
			endpoint:      "http://s3.com/mockBucket",
			wantedResult1: "http://s3.com/mockBucket",
			wantedResult2: "mockBucket",
			wantedResult3: "us-east-1",
			wantedIsErr:   false,
			wantedErrStr:  "",
		},
		{
			name:          "invalid endpoint",
			endpoint:      "http://mockBucket.s3.us-east-1.amazonaws.com\r\n",
			wantedResult1: "",
			wantedResult2: "",
			wantedResult3: "",
			wantedIsErr:   true,
			wantedErrStr:  "net/url: invalid control character in URL",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result1, result2, result3, err := parseEndpoint(tt.endpoint)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErrStr)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.wantedResult1, result1)
			assert.Equal(t, tt.wantedResult2, result2)
			assert.Equal(t, tt.wantedResult3, result3)
		})
	}
}

func TestS3Store_parseRegion(t *testing.T) {
	cases := []struct {
		name         string
		endpoint     string
		wantedResult string
	}{
		{
			name:         "1",
			endpoint:     "s3.ap-northeast-1.amazonaws.com",
			wantedResult: endpoints.ApNortheast1RegionID,
		},
		{
			name:         "2",
			endpoint:     "s3.dualstack.ap-east-1.amazonaws.com",
			wantedResult: endpoints.ApEast1RegionID,
		},
		{
			name:         "3",
			endpoint:     "amazonaws.com",
			wantedResult: endpoints.UsEast1RegionID,
		},
		{
			name:         "4",
			endpoint:     "external-1.amazonaws.com",
			wantedResult: endpoints.UsEast1RegionID,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRegion(tt.endpoint)
			assert.Equal(t, tt.wantedResult, result)
		})
	}
}

func TestS3Store_dialContext(t *testing.T) {
	cases := []struct {
		name         string
		network      string
		address      string
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name:         "1",
			network:      "tcp",
			address:      "1.1.1.1:853",
			wantedIsErr:  false,
			wantedErrStr: "",
		},
		{
			name:         "2",
			network:      "http",
			address:      "1.1.1.1:853",
			wantedIsErr:  true,
			wantedErrStr: "dial http: unknown network http",
		},
		{
			name:         "3",
			network:      "tcp",
			address:      "127",
			wantedIsErr:  true,
			wantedErrStr: "invalid address: 127",
		},
		{
			name:         "4",
			network:      "tcp",
			address:      "nonexistent-hostname:1000",
			wantedIsErr:  true,
			wantedErrStr: "lookup nonexistent-hostname",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dialContext(context.TODO(), tt.network, tt.address)
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErrStr)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
