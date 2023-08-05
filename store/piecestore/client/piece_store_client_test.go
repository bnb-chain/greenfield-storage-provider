package client

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/piece"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

func TestNewStoreClient(t *testing.T) {
	cases := []struct {
		name        string
		cfg         *storage.PieceStoreConfig
		wantedIsErr bool
		wantedErr   error
	}{
		{
			name: "1",
			cfg: &storage.PieceStoreConfig{
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
			name: "2",
			cfg: &storage.PieceStoreConfig{
				Shards: 0,
				Store: storage.ObjectStorageConfig{
					Storage:   storage.MemoryStore,
					BucketURL: "mock",
					IAMType:   "unknown",
				},
			},
			wantedIsErr: true,
			wantedErr:   errors.New("invalid iam type: unknown"),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewStoreClient(tt.cfg)
			assert.Equal(t, tt.wantedErr, err)
			if tt.wantedIsErr {
				assert.Nil(t, client)
			} else {
				assert.Equal(t, tt.cfg.Store.Storage, client.name)
			}
		})
	}
}

func TestGetPieceSuccessfully(t *testing.T) {
	cfg := &storage.PieceStoreConfig{
		Shards: 0,
		Store: storage.ObjectStorageConfig{
			Storage:   storage.MemoryStore,
			BucketURL: "mock",
			IAMType:   storage.AKSKIAMType,
		},
	}
	client, err := NewStoreClient(cfg)
	assert.Nil(t, err)
	ctrl := gomock.NewController(t)
	p := piece.NewMockPieceAPI(ctrl)
	p.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string, offset, limit int64) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("golang")), nil
		}).AnyTimes()
	client.ps = p
	data, err := client.GetPiece(context.Background(), "mock", 0, 0)
	assert.Nil(t, err)
	assert.Equal(t, []byte("golang"), data)
}

func TestGetPieceFailedToGet(t *testing.T) {
	cfg := &storage.PieceStoreConfig{
		Shards: 0,
		Store: storage.ObjectStorageConfig{
			Storage:   storage.MemoryStore,
			BucketURL: "mock",
			IAMType:   storage.AKSKIAMType,
		},
	}
	client, err := NewStoreClient(cfg)
	assert.Nil(t, err)
	ctrl := gomock.NewController(t)
	p := piece.NewMockPieceAPI(ctrl)
	p.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string, offset, limit int64) (closer io.ReadCloser, err error) {
			return nil, errors.New("invalid key")
		}).AnyTimes()
	client.ps = p
	data, err := client.GetPiece(context.Background(), "mock", 0, 0)
	assert.Nil(t, data)
	assert.Equal(t, errors.New("invalid key"), err)
}

func TestPutPieceSuccessfully(t *testing.T) {
	cfg := &storage.PieceStoreConfig{
		Shards: 0,
		Store: storage.ObjectStorageConfig{
			Storage:   storage.MemoryStore,
			BucketURL: "mock",
			IAMType:   storage.AKSKIAMType,
		},
	}
	client, err := NewStoreClient(cfg)
	assert.Nil(t, err)
	ctrl := gomock.NewController(t)
	p := piece.NewMockPieceAPI(ctrl)
	p.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string, reader io.Reader) error { return nil }).AnyTimes()
	client.ps = p
	err = client.PutPiece(context.Background(), "mock", []byte("1"))
	assert.Nil(t, err)
}

func TestPutPieceFailure(t *testing.T) {
	cfg := &storage.PieceStoreConfig{
		Shards: 0,
		Store: storage.ObjectStorageConfig{
			Storage:   storage.MemoryStore,
			BucketURL: "mock",
			IAMType:   storage.AKSKIAMType,
		},
	}
	client, err := NewStoreClient(cfg)
	assert.Nil(t, err)
	ctrl := gomock.NewController(t)
	p := piece.NewMockPieceAPI(ctrl)
	p.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string, reader io.Reader) error {
			return errors.New("failed to put piece")
		}).AnyTimes()
	client.ps = p
	err = client.PutPiece(context.Background(), "mock", []byte("1"))
	assert.Equal(t, errors.New("failed to put piece"), err)
}

func TestDeletePieceSuccessfully(t *testing.T) {
	cfg := &storage.PieceStoreConfig{
		Shards: 0,
		Store: storage.ObjectStorageConfig{
			Storage:   storage.MemoryStore,
			BucketURL: "mock",
			IAMType:   storage.AKSKIAMType,
		},
	}
	client, err := NewStoreClient(cfg)
	assert.Nil(t, err)
	ctrl := gomock.NewController(t)
	p := piece.NewMockPieceAPI(ctrl)
	p.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string, offset, limit int64) (closer io.ReadCloser, err error) {
			return io.NopCloser(strings.NewReader("get")), nil
		}).AnyTimes()
	p.EXPECT().Delete(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string) error { return nil }).AnyTimes()
	client.ps = p
	err = client.DeletePiece(context.Background(), "mock")
	assert.Nil(t, err)
}

func TestDeletePieceFailure(t *testing.T) {
	cfg := &storage.PieceStoreConfig{
		Shards: 0,
		Store: storage.ObjectStorageConfig{
			Storage:   storage.MemoryStore,
			BucketURL: "mock",
			IAMType:   storage.AKSKIAMType,
		},
	}
	client, err := NewStoreClient(cfg)
	assert.Nil(t, err)
	ctrl := gomock.NewController(t)
	p := piece.NewMockPieceAPI(ctrl)
	p.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string, offset, limit int64) (closer io.ReadCloser, err error) {
			return io.NopCloser(strings.NewReader("get")), nil
		}).AnyTimes()
	p.EXPECT().Delete(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string) error { return errors.New("failed to delete") }).AnyTimes()
	client.ps = p
	err = client.DeletePiece(context.Background(), "mock")
	assert.Equal(t, errors.New("failed to delete"), err)
}

func TestDeletePieceGetFailure(t *testing.T) {
	cfg := &storage.PieceStoreConfig{
		Shards: 0,
		Store: storage.ObjectStorageConfig{
			Storage:   storage.MemoryStore,
			BucketURL: "mock",
			IAMType:   storage.AKSKIAMType,
		},
	}
	client, err := NewStoreClient(cfg)
	assert.Nil(t, err)
	ctrl := gomock.NewController(t)
	p := piece.NewMockPieceAPI(ctrl)
	p.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string, offset, limit int64) (closer io.ReadCloser, err error) {
			return nil, errors.New("failed to get")
		}).AnyTimes()
	client.ps = p
	err = client.DeletePiece(context.Background(), "mock")
	assert.Equal(t, errors.New("failed to get"), err)
}
