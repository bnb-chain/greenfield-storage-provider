package storage

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/assert"
)

const mockB2Bucket = "b2Bucket"

func setupB2Test(t *testing.T) *b2Store {
	return &b2Store{s3Store{bucketName: mockB2Bucket}}
}

func TestB2Store_newB2Store(t *testing.T) {
	cases := []struct {
		name         string
		cfg          ObjectStorageConfig
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name: "succeed to new b2 store",
			cfg: ObjectStorageConfig{
				Storage:   B2Store,
				BucketURL: "https://s3.us-east-1.backblazeb2.com/b2Bucket",
				IAMType:   AKSKIAMType,
			},
			wantedIsErr:  false,
			wantedErrStr: emptyString,
		},
		{
			name: "failed to new b2 store",
			cfg: ObjectStorageConfig{
				Storage:   B2Store,
				BucketURL: "https://s3.us-east-1.backblazeb2.com/b2Bucket\r\n",
				IAMType:   AKSKIAMType,
			},
			wantedIsErr:  true,
			wantedErrStr: "net/url: invalid control character in URL",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := newB2Store(tt.cfg)
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

func TestB2Store_newB2SessionWithCache(t *testing.T) {
	_ = os.Setenv(B2AccessKey, mockAccessKey)
	_ = os.Setenv(B2SecretKey, mockSecretKey)
	_ = os.Setenv(B2SessionToken, mockSessionToken)
	defer os.Unsetenv(B2AccessKey)
	defer os.Unsetenv(B2SecretKey)
	defer os.Unsetenv(B2SessionToken)
	cfg := ObjectStorageConfig{
		Storage:   B2Store,
		BucketURL: "https://s3.us-east-1.backblazeb2.com/b2Bucket",
		IAMType:   AKSKIAMType,
	}
	result1, result2, err := b2SessionCache.newB2Session(cfg)
	defer b2SessionCache.clear()
	assert.NotNil(t, result1)
	assert.Equal(t, result2, "b2Bucket")
	assert.Nil(t, err)

	// get result from map cache
	_, bucketName1, err1 := b2SessionCache.newB2Session(cfg)
	if err1 != nil {
		t.Fatal(err)
	}
	if bucketName1 != "b2Bucket" {
		t.Fatalf("expected b2Bucket, got %v", bucketName1)
	}
}

func TestB2Store_newB2Session(t *testing.T) {
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
				Storage:   B2Store,
				BucketURL: "https://s3.us-east-1.backblazeb2.com/b2Bucket\r\n",
				IAMType:   AKSKIAMType,
			},
			wantedResult: "",
			wantedIsErr:  true,
			wantedErrStr: "net/url: invalid control character in URL",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result1, result2, err := b2SessionCache.newB2Session(tt.cfg)
			defer b2SessionCache.clear()
			assert.Equal(t, &session.Session{}, result1)
			assert.Equal(t, tt.wantedResult, result2)
			assert.Contains(t, err.Error(), tt.wantedErrStr)
		})
	}
}

func TestB2Store_String(t *testing.T) {
	m := setupB2Test(t)
	result := m.String()
	assert.Equal(t, "b2://b2Bucket/", result)
}

func TestB2Store_parseB2BucketURL(t *testing.T) {
	cases := []struct {
		name          string
		bucketURL     string
		wantedResult1 string
		wantedResult2 string
		wantedResult3 string
		wantedIsErr   bool
		wantedErrStr  string
	}{
		{
			name:          "correct b2 bucket url path style",
			bucketURL:     "https://s3.us-east-1.backblazeb2.com/b2Bucket",
			wantedResult1: "s3.us-east-1.backblazeb2.com",
			wantedResult2: "us-east-1",
			wantedResult3: "b2Bucket",
			wantedIsErr:   false,
			wantedErrStr:  "",
		},
		{
			name:          "correct b2 bucket url virtual style",
			bucketURL:     "http://b2Bucket.s3.us-east-1.backblazeb2.com",
			wantedResult1: "s3.us-east-1.backblazeb2.com",
			wantedResult2: "us-east-1",
			wantedResult3: "b2Bucket",
			wantedIsErr:   false,
			wantedErrStr:  "",
		},
		{
			name:          "invalid b2 bucket url",
			bucketURL:     "https://s3.us-east-1.backblazeb2.com/b2Bucket\n",
			wantedResult1: emptyString,
			wantedResult2: emptyString,
			wantedResult3: "",
			wantedIsErr:   true,
			wantedErrStr:  "net/url: invalid control character in URL",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result1, result2, result3, err := parseB2BucketURL(tt.bucketURL)
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
