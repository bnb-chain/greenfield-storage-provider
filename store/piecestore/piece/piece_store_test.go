package piece

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

const (
	mockBucketURL = "https://s3.us-east-1.amazonaws.com/test"
)

func TestNewPieceStore(t *testing.T) {
	cases := []struct {
		name        string
		pieceConfig *storage.PieceStoreConfig
		wantedIsErr bool
	}{
		{
			name: "Correct new piece store",
			pieceConfig: &storage.PieceStoreConfig{
				Shards: 0,
				Store: storage.ObjectStorageConfig{
					Storage:   storage.MemoryStore,
					BucketURL: "mock",
					IAMType:   storage.AKSKIAMType,
				},
			},
			wantedIsErr: false,
		},
		{
			name: "Failed to check config",
			pieceConfig: &storage.PieceStoreConfig{
				Shards: 257,
				Store: storage.ObjectStorageConfig{
					Storage:   storage.MemoryStore,
					BucketURL: "mock",
					IAMType:   storage.AKSKIAMType,
				},
			},
			wantedIsErr: true,
		},
		{
			name: "Failed to create storage",
			pieceConfig: &storage.PieceStoreConfig{
				Shards: 0,
				Store: storage.ObjectStorageConfig{
					Storage:   storage.S3Store,
					BucketURL: "mock",
					IAMType:   storage.AKSKIAMType,
				},
			},
			wantedIsErr: true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ps, err := NewPieceStore(tt.pieceConfig)
			if tt.wantedIsErr {
				assert.NotNil(t, err)
				assert.Nil(t, ps)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, ps)
			}
		})
	}
}

func Test_checkConfig(t *testing.T) {
	cases := []struct {
		name      string
		cfg       *storage.PieceStoreConfig
		wantedErr error
	}{
		{
			name: "Correct s3 piece store config",
			cfg: &storage.PieceStoreConfig{
				Shards: 0,
				Store: storage.ObjectStorageConfig{
					Storage:               storage.S3Store,
					BucketURL:             mockBucketURL,
					MaxRetries:            1,
					MinRetryDelay:         2,
					TLSInsecureSkipVerify: false,
					IAMType:               storage.AKSKIAMType,
				},
			},
			wantedErr: nil,
		},
		{
			name: "Correct file piece store config",
			cfg: &storage.PieceStoreConfig{
				Shards: 0,
				Store: storage.ObjectStorageConfig{
					Storage: storage.DiskFileStore,
					IAMType: storage.AKSKIAMType,
				},
			},
			wantedErr: nil,
		},
		{
			name: "Wrong shards config",
			cfg: &storage.PieceStoreConfig{
				Shards: 257,
			},
			wantedErr: errors.New("too many shards: 257"),
		},
		{
			name: "Wrong iam type config",
			cfg: &storage.PieceStoreConfig{
				Shards: 0,
				Store: storage.ObjectStorageConfig{
					IAMType: "unknown iam type",
				},
			},
			wantedErr: errors.New("invalid iam type: unknown iam type"),
		},
		{
			name: "Wrong max retries config",
			cfg: &storage.PieceStoreConfig{
				Shards: 0,
				Store: storage.ObjectStorageConfig{
					MaxRetries: -1,
					IAMType:    storage.AKSKIAMType,
				},
			},
			wantedErr: errors.New("MaxRetries should be equal or greater than zero"),
		},
		{
			name: "Wrong min retry delay config",
			cfg: &storage.PieceStoreConfig{
				Shards: 0,
				Store: storage.ObjectStorageConfig{
					MinRetryDelay: -1,
					IAMType:       storage.AKSKIAMType,
				},
			},
			wantedErr: errors.New("MinRetryDelay should be equal or greater than zero"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := checkConfig(tt.cfg)
			assert.Equal(t, err, tt.wantedErr)
		})
	}
}

func Test_createStorage(t *testing.T) {
	cases := []struct {
		name        string
		cfg         storage.PieceStoreConfig
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name: "no shards",
			cfg: storage.PieceStoreConfig{
				Shards: 0,
				Store: storage.ObjectStorageConfig{
					Storage:   storage.MemoryStore,
					BucketURL: "mock",
					IAMType:   storage.AKSKIAMType,
				},
			},
			wantedIsErr: false,
			wantedErr:   nil,
		},
		{
			name: "5 shards",
			cfg: storage.PieceStoreConfig{
				Shards: 5,
				Store: storage.ObjectStorageConfig{
					Storage:   storage.MemoryStore,
					BucketURL: "mock%d",
					IAMType:   storage.AKSKIAMType,
				},
			},
			wantedIsErr: false,
			wantedErr:   nil,
		},
		{
			name: "5 shards with wrong bucket url",
			cfg: storage.PieceStoreConfig{
				Shards: 5,
				Store: storage.ObjectStorageConfig{
					Storage:   storage.MemoryStore,
					BucketURL: "mock",
					IAMType:   storage.AKSKIAMType,
				},
			},
			wantedIsErr: true,
			wantedErr:   errors.New("can not generate different endpoint using [mock]"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ps, err := createStorage(tt.cfg)
			assert.Equal(t, tt.wantedErr, err)
			if tt.wantedIsErr {
				assert.Nil(t, ps)
			} else {
				assert.NotNil(t, ps)
			}
		})
	}
}

func Test_overrideConfigFromEnv(t *testing.T) {
	os.Setenv(storage.BucketURL, mockBucketURL)
	defer os.Unsetenv(storage.BucketURL)
	cfg := &storage.PieceStoreConfig{
		Shards: 0,
		Store:  storage.ObjectStorageConfig{},
	}
	overrideConfigFromEnv(cfg)
	assert.Equal(t, mockBucketURL, cfg.Store.BucketURL)
}

func Test_checkBucketSuccessfully(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := storage.NewMockObjectStorage(ctrl)
	m.EXPECT().String().Return("test").AnyTimes()
	m.EXPECT().HeadBucket(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
		return nil
	}).AnyTimes()
	err := checkBucket(context.Background(), m)
	assert.Nil(t, err)
}

func Test_checkBucketNoSuchBucket(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := storage.NewMockObjectStorage(ctrl)
	m.EXPECT().String().Return("test").AnyTimes()
	m.EXPECT().HeadBucket(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
		return storage.ErrNoSuchBucket
	}).AnyTimes()
	m.EXPECT().CreateBucket(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
		return nil
	})
	err := checkBucket(context.Background(), m)
	assert.Equal(t, nil, err)
}

func Test_checkBucketNoPermissionAccessBucket(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := storage.NewMockObjectStorage(ctrl)
	m.EXPECT().String().Return("test").AnyTimes()
	m.EXPECT().HeadBucket(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
		return errors.New("no access")
	}).AnyTimes()
	err := checkBucket(context.Background(), m)
	assert.Equal(t, storage.ErrNoPermissionAccessBucket, err)
}

func Test_checkBucketFailedToCreateBucket(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := storage.NewMockObjectStorage(ctrl)
	m.EXPECT().String().Return("test").AnyTimes()
	m.EXPECT().HeadBucket(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
		return storage.ErrNoSuchBucket
	}).AnyTimes()
	m.EXPECT().CreateBucket(gomock.Any()).DoAndReturn(func(ctx context.Context) error {
		return errors.New("failed to create bucket")
	})
	err := checkBucket(context.Background(), m)
	assert.NotNil(t, err)
}
