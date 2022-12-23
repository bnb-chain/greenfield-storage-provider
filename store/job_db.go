package store

import (
	"github.com/bnb-chain/inscription-storage-provider/pkg/job"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
)

type JobDB interface {
	CreateUploadPayloadJob(txHash []byte, objectInfo *types.ObjectInfo) (uint64, error)
	SetObjectCreateHeight(txHash []byte, height uint64) error

	GetJobContext(txHash []byte) (*types.JobContext, error)

	SetUploadPayloadJobState(jobID uint64, jobState string) error
	SetUploadPayloadJobJobError(jobID uint64, jobState string, jobErr string) error

	GetUploadPrimaryJob(jobID uint64) (*job.UploadSubJob, error)
	GetUploadSecondaryJob(jobID uint64) (*job.UploadSubJob, error)

	SetUploadPrimaryPieceJobState(jobID uint64, pieceJob *service.PieceJob, state string) error
	SetUploadSecondaryJobState(jobID uint64, pieceJob *service.PieceJob, state string) error
}
