package job

import types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"

var _ UploadPayload = &UploadPayloadJob{}

type UploadPayloadJob struct {
	UploadPrimaryJob   *UploadPrimaryJob
	UploadSecondaryJob *UploadSecondaryJob
}

func NewUploadPayloadJob() *UploadPayloadJob {
	return &UploadPayloadJob{}
}

func (job *UploadPayloadJob) DonePrimaryPieceJob(pieceKey, secondary string) error {
	return job.UploadPrimaryJob.DonePieceJob(pieceKey, secondary)
}

func (job *UploadPayloadJob) IsCompletedPrimaryJob() bool {
	return job.UploadPrimaryJob.IsCompleted()
}

func (job *UploadPayloadJob) PopPrimaryJob() []UploadPiece {
	return job.UploadPrimaryJob.PopPieceJob()
}

func (job *UploadPayloadJob) PopAccumulateSecondaryJob() []UploadPiece {
	return []UploadPiece{}
}

func (job *UploadPayloadJob) DoneSecondaryPieceJob(pieceKey, secondary string) error {
	return job.UploadSecondaryJob.DonePieceJob(pieceKey, secondary)
}

func (job *UploadPayloadJob) IsCompletedSecondaryJob() bool {
	return job.UploadSecondaryJob.IsCompleted()
}

func (job *UploadPayloadJob) GetSecondaryPieceJobBySegmentIdx(segmentKey string) []UploadPiece {
	return job.UploadSecondaryJob.GetPieceJobBySegmentIdx(segmentKey)
}

func (job *UploadPayloadJob) ReplaceSecondaryByBackup(oldSP string) ([]UploadPiece, error) {
	return job.UploadSecondaryJob.ReplaceSecondaryByBackup(oldSP)
}

func (job *UploadPayloadJob) IsCompleted() bool {
	return job.UploadSecondaryJob.IsCompleted()
}

var _ UploadPrimary = &UploadPrimaryJob{}

type UploadPrimaryJob struct {
	PieceJob        map[string]UploadPiece
	PieceJobCounter uint32
	Completed       uint32
}

func NewUploadPrimaryJob(object *types.ObjectInfo) *UploadPrimaryJob {
	return nil
}

func (job *UploadPrimaryJob) PopPieceJob() []UploadPiece {
	var pieceJobs []UploadPiece
	for _, piece := range job.PieceJob {
		if piece.GetDone() {
			continue
		}
		pieceJobs = append(pieceJobs, piece)
	}
	return pieceJobs
}

func (job *UploadPrimaryJob) DonePieceJob(pieceKey, secondary string) error {
	if piece, ok := job.PieceJob[pieceKey]; ok {
		if piece.GetDone() {
			return nil
		}
		if piece.GetSecondarySP() != secondary {
			return nil
		}
		piece.Done()
		job.Completed++
	} else {
		return nil
	}
	return nil
}

func (job *UploadPrimaryJob) IsCompleted() bool {
	return job.PieceJobCounter == job.Completed
}

var _ UploadSecondary = &UploadSecondaryJob{}

type UploadSecondaryJob struct {
	PieceJob        map[string]UploadPiece
	SPToPieceJob    map[string][]UploadPiece
	IdxToPieceJob   map[string][]UploadPiece
	BackupSecondary []string
	PieceJobCounter uint32
	Completed       uint32
}

func NewUploadSecondaryJob(object *types.ObjectInfo, secondarySP []string) (*UploadSecondaryJob, error) {
	return nil, nil
}

func (job *UploadSecondaryJob) GetPieceJobBySegmentIdx(segmentKey string) []UploadPiece {
	return job.IdxToPieceJob[segmentKey]
}

func (job *UploadSecondaryJob) DonePieceJob(pieceKey, secondary string) error {
	if piece, ok := job.PieceJob[pieceKey]; ok {
		if piece.GetDone() {
			return nil
		}
		if piece.GetSecondarySP() != secondary {
			return nil
		}
		piece.Done()
		job.Completed++
	} else {
		return nil
	}
	return nil
}

func (job *UploadSecondaryJob) IsCompleted() bool {
	return job.PieceJobCounter == job.Completed
}

func (job *UploadSecondaryJob) ReplaceSecondaryByBackup(oldSP string) ([]UploadPiece, error) {
	var pieceJobs []UploadPiece
	if len(job.BackupSecondary) == 0 {
		return pieceJobs, nil
	}
	newSP := job.BackupSecondary[0]
	job.BackupSecondary = job.BackupSecondary[1:]
	pieces, ok := job.SPToPieceJob[oldSP]
	if !ok {
		return pieceJobs, nil
	}
	for _, piece := range pieces {
		piece.SetSecondarySP(newSP)
	}
	job.SPToPieceJob[newSP] = append(job.SPToPieceJob[newSP], pieces...)
	return pieces, nil
}
