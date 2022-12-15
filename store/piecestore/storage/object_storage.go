package storage

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/bnb-chain/inscription-storage-provider/config"
	"github.com/bnb-chain/inscription-storage-provider/model"
	"github.com/bnb-chain/inscription-storage-provider/model/piecestore"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

func NewObjectStorage(cfg *config.ObjectStorage) (ObjectStorage, error) {
	if fn, ok := storageMap[strings.ToLower(cfg.Storage)]; ok {
		log.Infof("Creating %s storage at endpoint %s", cfg.Storage, cfg.BucketURL)
		return fn(cfg)
	}
	return nil, fmt.Errorf("Invalid object storage: %s", cfg.Storage)
}

type StorageFn func(cfg *config.ObjectStorage) (ObjectStorage, error)

var storageMap = map[string]StorageFn{
	"s3":     newS3Store,
	"file":   newDiskFileStore,
	"memory": newMemoryStore,
}

type DefaultObjectStorage struct{}

func (s DefaultObjectStorage) CreateBucket(ctx context.Context) error {
	return nil
}

func (s DefaultObjectStorage) ListObjects(ctx context.Context, prefix, marker, delimiter string, limit int64) ([]Object, error) {
	return nil, model.NotSupportedMethod
}

func (s DefaultObjectStorage) ListAllObjects(ctx context.Context, prefix, marker string) (<-chan Object, error) {
	return nil, model.NotSupportedMethod
}

type file struct {
	object
	group     string
	owner     string
	mode      os.FileMode
	isSymlink bool
}

func (f *file) Owner() string     { return f.owner }
func (f *file) Group() string     { return f.group }
func (f *file) Mode() os.FileMode { return f.mode }
func (f *file) IsSymlink() bool   { return f.isSymlink }

var bufPool = sync.Pool{
	New: func() any {
		buf := make([]byte, piecestore.BufPoolSize)
		return &buf
	},
}
