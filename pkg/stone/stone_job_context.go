package stone

import (
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/store"
	"sync"
)

type StoneJob interface{}

type JobContextWrapper struct {
	jobCtx *types.JobContext
	jobCh  chan StoneJob
	jobDB  store.JobDB
	metaDB store.MetaDB
	gcCh   chan uint64
	jobErr error
	mu     sync.RWMutex
}

func NewJobContextWrapper(jobCtx *types.JobContext, jobCh chan StoneJob, jobDB store.JobDB, metaDB store.MetaDB, gcCh chan uint64) *JobContextWrapper {
	return &JobContextWrapper{
		jobCtx: jobCtx,
		jobCh:  jobCh,
		jobDB:  jobDB,
		metaDB: metaDB,
		gcCh:   gcCh,
	}
}

func (wrapper *JobContextWrapper) SetJobContext(jobCtx *types.JobContext) {
	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()
	wrapper.jobCtx = jobCtx
}

func (wrapper *JobContextWrapper) GetObjectInfo() *types.ObjectInfo {
	wrapper.mu.RLock()
	defer wrapper.mu.RUnlock()
	return wrapper.jobCtx.ObjectInfo
}

func (wrapper *JobContextWrapper) GetBucketName() string {
	wrapper.mu.RLock()
	defer wrapper.mu.RUnlock()
	return wrapper.jobCtx.ObjectInfo.BucketName
}

func (wrapper *JobContextWrapper) GetObjectName() string {
	wrapper.mu.RLock()
	defer wrapper.mu.RUnlock()
	return wrapper.jobCtx.ObjectInfo.ObjectName
}

func (wrapper *JobContextWrapper) GetJobId() uint64 {
	wrapper.mu.RLock()
	defer wrapper.mu.RUnlock()
	return wrapper.jobCtx.GetJobId()
}

func (wrapper *JobContextWrapper) GetJobState() string {
	wrapper.mu.RLock()
	defer wrapper.mu.RUnlock()
	return types.JobState_name[int32(wrapper.jobCtx.GetJobState())]
}

func (wrapper *JobContextWrapper) SetJobState(state string) error {
	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()
	jobState, ok := types.JobState_value[state]
	if !ok {
		return nil
	}
	wrapper.jobCtx.JobState = types.JobState(jobState)
	return nil
}

func (wrapper *JobContextWrapper) GetJobDB() store.JobDB {
	return wrapper.jobDB
}

func (wrapper *JobContextWrapper) GetMetaDB() store.MetaDB {
	return wrapper.metaDB
}

func (wrapper *JobContextWrapper) SetIntegrityHash(primary types.StorageProviderInfo, secondary []types.StorageProviderInfo) error {
	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()
	return wrapper.jobCtx.ObjectInfo.SetIntegrityHash(primary, secondary)
}

func (wrapper *JobContextWrapper) SendJob(job StoneJob) {
	wrapper.jobCh <- job
}

func (wrapper *JobContextWrapper) InterruptJob(jobErr error) error {
	wrapper.mu.Lock()
	wrapper.jobCtx.JobState = types.JobState(types.JobState_value[types.JOB_STATE_ERROR])
	wrapper.jobErr = jobErr
	wrapper.jobCtx.JobErr = wrapper.jobCtx.JobErr + jobErr.Error() + "\n"
	wrapper.mu.Unlock()
	wrapper.jobDB.SetJobError(wrapper.jobCtx.JobId, types.JOB_STATE_ERROR, wrapper.jobCtx.JobErr)
	wrapper.gcCh <- wrapper.jobCtx.GetJobId()
	return nil
}

func (wrapper *JobContextWrapper) JobStateException() bool {
	return types.JobState_name[int32(wrapper.jobCtx.GetJobState())] == types.JOB_STATE_ERROR
}

func (wrapper *JobContextWrapper) JobError() error {
	wrapper.mu.RLock()
	defer wrapper.mu.RUnlock()
	return wrapper.jobErr
}
