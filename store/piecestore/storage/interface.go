package storage

import (
	"context"
	"io"
	"time"
)

// ObjectStorage is a common interface that must be implemented if some users want to use an object
// storage (such as S3, Azure Blob, Minio, OSS, COS, etc)
//
//go:generate mockgen -source=./interface.go -destination=./interface_mock.go -package=storage
type ObjectStorage interface {
	// String the description of an object storage
	String() string
	// CreateBucket create the bucket if not existed
	CreateBucket(ctx context.Context) error
	// GetObject gets data for the given object specified by key
	GetObject(ctx context.Context, key string, offset, limit int64) (io.ReadCloser, error)
	// PutObject puts data read from a reader to an object specified by key
	PutObject(ctx context.Context, key string, reader io.Reader) error
	// DeleteObject deletes an object
	DeleteObject(ctx context.Context, key string) error

	// HeadBucket determines if a bucket exists and have permission to access it
	HeadBucket(ctx context.Context) error
	// HeadObject returns some information about the object or an error if not found
	HeadObject(ctx context.Context, key string) (Object, error)
	// ListObjects lists returns a list of objects
	ListObjects(ctx context.Context, prefix, marker, delimiter string, limit int64) ([]Object, error)
	// ListAllObjects returns all the objects as a channel
	ListAllObjects(ctx context.Context, prefix, marker string) (<-chan Object, error)
}

// Object
type Object interface {
	Key() string
	Size() int64
	ModTime() time.Time
	IsSymlink() bool
}

type object struct {
	key     string
	size    int64
	modTime time.Time
	isDir   bool
}

func (o *object) Key() string        { return o.key }
func (o *object) Size() int64        { return o.size }
func (o *object) ModTime() time.Time { return o.modTime }
func (o *object) IsDir() bool        { return o.isDir }
func (o *object) IsSymlink() bool    { return false }
