package storage

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

func NewObjectStorage(cfg ObjectStorageConfig) (ObjectStorage, error) {
	if fn, ok := storageMap[strings.ToLower(cfg.Storage)]; ok {
		log.Debugf("create [%s] storage at endpoint %s", cfg.Storage, cfg.BucketURL)
		return fn(cfg)
	}
	return nil, fmt.Errorf("invalid object storage: %s", cfg.Storage)
}

type StorageFn func(cfg ObjectStorageConfig) (ObjectStorage, error)

var storageMap = map[string]StorageFn{
	S3Store:       newS3Store,
	OSSStore:      newOSSStore,
	B2Store:       newB2Store,
	MinioStore:    newMinioStore,
	LdfsStore:     newLdfsStore,
	DiskFileStore: newDiskFileStore,
	MemoryStore:   newMemoryStore,
}

type DefaultObjectStorage struct{}

func (s DefaultObjectStorage) CreateBucket(ctx context.Context) error {
	return nil
}

func (s DefaultObjectStorage) ListObjects(ctx context.Context, prefix, marker, delimiter string, limit int64) ([]Object, error) {
	return nil, ErrUnsupportedMethod
}

func (s DefaultObjectStorage) ListAllObjects(ctx context.Context, prefix, marker string) (<-chan Object, error) {
	return nil, ErrUnsupportedMethod
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
		buf := make([]byte, BufPoolSize)
		return &buf
	},
}

type objectStorageSecretKey struct {
	accessKey    string
	secretKey    string
	sessionToken string
}

func getSecretKeyFromEnv(accessKey, secretKey, sessionToken string) *objectStorageSecretKey {
	key := &objectStorageSecretKey{}
	if val, ok := os.LookupEnv(accessKey); ok {
		key.accessKey = val
	}
	if val, ok := os.LookupEnv(secretKey); ok {
		key.secretKey = val
	}
	if val, ok := os.LookupEnv(sessionToken); ok {
		key.sessionToken = val
	}
	return key
}

type ossStorageSecretKey struct {
	region       string
	accessKey    string
	secretKey    string
	sessionToken string
}

func getOSSSecretKeyFromEnv(region, accessKey, secretKey, sessionToken string) *ossStorageSecretKey {
	key := &ossStorageSecretKey{}
	if val, ok := os.LookupEnv(region); ok {
		key.region = val
	}
	if val, ok := os.LookupEnv(accessKey); ok {
		key.accessKey = val
	}
	if val, ok := os.LookupEnv(secretKey); ok {
		key.secretKey = val
	}
	if val, ok := os.LookupEnv(sessionToken); ok {
		key.sessionToken = val
	}
	return key
}
