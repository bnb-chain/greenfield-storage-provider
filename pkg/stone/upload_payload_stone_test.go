package stone

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb/jobmemory"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
)

var (
	txHash, _        = hex.DecodeString("245633f1d8b9adccf4a4eb7459e77ca240a6e4e7da3939328ce192239159ea85")
	height    uint64 = 1000
	objectID  uint64 = 19919
	jobCh            = make(chan StoneJob, 10)
	gcCh             = make(chan uint64, 10)
)

func InitENV() (*UploadPayloadStone, error) {
	object := &ptypes.ObjectInfo{
		Owner:          "test_owner",
		BucketName:     "test_bucket",
		ObjectName:     "test_object",
		Size_:          50 * 1024 * 1024,
		TxHash:         txHash,
		Height:         height,
		ObjectId:       objectID,
		RedundancyType: ptypes.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
		PrimarySp: &ptypes.StorageProviderInfo{
			SpId: "bnb-test-sp",
		},
	}
	jobDB := jobmemory.NewMemJobDB()
	jobID, err := jobDB.CreateUploadPayloadJob(object)
	if err != nil {
		return nil, err
	}
	jobCtx, err := jobDB.GetJobContext(jobID)
	if err != nil {
		return nil, err
	}
	stone, err := NewUploadPayloadStone(context.Background(), jobCtx, object, jobDB, nil, jobCh, gcCh)
	if err != nil {
		return nil, err
	}
	return stone, err
}

func TestFsmPrimaryDoing(t *testing.T) {
	stone, err := InitENV()
	assert.Equal(t, nil, err)
	assert.Equal(t, ptypes.JOB_STATE_UPLOAD_PRIMARY_DOING, stone.jobFsm.Current())
	primaryPieceJob := stone.job.PopPendingPrimarySPJob()
	assert.Equal(t, 4, len(primaryPieceJob.TargetIdx))
	secondaryPieceJob := stone.job.PopPendingSecondarySPJob()
	assert.Equal(t, 6, len(secondaryPieceJob.TargetIdx))
}

func TestFsmPrimaryDoingError(t *testing.T) {
	stone, err := InitENV()
	assert.Equal(t, nil, err)
	pieceJob := &stypes.PieceJob{
		StorageProviderSealInfo: &stypes.StorageProviderSealInfo{
			StorageProviderId: "bnb-test-sp",
			PieceIdx:          0,
		},
	}
	err = stone.ActionEvent(context.Background(), UploadPrimaryPieceDoneEvent, pieceJob)
	assert.Equal(t, merrors.ErrCheckSumCountMismatch, err)
	stone, err = InitENV()
	assert.Equal(t, nil, err)
	pieceJob.StorageProviderSealInfo.PieceChecksum = make([][]byte, 1)
	pieceJob.StorageProviderSealInfo.PieceChecksum[0] = []byte{123}
	err = stone.ActionEvent(context.Background(), UploadPrimaryPieceDoneEvent, pieceJob)
	assert.Equal(t, merrors.ErrCheckSumLengthMismatch, err)

	stone, err = InitENV()
	assert.Equal(t, nil, err)
	pieceJob.StorageProviderSealInfo.PieceChecksum = make([][]byte, 1)
	pieceJob.StorageProviderSealInfo.PieceChecksum[0] = hash.GenerateChecksum([]byte(time.Now().String()))
	err = stone.ActionEvent(context.Background(), UploadPrimaryPieceDoneEvent, pieceJob)
	assert.Equal(t, nil, err)
}

func TestFsmPrimaryDoingAndSecondaryDoingError(t *testing.T) {
	stone, err := InitENV()
	assert.Equal(t, nil, err)
	checkSum := hash.GenerateChecksum([]byte(time.Now().String()))
	primaryPieceJob := &stypes.PieceJob{
		StorageProviderSealInfo: &stypes.StorageProviderSealInfo{
			StorageProviderId: "bnb-test-sp",
		},
	}
	primaryPieceJob.StorageProviderSealInfo.PieceChecksum = make([][]byte, 1)
	primaryPieceJob.StorageProviderSealInfo.PieceChecksum[0] = checkSum
	primaryPieceJob.StorageProviderSealInfo.PieceIdx = 0
	err = stone.ActionEvent(context.Background(), UploadPrimaryPieceDoneEvent, primaryPieceJob)
	assert.Equal(t, nil, err)
	pendingPrimaryPieceJob := stone.job.PopPendingPrimarySPJob()
	assert.Equal(t, 3, len(pendingPrimaryPieceJob.TargetIdx))

	primaryPieceJob.StorageProviderSealInfo.PieceIdx = 1
	err = stone.ActionEvent(context.Background(), UploadPrimaryPieceDoneEvent, primaryPieceJob)
	assert.Equal(t, nil, err)
	pendingPrimaryPieceJob = stone.job.PopPendingPrimarySPJob()
	assert.Equal(t, 2, len(pendingPrimaryPieceJob.TargetIdx))

	primaryPieceJob.StorageProviderSealInfo.PieceIdx = 2
	err = stone.ActionEvent(context.Background(), UploadPrimaryPieceDoneEvent, primaryPieceJob)
	assert.Equal(t, nil, err)
	pendingPrimaryPieceJob = stone.job.PopPendingPrimarySPJob()
	assert.Equal(t, 1, len(pendingPrimaryPieceJob.TargetIdx))

	primaryPieceJob.StorageProviderSealInfo.PieceIdx = 3
	err = stone.ActionEvent(context.Background(), UploadPrimaryPieceDoneEvent, primaryPieceJob)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, stone.job.PrimarySPCompleted())

	secondaryPieceJob := &stypes.PieceJob{
		StorageProviderSealInfo: &stypes.StorageProviderSealInfo{
			StorageProviderId: "bnb-test-sp",
		},
	}
	secondaryPieceJob.StorageProviderSealInfo.IntegrityHash = checkSum
	secondaryPieceJob.StorageProviderSealInfo.Signature = checkSum
	secondaryPieceJob.StorageProviderSealInfo.PieceChecksum = [][]byte{checkSum, checkSum, checkSum, checkSum[0:10]}
	err = stone.ActionEvent(context.Background(), UploadSecondaryPieceDoneEvent, secondaryPieceJob)
	assert.Equal(t, merrors.ErrCheckSumLengthMismatch, err)
}
