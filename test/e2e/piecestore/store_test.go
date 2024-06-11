package piecestore_e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zkMeLabs/mechain-storage-provider/store/piecestore/storage"
)

func TestS3Store(t *testing.T) {
	// 1. init piece store
	handler, err := setup(t, storage.S3Store, s3BucketURL, storage.AKSKIAMType, 0)
	assert.Equal(t, err, nil)
	// 2. do some operations to test piece store api
	doOperations(t, handler)
}

func TestOSSStore(t *testing.T) {
	// 1. init piece store
	handler, err := setup(t, storage.OSSStore, ossBucketURL, storage.SAIAMType, 0)
	assert.Equal(t, err, nil)
	// 2. do some operations to test piece store api
	doOperations(t, handler)
}

func TestMinioStore(t *testing.T) {
	// 1. init piece store
	handler, err := setup(t, storage.MinioStore, s3BucketURL, storage.AKSKIAMType, 0)
	assert.Equal(t, err, nil)
	// 2. do some operations to test piece store api
	doOperations(t, handler)
}

func TestB2Store(t *testing.T) {
	// 1. init piece store
	handler, err := setup(t, storage.B2Store, b2BucketURL, storage.AKSKIAMType, 0)
	assert.Equal(t, err, nil)
	// 2. do some operations to test piece store api
	doOperations(t, handler)
}

func TestDiskFileStore(t *testing.T) {
	// 1. init piece store
	handler, err := setup(t, storage.DiskFileStore, "", storage.AKSKIAMType, 0)
	assert.Equal(t, err, nil)
	// 2. do some operations to test piece store api
	doOperations(t, handler)
}

func TestMemoryStore(t *testing.T) {
	// 1. init piece store
	handler, err := setup(t, storage.MemoryStore, "", storage.AKSKIAMType, 0)
	assert.Equal(t, err, nil)
	// 2. do some operations to test piece store api
	doOperations(t, handler)
}
