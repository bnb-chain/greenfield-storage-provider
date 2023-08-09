package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSharded(t *testing.T) {
	cases := []struct {
		name         string
		cfg          PieceStoreConfig
		wantedResult ObjectStorage
		wantedIsErr  bool
		wantedErr    error
	}{
		{
			name: "correct sharding cfg",
			cfg: PieceStoreConfig{
				Shards: 2,
				Store: ObjectStorageConfig{
					Storage:   MemoryStore,
					BucketURL: "test%d",
					IAMType:   AKSKIAMType,
				},
			},
			wantedIsErr: false,
			wantedErr:   nil,
		},
		{
			name: "invalid sharding bucket url",
			cfg: PieceStoreConfig{
				Shards: 2,
				Store: ObjectStorageConfig{
					Storage:   MemoryStore,
					BucketURL: "test",
					IAMType:   AKSKIAMType,
				},
			},
			wantedIsErr: true,
			wantedErr:   errors.New("can not generate different endpoint using [test]"),
		},
		{
			name: "invalid sharding storage type",
			cfg: PieceStoreConfig{
				Shards: 2,
				Store: ObjectStorageConfig{
					Storage:   "unknown",
					BucketURL: "test%d",
					IAMType:   AKSKIAMType,
				},
			},
			wantedIsErr: true,
			wantedErr:   errors.New("invalid object storage: unknown"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewSharded(tt.cfg)
			assert.Equal(t, tt.wantedErr, err)
			if tt.wantedIsErr {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestSharded_String(t *testing.T) {
	cfg := PieceStoreConfig{
		Shards: 2,
		Store: ObjectStorageConfig{
			Storage:   MemoryStore,
			BucketURL: "test%d",
			IAMType:   AKSKIAMType,
		},
	}
	s, err := NewSharded(cfg)
	assert.Nil(t, err)
	result := s.String()
	assert.Equal(t, "shard2://memory://test0/", result)
}

func TestSharded_CreateBucket(t *testing.T) {
	cases := []struct {
		name         string
		cfg          PieceStoreConfig
		wantedIsErr  bool
		wantedErrStr string
	}{
		{
			name: "1",
			cfg: PieceStoreConfig{
				Shards: 2,
				Store: ObjectStorageConfig{
					Storage:   MemoryStore,
					BucketURL: "test%d",
					IAMType:   AKSKIAMType,
				},
			},
			wantedIsErr:  false,
			wantedErrStr: emptyString,
		},
		{
			name: "2",
			cfg: PieceStoreConfig{
				Shards: 2,
				Store: ObjectStorageConfig{
					Storage:   DiskFileStore,
					BucketURL: "/root/unwritable_dir%d",
					IAMType:   AKSKIAMType,
				},
			},
			wantedIsErr:  true,
			wantedErrStr: "failed to create directory /root/unwritable_dir0 :",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewSharded(tt.cfg)
			assert.Nil(t, err)
			err = s.CreateBucket(context.TODO())
			if tt.wantedIsErr {
				assert.Contains(t, err.Error(), tt.wantedErrStr)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestSharded_GetObject(t *testing.T) {
	cases := []struct {
		name      string
		cfg       PieceStoreConfig
		key       string
		off       int64
		limit     int64
		wantedErr error
	}{
		{
			name: "1",
			cfg: PieceStoreConfig{
				Shards: 2,
				Store: ObjectStorageConfig{
					Storage:   MemoryStore,
					BucketURL: "test%d",
					IAMType:   AKSKIAMType,
				},
			},
			key:       mockKey,
			off:       0,
			limit:     -1,
			wantedErr: ErrNoSuchObject,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewSharded(tt.cfg)
			assert.Nil(t, err)
			rc, err := s.GetObject(context.TODO(), tt.key, tt.off, tt.limit)
			assert.Nil(t, rc)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestSharded_PutObject(t *testing.T) {
	cases := []struct {
		name      string
		cfg       PieceStoreConfig
		key       string
		body      io.Reader
		wantedErr error
	}{
		{
			name: "1",
			cfg: PieceStoreConfig{
				Shards: 2,
				Store: ObjectStorageConfig{
					Storage:   MemoryStore,
					BucketURL: "test%d",
					IAMType:   AKSKIAMType,
				},
			},
			key:       mockKey,
			body:      strings.NewReader("test"),
			wantedErr: nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewSharded(tt.cfg)
			assert.Nil(t, err)
			err = s.PutObject(context.TODO(), tt.key, tt.body)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestSharded_DeleteObject(t *testing.T) {
	cases := []struct {
		name      string
		cfg       PieceStoreConfig
		key       string
		wantedErr error
	}{
		{
			name: "1",
			cfg: PieceStoreConfig{
				Shards: 2,
				Store: ObjectStorageConfig{
					Storage:   MemoryStore,
					BucketURL: "test%d",
					IAMType:   AKSKIAMType,
				},
			},
			key:       mockKey,
			wantedErr: nil,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewSharded(tt.cfg)
			assert.Nil(t, err)
			err = s.DeleteObject(context.TODO(), tt.key)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestSharded_HeadBucket(t *testing.T) {
	cases := []struct {
		name      string
		cfg       PieceStoreConfig
		wantedErr error
	}{
		{
			name: "1",
			cfg: PieceStoreConfig{
				Shards: 2,
				Store: ObjectStorageConfig{
					Storage:   MemoryStore,
					BucketURL: "test%d",
					IAMType:   AKSKIAMType,
				},
			},
			wantedErr: nil,
		},
		{
			name: "2",
			cfg: PieceStoreConfig{
				Shards: 2,
				Store: ObjectStorageConfig{
					Storage:   DiskFileStore,
					BucketURL: "test%d",
					IAMType:   AKSKIAMType,
				},
			},
			wantedErr: ErrNoSuchBucket,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewSharded(tt.cfg)
			assert.Nil(t, err)
			err = s.HeadBucket(context.TODO())
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}

func TestSharded_HeadObject(t *testing.T) {
	cases := []struct {
		name      string
		cfg       PieceStoreConfig
		key       string
		wantedErr error
	}{
		{
			name: "1",
			cfg: PieceStoreConfig{
				Shards: 2,
				Store: ObjectStorageConfig{
					Storage:   MemoryStore,
					BucketURL: "test%d",
					IAMType:   AKSKIAMType,
				},
			},
			key:       mockKey,
			wantedErr: os.ErrNotExist,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewSharded(tt.cfg)
			assert.Nil(t, err)
			result, err := s.HeadObject(context.TODO(), tt.key)
			assert.Nil(t, result)
			assert.Equal(t, tt.wantedErr, err)
		})
	}
}
