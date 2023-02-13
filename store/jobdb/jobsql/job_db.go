package jobsql

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	"gorm.io/gorm"

	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/spdb"
)

var _ spdb.JobDB = &JobMetaImpl{}

// JobMetaImpl is an implement of JobDB interface
type JobMetaImpl struct {
	db *gorm.DB
}

// NewJobMetaImpl return a database instance
func NewJobMetaImpl(config *config.SqlDBConfig) (*JobMetaImpl, error) {
	db, err := InitDB(config)
	if err != nil {
		return nil, err
	}
	return &JobMetaImpl{db: db}, nil
}

// CreateUploadPayloadJob create DBJob record and DBObject record, Use JobID field for association.
func (jmi *JobMetaImpl) CreateUploadPayloadJob(info *ptypes.ObjectInfo) (uint64, error) {
	var (
		result             *gorm.DB
		insertJobRecord    *DBJob
		insertObjectRecord *DBObject
	)

	insertJobRecord = &DBJob{
		JobType:    uint32(ptypes.JobType_JOB_TYPE_CREATE_OBJECT),
		JobState:   uint32(ptypes.JobState_JOB_STATE_CREATE_OBJECT_DONE),
		CreateTime: time.Now(),
		ModifyTime: time.Now(),
	}
	result = jmi.db.Create(insertJobRecord)
	if result.Error != nil || result.RowsAffected != 1 {
		return 0, fmt.Errorf("insert job record failed, %s", result.Error)
	}
	insertObjectRecord = &DBObject{
		ObjectID:       info.ObjectId,
		JobID:          insertJobRecord.JobID,
		CreateHash:     hex.EncodeToString(info.TxHash),
		Height:         info.Height,
		Owner:          info.Owner,
		BucketName:     info.BucketName,
		ObjectName:     info.ObjectName,
		Size:           info.Size,
		IsPrivate:      info.IsPrivate,
		ContentType:    info.ContentType,
		PrimarySP:      "mock-sp-string", // todo: how to encode sp info?
		RedundancyType: uint32(info.RedundancyType),
	}
	result = jmi.db.Create(insertObjectRecord)
	if result.Error != nil || result.RowsAffected != 1 {
		return 0, fmt.Errorf("insert object record failed, %s", result.Error)
	}
	return insertJobRecord.JobID, nil
}

// GetObjectInfo query DBObject by objectID, and convert to types.ObjectInfo.
func (jmi *JobMetaImpl) GetObjectInfo(objectID uint64) (*ptypes.ObjectInfo, error) {
	var (
		result         *gorm.DB
		queryCondition *DBObject
		queryReturn    DBObject
	)

	// If the primary key is a number, the query will be written as follows:
	queryCondition = &DBObject{
		ObjectID: objectID,
	}
	result = jmi.db.Model(queryCondition).First(&queryReturn)
	if result.Error != nil {
		return nil, fmt.Errorf("select job record's failed, %s", result.Error)
	}
	txHash, err := hex.DecodeString(queryReturn.CreateHash)
	if err != nil {
		return nil, err
	}
	return &ptypes.ObjectInfo{
		JobId:          queryReturn.JobID,
		Owner:          queryReturn.Owner,
		BucketName:     queryReturn.BucketName,
		ObjectName:     queryReturn.ObjectName,
		Size:           queryReturn.Size,
		Checksum:       []byte(queryReturn.Checksum),
		IsPrivate:      queryReturn.IsPrivate,
		ContentType:    queryReturn.ContentType,
		PrimarySp:      nil, // todo: how to decode sp info
		Height:         queryReturn.Height,
		TxHash:         txHash,
		RedundancyType: ptypes.RedundancyType(queryReturn.RedundancyType),
		SecondarySps:   nil, // todo: how to fill
	}, nil
}

// GetJobContext query DBJob by jobID, and convert to types.JobContext.
func (jmi *JobMetaImpl) GetJobContext(jobID uint64) (*ptypes.JobContext, error) {
	var (
		result      *gorm.DB
		queryReturn DBJob
	)

	result = jmi.db.First(&queryReturn, "job_id = ?", jobID)
	if result.Error != nil {
		return nil, fmt.Errorf("select job record's failed, %s", result.Error)
	}
	// log.Infow("get job context", "id", jobID, "context", queryReturn)
	return &ptypes.JobContext{
		JobId:      queryReturn.JobID,
		JobType:    ptypes.JobType(queryReturn.JobType),
		JobState:   ptypes.JobState(queryReturn.JobState),
		JobErr:     queryReturn.JobErr,
		CreateTime: queryReturn.CreateTime.Unix(),
		ModifyTime: queryReturn.ModifyTime.Unix(),
	}, nil
}

// SetUploadPayloadJobState update DBJob record's state.
func (jmi *JobMetaImpl) SetUploadPayloadJobState(jobId uint64, jobState string, timestampSec int64) error {
	var (
		result         *gorm.DB
		queryCondition *DBJob
		updateFields   *DBJob
	)

	queryCondition = &DBJob{
		JobID: jobId,
	}
	updateFields = &DBJob{
		JobState:   uint32(ptypes.JobState_value[jobState]),
		ModifyTime: time.Unix(timestampSec, 0),
	}
	result = jmi.db.Model(queryCondition).Updates(updateFields)
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("update job record's state failed, %s", result.Error)
	}
	return nil
}

// SetUploadPayloadJobJobError update DBJob record's state and err.
func (jmi *JobMetaImpl) SetUploadPayloadJobJobError(jobID uint64, jobState string, jobErr string, timestampSec int64) error {
	var (
		result         *gorm.DB
		queryCondition *DBJob
		updateFields   *DBJob
	)

	queryCondition = &DBJob{
		JobID: jobID,
	}
	updateFields = &DBJob{
		JobState:   uint32(ptypes.JobState_value[jobState]),
		ModifyTime: time.Unix(timestampSec, 0),
		JobErr:     jobErr,
	}
	result = jmi.db.Model(queryCondition).Updates(updateFields)
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("update job record's state and err failed, %s", result.Error)
	}
	return nil
}

// GetPrimaryJob query DBPieceJob by objectID and primary type, and convert to PieceJob.
func (jmi *JobMetaImpl) GetPrimaryJob(objectId uint64) ([]*spdb.PieceJob, error) {
	var (
		result       *gorm.DB
		queryReturns []DBPieceJob
		pieceJobs    []*spdb.PieceJob
	)

	result = jmi.db.
		Where("object_id = ? AND piece_type = ?", objectId, ptypes.JobType_JOB_TYPE_UPLOAD_PRIMARY).
		Find(&queryReturns)
	if result.Error != nil {
		return pieceJobs, fmt.Errorf("select primary piece jobs failed, %s", result.Error)
	}
	for _, job := range queryReturns {
		pieceJobs = append(pieceJobs, &spdb.PieceJob{
			PieceId:         job.PieceIdx,
			Checksum:        [][]byte{[]byte(job.IntegrityHash)},
			StorageProvider: job.StorageProvider})
	}
	return pieceJobs, nil
}

// SetPrimaryPieceJobDone create primary DBPieceJob record.
func (jmi *JobMetaImpl) SetPrimaryPieceJobDone(objectID uint64, pj *spdb.PieceJob) error {
	var (
		result               *gorm.DB
		insertPieceJobRecord *DBPieceJob
	)

	insertPieceJobRecord = &DBPieceJob{
		ObjectID:        objectID,
		PieceType:       uint32(ptypes.JobType_JOB_TYPE_UPLOAD_PRIMARY),
		PieceIdx:        pj.PieceId,
		Checksum:        string(pj.Checksum[0]),
		PieceState:      0, // todo: fill what?
		StorageProvider: pj.StorageProvider,
		IntegrityHash:   "",
		Signature:       "",
	}
	result = jmi.db.Create(insertPieceJobRecord)
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("insert primary piece job record failed, %s", result.Error)
	}
	return nil
}

// SetSecondaryPieceJobDone create secondary DBPieceJob record.
func (jmi *JobMetaImpl) SetSecondaryPieceJobDone(objectID uint64, pj *spdb.PieceJob) error {
	var (
		result               *gorm.DB
		insertPieceJobRecord *DBPieceJob
	)

	insertPieceJobRecord = &DBPieceJob{
		ObjectID:        objectID,
		PieceType:       uint32(ptypes.JobType_JOB_TYPE_UPLOAD_SECONDARY_EC),
		PieceIdx:        pj.PieceId,
		Checksum:        string(pj.Checksum[0]),
		PieceState:      0, // todo: fill what?
		StorageProvider: pj.StorageProvider,
		IntegrityHash:   "",
		Signature:       "",
	}
	result = jmi.db.Create(insertPieceJobRecord)
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("insert secondary piece job record failed, %s", result.Error)
	}
	return nil
}

// GetSecondaryJob query DBPieceJob by objectID and secondary type, and convert to PieceJob.
func (jmi *JobMetaImpl) GetSecondaryJob(objectID uint64) ([]*spdb.PieceJob, error) {
	var (
		result       *gorm.DB
		queryReturns []DBPieceJob
		pieceJobs    []*spdb.PieceJob
	)

	result = jmi.db.
		Where("object_id = ? AND piece_type = ?", objectID, ptypes.JobType_JOB_TYPE_UPLOAD_SECONDARY_EC).
		Find(&queryReturns)
	if result.Error != nil {
		return pieceJobs, fmt.Errorf("select secondary piece jobs failed, %s", result.Error)
	}
	for _, job := range queryReturns {
		pieceJobs = append(pieceJobs, &spdb.PieceJob{
			PieceId:         job.PieceIdx,
			Checksum:        [][]byte{[]byte(job.IntegrityHash)},
			StorageProvider: job.StorageProvider})
	}
	return pieceJobs, nil
}

// ScanObjectInfo query scan DBObject, and convert to ObjectInfo.
func (jmi *JobMetaImpl) ScanObjectInfo(offset int, limit int) ([]*ptypes.ObjectInfo, error) {
	var (
		result       *gorm.DB
		queryReturns []DBObject
		objects      []*ptypes.ObjectInfo
	)

	result = jmi.db.Limit(limit).Offset(offset).Find(&queryReturns)
	if result.Error != nil {
		return objects, fmt.Errorf("select objects failed, %s", result.Error)
	}
	for _, object := range queryReturns {
		txHash, err := hex.DecodeString(object.CreateHash)
		if err != nil {
			return objects, err
		}
		objects = append(objects, &ptypes.ObjectInfo{
			JobId:          object.JobID,
			Owner:          object.Owner,
			BucketName:     object.BucketName,
			ObjectName:     object.ObjectName,
			Size:           object.Size,
			Checksum:       []byte(object.Checksum),
			IsPrivate:      object.IsPrivate,
			ContentType:    object.ContentType,
			PrimarySp:      nil, // todo: how to decode sp info
			Height:         object.Height,
			TxHash:         txHash,
			RedundancyType: ptypes.RedundancyType(object.RedundancyType),
			SecondarySps:   nil, // todo: how to fill
		})
	}
	return objects, nil
}

// ScanJob query scan DBJob, and convert to JobContext, [startObjectID, xx)
func (jmi *JobMetaImpl) ScanJob(startObjectID uint64, limit int) ([]*ptypes.JobContext, error) {
	var (
		result       *gorm.DB
		jobs         []*ptypes.JobContext
		queryReturns []DBJob
	)
	result = jmi.db.Where("job_id >= ?", startObjectID).Limit(limit).Find(&queryReturns)
	if result.Error != nil {
		return jobs, fmt.Errorf("select jobs failed, %s", result.Error)
	}
	for _, job := range queryReturns {
		jobs = append(jobs, &ptypes.JobContext{
			JobId:      job.JobID,
			JobType:    ptypes.JobType(job.JobType),
			JobState:   ptypes.JobState(job.JobState),
			JobErr:     job.JobErr,
			CreateTime: job.CreateTime.Unix(),
			ModifyTime: job.ModifyTime.Unix(),
		})
	}
	return jobs, nil
}

// GetObjectInfoByJob returns the object info by job id.
func (jmi *JobMetaImpl) GetObjectInfoByJob(jobID uint64) (*ptypes.ObjectInfo, error) {
	var (
		result      *gorm.DB
		queryReturn DBObject
	)

	result = jmi.db.Where("job_id = ?", jobID).Find(&queryReturn)
	if result.Error != nil {
		return nil, fmt.Errorf("select job record's failed, %s", result.Error)
	}
	txHash, err := hex.DecodeString(queryReturn.CreateHash)
	if err != nil {
		return nil, err
	}
	return &ptypes.ObjectInfo{
		JobId:          queryReturn.JobID,
		ObjectId:       queryReturn.ObjectID,
		Owner:          queryReturn.Owner,
		BucketName:     queryReturn.BucketName,
		ObjectName:     queryReturn.ObjectName,
		Size:           queryReturn.Size,
		Checksum:       []byte(queryReturn.Checksum),
		IsPrivate:      queryReturn.IsPrivate,
		ContentType:    queryReturn.ContentType,
		PrimarySp:      nil, // todo: how to decode sp info
		Height:         queryReturn.Height,
		TxHash:         txHash,
		RedundancyType: ptypes.RedundancyType(queryReturn.RedundancyType),
		SecondarySps:   nil, // todo: how to fill
	}, nil
}

// DeleteJob delete job by id, delete the related object and piece jobs.
func (jmi *JobMetaImpl) DeleteJob(jobID uint64) error {
	var (
		err      error
		objectID uint64
	)

	objectInfo, err := jmi.GetObjectInfoByJob(jobID)
	if err != nil {
		log.Infow("delete job context, due to get object info", "error", err, "job_id", jobID)
		return err
	}

	objectID = objectInfo.ObjectId
	log.Infow("delete job context", "job_id", jobID, "object_id", objectID)
	// jobID as primary key
	if err = jmi.db.Delete(&DBJob{}, jobID).Error; err != nil {
		log.Warnw("failed to delete job", "error", err)
		return err
	}

	// object id is not primary key
	if err = jmi.db.Delete(&DBObject{}, "object_id = ?", objectID).Error; err != nil {
		log.Warnw("failed to delete object", "object_id", objectID, "error", err)
		return err
	}
	if err = jmi.db.Delete(&DBPieceJob{}, "object_id = ?", objectID).Error; err != nil {
		log.Warnw("failed to delete piece job", "object_id", objectID, "error", err)
		return err
	}
	return nil
}

const iterBatchNum = 100

// jobIter is DBJobTable iterator
type jobIter struct {
	db         *JobMetaImpl
	startJobID uint64
	curJobs    []*ptypes.JobContext
	curIndex   int
	finished   bool
	err        error
}

func (it *jobIter) fetchNewBatch() error {
	it.curJobs, it.err = it.db.ScanJob(it.startJobID, iterBatchNum)
	if it.err != nil {
		return it.err
	}
	if len(it.curJobs) == 0 {
		it.finished = true
		return fmt.Errorf("has finished")
	}
	it.curIndex = 0
	it.startJobID = it.curJobs[len(it.curJobs)-1].JobId + 1
	return nil
}

// IsValid return true if current element is valid.
func (it *jobIter) IsValid() bool {
	if it.finished || it.err != nil {
		return false
	}
	if it.curJobs == nil { // first fetch
		if err := it.fetchNewBatch(); err != nil {
			return false
		}
	}
	if it.curIndex < len(it.curJobs) {
		return true
	}
	// fetch a new batch
	if err := it.fetchNewBatch(); err != nil {
		return false
	}
	return true
}

// Next move to next
func (it *jobIter) Next() {
	it.curIndex = it.curIndex + 1
}

// Error returns any accumulated error. Exhausting all the key/value pairs
// is not considered to be an error.
func (it *jobIter) Error() error {
	return it.err
}

// Key returns the key of the current key/value pair, or nil if done. The caller
// should not modify the contents of the returned slice, and its contents may
// change on the next call to Next.
func (it *jobIter) Key() interface{} {
	if it.curIndex < len(it.curJobs) {
		return it.curJobs[it.curIndex].JobId
	}
	return nil
}

// Value returns the value of the current key/value pair, or nil if done. The
// caller should not modify the contents of the returned slice, and its contents
// may change on the next call to Next.
func (it *jobIter) Value() interface{} {
	if it.curIndex < len(it.curJobs) {
		return it.curJobs[it.curIndex]
	}
	return nil
}

// Release releases associated resources. Release should always succeed and can
// be called multiple times without causing error.
func (it *jobIter) Release() {
	it.curJobs, it.db = nil, nil
}

// NewIterator creates an iterator over a subset,
// starting at a particular initial key.
func (jmi *JobMetaImpl) NewIterator(start interface{}) spdb.Iterator {
	return &jobIter{
		db:         jmi,
		startJobID: start.(uint64),
		curJobs:    nil,
		curIndex:   0,
		finished:   false,
		err:        nil,
	}
}

// jobBatch is a write-only memory batch that commits changes to its host
// database when Write is called. A batch cannot be used concurrently.
type jobBatch struct {
	db   *JobMetaImpl
	keys []uint64
	size int
}

// Put inserts the given value into the key-value data store.
func (b *jobBatch) Put(key interface{}, value interface{}) error {
	return errors.New("job db batch not support put")
}

// Delete removes the key from the key-value data store.
func (b *jobBatch) Delete(key interface{}) error {
	jobID := key.(uint64)
	b.keys = append(b.keys, jobID)
	b.size += 8
	return nil
}

// ValueSize retrieves the amount of data queued up for writing.
func (b *jobBatch) ValueSize() int {
	return b.size
}

// Write flushes any accumulated data to disk.
func (b *jobBatch) Write() error {
	for _, jobID := range b.keys {
		if err := b.db.DeleteJob(jobID); err != nil {
			return err
		}
	}
	return nil
}

// Reset resets the batch for reuse.
func (b *jobBatch) Reset() {
	b.keys = b.keys[:0]
	b.size = 0
}

// NewBatch creates a write-only database that buffers changes to its host db
// until a final write is called.
func (jmi *JobMetaImpl) NewBatch() spdb.Batch {
	return &jobBatch{
		db: jmi,
	}
}

// NewBatchWithSize creates a write-only database batch with pre-allocated buffer.
func (jmi *JobMetaImpl) NewBatchWithSize(size int) spdb.Batch {
	return &jobBatch{
		db:   jmi,
		size: size,
	}
}
