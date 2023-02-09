package jobmemory

import (
	"errors"
	"sort"
	"sync"

	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/spdb"
)

var _ spdb.JobDBV2 = &MemJobDB{}

// MemJobDB is a memory db, maintains job, object and piece job table.
type MemJobDB struct {
	JobCount               uint64
	JobTable               map[uint64]*ptypes.JobContext
	ObjectTable            map[uint64]*ptypes.ObjectInfo
	PrimaryPieceJobTable   map[uint64]map[uint32]spdb.PieceJob
	SecondaryPieceJobTable map[uint64]map[uint32]spdb.PieceJob
	JobToObject            map[uint64]uint64
	mu                     sync.RWMutex
}

// NewMemJobDB return a MemJobDBV2 instance.
func NewMemJobDB() *MemJobDB {
	return &MemJobDB{
		JobCount:               0,
		JobTable:               make(map[uint64]*ptypes.JobContext),
		ObjectTable:            make(map[uint64]*ptypes.ObjectInfo),
		PrimaryPieceJobTable:   make(map[uint64]map[uint32]spdb.PieceJob),
		SecondaryPieceJobTable: make(map[uint64]map[uint32]spdb.PieceJob),
		JobToObject:            make(map[uint64]uint64),
	}
}

// CreateUploadPayloadJobV2 create a job info for special object.
func (db *MemJobDB) CreateUploadPayloadJobV2(info *ptypes.ObjectInfo) (uint64, error) {
	if info == nil {
		return 0, errors.New("object info is nil")
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	db.JobTable[db.JobCount] = &ptypes.JobContext{
		JobId:    db.JobCount,
		JobState: ptypes.JobState_JOB_STATE_CREATE_OBJECT_DONE,
	}
	info.JobId = db.JobCount
	db.ObjectTable[info.GetObjectId()] = info
	db.JobToObject[db.JobCount] = info.GetObjectId()
	db.JobCount++
	return db.JobCount - 1, nil
}

// GetJobContextV2 returns the job info .
func (db *MemJobDB) GetJobContextV2(jobId uint64) (*ptypes.JobContext, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	job, ok := db.JobTable[jobId]
	if !ok {
		return nil, errors.New("job is not exist")
	}
	return job, nil
}

// GetObjectInfoByJobV2 returns the object info by job id.
func (db *MemJobDB) GetObjectInfoByJobV2(JobID uint64) (*ptypes.ObjectInfo, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	objectID, ok := db.JobToObject[JobID]
	if !ok {
		return nil, errors.New("job is not exist")
	}
	objectInfo, ok := db.ObjectTable[objectID]
	if !ok {
		return nil, errors.New("object is not exist")
	}
	return objectInfo, nil
}

// GetObjectInfoV2 returns the object info by object id.
func (db *MemJobDB) GetObjectInfoV2(objectID uint64) (*ptypes.ObjectInfo, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	objectInfo, ok := db.ObjectTable[objectID]
	if !ok {
		return nil, errors.New("object is not exist")
	}
	return objectInfo, nil
}

// SetUploadPayloadJobStateV2 set the job state.
func (db *MemJobDB) SetUploadPayloadJobStateV2(jobId uint64, state string, timestamp int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	job, ok := db.JobTable[jobId]
	if !ok {
		return errors.New("job is not exist")
	}
	jobState, ok := ptypes.JobState_value[state]
	if !ok {
		return errors.New("state is not correct job state")
	}
	job.JobState = (ptypes.JobState)(jobState)
	job.ModifyTime = timestamp
	db.JobTable[jobId] = job
	return nil
}

// SetUploadPayloadJobJobErrorV2 set the job error state and error message.
func (db *MemJobDB) SetUploadPayloadJobJobErrorV2(jobId uint64, state string, jobErr string, timestamp int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	job, ok := db.JobTable[jobId]
	if !ok {
		return errors.New("job is not exist")
	}
	jobState, ok := ptypes.JobState_value[state]
	if !ok {
		return errors.New("state is not correct job state")
	}
	job.JobState = (ptypes.JobState)(jobState)
	job.ModifyTime = timestamp
	job.JobErr = jobErr
	db.JobTable[jobId] = job
	return nil
}

// GetPrimaryJobV2 returns the primary piece jobs by object id.
func (db *MemJobDB) GetPrimaryJobV2(objectID uint64) ([]*spdb.PieceJob, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if _, ok := db.PrimaryPieceJobTable[objectID]; !ok {
		return []*spdb.PieceJob{}, nil
	}
	pieces := make([]*spdb.PieceJob, len(db.PrimaryPieceJobTable[objectID]))
	for idx, job := range db.PrimaryPieceJobTable[objectID] {
		pieces[idx] = &job
	}
	return pieces, nil
}

// GetSecondaryJobV2 returns the secondary piece jobs by object id.
func (db *MemJobDB) GetSecondaryJobV2(objectID uint64) ([]*spdb.PieceJob, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if _, ok := db.SecondaryPieceJobTable[objectID]; !ok {
		return []*spdb.PieceJob{}, nil
	}
	pieces := make([]*spdb.PieceJob, len(db.SecondaryPieceJobTable[objectID]))
	for idx, job := range db.SecondaryPieceJobTable[objectID] {
		pieces[idx] = &job
	}
	return pieces, nil
}

// SetPrimaryPieceJobDoneV2 set one primary piece job is completed.
func (db *MemJobDB) SetPrimaryPieceJobDoneV2(objectID uint64, piece *spdb.PieceJob) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, ok := db.PrimaryPieceJobTable[objectID]; !ok {
		db.PrimaryPieceJobTable[objectID] = make(map[uint32]spdb.PieceJob)
	}
	db.PrimaryPieceJobTable[objectID][piece.PieceId] = *piece
	return nil
}

// SetSecondaryPieceJobDoneV2 set one secondary piece job is completed.
func (db *MemJobDB) SetSecondaryPieceJobDoneV2(objectID uint64, piece *spdb.PieceJob) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, ok := db.SecondaryPieceJobTable[objectID]; !ok {
		db.SecondaryPieceJobTable[objectID] = make(map[uint32]spdb.PieceJob)
	}
	db.SecondaryPieceJobTable[objectID][piece.PieceId] = *piece
	return nil
}

// DeleteJobV2 delete job by id, delete the related object and piece jobs.
func (db *MemJobDB) DeleteJobV2(jobId uint64) error {
	delete(db.JobTable, jobId)
	objectID, ok := db.JobToObject[jobId]
	if !ok {
		return nil
	}
	delete(db.ObjectTable, objectID)
	delete(db.PrimaryPieceJobTable, objectID)
	delete(db.SecondaryPieceJobTable, objectID)
	return nil
}

// NewIterator creates an iterator over a subset,
// starting at a particular initial key.
func (db *MemJobDB) NewIterator(start interface{}) spdb.Iterator {
	db.mu.RLock()
	defer db.mu.RUnlock()
	minObjectID := start.(uint64)
	var keys []uint64
	for objectId, _ := range db.JobTable {
		if objectId >= minObjectID {
			keys = append(keys, objectId)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	var values []*ptypes.JobContext
	for _, objectId := range keys {
		values = append(values, db.JobTable[objectId])
	}
	return &iterator{
		keys:   keys,
		values: values,
	}
}

// NewBatch creates a write-only database that buffers changes to its host db
// until a final write is called.
func (db *MemJobDB) NewBatch() spdb.Batch {
	return &batch{
		db: db,
	}
}

// NewBatchWithSize creates a write-only database batch with pre-allocated buffer.
func (db *MemJobDB) NewBatchWithSize(size int) spdb.Batch {
	return &batch{
		db: db,
	}
}

// iterator can walk over the (potentially partial) keyspace of a memory key
// value store. Internally it is a deep copy of the entire iterated state,
// sorted by keys.
type iterator struct {
	inited bool
	keys   []uint64
	values []*ptypes.JobContext
}

// Next moves the iterator to the next key/value pair. It returns whether the
// iterator is exhausted.
func (it *iterator) Next() bool {
	if !it.inited {
		it.inited = true
		return len(it.keys) > 0
	}
	// Iterator already initialize, advance it
	if len(it.keys) > 0 {
		it.keys = it.keys[1:]
		it.values = it.values[1:]
	}
	return len(it.keys) > 0
}

// Error returns any accumulated error. Exhausting all the key/value pairs
// is not considered to be an error.
func (it *iterator) Error() error {
	return nil
}

// Key returns the key of the current key/value pair, or nil if done. The caller
// should not modify the contents of the returned slice, and its contents may
// change on the next call to Next.
func (it *iterator) Key() interface{} {
	if len(it.keys) > 0 {
		return it.keys[0]
	}
	return nil
}

// Value returns the value of the current key/value pair, or nil if done. The
// caller should not modify the contents of the returned slice, and its contents
// may change on the next call to Next.
func (it *iterator) Value() interface{} {
	if len(it.values) > 0 {
		return it.values[0]
	}
	return nil
}

// Release releases associated resources. Release should always succeed and can
// be called multiple times without causing error.
func (it *iterator) Release() {
	it.keys, it.values = nil, nil
}

// batch is a write-only memory batch that commits changes to its host
// database when Write is called. A batch cannot be used concurrently.
type batch struct {
	db   *MemJobDB
	keys []uint64
	size int
}

// Put inserts the given value into the key-value data store.
func (b *batch) Put(key interface{}, value interface{}) error {
	return errors.New("job db batch not support put")
}

// Delete removes the key from the key-value data store.
func (b *batch) Delete(key interface{}) error {
	jobID := key.(uint64)
	b.keys = append(b.keys, jobID)
	b.size += 8
	return nil
}

// ValueSize retrieves the amount of data queued up for writing.
func (b *batch) ValueSize() int {
	return b.size
}

// Write flushes any accumulated data to disk.
func (b *batch) Write() error {
	for _, jobID := range b.keys {
		b.db.DeleteJobV2(jobID)
	}
	return nil
}

// Reset resets the batch for reuse.
func (b *batch) Reset() {
	b.keys = b.keys[:0]
	b.size = 0
}
