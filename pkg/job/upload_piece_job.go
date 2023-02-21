package job

import (
	"sync"

	"github.com/bnb-chain/greenfield-storage-provider/model"
	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	stypes "github.com/bnb-chain/greenfield-storage-provider/service/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/util"
	"github.com/bnb-chain/greenfield-storage-provider/util/hash"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
)

type PieceJob interface {
	// Completed return whether all piece jobs is completed.
	Completed() bool
	// PopPendingJob return the uncompleted piece jobs.
	PopPendingJob() (pieceIdx []uint32)
	// Done completed one piece job and store the state to DB.
	Done(pieceJob *stypes.PieceJob) error
	// SealInfo return the storage provider info for seal.
	SealInfo() ([]*ptypes.StorageProviderInfo, error)
}

var _ PieceJob = &PrimaryJob{}

// PrimaryJob maintains primary segment piece job info.
type PrimaryJob struct {
	objectCtx        *ObjectInfoContext
	segmentPieceJobs []*spdb.PieceJob
	complete         int
	mu               sync.RWMutex
}

// NewPrimaryJob return the PrimaryJob instance.
func NewPrimaryJob(objectCtx *ObjectInfoContext) (*PrimaryJob, error) {
	job := &PrimaryJob{
		objectCtx: objectCtx,
	}
	segmentCount := util.ComputeSegmentCount(objectCtx.GetObjectSize())
	for idx := 0; idx < int(segmentCount); idx++ {
		job.segmentPieceJobs = append(job.segmentPieceJobs, &spdb.PieceJob{
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
		if pieceJob.PieceId >= segmentCount {
			return nil, merrors.ErrIndexOutOfBounds
		}
		job.segmentPieceJobs[pieceJob.PieceId] = pieceJob
	}
	return job, nil
}

// Completed return whether all segment piece jobs of the primary sp are completed.
func (job *PrimaryJob) Completed() bool {
	job.mu.RLock()
	defer job.mu.RUnlock()
	return job.complete == len(job.segmentPieceJobs)
}

// PopPendingJob return the uncompleted segment piece jobs of the primary sp.
func (job *PrimaryJob) PopPendingJob() (pieceIdx []uint32) {
	job.mu.RLock()
	defer job.mu.RUnlock()
	if job.complete == len(job.segmentPieceJobs) {
		return pieceIdx
	}
	for i, pieceJob := range job.segmentPieceJobs {
		if pieceJob.Done {
			continue
		}
		pieceIdx = append(pieceIdx, uint32(i))
	}
	return pieceIdx
}

// Done complete one segment piece job of the primary sp.
func (job *PrimaryJob) Done(remoteJob *stypes.PieceJob) error {
	job.mu.Lock()
	defer job.mu.Unlock()
	var (
		pieceIdx = remoteJob.GetStorageProviderSealInfo().GetPieceIdx()
		objectID = remoteJob.GetObjectId()
	)
	// 1. check job completed
	if job.complete == len(job.segmentPieceJobs) {
		log.Warnw("all primary piece jobs completed", "object_id", objectID, "piece_idx", pieceIdx)
		return nil
	}
	// 2. get local piece job
	if pieceIdx >= uint32(len(job.segmentPieceJobs)) {
		return merrors.ErrIndexOutOfBounds
	}
	localJob := job.segmentPieceJobs[pieceIdx]
	// 3. check piece job completed
	if localJob.Done {
		log.Warnw("piece job completed", "object_id", objectID, "piece_idx", pieceIdx)
		return nil
	}
	// 4. check piece job info
	if checksumCount := len(remoteJob.GetStorageProviderSealInfo().GetPieceChecksum()); checksumCount != 1 {
		log.Errorw("done primary segment piece error", "object_id", objectID, "piece_idx", pieceIdx,
			"want 1, checksum_count", checksumCount, "error", merrors.ErrCheckSumCountMismatch)
		return merrors.ErrCheckSumCountMismatch
	}
	if checksumLength := len(remoteJob.GetStorageProviderSealInfo().GetPieceChecksum()[0]); checksumLength != hash.LengthHash {
		log.Errorw("done primary segment piece error", "object_id", objectID, "piece_idx", pieceIdx,
			"piece_checksum_length", checksumLength, "error", merrors.ErrCheckSumLengthMismatch)
		return merrors.ErrCheckSumLengthMismatch
	}
	// 5. update piece state
	localJob.Checksum = remoteJob.GetStorageProviderSealInfo().GetPieceChecksum()
	localJob.StorageProvider = remoteJob.GetStorageProviderSealInfo().GetStorageProviderId()
	localJob.Done = true
	if err := job.objectCtx.SetPrimaryPieceJobDone(localJob); err != nil {
		log.Errorw("write primary piece to db error", "object_id", objectID, "piece_idx", pieceIdx, "error", err)
		localJob.Done = false
		return err
	}
	job.complete++
	return nil
}

// SealInfo return seal info of the primary sp.
func (job *PrimaryJob) SealInfo() ([]*ptypes.StorageProviderInfo, error) {
	job.mu.RLock()
	defer job.mu.RUnlock()
	var sealInfo []*ptypes.StorageProviderInfo
	if job.complete != len(job.segmentPieceJobs) {
		return sealInfo, merrors.ErrSpJobNotCompleted
	}
	var checksumList [][]byte
	for _, pieceJob := range job.segmentPieceJobs {
		checksumList = append(checksumList, pieceJob.Checksum[0])
	}
	sealInfo = append(sealInfo, &ptypes.StorageProviderInfo{
		SpId:          job.segmentPieceJobs[0].StorageProvider,
		IntegrityHash: hash.GenerateIntegrityHash(checksumList),
		PieceChecksum: checksumList,
	})

	return sealInfo, nil
}

var _ PieceJob = &SecondarySegmentPieceJob{}

// SecondarySegmentPieceJob maintains secondary segment piece job info.
// use for INLINE redundancy type and REPLICA redundancy type.
type SecondarySegmentPieceJob struct {
	objectCtx     *ObjectInfoContext
	copyPieceJobs []*spdb.PieceJob
	complete      int
	mu            sync.RWMutex
}

// NewSecondarySegmentPieceJob return the SecondarySegmentPieceJob instance.
func NewSecondarySegmentPieceJob(objectCtx *ObjectInfoContext) (*SecondarySegmentPieceJob, error) {
	job := &SecondarySegmentPieceJob{
		objectCtx: objectCtx,
	}
	// TODO:: the number of copy is configurable
	copyCount := model.EC_M + model.EC_K
	for idx := 0; idx < copyCount; idx++ {
		job.copyPieceJobs = append(job.copyPieceJobs, &spdb.PieceJob{
			PieceId:       uint32(idx),
			Checksum:      make([][]byte, 1),
			IntegrityHash: make([]byte, hash.LengthHash),
			Signature:     make([]byte, hash.LengthHash),
		})
	}
	pieceJobs, err := job.objectCtx.getSecondaryJob()
	if err != nil {
		return nil, err
	}
	for _, pieceJob := range pieceJobs {
		if pieceJob.PieceId >= uint32(copyCount) {
			return nil, merrors.ErrIndexOutOfBounds
		}
		job.copyPieceJobs[pieceJob.PieceId] = pieceJob
	}
	return job, nil
}

// Completed return whether all copy piece jobs of the secondary sp are completed.
func (job *SecondarySegmentPieceJob) Completed() bool {
	job.mu.RLock()
	defer job.mu.RUnlock()
	return job.complete == len(job.copyPieceJobs)
}

// PopPendingJob return the uncompleted copy piece jobs of the secondary sp.
func (job *SecondarySegmentPieceJob) PopPendingJob() (pieceIdx []uint32) {
	job.mu.RLock()
	defer job.mu.RUnlock()
	if job.complete == len(job.copyPieceJobs) {
		return pieceIdx
	}
	for i, pieceJob := range job.copyPieceJobs {
		if pieceJob.Done {
			continue
		}
		pieceIdx = append(pieceIdx, uint32(i))
	}
	return pieceIdx
}

// Done complete one copy piece job of the secondary sp.
func (job *SecondarySegmentPieceJob) Done(remoteJob *stypes.PieceJob) error {
	job.mu.Lock()
	defer job.mu.Unlock()
	var (
		copyIdx  = remoteJob.GetStorageProviderSealInfo().GetPieceIdx()
		objectID = remoteJob.GetObjectId()
	)
	// 1. check job completed
	if job.complete == len(job.copyPieceJobs) {
		log.Warnw("all secondary piece jobs completed", "object_id", objectID, "copy_idx", copyIdx)
		return nil
	}
	// 2. get local piece job
	if copyIdx >= uint32(len(job.copyPieceJobs)) {
		return merrors.ErrIndexOutOfBounds
	}
	localJob := job.copyPieceJobs[copyIdx]
	// 3. check piece job completed
	if localJob.Done {
		log.Warnw("piece job completed", "object_id", objectID, "copy_idx", copyIdx)
		return nil
	}
	// 4. check piece job info
	var (
		actualPieceCount   = uint32(len(remoteJob.GetStorageProviderSealInfo().GetPieceChecksum()))
		expectedPieceCount = util.ComputeSegmentCount(job.objectCtx.GetObjectSize())
	)
	if actualPieceCount != expectedPieceCount {
		log.Errorw("done secondary segment piece error", "object_id", objectID, "copy_idx", copyIdx,
			"expected_piece_count", expectedPieceCount, "actual_piece_count", actualPieceCount, "error", merrors.ErrCheckSumCountMismatch)
		return merrors.ErrCheckSumCountMismatch
	}
	for idx, checkSum := range remoteJob.GetStorageProviderSealInfo().GetPieceChecksum() {
		if len(checkSum) != hash.LengthHash {
			log.Errorw("done secondary segment piece error", "object_id", objectID, "copy_idx", copyIdx,
				"segment_idx", idx, "segment_check_sum_length", len(checkSum), "error", merrors.ErrCheckSumLengthMismatch)
			return merrors.ErrCheckSumLengthMismatch
		}
	}
	if integrityHashLength := len(remoteJob.GetStorageProviderSealInfo().GetIntegrityHash()); integrityHashLength != hash.LengthHash {
		log.Errorw("done secondary segment piece error", "object_id", objectID, "copy_idx", copyIdx,
			"integrity_hash_length", integrityHashLength, "error", merrors.ErrIntegrityHashLengthMismatch)
		return merrors.ErrIntegrityHashLengthMismatch
	}
	// TODO:: currrent signer service is not completed
	//if len(pieceJob.GetStorageProviderSealInfo().GetSignature()) != hash.LengthHash {
	//	log.Errorw("done segment piece error", "object_id", pieceJob.GetObjectId(),
	//		"second", job.secondary, "piece_idx", segmentPiece.PieceId, "error", merrors.ErrSignatureLengthMismatch)
	//	return merrors.ErrSignatureLengthMismatch
	//}
	localJob.IntegrityHash = remoteJob.GetStorageProviderSealInfo().GetIntegrityHash()
	localJob.Signature = remoteJob.GetStorageProviderSealInfo().GetSignature()
	localJob.Checksum = remoteJob.GetStorageProviderSealInfo().GetPieceChecksum()
	localJob.StorageProvider = remoteJob.GetStorageProviderSealInfo().GetStorageProviderId()
	localJob.Done = true
	if err := job.objectCtx.SetSecondaryPieceJobDone(localJob); err != nil {
		log.Errorw("write secondary piece to db error", "object_id", objectID, "copy_idx", copyIdx, "error", err)
		localJob.Done = false
		return err
	}
	job.complete++
	return nil
}

// SealInfo return seal info of the secondary sp.
func (job *SecondarySegmentPieceJob) SealInfo() ([]*ptypes.StorageProviderInfo, error) {
	job.mu.RLock()
	defer job.mu.RUnlock()
	var sealInfo []*ptypes.StorageProviderInfo
	if job.complete != len(job.copyPieceJobs) {
		return sealInfo, merrors.ErrSpJobNotCompleted
	}
	for _, pieceJob := range job.copyPieceJobs {
		sealInfo = append(sealInfo, &ptypes.StorageProviderInfo{
			SpId:          pieceJob.StorageProvider,
			Idx:           pieceJob.PieceId,
			IntegrityHash: pieceJob.IntegrityHash,
			PieceChecksum: pieceJob.Checksum,
			Signature:     pieceJob.Signature,
		})
	}
	return sealInfo, nil
}

var _ PieceJob = &SecondaryECPieceJob{}

// SecondaryECPieceJob maintains secondary ec piece job info.
type SecondaryECPieceJob struct {
	objectCtx   *ObjectInfoContext
	ecPieceJobs []*spdb.PieceJob
	complete    int
	mu          sync.RWMutex
}

// NewSecondaryECPieceJob return the SecondaryECPieceJob instance.
func NewSecondaryECPieceJob(objectCtx *ObjectInfoContext) (*SecondaryECPieceJob, error) {
	job := &SecondaryECPieceJob{
		objectCtx: objectCtx,
	}
	pieceJobCount := model.EC_M + model.EC_K
	for idx := 0; idx < pieceJobCount; idx++ {
		job.ecPieceJobs = append(job.ecPieceJobs, &spdb.PieceJob{
			PieceId:       uint32(idx),
			Checksum:      make([][]byte, 1),
			IntegrityHash: make([]byte, hash.LengthHash),
			Signature:     make([]byte, hash.LengthHash),
		})
	}
	pieceJobs, err := job.objectCtx.getSecondaryJob()
	if err != nil {
		return nil, err
	}
	for _, pieceJob := range pieceJobs {
		if pieceJob.PieceId >= uint32(pieceJobCount) {
			return nil, merrors.ErrIndexOutOfBounds
		}
		job.ecPieceJobs[pieceJob.PieceId] = pieceJob
	}
	return job, nil
}

// Completed return whether all ec piece jobs of the secondary sp are completed.
func (job *SecondaryECPieceJob) Completed() bool {
	job.mu.RLock()
	defer job.mu.RUnlock()
	return job.complete == len(job.ecPieceJobs)
}

// PopPendingJob return the uncompleted ec piece jobs of the secondary sp.
func (job *SecondaryECPieceJob) PopPendingJob() (pieceIdx []uint32) {
	job.mu.RLock()
	defer job.mu.RUnlock()
	if job.complete == len(job.ecPieceJobs) {
		return pieceIdx
	}
	for i, pieceJob := range job.ecPieceJobs {
		if pieceJob.Done {
			continue
		}
		pieceIdx = append(pieceIdx, uint32(i))
	}
	return pieceIdx
}

// Done complete one ec piece job of the secondary sp.
func (job *SecondaryECPieceJob) Done(remoteJob *stypes.PieceJob) error {
	job.mu.Lock()
	defer job.mu.Unlock()
	var (
		pieceIdx = remoteJob.GetStorageProviderSealInfo().GetPieceIdx()
		objectID = remoteJob.GetObjectId()
	)
	// 1. check job completed
	if job.complete == len(job.ecPieceJobs) {
		log.Warnw("all secondary piece jobs completed", "object_id", objectID, "piece_idx", pieceIdx)
		return nil
	}
	// 2. get local piece job
	if pieceIdx >= uint32(len(job.ecPieceJobs)) {
		return merrors.ErrIndexOutOfBounds
	}
	localJob := job.ecPieceJobs[pieceIdx]
	// 3. check piece job completed
	if localJob.Done {
		log.Warnw("piece job completed", "object_id", objectID, "piece_idx", pieceIdx)
		return nil
	}
	// 4. check piece job info
	var (
		actualPieceCount   = uint32(len(remoteJob.GetStorageProviderSealInfo().GetPieceChecksum()))
		expectedPieceCount = util.ComputeSegmentCount(job.objectCtx.GetObjectSize())
	)
	if actualPieceCount != expectedPieceCount {
		log.Errorw("done ec piece error", "object_id", objectID, "ec_piece_idx", pieceIdx,
			"expected_piece_count", expectedPieceCount, "actual_piece_count", actualPieceCount, "error", merrors.ErrCheckSumCountMismatch)
		return merrors.ErrCheckSumCountMismatch
	}
	for idx, checkSum := range remoteJob.GetStorageProviderSealInfo().GetPieceChecksum() {
		if len(checkSum) != hash.LengthHash {
			log.Errorw("done ec piece error", "object_id", objectID, "ec_idx", pieceIdx, "piece_idx", idx,
				"piece_check_sum_length", len(checkSum), "error", merrors.ErrCheckSumLengthMismatch)
			return merrors.ErrCheckSumLengthMismatch
		}
	}
	if integrityHashLength := len(remoteJob.GetStorageProviderSealInfo().GetIntegrityHash()); integrityHashLength != hash.LengthHash {
		log.Errorw("done ec piece error", "object_id", objectID, "ec_idx", pieceIdx,
			"integrity_hash_length", integrityHashLength, "error", merrors.ErrIntegrityHashLengthMismatch)
		return merrors.ErrIntegrityHashLengthMismatch
	}
	// TODO:: currrent signer service is not completed
	//if len(pieceJob.GetStorageProviderSealInfo().GetSignature()) != hash.LengthHash {
	//	log.Errorw("done ec piece error", "object_id", pieceJob.GetObjectId(),
	//		"piece_idx", ecPiece.PieceId, "error", merrors.ErrSignatureLengthMismatch)
	//	return merrors.ErrSignatureLengthMismatch
	//}
	localJob.Checksum = remoteJob.GetStorageProviderSealInfo().GetPieceChecksum()
	localJob.IntegrityHash = remoteJob.GetStorageProviderSealInfo().GetIntegrityHash()
	localJob.Signature = remoteJob.GetStorageProviderSealInfo().GetSignature()
	localJob.StorageProvider = remoteJob.GetStorageProviderSealInfo().GetStorageProviderId()
	localJob.Done = true
	if err := job.objectCtx.SetSecondaryPieceJobDone(localJob); err != nil {
		log.Infow("set secondary piece job to db error", "error", err)
		localJob.Done = false
		return err
	}
	job.complete++
	return nil
}

// SealInfo return seal info of the secondary sp.
func (job *SecondaryECPieceJob) SealInfo() ([]*ptypes.StorageProviderInfo, error) {
	job.mu.RLock()
	defer job.mu.RUnlock()
	var sealInfo []*ptypes.StorageProviderInfo
	if job.complete != len(job.ecPieceJobs) {
		return sealInfo, merrors.ErrSpJobNotCompleted
	}
	for _, pieceJob := range job.ecPieceJobs {
		sealInfo = append(sealInfo, &ptypes.StorageProviderInfo{
			SpId:          pieceJob.StorageProvider,
			Idx:           pieceJob.PieceId,
			IntegrityHash: pieceJob.IntegrityHash,
			PieceChecksum: pieceJob.Checksum,
			Signature:     pieceJob.Signature,
		})
	}
	return sealInfo, nil
}
