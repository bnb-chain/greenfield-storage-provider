package jobdb

import types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"

type PieceJob struct {
	PieceId         uint32
	CheckSum        []byte
	StorageProvider string
}

type JobDB interface {
	CreateUploadPayloadJob(txHash []byte, info *types.ObjectInfo) error
	SetObjectCreateHeight(txHash []byte, height uint64) error

	GetObjectInfo(txHash []byte) (*types.ObjectInfo, error)
	GetJobContext(jobId uint64) (*types.JobContext, error)

	SetUploadPayloadJobState(jobId uint64, state string, timestamp int64) error
	SetUploadPayloadJobJobError(jobID uint64, jobState string, jobErr string, timestamp int64) error

	GetPrimaryJob(txHash []byte) ([]*PieceJob, error)
	GetSecondaryJob(txHash []byte) ([]*PieceJob, error)
	SetPrimaryPieceJobDone(*PieceJob) error
	SetSecondaryPieceJobDone(*PieceJob) error
}
