package jobmemory

import (
	"errors"
	"sync"

	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/store/jobdb"
)

// MemJobDB is a memory db, maintains job, object and piece job table.
type MemJobDB struct {
	JobCount               uint64
	JobTable               map[uint64]types.JobContext
	ObjectTable            map[string]types.ObjectInfo
	PrimaryPieceJobTable   map[string]map[uint32]jobdb.PieceJob
	SecondaryPieceJobTable map[string]map[uint32]jobdb.PieceJob
	mu                     sync.RWMutex
}

// NewMemJobDB return a MemJobDB instance.
func NewMemJobDB() *MemJobDB {
	return &MemJobDB{
		JobCount:               0,
		JobTable:               make(map[uint64]types.JobContext),
		ObjectTable:            make(map[string]types.ObjectInfo),
		PrimaryPieceJobTable:   make(map[string]map[uint32]jobdb.PieceJob),
		SecondaryPieceJobTable: make(map[string]map[uint32]jobdb.PieceJob),
	}
}

// CreateUploadPayloadJob create a job info for special object.
func (db *MemJobDB) CreateUploadPayloadJob(txHash []byte, info *types.ObjectInfo) (uint64, error) {
	if info == nil {
		return 0, errors.New("object info is nil")
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	db.JobTable[db.JobCount] = types.JobContext{
		JobId:    db.JobCount,
		JobState: types.JobState_JOB_STATE_CREATE_OBJECT_DONE,
	}
	info.JobId = db.JobCount
	db.ObjectTable[string(txHash)] = *info
	db.JobCount++
	return db.JobCount - 1, nil
}

// SetObjectCreateHeightAndObjectID set the object create height and object resource id after successful chaining.
func (db *MemJobDB) SetObjectCreateHeightAndObjectID(txHash []byte, height uint64, objectID uint64) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	objectInfo, ok := db.ObjectTable[string(txHash)]
	if !ok {
		return errors.New("object is not exist")
	}
	objectInfo.ObjectId = objectID
	objectInfo.TxHash = txHash
	objectInfo.Height = height
	db.ObjectTable[string(txHash)] = objectInfo
	return nil
}

// GetObjectInfo returns the object info by create object transaction hash.
func (db *MemJobDB) GetObjectInfo(txHash []byte) (*types.ObjectInfo, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	objectInfo, ok := db.ObjectTable[string(txHash)]
	if !ok {
		return nil, errors.New("object is not exist")
	}
	return &objectInfo, nil
}

// GetJobContext returns the job info .
func (db *MemJobDB) GetJobContext(jobId uint64) (*types.JobContext, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	job, ok := db.JobTable[jobId]
	if !ok {
		return nil, errors.New("job is not exist")
	}
	return &job, nil
}

// SetUploadPayloadJobState set the job state.
func (db *MemJobDB) SetUploadPayloadJobState(jobId uint64, state string, timestamp int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	job, ok := db.JobTable[jobId]
	if !ok {
		return errors.New("job is not exist")
	}
	jobState, ok := types.JobState_value[state]
	if !ok {
		return errors.New("state is not correct job state")
	}
	job.JobState = (types.JobState)(jobState)
	job.ModifyTime = timestamp
	db.JobTable[jobId] = job
	return nil
}

// SetUploadPayloadJobJobError set the job error state and error message.
func (db *MemJobDB) SetUploadPayloadJobJobError(jobId uint64, state string, jobErr string, timestamp int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	job, ok := db.JobTable[jobId]
	if !ok {
		return errors.New("job is not exist")
	}
	jobState, ok := types.JobState_value[state]
	if !ok {
		return errors.New("state is not correct job state")
	}
	job.JobState = (types.JobState)(jobState)
	job.ModifyTime = timestamp
	job.JobErr = jobErr
	db.JobTable[jobId] = job
	return nil
}

// GetPrimaryJob returns the primary piece jobs by create object transaction hash.
func (db *MemJobDB) GetPrimaryJob(txHash []byte) ([]*jobdb.PieceJob, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if _, ok := db.PrimaryPieceJobTable[string(txHash)]; !ok {
		return []*jobdb.PieceJob{}, nil
	}
	pieces := make([]*jobdb.PieceJob, len(db.PrimaryPieceJobTable[string(txHash)]))
	for idx, job := range db.PrimaryPieceJobTable[string(txHash)] {
		pieces[idx] = &job
	}
	return pieces, nil
}

// GetSecondaryJob returns the secondary piece jobs by create object transaction hash.
func (db *MemJobDB) GetSecondaryJob(txHash []byte) ([]*jobdb.PieceJob, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if _, ok := db.SecondaryPieceJobTable[string(txHash)]; !ok {
		return []*jobdb.PieceJob{}, nil
	}
	pieces := make([]*jobdb.PieceJob, len(db.SecondaryPieceJobTable[string(txHash)]))
	for idx, job := range db.SecondaryPieceJobTable[string(txHash)] {
		pieces[idx] = &job
	}
	return pieces, nil
}

// SetPrimaryPieceJobDone set one primary piece job is completed.
func (db *MemJobDB) SetPrimaryPieceJobDone(txHash []byte, piece *jobdb.PieceJob) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, ok := db.PrimaryPieceJobTable[string(txHash)]; !ok {
		db.PrimaryPieceJobTable[string(txHash)] = make(map[uint32]jobdb.PieceJob)
	}
	db.PrimaryPieceJobTable[string(txHash)][piece.PieceId] = *piece
	return nil
}

// SetSecondaryPieceJobDone set one secondary piece job is completed.
func (db *MemJobDB) SetSecondaryPieceJobDone(txHash []byte, piece *jobdb.PieceJob) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, ok := db.SecondaryPieceJobTable[string(txHash)]; !ok {
		db.SecondaryPieceJobTable[string(txHash)] = make(map[uint32]jobdb.PieceJob)
	}
	db.SecondaryPieceJobTable[string(txHash)][piece.PieceId] = *piece
	return nil
}
