package e2e

import (
	"testing"

	"github.com/bnb-chain/inscription-storage-provider/store/piecestore/piece"
	"github.com/bnb-chain/inscription-storage-provider/store/piecestore/storage"
)

const (
	s3       = "s3"
	file     = "file"
	memory   = "memory"
	s3Bucket = "https://s3.us-east-1.amazonaws.com/tf-nodereal-prod-nodereal-storage-s3"
)

func setUp(t *testing.T, storageType, bucketURL string) (*piece.PieceStore, error) {
	t.Helper()
	return piece.NewPieceStore(&storage.PieceStoreConfig{
		Shards: 0,
		Store: &storage.ObjectStorageConfig{
			Storage:               storageType,
			BucketURL:             bucketURL,
			AccessKey:             "",
			SecretKey:             "",
			SessionToken:          "",
			NoSignRequest:         false,
			MaxRetries:            5,
			MinRetryDelay:         0,
			TlsInsecureSkipVerify: false,
			TestMode:              true,
		},
	})
}
