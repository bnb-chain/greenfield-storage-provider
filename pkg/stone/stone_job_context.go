package stone

import (
	"errors"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/store/jobdb"
	"sync"
	"time"
)

// JobContextWrapper maintain job context, goroutine safe
type JobContextWrapper struct {
	jobCtx *types.JobContext
	jobErr error
	jobDB  jobdb.JobDB
	mu     sync.RWMutex
}

// NewJobContextWrapper return the instance of JobContextWrapper
func NewJobContextWrapper(jobCtx *types.JobContext, jobDB jobdb.JobDB) *JobContextWrapper {
	return &JobContextWrapper{
		jobCtx: jobCtx,
		jobDB:  jobDB,
	}
}

// GetJobState return the state of job
func (wrapper *JobContextWrapper) GetJobState() (string, error) {
	wrapper.mu.RLock()
	defer wrapper.mu.RUnlock()
	state, ok := types.JobState_name[int32(wrapper.jobCtx.JobState)]
	if !ok {
		return "", errors.New("job state error")
	}
	return state, nil
}

// SetJobState set the state of job, if access DB error will return
func (wrapper *JobContextWrapper) SetJobState(state string) error {
	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()
	return wrapper.jobDB.SetUploadPayloadJobState(wrapper.jobCtx.JobId, state, time.Now().Unix())
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
	wrapper.jobCtx.JobErr = wrapper.jobCtx.JobErr + err.Error()
	if err := wrapper.jobDB.SetUploadPayloadJobJobError(wrapper.jobCtx.JobId,
		types.JOB_STATE_ERROR, wrapper.jobCtx.JobErr, time.Now().Unix()); err != nil {
		return err
	}
	return err
}

// ModifyTime return the last modify timestamp
func (wrapper *JobContextWrapper) ModifyTime() int64 {
	wrapper.mu.RLock()
	defer wrapper.mu.RUnlock()
	return wrapper.jobCtx.ModifyTime
}
