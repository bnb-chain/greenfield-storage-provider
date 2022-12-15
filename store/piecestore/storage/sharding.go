package storage

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"strings"

	"github.com/bnb-chain/inscription-storage-provider/store/piecestore/model"
)

type sharded struct {
	stores []ObjectStorage
	DefaultObjectStorage
}

func NewSharded(cfg *model.PieceStoreConfig) (ObjectStorage, error) {
	stores := make([]ObjectStorage, cfg.Shards)
	var err error
	for i := range stores {
		ep := fmt.Sprintf(cfg.Store.BucketURL, i)
		if strings.HasSuffix(ep, "%!(EXTRA int=0)") {
			return nil, fmt.Errorf("can not generate different endpoint using %s", cfg.Store.BucketURL)
		}
		stores[i], err = NewObjectStorage(&cfg.Store)
		if err != nil {
			return nil, err
		}
	}
	return &sharded{stores: stores}, nil
}

func (s *sharded) String() string {
	return fmt.Sprintf("shard%d://%s", len(s.stores), s.stores[0])
}

func (s *sharded) CreateBucket(ctx context.Context) error {
	for _, o := range s.stores {
		if err := o.CreateBucket(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *sharded) pick(key string) ObjectStorage {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	i := h.Sum32() % uint32(len(s.stores))
	return s.stores[i]
}

func (s *sharded) GetObject(ctx context.Context, key string, off, limit int64) (io.ReadCloser, error) {
	return s.pick(key).GetObject(ctx, key, off, limit)
}

func (s *sharded) PutObject(ctx context.Context, key string, body io.Reader) error {
	return s.pick(key).PutObject(ctx, key, body)
}

func (s *sharded) DeleteObject(ctx context.Context, key string) error {
	return s.pick(key).DeleteObject(ctx, key)
}

func (s *sharded) HeadBucket(ctx context.Context) error {
	for _, o := range s.stores {
		if err := o.HeadBucket(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *sharded) HeadObject(ctx context.Context, key string) (Object, error) {
	return s.pick(key).HeadObject(ctx, key)
}
