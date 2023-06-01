package storage

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
)

func setupB2Test(t *testing.T) *b2Store {
	return &b2Store{s3Store{bucketName: mockBucket}}
}

func TestB2_CreateSuccess(t *testing.T) {
	store := setupB2Test(t)
	cases := []struct {
		name      string
		req       s3.CreateBucketInput
		resp      s3.CreateBucketOutput
		wantedErr error
	}{
		{
			name:      "create bucket success",
			req:       s3.CreateBucketInput{Bucket: aws.String(mockBucket)},
			resp:      s3.CreateBucketOutput{Location: aws.String("zh")},
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

func TestB2_GetSuccess(t *testing.T) {
	store := setupB2Test(t)
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
			name:      "create object success",
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

func TestB2_PutSuccess(t *testing.T) {
	store := setupB2Test(t)
	cases := []struct {
		name      string
		key       string
		reader    io.Reader
		resp      s3.PutObjectOutput
		wantedErr error
	}{
		{
			name:      "put object success",
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

func TestB2_DeleteSuccess(t *testing.T) {
	store := setupB2Test(t)
	cases := []struct {
		name      string
		key       string
		req       s3.DeleteObjectInput
		resp      s3.DeleteObjectOutput
		wantedErr error
	}{
		{
			name:      "delete object success",
			key:       mockKey,
			req:       s3.DeleteObjectInput{Key: aws.String(mockKey)},
			resp:      s3.DeleteObjectOutput{},
			wantedErr: nil,
		},
		{
			name:      "delete non-existed object success",
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

func TestB2_HeadSuccess(t *testing.T) {
	store := setupB2Test(t)
	cases := []struct {
		name      string
		key       string
		resp      s3.HeadObjectOutput
		wantedErr error
	}{
		{
			name: "head object success",
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

func TestB2_ListSuccess(t *testing.T) {
	store := setupB2Test(t)
	s3Object := &s3.Object{
		Key:          aws.String(mockKey),
		LastModified: aws.Time(mockModifiedTime),
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
			name:      "list objects success",
			resp:      s3.ListObjectsOutput{Contents: objectsList},
			wantedErr: nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			store.api = mockS3Client{listObjectsResp: tt.resp}
			objs, err := store.ListObjects(context.TODO(), emptyString, emptyString, emptyString, 0)
			assert.Equal(t, mockKey, objs[0].Key())
			assert.Equal(t, mockModifiedTime, objs[0].ModTime())
			assert.Equal(t, int64(mockSize), objs[0].Size())
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestB2_ListAll(t *testing.T) {
	store := setupB2Test(t)
	_, err := store.ListAllObjects(context.TODO(), emptyString, emptyString)
	assert.Equal(t, ErrUnsupportedMethod, err)
}

func TestB2_CreateError(t *testing.T) {
	store := setupB2Test(t)
	cases := []struct {
		name      string
		resp      s3.CreateBucketOutput
		wantedErr error
	}{
		{
			name:      "create bucket error",
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

func TestB2_GetError(t *testing.T) {
	store := setupB2Test(t)
	cases := []struct {
		name      string
		key       string
		offset    int64
		limit     int64
		resp      s3.GetObjectOutput
		wantedErr error
	}{
		{
			name:      "get object error",
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

func TestB2_PutError(t *testing.T) {
	store := setupB2Test(t)
	cases := []struct {
		name      string
		key       string
		reader    io.Reader
		resp      s3.PutObjectOutput
		wantedErr error
	}{
		{
			name:      "put object error",
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

func TestB2_DeleteError(t *testing.T) {
	store := setupB2Test(t)
	cases := []struct {
		name      string
		key       string
		resp      s3.DeleteObjectOutput
		wantedErr error
	}{
		{
			name:      "delete object error",
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

func TestB2_HeadError(t *testing.T) {
	store := setupB2Test(t)
	cases := []struct {
		name      string
		key       string
		resp      s3.HeadObjectOutput
		wantedErr error
	}{
		{
			name:      "head object error",
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

func TestB2_ListError(t *testing.T) {
	store := setupB2Test(t)
	cases := []struct {
		name      string
		resp      s3.ListObjectsOutput
		wantedErr error
	}{
		{
			name:      "list objects error",
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
