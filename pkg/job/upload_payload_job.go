package job

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"

	"github.com/bnb-chain/inscription-storage-provider/model"
	"github.com/bnb-chain/inscription-storage-provider/model/piecestore"
	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
	service "github.com/bnb-chain/inscription-storage-provider/service/types/v1"
	"github.com/bnb-chain/inscription-storage-provider/store/jobdb"
	"github.com/bnb-chain/inscription-storage-provider/store/metadb"
	"github.com/bnb-chain/inscription-storage-provider/util"
)

// ObjectInfoContext maintains the object info, goroutine safe.
type ObjectInfoContext struct {
	object *types.ObjectInfo
	jobDB  jobdb.JobDB
	metaDB metadb.MetaDB
	mu     sync.RWMutex
}

// NewObjectInfoContext return the instance of ObjectInfoContext.
func NewObjectInfoContext(object *types.ObjectInfo, jobDB jobdb.JobDB, metaDB metadb.MetaDB) *ObjectInfoContext {
	return &ObjectInfoContext{
		object: object,
		jobDB:  jobDB,
		metaDB: metaDB,
	}
}

// GetObjectInfo return the object info.
func (ctx *ObjectInfoContext) GetObjectInfo() *types.ObjectInfo {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.object.SafeCopy()
}

// GetObjectID return the object resource id.
func (ctx *ObjectInfoContext) GetObjectID() uint64 {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.object.GetObjectId()
}

// GetObjectSize return the object size.
func (ctx *ObjectInfoContext) GetObjectSize() uint64 {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.object.GetSize()
}

// GetObjectRedundancyType return the object redundancy type.
func (ctx *ObjectInfoContext) GetObjectRedundancyType() types.RedundancyType {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.object.GetRedundancyType()
}

// TxHash return the CreateObjectTX hash.
func (ctx *ObjectInfoContext) TxHash() []byte {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.object.GetTxHash()
}

// GetPrimaryJob load the primary piece job from db and return.
func (ctx *ObjectInfoContext) GetPrimaryJob() (*UploadJob, error) {
	pieces, err := ctx.jobDB.GetPrimaryJob(ctx.object.TxHash)
	if err != nil {
		return nil, err
	}
	if len(pieces) == 0 {
		return nil, nil
	}
	job := &UploadJob{
		objectCtx:      ctx,
		redundancyType: ctx.object.RedundancyType,
	}
	for _, piece := range pieces {
		job.pieceJobs = append(job.pieceJobs, &UploadPieceJob{
			PieceId:         piece.PieceId,
			Checksum:        piece.Checksum,
			StorageProvider: piece.StorageProvider,
			Done:            true,
		})
		job.complete++
	}
	return job, nil
}

// GetSecondaryJob load the secondary piece job from db and return.
func (ctx *ObjectInfoContext) GetSecondaryJob() (*UploadJob, error) {
	pieces, err := ctx.jobDB.GetSecondaryJob(ctx.object.TxHash)
	if err != nil {
		return nil, err
	}
	if len(pieces) == 0 {
		return nil, nil
	}
	job := &UploadJob{
		objectCtx:      ctx,
		redundancyType: ctx.object.RedundancyType,
	}
	for _, piece := range pieces {
		job.pieceJobs = append(job.pieceJobs, &UploadPieceJob{
			PieceId:         piece.PieceId,
			Checksum:        piece.Checksum,
			StorageProvider: piece.StorageProvider,
			Done:            true,
		})
		job.complete++
	}
	return job, nil
}

// SetPrimaryPieceJobDone set the primary piece jod completed and update DB.
func (ctx *ObjectInfoContext) SetPrimaryPieceJobDone(piece *UploadPieceJob) error {
	job := &jobdb.PieceJob{
		PieceId:         piece.PieceId,
		Checksum:        piece.Checksum,
		StorageProvider: piece.StorageProvider,
	}
	return ctx.jobDB.SetPrimaryPieceJobDone(ctx.object.GetTxHash(), job)
}

// SetSecondaryPieceJobDone set the secondary piece jod completed and update DB.
func (ctx *ObjectInfoContext) SetSecondaryPieceJobDone(piece *UploadPieceJob) error {
	job := &jobdb.PieceJob{
		PieceId:         piece.PieceId,
		Checksum:        piece.Checksum,
		StorageProvider: piece.StorageProvider,
	}
	return ctx.jobDB.SetSecondaryPieceJobDone(ctx.object.GetTxHash(), job)
}

// SetSetIntegrityHash set integrity hash info to meta db.
func (ctx *ObjectInfoContext) SetSetIntegrityHash(meta *metadb.IntegrityMeta) error {
	return ctx.metaDB.SetIntegrityMeta(meta)
}

// UploadPayloadJob maintains the object info and piece job meta
type UploadPayloadJob struct {
	objectCtx    *ObjectInfoContext
	primaryJob   *UploadJob // the job of uploading primary storage provider
	secondaryJob *UploadJob // the job of uploading secondary storage provider
}

// NewUploadPayloadJob return the instance of UploadPayloadJob.
func NewUploadPayloadJob(objectCtx *ObjectInfoContext) (job *UploadPayloadJob, err error) {
	job = &UploadPayloadJob{
		objectCtx: objectCtx,
	}
	if job.primaryJob, err = NewUploadJob(objectCtx, true); err != nil {
		return nil, err
	}
	if job.secondaryJob, err = NewUploadJob(objectCtx, false); err != nil {
		return nil, err
	}
	return job, nil
}

// GetObjectCtx return the object context.
func (job *UploadPayloadJob) GetObjectCtx() *ObjectInfoContext {
	return job.objectCtx
}

// Completed return whether upload payload job is completed.
func (job *UploadPayloadJob) Completed() bool {
	return job.primaryJob.Completed() && job.secondaryJob.Completed()
}

// PrimarySPCompleted return whether upload primary storage provider job is completed.
func (job *UploadPayloadJob) PrimarySPCompleted() bool {
	return job.primaryJob.Completed()
}

// SecondarySPCompleted return whether upload secondary storage provider job is completed.
func (job *UploadPayloadJob) SecondarySPCompleted() bool {
	return job.secondaryJob.Completed()
}

// DonePrimarySPJob complete one primary piece job and update DB.
func (job *UploadPayloadJob) DonePrimarySPJob(pieceJob *service.PieceJob) error {
	return job.primaryJob.Done(pieceJob, true)
}

// DoneSecondarySPJob complete one secondary piece job and update DB.
func (job *UploadPayloadJob) DoneSecondarySPJob(pieceJob *service.PieceJob) error {
	return job.secondaryJob.Done(pieceJob, false)
}

// PopPendingPrimarySPJob return the uncompleted primary piece job.
func (job *UploadPayloadJob) PopPendingPrimarySPJob() *service.PieceJob {
	pieces := job.primaryJob.PopPendingJob()
	if len(pieces) == 0 {
		return nil
	}
	obj := job.objectCtx.GetObjectInfo()
	pieceJob := &service.PieceJob{
		BucketName:     obj.BucketName,
		ObjectName:     obj.ObjectName,
		TxHash:         obj.TxHash,
		ObjectId:       obj.ObjectId,
		PayloadSize:    obj.Size,
		TargetIdx:      pieces,
		RedundancyType: obj.RedundancyType,
	}
	return pieceJob
}

// PopPendingSecondarySPJob return the uncompleted secondary piece job.
func (job *UploadPayloadJob) PopPendingSecondarySPJob() *service.PieceJob {
	pieces := job.secondaryJob.PopPendingJob()
	if len(pieces) == 0 {
		return nil
	}
	obj := job.objectCtx.GetObjectInfo()
	pieceJob := &service.PieceJob{
		BucketName:     obj.BucketName,
		ObjectName:     obj.ObjectName,
		TxHash:         obj.TxHash,
		ObjectId:       obj.ObjectId,
		PayloadSize:    obj.Size,
		TargetIdx:      pieces,
		RedundancyType: obj.RedundancyType,
	}
	return pieceJob
}

// PrimarySPSealInfo return the primary storage provider seal info.
func (job *UploadPayloadJob) PrimarySPSealInfo() (*SealInfo, error) {
	if !job.primaryJob.Completed() {
		return nil, errors.New("primary job not completed")
	}
	if len(job.primaryJob.pieceJobs) == 0 {
		return nil, errors.New("primary job is empty")
	}
	sp := job.primaryJob.pieceJobs[0].StorageProvider
	sealInfo := &SealInfo{
		StorageProvider: sp,
	}
	objectId := job.objectCtx.GetObjectID()
	hash := sha256.New()
	for segmentIdx, pieceJob := range job.primaryJob.pieceJobs {
		if len(pieceJob.Checksum) != 1 {
			return sealInfo, errors.New("primary storage provider checksum error")
		}
		pieceKey := piecestore.EncodeSegmentPieceKey(objectId, uint32(segmentIdx))
		if segmentIdx >= len(pieceJob.Checksum) {
			return nil, errors.New("primary piece checksum idx is out of bound")
		}
		sealInfo.PieceCheckSum[pieceKey] = pieceJob.Checksum[segmentIdx]
		hash.Write(pieceJob.Checksum[0])
	}
	sealInfo.IntegrityHash = hash.Sum(nil)
	return sealInfo, nil
}

// SecondarySPSealInfo  return the secondary storage provider seal info.
func (job *UploadPayloadJob) SecondarySPSealInfo() ([]*SealInfo, error) {
	if !job.secondaryJob.Completed() {
		return []*SealInfo{}, errors.New("primary job not completed")
	}
	if len(job.secondaryJob.pieceJobs) == 0 {
		return []*SealInfo{}, nil
	}
	var sealInfos []*SealInfo
	for _, pieceJob := range job.secondaryJob.pieceJobs {
		sp := pieceJob.StorageProvider
		if _, ok := job.secondaryJob.spInfo[sp]; !ok {
			return sealInfos, errors.New("less storage provider seal info")
		}
		sealInfos = append(sealInfos, job.secondaryJob.spInfo[sp])
	}
	return sealInfos, nil
}

// SealInfo records the storage provider info for sealing.
type SealInfo struct {
	StorageProvider string
	PieceCheckSum   map[string][]byte
	IntegrityHash   []byte
	Signature       []byte
}

// UploadPieceJob stands one piece job.
// the piece job may be segment piece or ec piece (belong to one secondary).
type UploadPieceJob struct {
	PieceId         uint32
	Checksum        [][]byte
	StorageProvider string
	Done            bool
}

// UploadJob stands one upload job to the one storage provider.
// primary or secondary upload job.
type UploadJob struct {
	objectCtx      *ObjectInfoContext
	pieceJobs      []*UploadPieceJob
	spInfo         map[string]*SealInfo
	redundancyType types.RedundancyType
	complete       int
	mu             sync.RWMutex
}

// NewUploadJob return the instance of UploadJob
func NewUploadJob(objectCtx *ObjectInfoContext, primary bool) (*UploadJob, error) {
	if primary {
		job, err := objectCtx.GetPrimaryJob()
		if err != nil {
			return nil, err
		}
		if job != nil {
			return job, nil
		}
	} else {
		job, err := objectCtx.GetSecondaryJob()
		if err != nil {
			return nil, err
		}
		if job != nil {
			return job, nil
		}
	}
	var pieces []*UploadPieceJob
	var pieceCount uint32
	if primary {
		pieceCount = util.ComputeSegmentCount(objectCtx.GetObjectSize())
	} else {
		pieceCount = model.EC_M + model.EC_K
	}
	for i := 0; i < int(pieceCount); i++ {
		pieces = append(pieces, &UploadPieceJob{PieceId: uint32(i)})
	}
	return &UploadJob{
		objectCtx:      objectCtx,
		pieceJobs:      pieces,
		spInfo:         make(map[string]*SealInfo),
		redundancyType: objectCtx.GetObjectRedundancyType(),
	}, nil
}

// Completed whether upload job is completed.
func (job *UploadJob) Completed() bool {
	job.mu.RLock()
	defer job.mu.RUnlock()
	return job.complete == len(job.pieceJobs)
}

// Done completed one piece job and store the state to DB.
func (job *UploadJob) Done(pieceJob *service.PieceJob, primary bool) error {
	job.mu.Lock()
	defer job.mu.Unlock()
	if job.complete == len(job.pieceJobs) {
		return nil
	}
	pieceIdx := pieceJob.StorageProviderSealInfo.PieceIdx
	if pieceIdx > uint32(len(job.pieceJobs)) {
		return errors.New("piece idx out of bounds")
	}
	if job.pieceJobs[pieceIdx].Done {
		return fmt.Errorf("piece job repeat %d", pieceIdx)
	}
	uploadPieceJob := &UploadPieceJob{
		PieceId:         pieceIdx,
		Checksum:        pieceJob.StorageProviderSealInfo.PieceChecksum,
		StorageProvider: pieceJob.StorageProviderSealInfo.StorageProviderId,
		Done:            true,
	}
	if primary {
		if err := job.objectCtx.SetPrimaryPieceJobDone(uploadPieceJob); err != nil {
			return err
		}
	} else {
		if err := job.objectCtx.SetSecondaryPieceJobDone(uploadPieceJob); err != nil {
			return err
		}
	}
	job.pieceJobs[pieceIdx] = uploadPieceJob
	job.spInfo[pieceJob.StorageProviderSealInfo.StorageProviderId] = &SealInfo{
		StorageProvider: pieceJob.StorageProviderSealInfo.StorageProviderId,
		IntegrityHash:   pieceJob.StorageProviderSealInfo.IntegrityHash,
		Signature:       pieceJob.StorageProviderSealInfo.Signature,
	}
	job.complete++
	return nil
}

// PopPendingJob return the uncompleted piece jobs.
func (job *UploadJob) PopPendingJob() (pieceIdx []uint32) {
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
