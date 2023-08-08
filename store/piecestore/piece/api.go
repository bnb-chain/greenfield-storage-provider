package piece

import (
	"context"
	"io"

	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
)

// PieceAPI for mock use
//
//go:generate mockgen -source=./api.go -destination=./api_mock.go -package=piece
type PieceAPI interface {
	Get(ctx context.Context, key string, offset, limit int64) (io.ReadCloser, error)
	Put(ctx context.Context, key string, reader io.Reader) error
	Delete(ctx context.Context, key string) error
}

type PieceStore struct {
	storeAPI storage.ObjectStorage
}

// Get one piece from PieceStore
func (p *PieceStore) Get(ctx context.Context, key string, offset, limit int64) (io.ReadCloser, error) {
	return p.storeAPI.GetObject(ctx, key, offset, limit)
}

// Put one piece to PieceStore
func (p *PieceStore) Put(ctx context.Context, key string, reader io.Reader) error {
	return p.storeAPI.PutObject(ctx, key, reader)
}

// Delete one piece in PieceStore
func (p *PieceStore) Delete(ctx context.Context, key string) error {
	return p.storeAPI.DeleteObject(ctx, key)
}

// Head returns piece info in PieceStore
func (p *PieceStore) Head(ctx context.Context, key string) (storage.Object, error) {
	return p.storeAPI.HeadObject(ctx, key)
}
