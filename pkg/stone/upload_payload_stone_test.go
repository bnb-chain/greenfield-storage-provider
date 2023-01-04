package stone

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/store/jobdb"
)

var (
	txHash, _        = hex.DecodeString("245633f1d8b9adccf4a4eb7459e77ca240a6e4e7da3939328ce192239159ea85")
	height    uint64 = 1000
	objectID  uint64 = 19919
	jobCh            = make(chan StoneJob, 10)
	gcCh             = make(chan string, 10)
)

func InitENV() (*UploadPayloadStone, error) {
	object := &types.ObjectInfo{
		Owner:          "test_owner",
		BucketName:     "test_bucket",
		ObjectName:     "test_object",
		Size:           50 * 1024 * 1024,
		TxHash:         txHash,
		Height:         height,
		ObjectId:       objectID,
		RedundancyType: types.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED,
		PrimarySp: &types.StorageProviderInfo{
			SpId: "bnb-test-sp",
		},
	}
	jobDB := jobdb.NewMemJobDB()
	if err := jobDB.CreateUploadPayloadJob(txHash, object); err != nil {
		return nil, err
	}
	if err := jobDB.SetObjectCreateHeightAndObjectID(txHash, height, objectID); err != nil {
		return nil, err
	}
	jobID := jobDB.JobCount - 1
	jobCtx, err := jobDB.GetJobContext(jobID)
	if err != nil {
		return nil, err
	}
	stone, err := NewUploadPayloadStone(context.Background(), jobCtx, object, jobDB, jobDB, jobCh, gcCh)
	if err != nil {
		return nil, err
	}
	return stone, err
}

func Test_FSM_PRIMARY_DOING(t *testing.T) {
	stone, err := InitENV()
	assert.Equal(t, nil, err)
	assert.Equal(t, types.JOB_STATE_UPLOAD_PRIMARY_DOING, stone.jobFsm.Current())
	primaryPieceJob := stone.job.PopPendingPrimarySPJob()
	assert.Equal(t, 4, len(primaryPieceJob.TargetIdx))
	secondaryPieceJob := stone.job.PopPendingSecondarySPJob()
	assert.Equal(t, 4, len(secondaryPieceJob.TargetIdx))
}

func Test_FSM_PRIMARY_PIECE_JOB_DONE(t *testing.T) {
	stone, err := InitENV()
	assert.Equal(t, nil, err)
	primaryPieceJob := &service.PieceJob{
		BucketName: "test_bucket",
		ObjectName: "test_object",
		StorageProviderSealInfo: &service.StorageProviderSealInfo{
			StorageProviderId: "bnb-test-sp",
			PieceIdx:          0,
		},
	}
	err = stone.ActionEvent(context.Background(), UploadPrimaryPieceDoneEvent, primaryPieceJob)
	assert.Equal(t, nil, err)
	pendingPrimaryPieceJob := stone.job.PopPendingPrimarySPJob()
	assert.Equal(t, 3, len(pendingPrimaryPieceJob.TargetIdx))

	primaryPieceJob.StorageProviderSealInfo.PieceIdx = 1
	err = stone.ActionEvent(context.Background(), UploadPrimaryPieceDoneEvent, primaryPieceJob)
	pendingPrimaryPieceJob = stone.job.PopPendingPrimarySPJob()
	assert.Equal(t, 2, len(pendingPrimaryPieceJob.TargetIdx))

	primaryPieceJob.StorageProviderSealInfo.PieceIdx = 2
	err = stone.ActionEvent(context.Background(), UploadPrimaryPieceDoneEvent, primaryPieceJob)
	pendingPrimaryPieceJob = stone.job.PopPendingPrimarySPJob()
	assert.Equal(t, 1, len(pendingPrimaryPieceJob.TargetIdx))

	primaryPieceJob.StorageProviderSealInfo.PieceIdx = 3
	err = stone.ActionEvent(context.Background(), UploadPrimaryPieceDoneEvent, primaryPieceJob)
	pendingPrimaryPieceJob = stone.job.PopPendingPrimarySPJob()
	assert.Equal(t, (*service.PieceJob)(nil), pendingPrimaryPieceJob)

	assert.Equal(t, true, stone.PrimarySPJobDone())
	assert.Equal(t, types.JOB_STATE_UPLOAD_SECONDARY_DOING, stone.jobFsm.Current())
}

func Test_FSM_Secondary_PIECE_JOB_DONE(t *testing.T) {
	stone, err := InitENV()
	assert.Equal(t, nil, err)
	primaryPieceJob := &service.PieceJob{
		BucketName: "test_bucket",
		ObjectName: "test_object",
		StorageProviderSealInfo: &service.StorageProviderSealInfo{
			StorageProviderId: "bnb-test-sp",
			PieceIdx:          0,
			PieceCheckSum:     [][]byte{[]byte{1}},
		},
	}
	err = stone.ActionEvent(context.Background(), UploadPrimaryPieceDoneEvent, primaryPieceJob)
	assert.Equal(t, nil, err)
	primaryPieceJob.StorageProviderSealInfo.PieceIdx = 1
	err = stone.ActionEvent(context.Background(), UploadPrimaryPieceDoneEvent, primaryPieceJob)
	assert.Equal(t, nil, err)
	primaryPieceJob.StorageProviderSealInfo.PieceIdx = 2
	err = stone.ActionEvent(context.Background(), UploadPrimaryPieceDoneEvent, primaryPieceJob)
	assert.Equal(t, nil, err)
	primaryPieceJob.StorageProviderSealInfo.PieceIdx = 3
	err = stone.ActionEvent(context.Background(), UploadPrimaryPieceDoneEvent, primaryPieceJob)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, stone.PrimarySPJobDone())
	assert.Equal(t, types.JOB_STATE_UPLOAD_SECONDARY_DOING, stone.jobFsm.Current())

	secondaryPieceJob := &service.PieceJob{
		BucketName: "test_bucket",
		ObjectName: "test_object",
		StorageProviderSealInfo: &service.StorageProviderSealInfo{
			StorageProviderId: "bnb-test-sp",
			PieceIdx:          0,
			PieceCheckSum:     [][]byte{[]byte{1}, []byte{2}, []byte{3}, []byte{4}, []byte{5}, []byte{6}},
			IntegrityHash:     []byte{123},
			Signature:         []byte{123},
		},
	}
	err = stone.ActionEvent(context.Background(), UploadSecondaryPieceDoneEvent, secondaryPieceJob)
	assert.Equal(t, nil, err)
	pendingSecondaryPieceJob := stone.job.PopPendingSecondarySPJob()
	assert.Equal(t, 3, len(pendingSecondaryPieceJob.TargetIdx))
	secondaryPieceJob.StorageProviderSealInfo.PieceIdx = 1
	err = stone.ActionEvent(context.Background(), UploadSecondaryPieceDoneEvent, secondaryPieceJob)
	assert.Equal(t, nil, err)
	secondaryPieceJob.StorageProviderSealInfo.PieceIdx = 2
	err = stone.ActionEvent(context.Background(), UploadSecondaryPieceDoneEvent, secondaryPieceJob)
	assert.Equal(t, nil, err)
	secondaryPieceJob.StorageProviderSealInfo.PieceIdx = 3
	err = stone.ActionEvent(context.Background(), UploadSecondaryPieceDoneEvent, secondaryPieceJob)
	assert.Equal(t, nil, err)

	assert.Equal(t, true, stone.job.Completed())
	assert.Equal(t, types.JOB_STATE_SEAL_OBJECT_TX_DOING, stone.jobFsm.Current())
}

func Test_FSM_INTERRUPT(t *testing.T) {
	stone, err := InitENV()
	assert.Equal(t, nil, err)
	primaryPieceJob := &service.PieceJob{
		BucketName: "test_bucket",
		ObjectName: "test_object",
		StorageProviderSealInfo: &service.StorageProviderSealInfo{
			StorageProviderId: "bnb-test-sp",
			PieceIdx:          0,
			PieceCheckSum:     [][]byte{[]byte{1}},
		},
	}
	err = stone.ActionEvent(context.Background(), UploadPrimaryPieceDoneEvent, primaryPieceJob)
	assert.Equal(t, nil, err)
	err = errors.New("interrupt stone")
	stone.InterruptStone(context.Background(), err)
	assert.Equal(t, types.JOB_STATE_ERROR, stone.jobFsm.Current())
}
