package piece

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// NewPieceStore returns an instance of PieceStore
func NewPieceStore(pieceConfig *storage.PieceStoreConfig) (*PieceStore, error) {
	if err := overrideConfigFromEnv(pieceConfig); err != nil {
		log.Errorw("load piece store from env failed")
		return nil, err
	}
	cfg := checkConfig(pieceConfig)
	blob, err := createStorage(cfg)
	if err != nil {
		log.Errorw("create storage error", "error", err)
		return nil, err
	}
	log.Debugf("pieceStore is running", "Storage", cfg.Store.Storage, "BucketURL",
		cfg.Store.BucketURL)

	return &PieceStore{blob}, nil
}

// checkConfig checks config if right
func checkConfig(cfg *storage.PieceStoreConfig) *storage.PieceStoreConfig {
	if cfg.Shards > 256 {
		log.Panicf("too many shards: %d", cfg.Shards)
	}
	if cfg.Store.MaxRetries < 0 {
		log.Panic("MaxRetries should be equal or greater than zero")
	}
	if cfg.Store.MinRetryDelay < 0 {
		log.Panic("MinRetryDelay should be equal or greater than zero")
	}
	if cfg.Store.Storage == "file" {
		if cfg.Store.BucketURL == "" {
			cfg.Store.BucketURL = setDefaultFileStorePath()
		}
		p, err := filepath.Abs(cfg.Store.BucketURL)
		if err != nil {
			log.Panicw("Failed to get absolute path", "bucket", cfg.Store.BucketURL, "error", err)
		}
		cfg.Store.BucketURL = p
		cfg.Store.BucketURL += "/"
	}
	return cfg
}

func overrideConfigFromEnv(cfg *storage.PieceStoreConfig) error {
	if val, ok := os.LookupEnv(model.BucketURL); ok {
		cfg.Store.BucketURL = val
	} else {
		return fmt.Errorf("piece store config %s is not setted in environment", model.BucketURL)
	}
	return nil
}

func createStorage(cfg *storage.PieceStoreConfig) (storage.ObjectStorage, error) {
	var (
		object storage.ObjectStorage
		err    error
	)
	if cfg.Shards > 1 {
		object, err = storage.NewSharded(cfg)
	} else {
		object, err = storage.NewObjectStorage(cfg.Store)
	}
	if err != nil {
		log.Errorw("createStorage error", "error", err, "object", object)
		return nil, err
	}

	if err = checkBucket(context.Background(), object); err != nil {
		log.Errorw("checkBucket error, storage is not configured rightly ", "error", err,
			"object", object)
		return nil, err
	}

	return object, nil
}

// checkBucket checks bucket if exists
func checkBucket(ctx context.Context, store storage.ObjectStorage) error {
	if err := store.HeadBucket(ctx); err != nil {
		log.Errorw("HeadBucket error", "error", err)
		if errors.Is(err, merrors.BucketNotExisted) {
			if err2 := store.CreateBucket(ctx); err2 != nil {
				return fmt.Errorf("Failed to create bucket in %s: %s, previous err: %s", store, err2, err)
			}
			log.Info("Create bucket successfully!")
			return nil
		}
		return merrors.ErrNoPermissionAccessBucket
	}
	log.Debugf("HeadBucket succeeds in %s", store)
	return nil
}

func setDefaultFileStorePath() string {
	var defaultBucket = "/var/piecestore"
	switch runtime.GOOS {
	case "linux":
		if os.Getuid() == 0 {
			break
		}
		fallthrough
	case "darwin":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Panicf("error: %v", err)
		}
		defaultBucket = path.Join(homeDir, ".piecestore", "local")
	case "windows":
		defaultBucket = path.Join("C:/piecestore/local")
	default:
		log.Error("Unknown operating system!")
	}
	return defaultBucket
}
