package stone

import (
	"errors"
	"sync"
	"time"

	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb"
	"github.com/bnb-chain/greenfield-storage-provider/store/metadb"
)

// JobContextWrapper maintain job context, goroutine safe
type JobContextWrapper struct {
	jobCtx *ptypes.JobContext
	jobErr error
	jobDB  jobdb.JobDBV2
	metaDB metadb.MetaDB
	mu     sync.RWMutex
}

// NewJobContextWrapper return the instance of JobContextWrapper
func NewJobContextWrapper(jobCtx *ptypes.JobContext, jobDB jobdb.JobDBV2, metaDB metadb.MetaDB) *JobContextWrapper {
	return &JobContextWrapper{
		jobCtx: jobCtx,
		jobDB:  jobDB,
		metaDB: metaDB,
	}
}

// GetJobState return the state of job
func (wrapper *JobContextWrapper) GetJobState() (string, error) {
	wrapper.mu.RLock()
	defer wrapper.mu.RUnlock()
	state, ok := ptypes.JobState_name[int32(wrapper.jobCtx.GetJobState())]
	if !ok {
		return "", errors.New("job state error")
	}
	return state, nil
}

// SetJobState set the state of job, if access DB error will return
func (wrapper *JobContextWrapper) SetJobState(state string) error {
	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()
	wrapper.jobCtx.JobState = ptypes.JobState(ptypes.JobState_value[state])
	return wrapper.jobDB.SetUploadPayloadJobStateV2(wrapper.jobCtx.JobId, state, time.Now().Unix())
}

// JobErr return job error
func (wrapper *JobContextWrapper) JobErr() error {
	wrapper.mu.RLock()
	defer wrapper.mu.RUnlock()
	return wrapper.jobErr
}

// SetJobErr set the error of job and store the db
func (wrapper *JobContextWrapper) SetJobErr(err error) error {
	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()
	wrapper.jobErr = err
	if err == nil {
		wrapper.jobCtx.JobErr = ""
	} else {
		wrapper.jobCtx.JobErr = wrapper.jobCtx.JobErr + err.Error()
	}
	wrapper.jobCtx.JobState = ptypes.JobState_JOB_STATE_ERROR
	return wrapper.jobDB.SetUploadPayloadJobJobErrorV2(wrapper.jobCtx.JobId,
		ptypes.JOB_STATE_ERROR, wrapper.jobCtx.JobErr, time.Now().Unix())
}

// ModifyTime return the last modify timestamp
func (wrapper *JobContextWrapper) ModifyTime() int64 {
	wrapper.mu.RLock()
	defer wrapper.mu.RUnlock()
	return wrapper.jobCtx.ModifyTime
}

// JobContext return the copy of job context
func (wrapper *JobContextWrapper) JobContext() *ptypes.JobContext {
	wrapper.mu.RLock()
	defer wrapper.mu.RUnlock()
	return wrapper.jobCtx.SafeCopy()
}
