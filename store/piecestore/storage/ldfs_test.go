package storage

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/assert"
)

const mockLdfsBucket = "ldfsBucket"

func setupLdfsTest(t *testing.T) *ldfsStore {
	return &ldfsStore{s3Store{bucketName: mockLdfsBucket}}
}

func TestLdfsStore_newLdfsStore(t *testing.T) {
	cases := []struct {
		name         string
		cfg          ObjectStorageConfig
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name: "new ldfs store successfully",
			cfg: ObjectStorageConfig{
				Storage:   LdfsStore,
				BucketURL: "https://ldfs.us-east-1.aliyun.com/ldfsBucket",
				IAMType:   AKSKIAMType,
			},
			wantedIsErr:  false,
			wantedErrStr: emptyString,
		},
		{
			name: "failed to new ldfs store",
			cfg: ObjectStorageConfig{
				Storage:   LdfsStore,
				BucketURL: "https://ldfs.us-east-1.aliyun.com/ldfsBucket\r\n",
				IAMType:   AKSKIAMType,
			},
			wantedIsErr:  true,
			wantedErrStr: "net/url: invalid control character in URL",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := newLdfsStore(tt.cfg)
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

func TestLdfsStore_newLdfsSessionWithCache(t *testing.T) {
	cfg := ObjectStorageConfig{
		Storage:   LdfsStore,
		BucketURL: "https://ldfs.us-east-1.aliyun.com/ldfsBucket",
		IAMType:   AKSKIAMType,
	}
	result1, result2, err := ldfsSessionCache.newLdfsSession(cfg)
	defer ldfsSessionCache.clear()
	assert.NotNil(t, result1)
	assert.Equal(t, result2, "ldfsBucket")
	assert.Nil(t, err)

	// get result from map cache
	_, bucketName1, err1 := ldfsSessionCache.newLdfsSession(cfg)
	if err1 != nil {
		t.Fatal(err)
	}
	if bucketName1 != "ldfsBucket" {
		t.Fatalf("expected ldfsBucket, got %v", bucketName1)
	}
}

func TestLdfsStore_newLdfsSession(t *testing.T) {
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
				Storage:   LdfsStore,
				BucketURL: "https://ldfs.us-east-1.aliyun.com/ldfsBucket\r\n",
				IAMType:   AKSKIAMType,
			},
			wantedResult: "",
			wantedIsErr:  true,
			wantedErrStr: "net/url: invalid control character in URL",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result1, result2, err := ldfsSessionCache.newLdfsSession(tt.cfg)
			defer ldfsSessionCache.clear()
			assert.Equal(t, &session.Session{}, result1)
			assert.Equal(t, tt.wantedResult, result2)
			assert.Contains(t, err.Error(), tt.wantedErrStr)
		})
	}
}

func TestLdfsStore_String(t *testing.T) {
	m := setupLdfsTest(t)
	result := m.String()
	assert.Equal(t, "ldfs://ldfsBucket/", result)
}

func TestLdfsStore_parseLdfsBucketURL(t *testing.T) {
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
			name:          "correct ldfs bucket url",
			bucketURL:     "https://ldfs.us-east-1.aliyun.com/ldfsBucket",
			wantedResult1: "ldfs.us-east-1.aliyun.com",
			wantedResult2: "ldfsBucket",
			wantedResult3: true,
			wantedIsErr:   false,
			wantedErrStr:  "",
		},
		{
			name:          "ldfs bucket url without scheme",
			bucketURL:     "ldfs.us-east-1.aliyun.com/ldfsBucket",
			wantedResult1: "ldfs.us-east-1.aliyun.com",
			wantedResult2: "ldfsBucket",
			wantedResult3: false,
			wantedIsErr:   false,
			wantedErrStr:  "",
		},
		{
			name:          "invalid ldfs bucket url",
			bucketURL:     "http://ldfs.us-east-1.aliyun.com/ldfsBucket\n",
			wantedResult1: emptyString,
			wantedResult2: emptyString,
			wantedResult3: false,
			wantedIsErr:   true,
			wantedErrStr:  "net/url: invalid control character in URL",
		},
		{
			name:          "no buket name provided in bucket url",
			bucketURL:     "http://ldfs.us-east-1.aliyun.com",
			wantedResult1: emptyString,
			wantedResult2: emptyString,
			wantedResult3: false,
			wantedIsErr:   true,
			wantedErrStr:  "no bucket name provided in",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result1, result2, result3, err := parseLdfsBucketURL(tt.bucketURL)
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
