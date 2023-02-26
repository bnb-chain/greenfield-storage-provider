package piecestore_e2e

import (
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/piece"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

const (
	s3       = "s3"
	minio    = "minio"
	file     = "file"
	memory   = "memory"
	s3Bucket = "https://s3.us-east-1.amazonaws.com/test"
	// virtualPath = "https://test.s3.us-east-1.amazonaws.com"
)

func setUp(t *testing.T, storageType, bucketURL string) (*piece.PieceStore, error) {
	t.Helper()
	return piece.NewPieceStore(&storage.PieceStoreConfig{
		Shards: 5,
		Store: storage.ObjectStorageConfig{
			Storage:               storageType,
			BucketURL:             bucketURL,
			MaxRetries:            5,
			MinRetryDelay:         0,
			TLSInsecureSkipVerify: false,
			TestMode:              true,
		},
	})
}
