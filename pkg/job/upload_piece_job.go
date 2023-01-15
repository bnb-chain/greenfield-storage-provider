package job

import (
	"sync"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	types "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

// UploadSpJob stands one upload job for the one storage provider.
type UploadSpJob struct {
	objectCtx  *ObjectInfoContext
	pieceJobs  []*jobdb.PieceJob
	pieceType  model.PieceType
	redundancy types.RedundancyType
	secondary  bool
	complete   int
	mu         sync.RWMutex
}

func NewSegmentUploadSpJob(objectCtx *ObjectInfoContext, secondary bool) (*UploadSpJob, error) {
	job := &UploadSpJob{
		objectCtx:  objectCtx,
		secondary:  secondary,
		pieceType:  model.SegmentPieceType,
		redundancy: objectCtx.GetObjectRedundancyType(),
	}
	pieceCount := util.ComputeSegmentCount(objectCtx.GetObjectSize())
	if job.secondary && job.redundancy == types.RedundancyType_REDUNDANCY_TYPE_REPLICA_TYPE {
		pieceCount = model.EC_M + model.EC_K
	}
	for idx := 0; idx < int(pieceCount); idx++ {
		job.pieceJobs = append(job.pieceJobs, &jobdb.PieceJob{
			PieceId:       uint32(idx),
			Checksum:      make([][]byte, 1),
			IntegrityHash: make([]byte, hash.LengthHash),
			Signature:     make([]byte, hash.LengthHash),
		})
	}
	pieceJobs, err := job.objectCtx.getPrimaryPieceJob()
	if err != nil {
		return nil, err
	}
	for _, pieceJob := range pieceJobs {
		if pieceJob.PieceId >= pieceCount {
			return nil, merrors.ErrIndexOutOfBounds
		}
		job.pieceJobs[pieceJob.PieceId] = pieceJob
	}
	return job, nil
}

func NewECUploadSpJob(objectCtx *ObjectInfoContext, secondary bool) (*UploadSpJob, error) {
	job := &UploadSpJob{
		objectCtx:  objectCtx,
		secondary:  secondary,
		pieceType:  model.ECPieceType,
		redundancy: objectCtx.GetObjectRedundancyType(),
	}
	pieceCount := model.EC_M + model.EC_K
	segmentCount := util.ComputeSegmentCount(objectCtx.GetObjectSize())
	for pieceIdx := 0; pieceIdx < int(pieceCount); pieceIdx++ {
		job.pieceJobs = append(job.pieceJobs, &jobdb.PieceJob{
			PieceId:       uint32(pieceIdx),
			Checksum:      make([][]byte, segmentCount),
			IntegrityHash: make([]byte, hash.LengthHash),
			Signature:     make([]byte, hash.LengthHash),
		})
	}
	pieceJobs, err := job.objectCtx.getSecondaryJob()
	if err != nil {
		return nil, err
	}
	for _, pieceJob := range pieceJobs {
		if pieceJob.PieceId >= segmentCount {
			return nil, merrors.ErrIndexOutOfBounds
		}
		job.pieceJobs[pieceJob.PieceId] = pieceJob
	}
	return job, nil
}

// Completed whether upload job is completed.
func (job *UploadSpJob) Completed() bool {
	job.mu.RLock()
	defer job.mu.RUnlock()
	return job.complete == len(job.pieceJobs)
}

// PopPendingJob return the uncompleted piece jobs.
func (job *UploadSpJob) PopPendingJob() (pieceIdx []uint32) {
	job.mu.RLock()
	defer job.mu.RUnlock()
	if job.complete == len(job.pieceJobs) {
		return pieceIdx
	}
	for i, pieceJob := range job.pieceJobs {
		if pieceJob.Done {
			continue
		}
		pieceIdx = append(pieceIdx, uint32(i))
	}
	return pieceIdx
}

// Done completed one piece job and store the state to DB.
func (job *UploadSpJob) Done(pieceJob *service.PieceJob) error {
	job.mu.Lock()
	defer job.mu.Unlock()
	// 1. check job weather has completed
	if job.complete == len(job.pieceJobs) {
		log.Warnw("upload storage provider already completed", "object_id", pieceJob.GetObjectId(), "secondary_sp", job.secondary,
			"piece_idx", pieceJob.GetStorageProviderSealInfo().GetPieceIdx(), "piece_type", job.pieceType)
		return nil
	}
	// 2. get piece job
	pieceIdx := pieceJob.GetStorageProviderSealInfo().GetPieceIdx()
	if pieceIdx >= uint32(len(job.pieceJobs)) {
		return merrors.ErrIndexOutOfBounds
	}
	piece := job.pieceJobs[pieceIdx]
	// 3. check piece job weather has completed
	if piece.Done {
		log.Warnw("piece job already completed", "object_id", pieceJob.GetObjectId(), "secondary_sp", job.secondary,
			"piece_idx", pieceJob.GetStorageProviderSealInfo().GetPieceIdx(), "piece_type", job.pieceType)
		return nil
	}
	// 4. update piece state
	var err error
	if job.pieceType == model.SegmentPieceType {
		err = job.doneSegment(piece, pieceJob)
	} else {
		err = job.doneEC(piece, pieceJob)
	}
	return err
}

// donePrimary update primary piece job state, include memory and db.
func (job *UploadSpJob) doneSegment(segmentPiece *jobdb.PieceJob, pieceJob *service.PieceJob) error {
	pieceCount := uint32(len(pieceJob.GetStorageProviderSealInfo().GetPieceChecksum()))
	segmentCount := util.ComputeSegmentCount(job.objectCtx.GetObjectSize())
	if job.redundancy == types.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED {
		if pieceCount != 1 {
			log.Errorw("done segment piece error", "object_id", pieceJob.GetObjectId(), "second", job.secondary,
				"piece_idx", segmentPiece.PieceId, "want 1, checksum_count", pieceCount, "error", merrors.ErrCheckSumCountMismatch)
			return merrors.ErrCheckSumCountMismatch
		}
		pieceCheckSumLen := len(pieceJob.GetStorageProviderSealInfo().GetPieceChecksum()[0])
		if pieceCheckSumLen != hash.LengthHash {
			log.Errorw("done segment piece error", "object_id", pieceJob.GetObjectId(), "second", job.secondary,
				"piece_idx", segmentPiece.PieceId, "piece_checksum_length", pieceCheckSumLen, "error", merrors.ErrCheckSumLengthMismatch)
			return merrors.ErrCheckSumLengthMismatch
		}
	}
	if job.redundancy == types.RedundancyType_REDUNDANCY_TYPE_EC_TYPE_UNSPECIFIED {
		if pieceCount != segmentCount {
			log.Errorw("done ec piece error", "object_id", pieceJob.GetObjectId(), "piece_idx", segmentPiece.PieceId,
				"want_count", segmentCount, "piece_count", pieceCount, "error", merrors.ErrCheckSumCountMismatch)
			return merrors.ErrCheckSumCountMismatch
		}
		for idx, checkSum := range pieceJob.GetStorageProviderSealInfo().GetPieceChecksum() {
			if len(checkSum) != hash.LengthHash {
				log.Errorw("done ec piece error", "object_id", pieceJob.GetObjectId(), "piece_idx", segmentPiece.PieceId,
					"checksum_idx", idx, "piece_check_sum_lengeth", len(checkSum), "error", merrors.ErrCheckSumLengthMismatch)
				return merrors.ErrCheckSumLengthMismatch
			}
		}
	}
	if job.secondary {
		integrityHashLen := len(pieceJob.GetStorageProviderSealInfo().GetIntegrityHash())
		if integrityHashLen != hash.LengthHash {
			log.Errorw("done segment piece error", "object_id", pieceJob.GetObjectId(), "second", job.secondary,
				"piece_idx", segmentPiece.PieceId, "integrity_hash_length", integrityHashLen, "error", merrors.ErrIntegrityHashLengthMismatch)
			return merrors.ErrIntegrityHashLengthMismatch
		}
		// TODO:: currrent signer service is not completed
		//if len(pieceJob.GetStorageProviderSealInfo().GetSignature()) != hash.LengthHash {
		//	log.Errorw("done segment piece error", "object_id", pieceJob.GetObjectId(),
		//		"second", job.secondary, "piece_idx", segmentPiece.PieceId, "error", merrors.ErrSignatureLengthMismatch)
		//	return merrors.ErrSignatureLengthMismatch
		//}
		segmentPiece.IntegrityHash = pieceJob.GetStorageProviderSealInfo().GetIntegrityHash()
		segmentPiece.Signature = pieceJob.GetStorageProviderSealInfo().GetSignature()
	}
	segmentPiece.Checksum = pieceJob.GetStorageProviderSealInfo().GetPieceChecksum()
	segmentPiece.StorageProvider = pieceJob.GetStorageProviderSealInfo().GetStorageProviderId()
	segmentPiece.Done = true
	if job.secondary {
		if err := job.objectCtx.SetSecondaryPieceJobDone(segmentPiece); err != nil {
			log.Errorw("write secondary piece to db error", "object_id", pieceJob.GetObjectId(),
				"piece_idx", segmentPiece.PieceId, "error", err)
			return err
		}
	} else {
		if err := job.objectCtx.SetPrimaryPieceJobDone(segmentPiece); err != nil {
			log.Errorw("write primary piece to db error", "object_id", pieceJob.GetObjectId(),
				"piece_idx", segmentPiece.PieceId, "error", err)
			return err
		}
	}
	job.complete++
	return nil
}

// doneSecondary update primary piece job state, include memory and db.
func (job *UploadSpJob) doneEC(ecPiece *jobdb.PieceJob, pieceJob *service.PieceJob) error {
	pieceCount := uint32(len(pieceJob.GetStorageProviderSealInfo().GetPieceChecksum()))
	segmentCount := util.ComputeSegmentCount(job.objectCtx.GetObjectSize())
	if pieceCount != segmentCount {
		log.Errorw("done ec piece error", "object_id", pieceJob.GetObjectId(), "piece_idx", ecPiece.PieceId,
			"want_count", segmentCount, "piece_count", pieceCount, "error", merrors.ErrCheckSumCountMismatch)
		return merrors.ErrCheckSumCountMismatch
	}
	for idx, checkSum := range pieceJob.GetStorageProviderSealInfo().GetPieceChecksum() {
		if len(checkSum) != hash.LengthHash {
			log.Errorw("done ec piece error", "object_id", pieceJob.GetObjectId(), "piece_idx", ecPiece.PieceId,
				"checksum_idx", idx, "piece_check_sum_lengeth", len(checkSum), "error", merrors.ErrCheckSumLengthMismatch)
			return merrors.ErrCheckSumLengthMismatch
		}
	}
	integrityHashLen := len(pieceJob.GetStorageProviderSealInfo().GetIntegrityHash())
	if integrityHashLen != hash.LengthHash {
		log.Errorw("done ec piece error", "object_id", pieceJob.GetObjectId(), "piece_idx", ecPiece.PieceId,
			"integrity_hash_length", integrityHashLen, "error", merrors.ErrIntegrityHashLengthMismatch)
		return merrors.ErrIntegrityHashLengthMismatch
	}
	// TODO:: currrent signer service is not completed
	//if len(pieceJob.GetStorageProviderSealInfo().GetSignature()) != hash.LengthHash {
	//	log.Errorw("done ec piece error", "object_id", pieceJob.GetObjectId(),
	//		"piece_idx", ecPiece.PieceId, "error", merrors.ErrSignatureLengthMismatch)
	//	return merrors.ErrSignatureLengthMismatch
	//}
	ecPiece.Checksum = pieceJob.GetStorageProviderSealInfo().GetPieceChecksum()
	ecPiece.IntegrityHash = pieceJob.GetStorageProviderSealInfo().GetIntegrityHash()
	ecPiece.Signature = pieceJob.GetStorageProviderSealInfo().GetSignature()
	ecPiece.StorageProvider = pieceJob.GetStorageProviderSealInfo().GetStorageProviderId()
	ecPiece.Done = true
	if err := job.objectCtx.SetSecondaryPieceJobDone(ecPiece); err != nil {
		log.Infow("set secondary piece job to db error", "error", err)
		return err
	}
	job.complete++
	return nil
}

// SealInfo return the info for seal.
func (job *UploadSpJob) SealInfo() ([]*types.StorageProviderInfo, error) {
	job.mu.RLock()
	defer job.mu.RUnlock()
	if job.complete != len(job.pieceJobs) {
		return nil, merrors.ErrSpJobNotCompleted
	}
	var sealInfo []*types.StorageProviderInfo
	if job.secondary {
		sealInfo = job.sealSecondary()
	} else {
		sealInfo = append(sealInfo, job.sealPrimary())
	}
	return sealInfo, nil
}

// PrimarySealInfo compute the primary integrity hash.
func (job *UploadSpJob) sealPrimary() *types.StorageProviderInfo {
	var checksumList [][]byte
	for _, pieceJob := range job.pieceJobs {
		checksumList = append(checksumList, pieceJob.Checksum[0])
	}
	// TODO:: sign the primary integrity hash in stone hub level.
	return &types.StorageProviderInfo{
		SpId:     job.pieceJobs[0].StorageProvider,
		Checksum: hash.GenerateIntegrityHash(checksumList),
	}
}

// SecondarySealInfo return secondary info for seal, the stone node service report.
func (job *UploadSpJob) sealSecondary() []*types.StorageProviderInfo {
	var sealInfo []*types.StorageProviderInfo
	for _, pieceJob := range job.pieceJobs {
		sealInfo = append(sealInfo, &types.StorageProviderInfo{
			SpId:      pieceJob.StorageProvider,
			Idx:       pieceJob.PieceId,
			Checksum:  pieceJob.IntegrityHash,
			Signature: pieceJob.Signature,
		})
	}
	return sealInfo
}
