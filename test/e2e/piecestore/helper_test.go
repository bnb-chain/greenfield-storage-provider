package piecestore_e2e

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	mpiecestore "github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/piece"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

const (
	pieceKey = "hello.txt"
	s3Bucket = "https://s3.us-east-1.amazonaws.com/test"
	// virtualPath = "https://test.s3.us-east-1.amazonaws.com"
)

func setup(t *testing.T, storageType, bucketURL string, shards int) (*piece.PieceStore, error) {
	t.Helper()
	return piece.NewPieceStore(&storage.PieceStoreConfig{
		Shards: shards,
		Store: storage.ObjectStorageConfig{
			Storage:    storageType,
			BucketURL:  bucketURL,
			MaxRetries: 5,
			IAMType:    mpiecestore.AKSKIAMType,
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
	log.Info("Get piece info")
	object, err := handler.GetPieceInfo(context.Background(), pieceKey)
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
