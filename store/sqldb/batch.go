package sqldb

// IdealBatchSize defines the size of the data batches should ideally add in one
// write.
// const IdealBatchSize = 100 * 1024

// Batch is a write-only database that commits changes to its host database
// when Write is called. A batch cannot be used concurrently.
//
//go:generate mockgen -source=./batch.go -destination=./batch_mock.go -package=sqldb
type Batch interface {
	// Put inserts the given value into the key-value data store.
	Put(key interface{}, value interface{}) error

	// Delete removes the key from the key-value data store.
	Delete(key interface{}) error

	// ValueSize retrieves the amount of data queued up for writing.
	ValueSize() int

	// Write flushes any accumulated data to disk.
	Write() error

	// Reset resets the batch for reuse.
	Reset()
}

// Batcher wraps the NewBatch method of a backing data store.
type Batcher interface {
	// NewBatch creates a write-only database that buffers changes to its host db
	// until a final write is called.
	NewBatch() Batch

	// NewBatchWithSize creates a write-only database batch with pre-allocated buffer.
	NewBatchWithSize(size int) Batch
}
