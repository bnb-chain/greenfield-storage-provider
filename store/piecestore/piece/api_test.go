package piece

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

const key = "mockKey"

func setup() (*PieceStore, error) {
	cfg := &storage.PieceStoreConfig{
		Shards: 0,
		Store: storage.ObjectStorageConfig{
			Storage:   storage.MemoryStore,
			BucketURL: "mock",
			IAMType:   storage.AKSKIAMType,
		},
	}
	return NewPieceStore(cfg)
}

func TestGet(t *testing.T) {
	ps, err := setup()
	assert.Nil(t, err)
	ctrl := gomock.NewController(t)
	m := storage.NewMockObjectStorage(ctrl)
	m.EXPECT().String().Return("test").AnyTimes()
	m.EXPECT().GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string, offset, limit int64) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("memory get")), nil
		}).AnyTimes()
	ps.storeAPI = m

	rc, err := ps.Get(context.Background(), key, 0, 0)
	assert.Nil(t, err)
	data, err := io.ReadAll(rc)
	assert.Nil(t, err)
	assert.Equal(t, "memory get", string(data))
}

func TestPut(t *testing.T) {
	ps, err := setup()
	assert.Nil(t, err)
	ctrl := gomock.NewController(t)
	m := storage.NewMockObjectStorage(ctrl)
	m.EXPECT().String().Return("test").AnyTimes()
	m.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string, reader io.Reader) error { return nil }).AnyTimes()
	ps.storeAPI = m

	err = ps.Put(context.Background(), key, nil)
	assert.Nil(t, err)
}

func TestDelete(t *testing.T) {
	ps, err := setup()
	assert.Nil(t, err)
	ctrl := gomock.NewController(t)
	m := storage.NewMockObjectStorage(ctrl)
	m.EXPECT().String().Return("test").AnyTimes()
	m.EXPECT().DeleteObject(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string) error { return nil }).AnyTimes()
	ps.storeAPI = m

	err = ps.Delete(context.Background(), key)
	assert.Nil(t, err)
}

func TestHead(t *testing.T) {
	ps, err := setup()
	assert.Nil(t, err)
	ctrl := gomock.NewController(t)
	object := storage.NewMockObject(ctrl)
	object.EXPECT().Key().Return("golang").AnyTimes()
	m := storage.NewMockObjectStorage(ctrl)
	m.EXPECT().String().Return("test").AnyTimes()
	m.EXPECT().HeadObject(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string) (storage.Object, error) {
			return object, nil
		}).AnyTimes()
	ps.storeAPI = m

	obj, err := ps.Head(context.Background(), key)
	assert.Nil(t, err)
	assert.Equal(t, "golang", obj.Key())
}
