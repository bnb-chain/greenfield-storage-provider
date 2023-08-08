package storage

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/assert"
)

const mockMinioBucket = "minioBucket"

func setupMinioTest(t *testing.T) *minioStore {
	return &minioStore{s3Store{bucketName: mockMinioBucket}}
}

func TestMinioStore_newMinioStore(t *testing.T) {
	cases := []struct {
		name         string
		cfg          ObjectStorageConfig
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name: "new minio store successfully",
			cfg: ObjectStorageConfig{
				Storage:   MinioStore,
				BucketURL: "http://minio.us-east-1.amazonaws.com/minioBucket",
				IAMType:   AKSKIAMType,
			},
			wantedIsErr:  false,
			wantedErrStr: emptyString,
		},
		{
			name: "failed to new minio store",
			cfg: ObjectStorageConfig{
				Storage:   MinioStore,
				BucketURL: "http://minio.us-east-1.amazonaws.com/minioBucket\r\n",
				IAMType:   AKSKIAMType,
			},
			wantedIsErr:  true,
			wantedErrStr: "net/url: invalid control character in URL",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := newMinioStore(tt.cfg)
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

func TestMinioStore_newMinioSessionWithCache(t *testing.T) {
	_ = os.Setenv(MinioAccessKey, mockAccessKey)
	_ = os.Setenv(MinioSecretKey, mockSecretKey)
	_ = os.Setenv(MinioSessionToken, mockSessionToken)
	defer os.Unsetenv(MinioAccessKey)
	defer os.Unsetenv(MinioSecretKey)
	defer os.Unsetenv(MinioSessionToken)
	cfg := ObjectStorageConfig{
		Storage:   MinioStore,
		BucketURL: "https://minio.us-east-1.amazonaws.com/minioBucket",
		IAMType:   AKSKIAMType,
	}
	result1, result2, err := minioSessionCache.newMinioSession(cfg)
	defer minioSessionCache.clear()
	assert.NotNil(t, result1)
	assert.Equal(t, result2, "minioBucket")
	assert.Nil(t, err)

	// get result from map cache
	_, bucketName1, err1 := minioSessionCache.newMinioSession(cfg)
	if err1 != nil {
		t.Fatal(err)
	}
	if bucketName1 != "minioBucket" {
		t.Fatalf("expected minioBucket, got %v", bucketName1)
	}
}

func TestMinioStore_newMinioSession(t *testing.T) {
	cases := []struct {
		name         string
		cfg          ObjectStorageConfig
		wantedResult string
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name: "invalid iam type",
			cfg: ObjectStorageConfig{
				Storage:   MinioStore,
				BucketURL: "http://mockBucket.s3.us-east-1.amazonaws.com",
				IAMType:   "unknown",
			},
			wantedResult: "",
			wantedIsErr:  true,
			wantedErrStr: "minio now only supports AKSK iam type",
		},
		{
			name: "wrong bucket url",
			cfg: ObjectStorageConfig{
				Storage:   MinioStore,
				BucketURL: "http://mockBucket.s3.us-east-1.amazonaws.com\r\n",
				IAMType:   AKSKIAMType,
			},
			wantedResult: "",
			wantedIsErr:  true,
			wantedErrStr: "net/url: invalid control character in URL",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result1, result2, err := minioSessionCache.newMinioSession(tt.cfg)
			defer minioSessionCache.clear()
			assert.Equal(t, &session.Session{}, result1)
			assert.Equal(t, tt.wantedResult, result2)
			assert.Contains(t, err.Error(), tt.wantedErrStr)
		})
	}
}

func TestMinioStore_String(t *testing.T) {
	m := setupMinioTest(t)
	result := m.String()
	assert.Equal(t, "minio://minioBucket/", result)
}

func TestMinioStore_parseMinioBucketURL(t *testing.T) {
	cases := []struct {
		name          string
		bucketURL     string
		wantedResult1 string
		wantedResult2 string
		wantedResult3 bool
		wantedIsErr   bool
		wantedErrStr  string
	}{
		{
			name:          "correct minio bucket url",
			bucketURL:     "https://minio.us-east-1.amazonaws.com/minioBucket",
			wantedResult1: "minio.us-east-1.amazonaws.com",
			wantedResult2: "minioBucket",
			wantedResult3: true,
			wantedIsErr:   false,
			wantedErrStr:  "",
		},
		{
			name:          "minio bucket url without scheme",
			bucketURL:     "minio.us-east-1.amazonaws.com/minioBucket",
			wantedResult1: "minio.us-east-1.amazonaws.com",
			wantedResult2: "minioBucket",
			wantedResult3: false,
			wantedIsErr:   false,
			wantedErrStr:  "",
		},
		{
			name:          "invalid minio bucket url",
			bucketURL:     "http://minio.us-east-1.amazonaws.com/minioBucket\n",
			wantedResult1: emptyString,
			wantedResult2: emptyString,
			wantedResult3: false,
			wantedIsErr:   true,
			wantedErrStr:  "net/url: invalid control character in URL",
		},
		{
			name:          "no buket name provided in bucket url",
			bucketURL:     "http://minio.us-east-1.amazonaws.com",
			wantedResult1: emptyString,
			wantedResult2: emptyString,
			wantedResult3: false,
			wantedIsErr:   true,
			wantedErrStr:  "no bucket name provided in",
		},
		{
			name:          "bucket name with extra key",
			bucketURL:     "http://minio.us-east-1.amazonaws.com/minio/abc",
			wantedResult1: "minio.us-east-1.amazonaws.com",
			wantedResult2: "abc",
			wantedResult3: false,
			wantedIsErr:   false,
			wantedErrStr:  "",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result1, result2, result3, err := parseMinioBucketURL(tt.bucketURL)
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
