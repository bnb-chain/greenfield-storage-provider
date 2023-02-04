package job

import (
	"testing"
	"time"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	stypesv1pb "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/stretchr/testify/assert"

	ptypesv1pb "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb/jobmemory"
)

func InitEnv(rType ptypesv1pb.RedundancyType) (*UploadPayloadJob, *ptypesv1pb.ObjectInfo) {
	objectSize := 50 * 1024 * 1024
	switch rType {
	case ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE:
		objectSize = 1 * 1024 * 1024
	}
	object := &ptypesv1pb.ObjectInfo{
		Size:           uint64(objectSize),
		ObjectId:       1,
		RedundancyType: rType,
	}
	job, _ := NewUploadPayloadJob(NewObjectInfoContext(object, jobmemory.NewMemJobDBV2(), nil))
	return job, object
}

func TestInitUploadPayloadJob(t *testing.T) {
	job, _ := InitEnv(ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED)
	assert.Equal(t, len(job.primaryJob.PopPendingJob()), 4)
	assert.Equal(t, len(job.secondaryJob.PopPendingJob()), 6)

	job, _ = InitEnv(ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE)
	assert.Equal(t, len(job.primaryJob.PopPendingJob()), 1)
	assert.Equal(t, len(job.secondaryJob.PopPendingJob()), 6)

	job, _ = InitEnv(ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE)
	assert.Equal(t, len(job.primaryJob.PopPendingJob()), 4)
	assert.Equal(t, len(job.secondaryJob.PopPendingJob()), 6)
}

func TestDoneReplicatePieceJob(t *testing.T) {
	job, object := InitEnv(ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE)
	pieceJob := &stypesv1pb.PieceJob{
		ObjectId:       object.GetObjectId(),
		PayloadSize:    object.GetSize(),
		RedundancyType: object.GetRedundancyType(),
	}
	pieceJob.StorageProviderSealInfo = &stypesv1pb.StorageProviderSealInfo{
		StorageProviderId: "test-storage-provider",
		PieceChecksum:     [][]byte{hash.GenerateChecksum([]byte(time.Now().String()))},
		IntegrityHash:     hash.GenerateChecksum([]byte(time.Now().String())),
		Signature:         hash.GenerateChecksum([]byte(time.Now().String())),
	}
	pieceJob.StorageProviderSealInfo.PieceIdx = 0
	job.DonePrimarySPJob(pieceJob)
	assert.Equal(t, 3, len(job.primaryJob.PopPendingJob()))
	pieceJob.StorageProviderSealInfo.PieceIdx = 3
	job.DonePrimarySPJob(pieceJob)
	assert.Equal(t, 2, len(job.primaryJob.PopPendingJob()))
	pieceJob.StorageProviderSealInfo.PieceIdx = 2
	job.DonePrimarySPJob(pieceJob)
	assert.Equal(t, 1, len(job.primaryJob.PopPendingJob()))
	pieceJob.StorageProviderSealInfo.PieceIdx = 1
	job.DonePrimarySPJob(pieceJob)
	assert.Equal(t, true, job.primaryJob.Completed())

	pieceJob.StorageProviderSealInfo.PieceChecksum = append(pieceJob.StorageProviderSealInfo.PieceChecksum,
		hash.GenerateChecksum([]byte(time.Now().String())))
	pieceJob.StorageProviderSealInfo.PieceChecksum = append(pieceJob.StorageProviderSealInfo.PieceChecksum,
		hash.GenerateChecksum([]byte(time.Now().String())))
	pieceJob.StorageProviderSealInfo.PieceChecksum = append(pieceJob.StorageProviderSealInfo.PieceChecksum,
		hash.GenerateChecksum([]byte(time.Now().String())))
	assert.Equal(t, 6, len(job.secondaryJob.PopPendingJob()))
	pieceJob.StorageProviderSealInfo.PieceIdx = 0
	job.DoneSecondarySPJob(pieceJob)
	assert.Equal(t, 5, len(job.secondaryJob.PopPendingJob()))
	pieceJob.StorageProviderSealInfo.PieceIdx = 3
	job.DoneSecondarySPJob(pieceJob)
	assert.Equal(t, 4, len(job.secondaryJob.PopPendingJob()))
	pieceJob.StorageProviderSealInfo.PieceIdx = 2
	job.DoneSecondarySPJob(pieceJob)
	assert.Equal(t, 3, len(job.secondaryJob.PopPendingJob()))
	pieceJob.StorageProviderSealInfo.PieceIdx = 1
	job.DoneSecondarySPJob(pieceJob)
	assert.Equal(t, 2, len(job.secondaryJob.PopPendingJob()))
	pieceJob.StorageProviderSealInfo.PieceIdx = 4
	job.DoneSecondarySPJob(pieceJob)
	assert.Equal(t, 1, len(job.secondaryJob.PopPendingJob()))
	pieceJob.StorageProviderSealInfo.PieceIdx = 5
	job.DoneSecondarySPJob(pieceJob)
	//assert.Equal(t, true, job.secondaryJob.Completed())
}

func TestDoneInlinePieceJob(t *testing.T) {
	job, object := InitEnv(ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_INLINE_TYPE)
	pieceJob := &stypesv1pb.PieceJob{
		ObjectId:       object.GetObjectId(),
		PayloadSize:    object.GetSize(),
		RedundancyType: object.GetRedundancyType(),
	}
	intergrity := hash.GenerateChecksum([]byte(time.Now().String()))
	pieceCheckSum := hash.GenerateChecksum([]byte(time.Now().String()))
	signature := hash.GenerateChecksum([]byte(time.Now().String()))
	pieceJob.StorageProviderSealInfo = &stypesv1pb.StorageProviderSealInfo{
		StorageProviderId: "test-storage-provider",
		PieceChecksum:     [][]byte{pieceCheckSum},
		IntegrityHash:     intergrity,
		Signature:         signature,
	}
	pieceJob.StorageProviderSealInfo.PieceIdx = 0
	job.DonePrimarySPJob(pieceJob)
	assert.Equal(t, 0, len(job.primaryJob.PopPendingJob()))
	pieceJob.StorageProviderSealInfo.PieceIdx = 0
	job.DoneSecondarySPJob(pieceJob)
	assert.Equal(t, 5, len(job.secondaryJob.PopPendingJob()))
}

func TestDoneECPieceJob(t *testing.T) {
	job, object := InitEnv(ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED)
	pieceJob := &stypesv1pb.PieceJob{
		ObjectId:       object.GetObjectId(),
		PayloadSize:    object.GetSize(),
		RedundancyType: object.GetRedundancyType(),
	}
	pieceJob.StorageProviderSealInfo = &stypesv1pb.StorageProviderSealInfo{
		StorageProviderId: "test-storage-provider",
		PieceChecksum: [][]byte{hash.GenerateChecksum([]byte(time.Now().String())),
			hash.GenerateChecksum([]byte(time.Now().String())),
			hash.GenerateChecksum([]byte(time.Now().String())),
			hash.GenerateChecksum([]byte(time.Now().String()))},
		IntegrityHash: hash.GenerateChecksum([]byte(time.Now().String())),
		Signature:     hash.GenerateChecksum([]byte(time.Now().String())),
	}
	pieceJob.StorageProviderSealInfo.PieceIdx = 0
	job.DoneSecondarySPJob(pieceJob)
	assert.Equal(t, 5, len(job.secondaryJob.PopPendingJob()))
	pieceJob.StorageProviderSealInfo.PieceIdx = 5
	job.DoneSecondarySPJob(pieceJob)
	assert.Equal(t, 4, len(job.secondaryJob.PopPendingJob()))
	pieceJob.StorageProviderSealInfo.PieceIdx = 4
	job.DoneSecondarySPJob(pieceJob)
	assert.Equal(t, 3, len(job.secondaryJob.PopPendingJob()))
	pieceJob.StorageProviderSealInfo.PieceIdx = 3
	job.DoneSecondarySPJob(pieceJob)
	assert.Equal(t, 2, len(job.secondaryJob.PopPendingJob()))
	pieceJob.StorageProviderSealInfo.PieceIdx = 2
	job.DoneSecondarySPJob(pieceJob)
	assert.Equal(t, 1, len(job.secondaryJob.PopPendingJob()))
	pieceJob.StorageProviderSealInfo.PieceIdx = 1
	job.DoneSecondarySPJob(pieceJob)
	assert.Equal(t, true, job.secondaryJob.Completed())
}

func TestSegmentPieceError(t *testing.T) {
	job, object := InitEnv(ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED)
	pieceJob := &stypesv1pb.PieceJob{
		ObjectId:       object.GetObjectId(),
		PayloadSize:    object.GetSize(),
		RedundancyType: object.GetRedundancyType(),
	}
	badCheckSum := hash.GenerateChecksum([]byte(time.Now().String()))[0:10]
	pieceJob.StorageProviderSealInfo = &stypesv1pb.StorageProviderSealInfo{
		StorageProviderId: "test-storage-provider",
	}
	pieceJob.StorageProviderSealInfo.PieceIdx = 0
	err := job.DonePrimarySPJob(pieceJob)
	assert.Equal(t, merrors.ErrCheckSumCountMismatch, err)
	pieceJob.StorageProviderSealInfo.PieceChecksum = append(pieceJob.StorageProviderSealInfo.PieceChecksum, badCheckSum)
	err = job.DonePrimarySPJob(pieceJob)
	assert.Equal(t, merrors.ErrCheckSumLengthMismatch, err)

}

func TestECPieceError(t *testing.T) {
	job, object := InitEnv(ptypesv1pb.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED)
	pieceJob := &stypesv1pb.PieceJob{
		ObjectId:       object.GetObjectId(),
		PayloadSize:    object.GetSize(),
		RedundancyType: object.GetRedundancyType(),
	}
	badCheckSum := hash.GenerateChecksum([]byte(time.Now().String()))[0:10]
	checkSum := hash.GenerateChecksum([]byte(time.Now().String()))
	pieceJob.StorageProviderSealInfo = &stypesv1pb.StorageProviderSealInfo{
		StorageProviderId: "test-storage-provider",
	}
	pieceJob.StorageProviderSealInfo.PieceIdx = 0
	pieceJob.StorageProviderSealInfo.PieceChecksum = append(pieceJob.StorageProviderSealInfo.PieceChecksum, checkSum)
	pieceJob.StorageProviderSealInfo.PieceChecksum = append(pieceJob.StorageProviderSealInfo.PieceChecksum, checkSum)
	pieceJob.StorageProviderSealInfo.PieceChecksum = append(pieceJob.StorageProviderSealInfo.PieceChecksum, badCheckSum)
	pieceJob.StorageProviderSealInfo.PieceChecksum = append(pieceJob.StorageProviderSealInfo.PieceChecksum, checkSum)
	err := job.DoneSecondarySPJob(pieceJob)
	assert.Equal(t, merrors.ErrCheckSumLengthMismatch, err)
}
