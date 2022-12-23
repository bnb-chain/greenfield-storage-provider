package stone

import (
	"errors"
	"github.com/bnb-chain/inscription-storage-provider/pkg/job"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/store"
	"sync"
)

type JobContextWrapper struct {
	jobCtx *types.JobContext
	jobErr error
	jobDB  store.JobDB
	metaDB store.MetaDB
	mu     sync.RWMutex
}

func NewJobContextWrapper(jobCtx *types.JobContext, jobDB store.JobDB, metaDB store.MetaDB) *JobContextWrapper {
	return &JobContextWrapper{
		jobCtx: jobCtx,
		jobDB:  jobDB,
		metaDB: metaDB,
	}
}

func (wrapper *JobContextWrapper) GetJobState() (string, error) {
	wrapper.mu.RLock()
	defer wrapper.mu.RUnlock()
	state, ok := types.JobState_name[int32(wrapper.jobCtx.JobState)]
	if !ok {
		// return error
		return "", errors.New("")
	}
	return state, nil
}

func (wrapper *JobContextWrapper) SetJobState(state string) error {
	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()
	return wrapper.jobDB.SetUploadPayloadJobState(wrapper.jobCtx.JobId, state)
}

func (wrapper *JobContextWrapper) JobErr() error {
	wrapper.mu.RLock()
	defer wrapper.mu.RUnlock()
	return wrapper.jobErr
}

func (wrapper *JobContextWrapper) SetJobErr(err error) error {
	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()
	wrapper.jobErr = err
	wrapper.jobCtx.JobErr = wrapper.jobCtx.JobErr + err.Error()
	if err := wrapper.jobDB.SetUploadPayloadJobJobError(wrapper.jobCtx.JobId,
		types.JOB_STATE_ERROR, wrapper.jobCtx.JobErr); err != nil {
		// log error
	}
	return err
}

func (wrapper *JobContextWrapper) GetUploadPrimaryJob() (*job.UploadSubJob, error) {
	wrapper.mu.RLock()
	jobID := wrapper.jobCtx.JobId
	wrapper.mu.RUnlock()
	return wrapper.jobDB.GetUploadPrimaryJob(jobID)
}

func (wrapper *JobContextWrapper) GetUploadSecondaryJob() (*job.UploadSubJob, error) {
	wrapper.mu.RLock()
	jobID := wrapper.jobCtx.JobId
	wrapper.mu.RUnlock()
	return wrapper.jobDB.GetUploadSecondaryJob(jobID)
}

func (wrapper *JobContextWrapper) SetPrimaryPieceJobState(pieceJob *service.PieceJob, state string) error {
	wrapper.mu.RLock()
	jobID := wrapper.jobCtx.JobId
	wrapper.mu.RUnlock()
	return wrapper.jobDB.SetUploadPrimaryPieceJobState(jobID, pieceJob, state)
}

func (wrapper *JobContextWrapper) SetSecondaryJobState(pieceJob *service.PieceJob, state string) error {
	wrapper.mu.RLock()
	for _, spSealInfo := range pieceJob.StorageProviderSealInfo {
		for i, idx := range spSealInfo.PieceIdx {
			sp := &types.StorageProviderInfo{
				SpId:      spSealInfo.StorageProviderId,
				Idx:       idx,
				Checksum:  spSealInfo.CheckSum[i],
				Signature: spSealInfo.Signature,
			}
			wrapper.jobCtx.ObjectInfo.SecondarySps[spSealInfo.StorageProviderId] = sp
		}
	}
	jobID := wrapper.jobCtx.JobId
	bucketName := wrapper.jobCtx.ObjectInfo.BucketName
	objectName := wrapper.jobCtx.ObjectInfo.ObjectName
	wrapper.mu.RUnlock()
	if err := wrapper.metaDB.SetIntegrityHash(bucketName, objectName, pieceJob); err != nil {
		return err
	}
	if err := wrapper.jobDB.SetUploadSecondaryJobState(jobID, pieceJob, state); err != nil {
		return err
	}
	return nil
}
