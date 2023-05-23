package storage

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	cfg = ObjectStorageConfig{
		Storage:   "b2",
		BucketURL: "https://s3.us-east-005.backblazeb2.com/greenfieldsp",
	}

	ctx = context.Background()

	key     = "test.txt"
	content = "S3 Compatible API"
)

func Test_B2Store_CreateBucket(t *testing.T) {
	b2, err := NewObjectStorage(cfg)
	assert.NoError(t, err)

	err = b2.CreateBucket(ctx)
	assert.NoError(t, err)
}

func Test_B2Store_PutObject(t *testing.T) {
	b2, err := NewObjectStorage(cfg)
	assert.NoError(t, err)

	err = b2.PutObject(ctx, key, strings.NewReader(content))
	assert.NoError(t, err)
}

func Test_B2Store_GetObject(t *testing.T) {
	b2, err := NewObjectStorage(cfg)
	assert.NoError(t, err)

	data, err := b2.GetObject(ctx, key, 0, 0)
	assert.NoError(t, err)

	data1, err := io.ReadAll(data)
	assert.NoError(t, err)

	assert.Equal(t, content, string(data1))
}

// DeleteObject deletes an object from the bucket.
func Test_B2Store_DeleteObject(t *testing.T) {
	b2, err := NewObjectStorage(cfg)
	assert.NoError(t, err)

	err = b2.DeleteObject(ctx, key)
	assert.NoError(t, err)
}

// HeadBucket determines if a bucket exists and have permission to access it
func Test_B2Store_HeadBucket(t *testing.T) {
	b2, err := NewObjectStorage(cfg)
	assert.NoError(t, err)

	err = b2.HeadBucket(ctx)
	assert.NoError(t, err)
}

// HeadObject returns some information about the object or an error if not found
func Test_B2Store_HeadObject(t *testing.T) {
	b2, err := NewObjectStorage(cfg)
	assert.NoError(t, err)

	_, err = b2.HeadObject(ctx, key)
	assert.NoError(t, err)
}

// ListObjects lists returns a list of objects
func Test_B2Store_ListObjects(t *testing.T) {
	b2, err := NewObjectStorage(cfg)
	assert.NoError(t, err)

	_, err = b2.ListObjects(ctx, "", "", "", 0)
	assert.NoError(t, err)
}
