package store

import (
	"github.com/bnb-chain/inscription-storage-provider/pkg/job"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
)

type JobDB interface {
	GetJobContext(jobID uint64) (*types.JobContext, error)
	SetUploadJobState(jobID uint64, jobState string) error
	SetJobError(jobID uint64, jobState string, jobErr string) error

	GetPrimaryJob(jobID uint64) (job.UploadPrimaryJob, error)
	SetPrimaryJobState(jobID uint64, jobState string) error
	SetPrimaryPieceJobState(jobID uint64, piece string, jobState string) error
	
	SetSecondaryJob(jobID uint64, secondaryJob *job.UploadSecondaryJob) error
}
