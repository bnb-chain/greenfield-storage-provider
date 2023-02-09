package jobmemory

import (
	"errors"
	"sync"

	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb"
)

var _ jobdb.JobDBV2 = &MemJobDB{}

// MemJobDB is a memory db, maintains job, object and piece job table.
type MemJobDB struct {
	JobCount               uint64
	JobTable               map[uint64]*ptypes.JobContext
	ObjectTable            map[uint64]*ptypes.ObjectInfo
	PrimaryPieceJobTable   map[uint64]map[uint32]jobdb.PieceJob
	SecondaryPieceJobTable map[uint64]map[uint32]jobdb.PieceJob
	mu                     sync.RWMutex
}

// NewMemJobDB return a MemJobDBV2 instance.
func NewMemJobDB() *MemJobDB {
	return &MemJobDB{
		JobCount:               0,
		JobTable:               make(map[uint64]*ptypes.JobContext),
		ObjectTable:            make(map[uint64]*ptypes.ObjectInfo),
		PrimaryPieceJobTable:   make(map[uint64]map[uint32]jobdb.PieceJob),
		SecondaryPieceJobTable: make(map[uint64]map[uint32]jobdb.PieceJob),
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

// GetPrimaryJobV2 returns the primary piece jobs by by object id.
func (db *MemJobDB) GetPrimaryJobV2(objectID uint64) ([]*jobdb.PieceJob, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if _, ok := db.PrimaryPieceJobTable[objectID]; !ok {
		return []*jobdb.PieceJob{}, nil
	}
	pieces := make([]*jobdb.PieceJob, len(db.PrimaryPieceJobTable[objectID]))
	for idx, job := range db.PrimaryPieceJobTable[objectID] {
		pieces[idx] = &job
	}
	return pieces, nil
}

// GetSecondaryJobV2 returns the secondary piece jobs by object id.
func (db *MemJobDB) GetSecondaryJobV2(objectID uint64) ([]*jobdb.PieceJob, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if _, ok := db.SecondaryPieceJobTable[objectID]; !ok {
		return []*jobdb.PieceJob{}, nil
	}
	pieces := make([]*jobdb.PieceJob, len(db.SecondaryPieceJobTable[objectID]))
	for idx, job := range db.SecondaryPieceJobTable[objectID] {
		pieces[idx] = &job
	}
	return pieces, nil
}

// SetPrimaryPieceJobDoneV2 set one primary piece job is completed.
func (db *MemJobDB) SetPrimaryPieceJobDoneV2(objectID uint64, piece *jobdb.PieceJob) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, ok := db.PrimaryPieceJobTable[objectID]; !ok {
		db.PrimaryPieceJobTable[objectID] = make(map[uint32]jobdb.PieceJob)
	}
	db.PrimaryPieceJobTable[objectID][piece.PieceId] = *piece
	return nil
}

// SetSecondaryPieceJobDoneV2 set one secondary piece job is completed.
func (db *MemJobDB) SetSecondaryPieceJobDoneV2(objectID uint64, piece *jobdb.PieceJob) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, ok := db.SecondaryPieceJobTable[objectID]; !ok {
		db.SecondaryPieceJobTable[objectID] = make(map[uint32]jobdb.PieceJob)
	}
	db.SecondaryPieceJobTable[objectID][piece.PieceId] = *piece
	return nil
}
