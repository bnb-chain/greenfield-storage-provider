package e2e

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

func TestDiskFileStore(t *testing.T) {
	// 1. init PieceStore
	handler, err := setUp(t, file, "")
	assert.Equal(t, err, nil)

	// 2. put piece
	log.Info("Put piece")
	f, err := os.Open("./testdata/hello.txt")
	assert.Equal(t, err, nil)
	err = handler.Put(context.Background(), "hello.txt", f)
	assert.Equal(t, err, nil)

	// 3. get piece
	log.Info("Get piece")
	rc, err := handler.Get(context.Background(), "hello.txt", 0, -1)
	assert.Equal(t, err, nil)
	data, err := io.ReadAll(rc)
	assert.Equal(t, err, nil)
	assert.Equal(t, string(data), "Hello, World!\n")

	// 4. delete piece
	log.Info("Delete piece")
	err = handler.Delete(context.Background(), "hello.txt")
	assert.Equal(t, err, nil)
}
