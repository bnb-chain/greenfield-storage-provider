package piece

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage"
	weaveVMtypes "github.com/bnb-chain/greenfield-storage-provider/store/piecestore/storage/weavevm/types"
)

// NewPieceStore returns an instance of PieceStore
func NewPieceStore(pieceConfig *storage.PieceStoreConfig) (*PieceStore, error) {
	if err := checkConfig(pieceConfig); err != nil {
		log.Errorw("failed to check piece store config", "error", err)
		return nil, err
	}
	blob, err := createStorage(*pieceConfig)
	if err != nil {
		log.Errorw("failed to create storage", "error", err)
		return nil, err
	}
	log.Infow("piece store is running", "storage type", pieceConfig.Store.Storage,
		"shards", pieceConfig.Shards)

	return &PieceStore{blob}, nil
}

// checkConfig checks config if right
func checkConfig(cfg *storage.PieceStoreConfig) error {
	overrideConfigFromEnv(cfg)
	if cfg.Shards > 256 {
		return fmt.Errorf("too many shards: %d", cfg.Shards)
	}
	if cfg.Store.IAMType != storage.AKSKIAMType && cfg.Store.IAMType != storage.SAIAMType {
		return fmt.Errorf("invalid iam type: %s", cfg.Store.IAMType)
	}
	if cfg.Store.MaxRetries < 0 {
		return fmt.Errorf("MaxRetries should be equal or greater than zero")
	}
	if cfg.Store.MinRetryDelay < 0 {
		return fmt.Errorf("MinRetryDelay should be equal or greater than zero")
	}
	if cfg.Store.Storage == storage.DiskFileStore {
		if cfg.Store.BucketURL == "" {
			cfg.Store.BucketURL = setDefaultFileStorePath()
		}
		p, err := filepath.Abs(cfg.Store.BucketURL)
		if err != nil {
			log.Errorw("failed to get absolute path", "bucket", cfg.Store.BucketURL, "error", err)
			return err
		}
		cfg.Store.BucketURL = p
		cfg.Store.BucketURL += "/"
	}

	if cfg.Store.Storage == storage.WeavevmStore {
		weavevmConfig, err := getWeaveVMConfigFromEnv()
		if err != nil {
			return fmt.Errorf("failed to get weavevm config from env vars: %w", err)
		}
		cfg.Store.WeavevmConfig = *weavevmConfig
	}
	return nil
}

func overrideConfigFromEnv(cfg *storage.PieceStoreConfig) {
	if val, ok := os.LookupEnv(storage.BucketURL); ok {
		cfg.Store.BucketURL = val
	}
}

func getWeaveVMConfigFromEnv() (*weaveVMtypes.Config, error) {
	cfg := &weaveVMtypes.Config{
		// Set defaults
		RetryAttempts: 3,
		RetryDelay:    time.Millisecond * 100,
		Timeout:       time.Second * 3,
	}

	// Required: Endpoint
	endpoint := os.Getenv(storage.WeaveVMEndpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("%s environment variable is required", storage.WeaveVMEndpoint)
	}
	cfg.Endpoint = endpoint

	// Required: ChainID
	chainID := os.Getenv(storage.WeaveVMChainID)
	if chainID == "" {
		return nil, fmt.Errorf("%s environment variable is required", storage.WeaveVMChainID)
	}
	id, err := strconv.ParseInt(chainID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", storage.WeaveVMChainID, err)
	}
	cfg.ChainID = id

	// Authentication method
	web3Endpoint := os.Getenv(storage.WeaveVMWeb3Endpoint)
	privateKey := os.Getenv(storage.WeaveVMPrivateKey)

	if web3Endpoint != "" {
		cfg.Web3SignerEndpoint = web3Endpoint

		// Web3Signer TLS config - all files must be provided if using TLS
		tlsCert := os.Getenv(storage.WeaveVMWeb3TLSCert)
		tlsKey := os.Getenv(storage.WeaveVMWeb3TLSKey)
		tlsCACert := os.Getenv(storage.WeaveVMWeb3TLSCACert)

		if tlsCert != "" || tlsKey != "" || tlsCACert != "" {
			// If any TLS env var is set, all must be set
			if tlsCert == "" || tlsKey == "" || tlsCACert == "" {
				return nil, fmt.Errorf("all TLS files must be provided when using Web3Signer with TLS: cert, key and CA cert")
			}
			cfg.Web3SignerTLSCertFile = tlsCert
			cfg.Web3SignerTLSKeyFile = tlsKey
			cfg.Web3SignerTLSCACertFile = tlsCACert
		}
	} else if privateKey != "" {
		cfg.PrivateKeyHex = privateKey
	} else {
		return nil, fmt.Errorf("either %s or %s must be provided", storage.WeaveVMWeb3Endpoint, storage.WeaveVMPrivateKey)
	}

	// Optional params
	if timeoutStr := os.Getenv(storage.WeaveVMTimeout); timeoutStr != "" {
		timeout, err := strconv.ParseInt(timeoutStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid %s: %w", storage.WeaveVMTimeout, err)
		}
		cfg.Timeout = time.Second * time.Duration(timeout)
	}

	if attemptsStr := os.Getenv(storage.WeaveVMRetryAttempts); attemptsStr != "" {
		attempts, err := strconv.Atoi(attemptsStr)
		if err != nil {
			return nil, fmt.Errorf("invalid %s: %w", storage.WeaveVMRetryAttempts, err)
		}
		if attempts < 0 {
			return nil, fmt.Errorf("%s must be non-negative", storage.WeaveVMRetryAttempts)
		}
		cfg.RetryAttempts = attempts
	}

	if delayStr := os.Getenv(storage.WeaveVMRetryDelay); delayStr != "" {
		delay, err := strconv.ParseInt(delayStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid %s: %w", storage.WeaveVMRetryDelay, err)
		}
		if delay < 0 {
			return nil, fmt.Errorf("%s must be non-negative", storage.WeaveVMRetryDelay)
		}
		cfg.RetryDelay = time.Millisecond * time.Duration(delay)
	}

	return cfg, nil
}

func createStorage(cfg storage.PieceStoreConfig) (storage.ObjectStorage, error) {
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
		log.Errorw("failed to create storage", "error", err, "object", object)
		return nil, err
	}

	if err = checkBucket(context.Background(), object); err != nil {
		log.Errorw("failed to check bucket due to storage is not configured rightly ", "error", err,
			"object", object)
		return nil, err
	}

	return object, nil
}

// checkBucket checks bucket if exists
func checkBucket(ctx context.Context, store storage.ObjectStorage) error {
	if err := store.HeadBucket(ctx); err != nil {
		log.Errorw("failed to head bucket", "error", err)
		if errors.Is(err, storage.ErrNoSuchBucket) {
			if err2 := store.CreateBucket(ctx); err2 != nil {
				return fmt.Errorf("failed to create bucket in %s: %s, previous err: %s", store, err2, err)
			}
			log.Info("create bucket successfully!")
			return nil
		}
		return storage.ErrNoPermissionAccessBucket
	}
	log.Debugf("succeed to head bucket in %s", store)
	return nil
}

func setDefaultFileStorePath() string {
	defaultBucket := "/var/piecestore"
	switch runtime.GOOS {
	case "linux":
		if os.Getuid() == 0 {
			break
		}
		fallthrough
	case "darwin":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Panicw("failed to get current user's home directory", "error", err)
		}
		defaultBucket = path.Join(homeDir, ".piecestore", "local")
	case "windows":
		defaultBucket = path.Join("C:/piecestore/local")
	default:
		log.Panic("Unknown operating system!")
	}
	return defaultBucket
}
