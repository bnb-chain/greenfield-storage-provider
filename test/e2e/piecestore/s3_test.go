package piecestore_e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

func TestS3Store(t *testing.T) {
	// 1. init PieceStore
	handler, err := setup(t, storage.S3Store, s3BucketURL, 0)
	assert.Equal(t, err, nil)

	doOperations(t, handler)
}
