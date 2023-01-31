package jobdb

import ptypesv1pb "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"

// PieceJob record the piece job context, interact with db.
// For primary:
// a. one piece job stands for one segment.
// For secondary:
// a. ec type: one piece job stands for a set of ec piece that they have the same ec idx.
// b. inline or replicate: one piece job stands for one replicate that store the one secondary sp.
type PieceJob struct {
	PieceId uint32
	// for primary pieces the length of checksum always equal 1.
	// for secondary piece the length of checksum should equal segment count.
	Checksum [][]byte
	// only secondary has piece integrity hash, because the primary
	// need compute after all primary pieces are done
	IntegrityHash []byte
	// only secondary has piece signature, because the primary need
	// compute after all primary pieces are done, before seal object
	Signature       []byte
	StorageProvider string
	Done            bool
}

type JobDB interface {
	CreateUploadPayloadJob(txHash []byte, info *ptypesv1pb.ObjectInfo) (uint64, error)
	SetObjectCreateHeightAndObjectID(txHash []byte, height uint64, objectID uint64) error

	GetObjectInfo(txHash []byte) (*ptypesv1pb.ObjectInfo, error)
	GetJobContext(jobId uint64) (*ptypesv1pb.JobContext, error)

	SetUploadPayloadJobState(jobId uint64, state string, timestamp int64) error
	SetUploadPayloadJobJobError(jobID uint64, jobState string, jobErr string, timestamp int64) error

	GetPrimaryJob(txHash []byte) ([]*PieceJob, error)
	GetSecondaryJob(txHash []byte) ([]*PieceJob, error)
	SetPrimaryPieceJobDone(txHash []byte, piece *PieceJob) error
	SetSecondaryPieceJobDone(txHash []byte, piece *PieceJob) error
}
