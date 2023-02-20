package metalevel

import (
	"encoding/json"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/spdb"
)

var _ spdb.MetaDB = &Database{}

// Database is a persistent key-value store.
type Database struct {
	Path      string
	Namespace string
	db        *leveldb.DB // LevelDB instance
}

var (
	once   sync.Once
	metaDB *Database
)

// NewMetaDB call NewCustomMetaDB return Database instance.
func NewMetaDB(config *config.LevelDBConfig) (*Database, error) {
	var err error
	once.Do(func() {
		metaDB, err = newCustomMetaDB(config.Path, config.NameSpace, config.Cache, config.FileHandles, config.ReadOnly)
	})
	return metaDB, err
}

// NewCustomMetaDB return Database instance.
func newCustomMetaDB(path string, namespace string, cache int, handles int, readonly bool) (*Database, error) {
	// init options
	optionFunc := func() *opt.Options {
		options := &opt.Options{
			Filter: filter.NewBloomFilter(10),
		}
		// Ensure we have some minimal caching and file guarantees
		if cache < config.MinCache {
			cache = config.MinCache
		}
		if handles < config.MinHandles {
			handles = config.MinHandles
		}
		// Set default options
		options.OpenFilesCacheCapacity = handles
		options.BlockCacheCapacity = cache / 2 * opt.MiB
		options.WriteBuffer = cache / 4 * opt.MiB // Two of these are used internally
		if readonly {
			options.ReadOnly = true
		}
		return options
	}
	options := optionFunc()

	// Open the db and recover any potential corruptions
	db, err := leveldb.OpenFile(path, options)
	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		db, err = leveldb.RecoverFile(path, nil)
	}
	if err != nil {
		return nil, err
	}

	// Assemble the wrapper with all the registered metrics
	ldb := &Database{
		Path:      path,
		Namespace: namespace,
		db:        db,
	}
	return ldb, nil
}

// Close the leveldb resource.
func (db *Database) Close() error {
	return db.db.Close()
}

// SetIntegrityMeta put integrity hash info to db.
func (db *Database) SetIntegrityMeta(meta *spdb.IntegrityMeta) error {
	if meta == nil {
		return errors.New("primary integrity meta is nil")
	}
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return db.db.Put(IntegrityMetaKey(db.Namespace,
		meta.ObjectID, meta.IsPrimary, meta.RedundancyType, meta.EcIdx), data, nil)
}

// GetIntegrityMeta return the integrity hash info

func (db *Database) GetIntegrityMeta(queryCondition *spdb.IntegrityMeta) (*spdb.IntegrityMeta, error) {
	data, err := db.db.Get(IntegrityMetaKey(db.Namespace,
		queryCondition.ObjectID, queryCondition.IsPrimary, queryCondition.RedundancyType, queryCondition.EcIdx), nil)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("integrity info not exits")
	}
	var metaReturn spdb.IntegrityMeta
	err = json.Unmarshal(data, &metaReturn)
	return &metaReturn, err
}

// SetUploadPayloadAskingMeta put payload asking info to db.
func (db *Database) SetUploadPayloadAskingMeta(meta *spdb.UploadPayloadAskingMeta) error {
	if meta == nil {
		return errors.New("upload payload meta is nil")
	}
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return db.db.Put(UploadPayloadAsingKey(db.Namespace, meta.BucketName, meta.ObjectName), data, nil)
}

// GetUploadPayloadAskingMeta return the payload asking info.
func (db *Database) GetUploadPayloadAskingMeta(bucketName, objectName string) (*spdb.UploadPayloadAskingMeta, error) {
	data, err := db.db.Get(UploadPayloadAsingKey(db.Namespace, bucketName, objectName), nil)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("upload payload meta not exits")
	}
	var meta spdb.UploadPayloadAskingMeta
	err = json.Unmarshal(data, &meta)
	return &meta, err
}
