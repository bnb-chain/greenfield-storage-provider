package metalevel

import (
	"encoding/json"
	"sync"

	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/spdb"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const (
	// minCache is the minimum amount of memory in megabytes to allocate to leveldb
	// read and write caching, split half and half.
	minCache = 16

	// minHandles is the minimum number of files handles to allocate to the open
	// database files.
	minHandles = 16
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
		if cache < minCache {
			cache = minCache
		}
		if handles < minHandles {
			handles = minHandles
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
	return db.db.Put(IntegrityMetaKey(db.Namespace, meta.ObjectID), data, nil)
}

// GetIntegrityMeta return the integrity hash info

func (db *Database) GetIntegrityMeta(objectID uint64) (*spdb.IntegrityMeta, error) {
	data, err := db.db.Get(IntegrityMetaKey(db.Namespace, objectID), nil)
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
