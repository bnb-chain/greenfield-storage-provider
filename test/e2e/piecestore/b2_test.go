package piecestore_e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

const (
	b2BucketURL = "https://s3.us-east-005.backblazeb2.com/greenfieldsp"
)

func TestB2Store(t *testing.T) {
	// 1. init PieceStore
	handler, err := setup(t, storage.B2Store, b2BucketURL, 0)
	assert.Equal(t, err, nil)

	doOperations(t, handler)
}
