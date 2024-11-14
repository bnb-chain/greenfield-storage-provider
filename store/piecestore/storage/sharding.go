package storage

import (
	"context"
	"fmt"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"hash/fnv"
	"io"
	"math"
	"strings"
)

type sharded struct {
	stores []ObjectStorage
	DefaultObjectStorage
}

func NewSharded(cfg PieceStoreConfig) (ObjectStorage, error) {
	stores := make([]ObjectStorage, cfg.Shards)
	var err error
	shardingURL := cfg.Store.BucketURL
	for i := range stores {
		ep := fmt.Sprintf(shardingURL, i)
		if strings.HasSuffix(ep, "%!(EXTRA int=0)") {
			return nil, fmt.Errorf("can not generate different endpoint using [%s]", shardingURL)
		}
		cfg.Store.BucketURL = ep
		stores[i], err = NewObjectStorage(cfg.Store)
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

func (s *sharded) DeleteObjectsByPrefix(ctx context.Context, key string) (uint64, error) {
	objs, err := s.pick(key).ListObjects(ctx, key, "", "", math.MaxUint64)
	if err != nil {
		log.Errorw("DeleteObjectsByPrefix list objects error", "error", err)
		return 0, err
	}

	var size uint64
	for _, obj := range objs {
		err = s.pick(obj.Key()).DeleteObject(ctx, obj.Key())
		if err != nil {
			log.Errorw("remove single file by prefix error", "error", err)
		} else {
			size += uint64(obj.Size())
		}
	}
	return size, nil
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
