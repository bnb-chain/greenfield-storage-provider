package piecestore_e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

func TestMemoryStore(t *testing.T) {
	// 1. init PieceStore
	handler, err := setup(t, storage.MemoryStore, "", 0)
	assert.Equal(t, err, nil)

	doOperations(t, handler)
}
