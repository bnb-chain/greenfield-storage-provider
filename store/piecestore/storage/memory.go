package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

type memoryStore struct {
	name    string
	objects map[string]*memoryObject
	sync.Mutex
	DefaultObjectStorage
}

type memoryObject struct {
	data    []byte
	modTime time.Time
}

func newMemoryStore(cfg ObjectStorageConfig) (ObjectStorage, error) {
	store := &memoryStore{name: cfg.BucketURL}
	store.objects = make(map[string]*memoryObject)
	return store, nil
}

func (m *memoryStore) String() string {
	return fmt.Sprintf("memory://%s/", m.name)
}

func (m *memoryStore) GetObject(ctx context.Context, key string, offset, limit int64) (io.ReadCloser, error) {
	m.Lock()
	defer m.Unlock()
	// Minimum length is 1
	if key == "" {
		return nil, ErrInvalidObjectKey
	}
	d, ok := m.objects[key]
	if !ok {
		return nil, ErrNoSuchObject
	}

	if offset > int64(len(d.data)) {
		offset = int64(len(d.data))
	}
	data := d.data[offset:]
	log.Infow("uu", "limit", limit, "length", len(data))
	if limit > 0 && limit < int64(len(data)) {
		data = data[:limit]
	}
	return io.NopCloser(bytes.NewBuffer(data)), nil
}

func (m *memoryStore) PutObject(ctx context.Context, key string, reader io.Reader) error {
	m.Lock()
	defer m.Unlock()
	// Minimum length is 1
	if key == "" {
		return ErrInvalidObjectKey
	}
	if _, ok := m.objects[key]; ok {
		log.Info("overwrite key: ", key)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	m.objects[key] = &memoryObject{data: data, modTime: time.Now()}
	return nil
}

func (m *memoryStore) DeleteObject(ctx context.Context, key string) error {
	m.Lock()
	defer m.Unlock()
	delete(m.objects, key)
	return nil
}

func (m *memoryStore) DeleteObjectsByPrefix(ctx context.Context, key string) (uint64, error) {
	var (
		continueDeleteObject = true
		batchSize            = int64(1000)
		size                 uint64
	)

	for continueDeleteObject {
		// batch list and delete objects
		objs, err := m.ListObjects(ctx, key, "", "", batchSize)
		if err != nil {
			log.Errorw("DeleteObjectsByPrefix read directory error", "error", err)
			return size, err
		}

		// if the object listed here is less than required batch size, meaning it is the last page
		if int64(len(objs)) < batchSize {
			continueDeleteObject = false
		}

		for _, obj := range objs {
			deleteErr := m.DeleteObject(ctx, obj.Key())
			if deleteErr != nil {
				log.Errorw("remove single file by prefix error", "error", err)
			} else {
				size += uint64(obj.Size())
			}
		}
	}

	return size, nil
}

func (m *memoryStore) HeadBucket(ctx context.Context) error {
	return nil
}

func (m *memoryStore) HeadObject(ctx context.Context, key string) (Object, error) {
	m.Lock()
	defer m.Unlock()
	// Minimum length is 1
	if key == "" {
		return nil, ErrInvalidObjectKey
	}
	o, ok := m.objects[key]
	if !ok {
		return nil, os.ErrNotExist
	}

	return &object{
		key,
		int64(len(o.data)),
		o.modTime,
		false,
	}, nil
}

func (m *memoryStore) ListObjects(ctx context.Context, prefix, marker, delimiter string, limit int64) ([]Object, error) {
	if delimiter != "" {
		return nil, ErrUnsupportedDelimiter
	}
	m.Lock()
	defer m.Unlock()

	objs := make([]Object, 0)
	for k := range m.objects {
		if strings.HasPrefix(k, prefix) && k > marker {
			o := m.objects[k]
			f := &object{
				k,
				int64(len(o.data)),
				o.modTime,
				false,
			}
			objs = append(objs, f)
		}
	}
	sort.Slice(objs, func(i, j int) bool {
		return objs[i].Key() < objs[j].Key()
	})
	if int64(len(objs)) > limit {
		objs = objs[:limit]
	}
	return objs, nil
}
