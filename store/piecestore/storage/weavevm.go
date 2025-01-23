package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

type weavevmStore struct {
}

func newWeaveVMmStore(cfg ObjectStorageConfig) (ObjectStorage, error) {
	awsSession, bucket, err := s3SessionCache.newSession(cfg)
	if err != nil {
		log.Errorw("failed to new s3 session", "error", err)
		return nil, err
	}
	log.Infow("new S3 store succeeds", "bucket", bucket)

	return &s3Store{bucketName: bucket, api: s3.New(awsSession)}, nil
}

// show address on weavevm?
func (s *weavevmStore) String() string {
	return fmt.Sprintf("weavevm")
}

// TODO: 	return nil, ErrUnsupportedMethod ?
func (s *weavevmStore) CreateBucket(ctx context.Context) error {
	log.Debugw("create bucket operation on WeaveVM store - non supported")
	return nil
}

func (s *weavevmStore) GetObject(ctx context.Context, key string, offset, limit int64) (io.ReadCloser, error) {
	return nil, nil
}

// hmmm, key incoming?
func (s *weavevmStore) PutObject(ctx context.Context, key string, reader io.Reader) error {
	return nil
}

// TODO: 	return nil, ErrUnsupportedMethod ?
func (s *weavevmStore) DeleteObject(ctx context.Context, key string) error {
	log.Debugw("delete operation on WeaveVM store - data remains permanently available", "key", key)
	return nil
}

// TODO: 	return nil, ErrUnsupportedMethod ?
func (s *weavevmStore) DeleteObjectsByPrefix(ctx context.Context, key string) (uint64, error) {
	log.Debugw("bulk delete operation on WeaveVM store - data remains permanently available", "key", key)
	return 0, nil
}

func (s *weavevmStore) HeadBucket(ctx context.Context) error {
	return nil
}

func (s *weavevmStore) HeadObject(ctx context.Context, key string) (Object, error) {

	return nil, nil
}

func (s *weavevmStore) ListObjects(ctx context.Context, prefix, marker, delimiter string, limit int64) ([]Object, error) {
	return nil, ErrUnsupportedMethod

}

func (s *weavevmStore) ListAllObjects(ctx context.Context, prefix, marker string) (<-chan Object, error) {
	return nil, ErrUnsupportedMethod
}
