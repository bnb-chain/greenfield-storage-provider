package netdb

import (
	"encoding/binary"
	"encoding/json"
	"sync"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/store/spdb"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/bnb-chain/greenfield-storage-provider/store/config"
)

const (
	NodeIdPrefix  = "p2p_node_id"
	NodeIdxPrefix = "p2p_node_idx"
)

//ID        uint `gorm:"primarykey"`
//CreatedAt time.Time
//UpdatedAt time.Time
//DeletedAt DeletedAt `gorm:"index"`

// Database is a persistent key-value store.
type Database struct {
	Path      string
	Namespace string
	db        *leveldb.DB // LevelDB instance
	idx       uint
}

var (
	once   sync.Once
	metaDB *Database
)

var _ spdb.P2PNodeDB = &Database{}

// NewNetDB call NewCustomNetDB return Database instance.
func NewNetDB(config *config.LevelDBConfig) (*Database, error) {
	var err error
	once.Do(func() {
		metaDB, err = newCustomNetDB(config.Path, config.NameSpace, config.Cache, config.FileHandles, config.ReadOnly)
	})
	return metaDB, err
}

// newCustomNetDB return Database instance.
func newCustomNetDB(path string, namespace string, cache int, handles int, readonly bool) (*Database, error) {
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

func (db *Database) Get(nodeId string) (spdb.Provider, error) {
	data, err := db.db.Get([]byte(NodeIdPrefix+nodeId), nil)
	if err != nil {
		return spdb.Provider{}, err
	}
	var provider spdb.Provider
	err = json.Unmarshal(data, &provider)
	return provider, err
}

func (db *Database) Create(provider *spdb.Provider) error {
	provider.ID = db.idx
	provider.CreatedAt = time.Now()
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(provider.ID))
	if err := db.db.Put([]byte(NodeIdxPrefix), buf, nil); err != nil {
		return err
	}
	db.idx++
	data, err := json.Marshal(provider)
	if err != nil {
		return err
	}
	return db.db.Put([]byte(NodeIdPrefix+provider.NodeId), data, nil)
}

func (db *Database) Delete(nodeId string) error {
	return db.db.Delete([]byte(NodeIdPrefix+nodeId), nil)
}

func (db *Database) FetchAll() ([]spdb.Provider, error) {
	var providers []spdb.Provider
	iter := db.db.NewIterator(util.BytesPrefix([]byte(NodeIdPrefix)), nil)
	if iter.Next() {
		if err := iter.Error(); err != nil {
			return providers, err
		}
		value := iter.Value()
		var provider spdb.Provider
		if err := json.Unmarshal(value, &provider); err != nil {
			return providers, err
		}
		providers = append(providers, provider)
	}
	iter.Release()
	return providers, nil
}
