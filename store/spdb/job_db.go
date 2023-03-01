package spdb

// import (
//
//	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
//
// )
//
// // PieceJob record the piece job context, interact with db.
// // For primary:
// // a. one piece job stands for one segment.
// // For secondary:
// // a. ec type: one piece job stands for a set of ec piece that they have the same ec idx.
// // b. inline or replicate: one piece job stands for one replicate that store the one secondary sp.
//
//	type PieceJob struct {
//		PieceId uint32
//		// for primary pieces the length of checksum always equal 1.
//		// for secondary piece the length of checksum should equal segment count.
//		Checksum [][]byte
//		// only secondary has piece integrity hash, because the primary
//		// need compute after all primary pieces are done
//		IntegrityHash []byte
//		// only secondary has piece signature, because the primary need
//		// compute after all primary pieces are done, before seal object
//		Signature       []byte
//		StorageProvider string
//		Done            bool
//	}
//
// // JobDB use objectID as primary key,adapt for changing light client to heavy client
type JobDB interface{}

//	CreateUploadPayloadJob(info *ptypes.ObjectInfo) (uint64, error)
//
//	GetJobContext(jobID uint64) (*ptypes.JobContext, error)
//	GetObjectInfo(objectID uint64) (*ptypes.ObjectInfo, error)
//	GetObjectInfoByJob(jobID uint64) (*ptypes.ObjectInfo, error)
//
//	SetUploadPayloadJobState(jobID uint64, state string, timestamp int64) error
//	SetUploadPayloadJobJobError(jobID uint64, jobState string, jobErr string, timestamp int64) error
//
//	GetPrimaryJob(objectID uint64) ([]*PieceJob, error)
//	GetSecondaryJob(objectID uint64) ([]*PieceJob, error)
//	SetPrimaryPieceJobDone(objectID uint64, piece *PieceJob) error
//	SetSecondaryPieceJobDone(objectID uint64, piece *PieceJob) error
//
//	DeleteJob(jobID uint64) error
//
//	Iteratee
//	Batcher
//}
