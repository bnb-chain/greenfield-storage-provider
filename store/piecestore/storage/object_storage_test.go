package storage

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/stretchr/testify/assert"
)

func TestNewObjectStorage(t *testing.T) {
	cases := []struct {
		name        string
		cfg         ObjectStorageConfig
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name: "memory storage",
			cfg: ObjectStorageConfig{
				Storage:   MemoryStore,
				BucketURL: "mock",
				IAMType:   AKSKIAMType,
			},
			wantedIsErr: false,
			wantedErr:   nil,
		},
		{
			name: "invalid storage",
			cfg: ObjectStorageConfig{
				Storage:   "unknown",
				BucketURL: "mock",
				IAMType:   AKSKIAMType,
			},
			wantedIsErr: true,
			wantedErr:   errors.New("invalid object storage: unknown"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewObjectStorage(tt.cfg)
			assert.Equal(t, tt.wantedErr, err)
			if tt.wantedIsErr {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestDefaultObjectStorage_CreateBucket(t *testing.T) {
	d := DefaultObjectStorage{}
	err := d.CreateBucket(context.Background())
	assert.Nil(t, err)
}

func Test_file(t *testing.T) {
	f := &file{
		object: object{key: "mock", size: 10, modTime: time.Date(2023, 8, 3, 8, 0, 0, 0, time.UTC)},
		group:  "1",
		owner:  "2",
		mode:   0755,
	}
	owner := f.Owner()
	assert.Equal(t, f.owner, owner)
	group := f.Group()
	assert.Equal(t, f.group, group)
	mode := f.Mode()
	assert.Equal(t, f.mode, mode)
	isSymlink := f.IsSymlink()
	assert.Equal(t, f.isSymlink, isSymlink)
	isDir := f.IsDir()
	assert.Equal(t, f.isDir, isDir)
}

func Test_getSecretKeyFromEnv(t *testing.T) {
	var (
		ak = "ak"
		sk = "sk"
		st = "st"
	)
	os.Setenv(ak, mockAccessKey)
	os.Setenv(sk, mockSecretKey)
	os.Setenv(st, mockSessionToken)
	defer os.Unsetenv(ak)
	defer os.Unsetenv(sk)
	defer os.Unsetenv(st)

	result := getSecretKeyFromEnv(ak, sk, st)
	assert.Equal(t, mockAccessKey, result.accessKey)
}

func Test_getAliyunSecretKeyFromEnv(t *testing.T) {
	var (
		region = "region"
		ak     = "ak"
		sk     = "sk"
		st     = "st"
	)
	os.Setenv(region, endpoints.UsEast1RegionID)
	os.Setenv(ak, mockAccessKey)
	os.Setenv(sk, mockSecretKey)
	os.Setenv(st, mockSessionToken)
	defer os.Unsetenv(ak)
	defer os.Unsetenv(sk)
	defer os.Unsetenv(st)

	result := getAliyunSecretKeyFromEnv(region, ak, sk, st)
	assert.Equal(t, endpoints.UsEast1RegionID, result.region)
}
