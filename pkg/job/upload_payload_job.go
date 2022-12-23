package job

import (
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"sync"
)

type UploadPayloadJob struct {
	primaryJob   *UploadSubJob
	secondaryJob *UploadSubJob
}

func (job *UploadPayloadJob) IsCompletedPrimaryJob() bool {
	return job.primaryJob.Done()
}

func (job *UploadPayloadJob) IsCompletedSecondaryJob() bool {
	if job.secondaryJob == nil {
		return true
	}
	return job.secondaryJob.Done()
}

func (job *UploadPayloadJob) SetUploadPrimaryJob(primaryJob *UploadSubJob) error {
	if job.primaryJob != nil {
		// return error
		return nil
	}
	job.primaryJob = primaryJob
	return nil
}

func (job *UploadPayloadJob) SetUploadSecondaryJob(secondaryJob *UploadSubJob) error {
	if job.secondaryJob != nil {
		// return error
		return nil
	}
	job.secondaryJob = secondaryJob
	return nil
}

func (job *UploadPayloadJob) DonePrimaryPieceJob(piece *service.PieceJob) error {
	return job.primaryJob.DonePieceJob(piece)
}

func (job *UploadPayloadJob) DoneSecondaryPieceJob(piece *service.PieceJob) error {
	return job.secondaryJob.DonePieceJob(piece)
}

func (job *UploadPayloadJob) PopPendingPrimaryJob() []uint32 {
	return job.primaryJob.PopPendingPieceJob()
}

func (job *UploadPayloadJob) PopPendingSecondaryJob() []uint32 {
	return job.secondaryJob.PopPendingPieceJob()
}

type UploadSubJob struct {
	pieceJob  []bool
	checkSum  [][]byte
	completed uint32
	mu        sync.RWMutex
}

func (job *UploadSubJob) Done() bool {
	job.mu.RLock()
	defer job.mu.RUnlock()
	return job.completed == uint32(len(job.pieceJob))
}

func (job *UploadSubJob) DonePieceJob(piece *service.PieceJob) error {
	job.mu.Lock()
	defer job.mu.Unlock()
	if !piece.GetDone() {
		// return error
		return nil
	}
	for _, sp := range piece.StorageProviderSealInfo {
		for i, piece := range sp.PieceIdx {
			if piece < 0 || piece > uint32(len(job.pieceJob)) {
				// return error
				return nil
			}
			if job.pieceJob[piece] {
				continue
			}
			job.pieceJob[piece] = true
			job.checkSum[piece] = sp.CheckSum[i]
			job.completed++
		}
	}
	return nil
}

func (job *UploadSubJob) PopPendingPieceJob() []uint32 {
	job.mu.RLock()
	defer job.mu.RUnlock()
	var pieceIdx []uint32
	for idx, piece := range job.pieceJob {
		if piece {
			continue
		}
		pieceIdx = append(pieceIdx, uint32(idx))
	}
	return pieceIdx
}
