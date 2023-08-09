package storage

import (
	"errors"
	"os"
	"testing"

	credentials_aliyun "github.com/aliyun/credentials-go/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/assert"
)

const mockOssBucket = "ossBucket"

func setupOssTest(t *testing.T) *aliyunfsStore {
	return &aliyunfsStore{s3Store{bucketName: mockOssBucket}}
}

func TestOssStore_newAliyunCredProvider(t *testing.T) {
	_ = os.Setenv(credentials_aliyun.EnvVarAccessKeyId, "accesskey")
	_ = os.Setenv(credentials_aliyun.EnvVarAccessKeySecret, "accesssecret")
	defer os.Unsetenv(credentials_aliyun.EnvVarAccessKeyId)
	defer os.Unsetenv(credentials_aliyun.EnvVarAccessKeySecret)
	cred, err := credentials_aliyun.NewCredential(nil)
	assert.Nil(t, err)
	p := newAliyunCredProvider(cred)
	assert.NotNil(t, p)
}

func TestOssStore_RetrieveAndExpired(t *testing.T) {
	_ = os.Setenv(credentials_aliyun.EnvVarAccessKeyId, "accesskey")
	_ = os.Setenv(credentials_aliyun.EnvVarAccessKeySecret, "accesssecret")
	defer os.Unsetenv(credentials_aliyun.EnvVarAccessKeyId)
	defer os.Unsetenv(credentials_aliyun.EnvVarAccessKeySecret)
	cred, err := credentials_aliyun.NewCredential(nil)
	assert.Nil(t, err)
	p := newAliyunCredProvider(cred)
	assert.NotNil(t, p)
	v, err := p.Retrieve()
	assert.NotNil(t, v)
	assert.Nil(t, err)
	ok := p.IsExpired()
	assert.Equal(t, false, ok)
}

func TestOssStore_newAliyunfsStore(t *testing.T) {
	cases := []struct {
		name         string
		cfg          ObjectStorageConfig
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name: "new oss store successfully",
			cfg: ObjectStorageConfig{
				Storage:   AliyunfsStore,
				BucketURL: "https://oss-cn-hangzhou.aliyuncs.com/ossBucket",
				IAMType:   AKSKIAMType,
			},
			wantedIsErr:  false,
			wantedErrStr: emptyString,
		},
		{
			name: "failed to new oss store",
			cfg: ObjectStorageConfig{
				Storage:   AliyunfsStore,
				BucketURL: "https://oss-cn-hangzhou.aliyuncs.com/ossBucket\r\n",
				IAMType:   AKSKIAMType,
			},
			wantedIsErr:  true,
			wantedErrStr: "net/url: invalid control character in URL",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := newAliyunfsStore(tt.cfg)
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

func TestOssStore_newAliyunfsSessionAKSKWithCache(t *testing.T) {
	_ = os.Setenv(AliyunAccessKey, mockAccessKey)
	_ = os.Setenv(AliyunSecretKey, mockSecretKey)
	_ = os.Setenv(AliyunSessionToken, mockSessionToken)
	defer os.Unsetenv(AliyunAccessKey)
	defer os.Unsetenv(AliyunSecretKey)
	defer os.Unsetenv(AliyunSessionToken)
	cfg := ObjectStorageConfig{
		Storage:   AliyunfsStore,
		BucketURL: "https://oss-cn-hangzhou.aliyuncs.com/ossBucket",
		IAMType:   AKSKIAMType,
	}
	result1, result2, err := aliyunfsSessionCache.newAliyunfsSession(cfg)
	defer aliyunfsSessionCache.clear()
	assert.NotNil(t, result1)
	assert.Equal(t, result2, "ossBucket")
	assert.Nil(t, err)

	// get result from map cache
	_, bucketName1, err1 := aliyunfsSessionCache.newAliyunfsSession(cfg)
	if err1 != nil {
		t.Fatal(err)
	}
	if bucketName1 != "ossBucket" {
		t.Fatalf("expected ossBucket, got %v", bucketName1)
	}
}

func TestOssStore_newSessionWithNoSignRequest(t *testing.T) {
	_ = os.Setenv(AliyunAccessKey, "NoSignRequest")
	defer os.Unsetenv(AWSAccessKey)

	sess, _, err := aliyunfsSessionCache.newAliyunfsSession(ObjectStorageConfig{
		Storage:   AliyunfsStore,
		BucketURL: mockEndpoint,
		IAMType:   AKSKIAMType,
	})
	defer aliyunfsSessionCache.clear()
	if err != nil {
		t.Fatal(err)
	}

	got := sess.Config.Credentials
	expected := credentials.AnonymousCredentials
	if expected != got {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestOssStore_newSessionWithSATypeSuccess(t *testing.T) {
	defer aliyunfsSessionCache.clear()
	cfg := ObjectStorageConfig{
		Storage:   AliyunfsStore,
		BucketURL: "https://oss-cn-hangzhou.aliyuncs.com/ossBucket",
		IAMType:   SAIAMType,
	}
	_ = os.Setenv(credentials_aliyun.EnvVarAccessKeyId, "accesskey")
	_ = os.Setenv(credentials_aliyun.EnvVarAccessKeySecret, "accesssecret")
	_ = os.Setenv(AliyunRoleARN, "mockOssRoleARN")
	_ = os.Setenv(AliyunWebIdentityTokenFile, "mockOssWebIdentityTokenFile")
	defer os.Unsetenv(credentials_aliyun.EnvVarAccessKeyId)
	defer os.Unsetenv(credentials_aliyun.EnvVarAccessKeySecret)
	defer os.Unsetenv(AliyunRoleARN)
	defer os.Unsetenv(AliyunWebIdentityTokenFile)
	_, bucketName, err := aliyunfsSessionCache.newAliyunfsSession(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if bucketName != "ossBucket" {
		t.Fatalf("expected ossBucket, got %v", bucketName)
	}

	// get result from map cache
	_, bucketName1, err1 := aliyunfsSessionCache.newAliyunfsSession(cfg)
	if err1 != nil {
		t.Fatal(err)
	}
	if bucketName1 != "ossBucket" {
		t.Fatalf("expected ossBucket, got %v", bucketName1)
	}
}

func TestOssStore_newSessionWithSATypeFailure(t *testing.T) {
	defer aliyunfsSessionCache.clear()
	cfg := ObjectStorageConfig{
		Storage:   AliyunfsStore,
		BucketURL: "https://oss-cn-hangzhou.aliyuncs.com/ossBucket",
		IAMType:   SAIAMType,
	}
	_ = os.Setenv(AliyunRoleARN, "mockOssRoleARN")
	_ = os.Setenv(AliyunWebIdentityTokenFile, "mockOssWebIdentityTokenFile")
	defer os.Unsetenv(AliyunRoleARN)
	defer os.Unsetenv(AliyunWebIdentityTokenFile)
	_, bucketName, err := aliyunfsSessionCache.newAliyunfsSession(cfg)
	assert.Equal(t, emptyString, bucketName)
	assert.Equal(t, err, errors.New("No credential found"))
}

func TestOssStore_newAliyunfsSession(t *testing.T) {
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
				Storage:   AliyunfsStore,
				BucketURL: "https://oss-cn-hangzhou.aliyuncs.com/ossBucket",
				IAMType:   "unknown",
			},
			wantedResult: "",
			wantedIsErr:  true,
			wantedErrStr: "unknown IAM type: unknown",
		},
		{
			name: "cannot access oss use sa",
			cfg: ObjectStorageConfig{
				Storage:   AliyunfsStore,
				BucketURL: "https://oss-cn-hangzhou.aliyuncs.com/ossBucket",
				IAMType:   SAIAMType,
			},
			wantedResult: "",
			wantedIsErr:  true,
			wantedErrStr: "failed to use sa to access aliyunfs",
		},
		{
			name: "wrong bucket url",
			cfg: ObjectStorageConfig{
				Storage:   AliyunfsStore,
				BucketURL: "https://oss-cn-hangzhou.aliyuncs.com/ossBucket\r\n",
				IAMType:   AKSKIAMType,
			},
			wantedResult: "",
			wantedIsErr:  true,
			wantedErrStr: "net/url: invalid control character in URL",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result1, result2, err := aliyunfsSessionCache.newAliyunfsSession(tt.cfg)
			defer aliyunfsSessionCache.clear()
			assert.Equal(t, &session.Session{}, result1)
			assert.Equal(t, tt.wantedResult, result2)
			assert.Contains(t, err.Error(), tt.wantedErrStr)
		})
	}
}

func TestOssStore_String(t *testing.T) {
	m := setupOssTest(t)
	result := m.String()
	assert.Equal(t, "aliyunfs://ossBucket/", result)
}

func TestOssStore_parseAliyunfsBucketURL(t *testing.T) {
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
			name:          "correct oss bucket url",
			bucketURL:     "https://oss-cn-hangzhou.aliyuncs.com/ossBucket",
			wantedResult1: "oss-cn-hangzhou.aliyuncs.com",
			wantedResult2: "ossBucket",
			wantedResult3: true,
			wantedIsErr:   false,
			wantedErrStr:  "",
		},
		{
			name:          "oss bucket url without scheme",
			bucketURL:     "oss-cn-hangzhou.aliyuncs.com/ossBucket",
			wantedResult1: "oss-cn-hangzhou.aliyuncs.com",
			wantedResult2: "ossBucket",
			wantedResult3: false,
			wantedIsErr:   false,
			wantedErrStr:  "",
		},
		{
			name:          "invalid oss bucket url",
			bucketURL:     "https://examplebucket.oss-cn-hangzhou.aliyuncs.com/\n",
			wantedResult1: emptyString,
			wantedResult2: emptyString,
			wantedResult3: false,
			wantedIsErr:   true,
			wantedErrStr:  "net/url: invalid control character in URL",
		},
		{
			name:          "no buket name provided in bucket url",
			bucketURL:     "http://oss-cn-hangzhou.aliyuncs.com",
			wantedResult1: emptyString,
			wantedResult2: emptyString,
			wantedResult3: false,
			wantedIsErr:   true,
			wantedErrStr:  "no bucket name provided in",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result1, result2, result3, err := parseAliyunfsBucketURL(tt.bucketURL)
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

func Test_checkAliyunAvailableWithEnv(t *testing.T) {
	_ = os.Setenv(AliyunRoleARN, "mockOssRoleARN")
	_ = os.Setenv(AliyunWebIdentityTokenFile, "mockOssWebIdentityTokenFile")
	defer os.Unsetenv(AliyunRoleARN)
	defer os.Unsetenv(AliyunWebIdentityTokenFile)
	result1, result2, result3 := checkAliyunAvailable()
	assert.Equal(t, true, result1)
	assert.Equal(t, "mockOssRoleARN", result2)
	assert.Equal(t, "mockOssWebIdentityTokenFile", result3)
}

func Test_checkAliyunAvailableWithoutEnv(t *testing.T) {
	result1, result2, result3 := checkAliyunAvailable()
	assert.Equal(t, false, result1)
	assert.Equal(t, "", result2)
	assert.Equal(t, "", result3)
}
