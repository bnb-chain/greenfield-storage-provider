package job

import (
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	types "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
)

// UploadPayloadJob maintains the object info and piece job meta
type UploadPayloadJob struct {
	objectCtx    *ObjectInfoContext
	primaryJob   *UploadSpJob // the job of uploading primary storage provider
	secondaryJob *UploadSpJob // the job of uploading secondary storage provider
}

// NewUploadPayloadJob return the instance of UploadPayloadJob.
func NewUploadPayloadJob(objectCtx *ObjectInfoContext) (job *UploadPayloadJob, err error) {
	job = &UploadPayloadJob{
		objectCtx: objectCtx,
	}
	if job.primaryJob, err = NewSegmentUploadSpJob(objectCtx, false); err != nil {
		return nil, err
	}
	switch objectCtx.GetObjectRedundancyType() {
	case types.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED:
		if job.secondaryJob, err = NewECUploadSpJob(objectCtx, true); err != nil {
			return nil, err
		}
	case types.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
		if job.secondaryJob, err = NewSegmentUploadSpJob(objectCtx, true); err != nil {
			return nil, err
		}
	case types.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE:
		if job.secondaryJob, err = NewSegmentUploadSpJob(objectCtx, true); err != nil {
			return nil, err
		}
	default:
		return nil, merrors.ErrRedundancyType
	}
	return job, nil
}

// GetObjectCtx return the object context.
func (job *UploadPayloadJob) GetObjectCtx() *ObjectInfoContext {
	return job.objectCtx
}

// Completed return whether upload payload job is completed.
func (job *UploadPayloadJob) Completed() bool {
	return job.primaryJob.Completed() && job.secondaryJob.Completed()
}

// PrimarySPCompleted return whether upload primary storage provider job is completed.
func (job *UploadPayloadJob) PrimarySPCompleted() bool {
	return job.primaryJob.Completed()
}

// SecondarySPCompleted return whether upload secondary storage provider job is completed.
func (job *UploadPayloadJob) SecondarySPCompleted() bool {
	return job.secondaryJob.Completed()
}

// DonePrimarySPJob complete one primary piece job and update DB.
func (job *UploadPayloadJob) DonePrimarySPJob(pieceJob *service.PieceJob) error {
	return job.primaryJob.Done(pieceJob)
}

// DoneSecondarySPJob complete one secondary piece job and update DB.
func (job *UploadPayloadJob) DoneSecondarySPJob(pieceJob *service.PieceJob) error {
	return job.secondaryJob.Done(pieceJob)
}

// PopPendingPrimarySPJob return the uncompleted primary piece job.
func (job *UploadPayloadJob) PopPendingPrimarySPJob() *service.PieceJob {
	pieces := job.primaryJob.PopPendingJob()
	if len(pieces) == 0 {
		return nil
	}
	obj := job.objectCtx.GetObjectInfo()
	pieceJob := &service.PieceJob{
		ObjectId:       obj.ObjectId,
		PayloadSize:    obj.Size,
		TargetIdx:      pieces,
		RedundancyType: obj.RedundancyType,
	}
	return pieceJob
}

// PopPendingSecondarySPJob return the uncompleted secondary piece job.
func (job *UploadPayloadJob) PopPendingSecondarySPJob() *service.PieceJob {
	pieces := job.secondaryJob.PopPendingJob()
	if len(pieces) == 0 {
		return nil
	}
	obj := job.objectCtx.GetObjectInfo()
	pieceJob := &service.PieceJob{
		ObjectId:       obj.ObjectId,
		PayloadSize:    obj.Size,
		TargetIdx:      pieces,
		RedundancyType: obj.RedundancyType,
	}
	return pieceJob
}

// PrimarySPSealInfo return the primary storage provider seal info.
func (job *UploadPayloadJob) PrimarySPSealInfo() ([]*types.StorageProviderInfo, error) {
	return job.primaryJob.SealInfo()
}

// SecondarySPSealInfo  return the secondary storage provider seal info.
func (job *UploadPayloadJob) SecondarySPSealInfo() ([]*types.StorageProviderInfo, error) {
	return job.secondaryJob.SealInfo()
}
