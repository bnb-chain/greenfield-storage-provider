package job

type UploadPayload interface {
	PopPrimaryJob() []UploadPiece
	DonePrimaryPieceJob(pieceKey, secondary string) error
	IsCompletedPrimaryJob() bool

	GetSecondaryPieceJobBySegmentIdx(segmentKey string) []UploadPiece
	DoneSecondaryPieceJob(pieceKey, secondary string) error
	IsCompletedSecondaryJob() bool
	ReplaceSecondaryByBackup(oldSP string) ([]UploadPiece, error)

	IsCompleted() bool
}

type UploadPrimary interface {
	PopPieceJob() []UploadPiece
	DonePieceJob(pieceKey, secondary string) error
	IsCompleted() bool
}

type UploadSecondary interface {
	DonePieceJob(pieceKey, secondary string) error
	IsCompleted() bool
	ReplaceSecondaryByBackup(oldSP string) ([]UploadPiece, error)
	GetPieceJobBySegmentIdx(segmentKey string) []UploadPiece
}

type UploadPiece interface {
	GetDone() bool
	Done()
	SetSecondarySP(sp string)
	GetSecondarySP() string
	UploadPieceKey() string
	GetSP() string
}
