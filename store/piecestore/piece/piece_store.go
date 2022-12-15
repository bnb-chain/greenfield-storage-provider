package piece

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/bnb-chain/inscription-storage-provider/config"
	"github.com/bnb-chain/inscription-storage-provider/model"
	"github.com/bnb-chain/inscription-storage-provider/store/piecestore/storage"
	"github.com/bnb-chain/inscription-storage-provider/util/log"
)

var (
	serviceName = "piece_store"
)

// NewPieceStore returns an instance of PieceStore
func NewPieceStore(filePath string) (*PieceStore, error) {
	cfg := checkConfig(config.LoadConfig(filePath))
	initLog(&cfg.Log)
	blob, err := createStorage(&cfg.PieceStore)
	if err != nil {
		log.Panicw("create storage error", "error", err)
		return nil, err
	}
	log.Infow("PieceStore is running", "Storage", cfg.PieceStore.Store.Storage, "BucketURL",
		cfg.PieceStore.Store.BucketURL)

	return &PieceStore{blob}, nil
}

// checkConfig checks config if right
func checkConfig(cfg *config.Config) *config.Config {
	if cfg.PieceStore.Shards > 256 {
		log.Panicf("too many shards: %d", cfg.PieceStore.Shards)
	}
	if cfg.PieceStore.Store.MaxRetries < 0 {
		log.Panic("MaxRetries should be equal or greater than zero")
	}
	if cfg.PieceStore.Store.MinRetryDelay < 0 {
		log.Panic("MinRetryDelay should be equal or greater than zero")
	}
	if cfg.PieceStore.Store.Storage == "file" {
		if cfg.PieceStore.Store.BucketURL == "" {
			cfg.PieceStore.Store.BucketURL = setDefaultFileStorePath()
		}
		p, err := filepath.Abs(cfg.PieceStore.Store.BucketURL)
		if err != nil {
			log.Panicw("Failed to get absolute path", "bucket", cfg.PieceStore.Store.BucketURL, "error", err)
		}
		cfg.PieceStore.Store.BucketURL = p
		cfg.PieceStore.Store.BucketURL += "/"
	}
	return cfg
}

// initLog initialize log config
func initLog(cfg *config.LogConfig) {
	lvl, _ := log.ParseLevel(cfg.Level)
	log.Init(lvl, log.StandardizePath(cfg.FilePath, serviceName))
}

func createStorage(cfg *config.PieceStoreConfig) (storage.ObjectStorage, error) {
	var (
		object storage.ObjectStorage
		err    error
	)
	if cfg.Shards > 1 {
		object, err = storage.NewSharded(cfg)
	} else {
		object, err = storage.NewObjectStorage(&cfg.Store)
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
		if errors.Is(err, model.BucketNotExisted) {
			if err2 := store.CreateBucket(ctx); err2 != nil {
				return fmt.Errorf("Failed to create bucket in %s: %s, previous err: %s", store, err2, err)
			}
			log.Info("Create bucket successfully!")
			return nil
		}
		return fmt.Errorf("Check if you have the permission to access the bucket")
	}
	log.Infof("HeadBucket succeeds in %s, bucketName %s", store)
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
