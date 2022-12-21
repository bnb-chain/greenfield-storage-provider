package storage

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	errors2 "github.com/bnb-chain/inscription-storage-provider/model"
	"github.com/stretchr/testify/assert"
)

const (
	endPoint     = "https://s3.mock-region-1.amazonaws.com/mockBucket"
	accessKey    = "mockAccessKey"
	secretKey    = "mockSecretKey"
	sessionToken = "mockSessionToken"
	mockBucket   = "mockBucket"
	mockKey      = "mock"
	mockSize     = 10
	emptyString  = ""
)

var modifiedTime = time.Date(2022, time.July, 1, 10, 0, 0, 0, time.UTC)

type mockS3Client struct {
	s3iface.S3API
	createBucketReq  s3.CreateBucketInput
	createBucketResp s3.CreateBucketOutput
	headObjectResp   s3.HeadObjectOutput
	getObjectResp    s3.GetObjectOutput
	putObjectResp    s3.PutObjectOutput
	deleteObjectReq  s3.DeleteObjectInput
	deleteObjectResp s3.DeleteObjectOutput
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

func (m mockS3Client) HeadObjectWithContext(aws.Context, *s3.HeadObjectInput, ...request.Option) (
	*s3.HeadObjectOutput, error) {
	return &m.headObjectResp, nil
}

func (m mockS3Client) ListObjectsWithContext(aws.Context, *s3.ListObjectsInput, ...request.Option) (
	*s3.ListObjectsOutput, error) {
	return &m.listObjectsResp, nil
}

func setupS3Test(t *testing.T) *s3Store {
	return &s3Store{bucketName: mockBucket}
}

func TestS3_String(t *testing.T) {
	store := setupS3Test(t)
	store.api = mockS3Client{}
	result := store.String()
	assert.Equal(t, "s3://mockBucket/", result)
}

func TestS3_CreateSuccess(t *testing.T) {
	store := setupS3Test(t)
	cases := []struct {
		name      string
		req       s3.CreateBucketInput
		resp      s3.CreateBucketOutput
		wantedErr error
	}{
		{
			name:      "1",
			req:       s3.CreateBucketInput{Bucket: aws.String(mockBucket)},
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

func TestS3_GetSuccess(t *testing.T) {
	store := setupS3Test(t)
	body := io.NopCloser(strings.NewReader("s3 get"))
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
			resp:      s3.GetObjectOutput{Body: body},
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
				t.Fatalf("Get io.ReadAll error: %s", err)
			}
			assert.Equal(t, "s3 get", string(data1))
		})
	}
}

func TestS3_PutSuccess(t *testing.T) {
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
			wantedErr: nil,
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

func TestS3_DeleteSuccess(t *testing.T) {
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

func TestS3_HeadSuccess(t *testing.T) {
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
				LastModified:  aws.Time(modifiedTime),
			},
			wantedErr: nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store.api = mockS3Client{headObjectResp: tt.resp}
			obj, err := store.HeadObject(context.TODO(), tt.key)
			assert.Equal(t, int64(mockSize), obj.Size())
			assert.Equal(t, modifiedTime, obj.ModTime())
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestS3_ListSuccess(t *testing.T) {
	store := setupS3Test(t)
	s3Object := &s3.Object{
		Key:          aws.String(mockKey),
		LastModified: aws.Time(modifiedTime),
		Size:         aws.Int64(mockSize),
	}
	objectsList := make([]*s3.Object, 0)
	objectsList = append(objectsList, s3Object)
	cases := []struct {
		name      string
		resp      s3.ListObjectsOutput
		wantedErr error
	}{
		{
			name:      "1",
			resp:      s3.ListObjectsOutput{Contents: objectsList},
			wantedErr: nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store.api = mockS3Client{listObjectsResp: tt.resp}
			objs, err := store.ListObjects(context.TODO(), emptyString, emptyString, emptyString, 0)
			assert.Equal(t, mockKey, objs[0].Key())
			assert.Equal(t, modifiedTime, objs[0].ModTime())
			assert.Equal(t, int64(mockSize), objs[0].Size())
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestS3_ListAll(t *testing.T) {
	store := setupS3Test(t)
	_, err := store.ListAllObjects(context.TODO(), emptyString, emptyString)
	assert.Equal(t, errors2.NotSupportedMethod, err)
}

type mockS3ClientError struct {
	s3iface.S3API
	createBucketResp s3.CreateBucketOutput
	headObjectResp   s3.HeadObjectOutput
	getObjectResp    s3.GetObjectOutput
	putObjectResp    s3.PutObjectOutput
	deleteObjectResp s3.DeleteObjectOutput
	listObjectsResp  s3.ListObjectsOutput
}

func (m mockS3ClientError) CreateBucketWithContext(aws.Context, *s3.CreateBucketInput, ...request.Option) (
	*s3.CreateBucketOutput, error) {
	return nil, errors.New("Create bucket error")
}

func (m mockS3ClientError) GetObjectWithContext(aws.Context, *s3.GetObjectInput, ...request.Option) (
	*s3.GetObjectOutput, error) {
	return nil, errors.New("Get object error")
}

func (m mockS3ClientError) PutObjectWithContext(aws.Context, *s3.PutObjectInput, ...request.Option) (
	*s3.PutObjectOutput, error) {
	return nil, errors.New("Put object error")
}

func (m mockS3ClientError) DeleteObjectWithContext(aws.Context, *s3.DeleteObjectInput, ...request.Option) (
	*s3.DeleteObjectOutput, error) {
	return nil, errors.New("Delete object error")
}

func (m mockS3ClientError) HeadObjectWithContext(aws.Context, *s3.HeadObjectInput, ...request.Option) (
	*s3.HeadObjectOutput, error) {
	return nil, errors.New("Head object error")
}

func (m mockS3ClientError) ListObjectsWithContext(aws.Context, *s3.ListObjectsInput, ...request.Option) (
	*s3.ListObjectsOutput, error) {
	return nil, errors.New("List objects error")
}

func TestS3_CreateError(t *testing.T) {
	store := setupS3Test(t)
	cases := []struct {
		name      string
		resp      s3.CreateBucketOutput
		wantedErr error
	}{
		{
			name:      "1",
			resp:      s3.CreateBucketOutput{},
			wantedErr: errors.New("Create bucket error"),
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

func TestS3_GetError(t *testing.T) {
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
			wantedErr: errors.New("Get object error"),
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

func TestS3_PutError(t *testing.T) {
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
			wantedErr: errors.New("Put object error"),
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

func TestS3_DeleteError(t *testing.T) {
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
			wantedErr: errors.New("Delete object error"),
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

func TestS3_HeadError(t *testing.T) {
	store := setupS3Test(t)
	cases := []struct {
		name      string
		key       string
		resp      s3.HeadObjectOutput
		wantedErr error
	}{
		{
			name:      "1",
			key:       mockKey,
			resp:      s3.HeadObjectOutput{},
			wantedErr: errors.New("Head object error"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store.api = mockS3ClientError{}
			obj, err := store.HeadObject(context.TODO(), tt.key)
			assert.Equal(t, nil, obj)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestS3_ListError(t *testing.T) {
	store := setupS3Test(t)
	cases := []struct {
		name      string
		resp      s3.ListObjectsOutput
		wantedErr error
	}{
		{
			name:      "1",
			resp:      s3.ListObjectsOutput{},
			wantedErr: errors.New("List objects error"),
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
