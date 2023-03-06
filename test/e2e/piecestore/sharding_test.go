package piecestore_e2e

import (
	"testing"

	mpiecestore "github.com/bnb-chain/greenfield-storage-provider/model/piecestore"
	"github.com/stretchr/testify/assert"
)

func TestSharding(t *testing.T) {
	// 1. init PieceStore
	handler, err := setup(t, mpiecestore.DiskFileStore, "", 3)
	assert.Equal(t, err, nil)

	doOperations(t, handler)
}
