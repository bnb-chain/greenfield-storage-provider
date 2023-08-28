package piecestore_e2e

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/piece"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

var shardNum = 5

func TestSharding(t *testing.T) {
	// init PieceStore
	handler, err := setup(t, storage.DiskFileStore, "./data/test%d", storage.AKSKIAMType, shardNum)
	assert.Equal(t, err, nil)

	err = createFiles(shardNum)
	assert.Equal(t, nil, err)
	doShardingOperations(t, handler, shardNum)

	err = removeFiles(shardNum)
	assert.Equal(t, nil, err)
}

func doShardingOperations(t *testing.T, handler *piece.PieceStore, shards int) {
	for i := 0; i < shards; i++ {
		key := fmt.Sprintf("shards%d.txt", i)
		// 1. put piece
		log.Info("Put piece")
		f, err := os.Open(fmt.Sprintf("./testdata/shards%d.txt", i))
		assert.Equal(t, nil, err)
		err = handler.Put(context.Background(), key, f)
		assert.Equal(t, nil, err)

		// 2. get piece
		log.Info("Get piece")
		rc, err := handler.Get(context.Background(), key, 0, -1)
		assert.Equal(t, nil, err)
		data, err := io.ReadAll(rc)
		assert.Equal(t, nil, err)
		assert.Contains(t, string(data), "test sharding func:")

		// 3. head object
		log.Info("Get piece info")
		object, err := handler.Head(context.Background(), key)
		assert.Equal(t, nil, err)
		assert.Equal(t, key, object.Key())
		assert.Equal(t, false, object.IsSymlink())
		log.Infow("piece info", "key", object.Key(), "size", object.Size(), "mod time", object.ModTime(),
			"is symlink", object.IsSymlink())

		// 3. delete piece
		log.Info("Delete piece")
		err = handler.Delete(context.Background(), key)
		assert.Equal(t, nil, err)
	}
}

func createFiles(fileNum int) error {
	for i := 0; i < fileNum; i++ {
		f, err := os.Create(fmt.Sprintf("./testdata/shards%d.txt", i))
		if err != nil {
			log.Errorw("failed to create files", "error", err)
			return err
		}
		defer f.Close()
		writer := bufio.NewWriter(f)
		writer.WriteString(fmt.Sprintf("test sharding func: %d\n", i))
		writer.Flush()
	}
	return nil
}

func removeFiles(fileNum int) error {
	for i := 0; i < fileNum; i++ {
		if err := os.Remove(fmt.Sprintf("./testdata/shards%d.txt", i)); err != nil {
			return err
		}
	}
	return nil
}
