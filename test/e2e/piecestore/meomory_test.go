package piecestore_e2e

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

func TestMemoryStore(t *testing.T) {
	// 1. init PieceStore
	handler, err := setUp(t, memory, "")
	assert.Equal(t, err, nil)

	// 2. put piece
	log.Info("Put piece")
	f, err := os.Open("./testdata/hello.txt")
	assert.Equal(t, nil, err)
	err = handler.Put(context.Background(), "hello.txt", f)
	assert.Equal(t, nil, err)

	// 3. get piece
	log.Info("Get piece")
	rc, err := handler.Get(context.Background(), "hello.txt", 0, -1)
	assert.Equal(t, nil, err)
	data, err := io.ReadAll(rc)
	assert.Equal(t, nil, err)
	assert.Equal(t, string(data), "Hello, World!\n")

	// 4. delete piece
	log.Info("Delete piece")
	err = handler.Delete(context.Background(), "hello.txt")
	assert.Equal(t, nil, err)
}
