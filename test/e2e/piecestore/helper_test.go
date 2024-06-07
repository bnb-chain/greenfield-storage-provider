package piecestore_e2e

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zkMeLabs/mechain-storage-provider/pkg/log"
	"github.com/zkMeLabs/mechain-storage-provider/store/piecestore/piece"
	"github.com/zkMeLabs/mechain-storage-provider/store/piecestore/storage"
)

const (
	pieceKey     = "hello.txt"
	s3BucketURL  = "https://s3.us-east-1.amazonaws.com/test"
	ossBucketURL = "https://example.oss-ap-northeast-1.aliyuncs.com/"
	b2BucketURL  = "https://s3.us-east-005.backblazeb2.com/greenfieldsp"
	// virtualPath = "https://test.s3.us-east-1.amazonaws.com"
)

func setup(t *testing.T, storageType, bucketURL, iamType string, shards int) (*piece.PieceStore, error) {
	t.Helper()
	return piece.NewPieceStore(&storage.PieceStoreConfig{
		Shards: shards,
		Store: storage.ObjectStorageConfig{
			Storage:    storageType,
			BucketURL:  bucketURL,
			MaxRetries: 5,
			IAMType:    iamType,
		},
	})
}

func doOperations(t *testing.T, handler *piece.PieceStore) {
	// 1. put piece
	log.Info("Put piece")
	f, err := os.Open("./testdata/hello.txt")
	assert.Equal(t, nil, err)
	err = handler.Put(context.Background(), pieceKey, f)
	assert.Equal(t, nil, err)

	// 2. get piece
	log.Info("Get piece")
	rc, err := handler.Get(context.Background(), pieceKey, 0, -1)
	assert.Equal(t, nil, err)
	data, err := io.ReadAll(rc)
	assert.Equal(t, nil, err)
	assert.Contains(t, string(data), "Hello, World!\n")

	// 3. head object
	log.Info("Head piece info")
	object, err := handler.Head(context.Background(), pieceKey)
	assert.Equal(t, nil, err)
	assert.Equal(t, pieceKey, object.Key())
	assert.Equal(t, false, object.IsSymlink())
	log.Infow("piece info", "key", object.Key(), "size", object.Size(), "mod time", object.ModTime(),
		"is symlink", object.IsSymlink())

	// 3. delete piece
	log.Info("Delete piece")
	err = handler.Delete(context.Background(), pieceKey)
	assert.Equal(t, nil, err)
}
