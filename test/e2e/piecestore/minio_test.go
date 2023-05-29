package piecestore_e2e

import (
	"testing"

	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
	"github.com/stretchr/testify/assert"
)

func TestMinioStore(t *testing.T) {
	// 1. init PieceStore
	handler, err := setup(t, storage.MinioStore, s3BucketURL, 0)
	assert.Equal(t, err, nil)

	doOperations(t, handler)
}
