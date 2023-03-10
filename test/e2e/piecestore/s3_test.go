package piecestore_e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mpiecestore "github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
)

func TestS3Store(t *testing.T) {
	// 1. init PieceStore
	handler, err := setup(t, mpiecestore.S3Store, s3Bucket, 0)
	assert.Equal(t, err, nil)

	doOperations(t, handler)
}
