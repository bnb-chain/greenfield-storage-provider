package jobsql

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"gorm.io/gorm"

	ptypes "github.com/bnb-chain/greenfield-storage-provider/pkg/types/v1"
	"github.com/bnb-chain/greenfield-storage-provider/store/jobdb"
)

var _ jobdb.JobDB = &JobMetaImpl{}

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

// v1 interface implement

// CreateUploadPayloadJob create DBJob record and DBObject record, Use JobID field for association.
func (jmi *JobMetaImpl) CreateUploadPayloadJob(txHash []byte, info *ptypes.ObjectInfo) (uint64, error) {
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
		CreateHash:     hex.EncodeToString(txHash),
		JobID:          insertJobRecord.JobID,
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

// SetObjectCreateHeight update DBObject record's height.
func (jmi *JobMetaImpl) SetObjectCreateHeight(txHash []byte, height uint64) error {
	var (
		result         *gorm.DB
		queryCondition *DBObject
		updateFields   *DBObject
	)

	queryCondition = &DBObject{
		CreateHash: hex.EncodeToString(txHash),
	}
	updateFields = &DBObject{
		Height: height,
	}
	result = jmi.db.Model(queryCondition).Updates(updateFields)
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("update object record's height failed, %s", result.Error)
	}
	return nil
}

// GetObjectInfo query DBObject by txHash, and convert to types.ObjectInfo.
func (jmi *JobMetaImpl) GetObjectInfo(txHash []byte) (*ptypes.ObjectInfo, error) {
	var (
		result      *gorm.DB
		queryReturn DBObject
	)

	// If the primary key is a string, the query will be written as follows:
	result = jmi.db.First(&queryReturn, "create_hash = ?", hex.EncodeToString(txHash))
	if result.Error != nil {
		return nil, fmt.Errorf("select object record's failed, %s", result.Error)
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

// ScanObjectInfo query scan DBObject, and convert to ObjectInfo.
func (jmi *JobMetaImpl) ScanObjectInfo(offset int, limit int) ([]*ptypes.ObjectInfo, error) {
	var (
		result       *gorm.DB
		queryReturns []DBObject
		objects      []*ptypes.ObjectInfo
	)

	result = jmi.db.Limit(limit).Offset(offset).Find(&queryReturns)
	if result.Error != nil {
		return objects, fmt.Errorf("select primary piece jobs failed, %s", result.Error)
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

// GetJobContext query DBJob by jobID, and convert to types.JobContext.
func (jmi *JobMetaImpl) GetJobContext(jobId uint64) (*ptypes.JobContext, error) {
	var (
		result         *gorm.DB
		queryCondition *DBJob
		queryReturn    DBJob
	)

	// If the primary key is a number, the query will be written as follows:
	queryCondition = &DBJob{
		JobID: jobId,
	}
	result = jmi.db.Model(queryCondition).First(&queryReturn)
	if result.Error != nil {
		return nil, fmt.Errorf("select job record's failed, %s", result.Error)
	}
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

// SetPrimaryPieceJobDone create primary DBPieceJob record.
func (jmi *JobMetaImpl) SetPrimaryPieceJobDone(txHash []byte, pj *jobdb.PieceJob) error {
	var (
		result               *gorm.DB
		insertPieceJobRecord *DBPieceJob
	)

	insertPieceJobRecord = &DBPieceJob{
		CreateHash:      hex.EncodeToString(txHash),
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

// GetPrimaryJob query DBPieceJob by txHash and primary type, and convert to PieceJob.
func (jmi *JobMetaImpl) GetPrimaryJob(txHash []byte) ([]*jobdb.PieceJob, error) {
	var (
		result       *gorm.DB
		queryReturns []DBPieceJob
		pieceJobs    []*jobdb.PieceJob
	)

	result = jmi.db.
		Where("create_hash = ? AND piece_type = ?", hex.EncodeToString(txHash), ptypes.JobType_JOB_TYPE_UPLOAD_PRIMARY).
		Find(&queryReturns)
	if result.Error != nil {
		return pieceJobs, fmt.Errorf("select primary piece jobs failed, %s", result.Error)
	}
	for _, job := range queryReturns {
		pieceJobs = append(pieceJobs, &jobdb.PieceJob{
			PieceId:         job.PieceIdx,
			Checksum:        [][]byte{[]byte(job.IntegrityHash)},
			StorageProvider: job.StorageProvider})
	}
	return pieceJobs, nil
}

// SetSecondaryPieceJobDone create secondary DBPieceJob record.
func (jmi *JobMetaImpl) SetSecondaryPieceJobDone(txHash []byte, pj *jobdb.PieceJob) error {
	var (
		result               *gorm.DB
		insertPieceJobRecord *DBPieceJob
	)

	insertPieceJobRecord = &DBPieceJob{
		CreateHash:      hex.EncodeToString(txHash),
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

// GetSecondaryJob query DBPieceJob by txHash and secondary type, and convert to PieceJob.
func (jmi *JobMetaImpl) GetSecondaryJob(txHash []byte) ([]*jobdb.PieceJob, error) {
	var (
		result       *gorm.DB
		queryReturns []DBPieceJob
		pieceJobs    []*jobdb.PieceJob
	)

	result = jmi.db.
		Where("create_hash = ? AND piece_type = ?", hex.EncodeToString(txHash), ptypes.JobType_JOB_TYPE_UPLOAD_SECONDARY_EC).
		Find(&queryReturns)
	if result.Error != nil {
		return pieceJobs, fmt.Errorf("select secondary piece jobs failed, %s", result.Error)
	}
	for _, job := range queryReturns {
		pieceJobs = append(pieceJobs, &jobdb.PieceJob{
			PieceId:         job.PieceIdx,
			Checksum:        [][]byte{[]byte(job.IntegrityHash)},
			StorageProvider: job.StorageProvider})
	}
	return pieceJobs, nil
}

// SetObjectCreateHeightAndObjectID set object height and id
func (jmi *JobMetaImpl) SetObjectCreateHeightAndObjectID(txHash []byte, height uint64, objectID uint64) error {
	var (
		result         *gorm.DB
		queryCondition *DBObject
		updateFields   *DBObject
	)

	queryCondition = &DBObject{
		CreateHash: hex.EncodeToString(txHash),
	}
	updateFields = &DBObject{
		Height:   height,
		ObjectID: objectID,
	}
	result = jmi.db.Model(queryCondition).Updates(updateFields)
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("update object record's height and objectid failed, %s", result.Error)
	}
	return nil
}

// v2 interface implement

// CreateUploadPayloadJobV2 create DBJobV2 record and DBObjectV2 record, Use JobID field for association.
func (jmi *JobMetaImpl) CreateUploadPayloadJobV2(info *ptypes.ObjectInfo) (uint64, error) {
	var (
		result             *gorm.DB
		insertJobRecord    *DBJobV2
		insertObjectRecord *DBObjectV2
	)

	insertJobRecord = &DBJobV2{
		JobType:    uint32(ptypes.JobType_JOB_TYPE_CREATE_OBJECT),
		JobState:   uint32(ptypes.JobState_JOB_STATE_CREATE_OBJECT_DONE),
		CreateTime: time.Now(),
		ModifyTime: time.Now(),
	}
	result = jmi.db.Create(insertJobRecord)
	if result.Error != nil || result.RowsAffected != 1 {
		return 0, fmt.Errorf("insert job record failed, %s", result.Error)
	}
	insertObjectRecord = &DBObjectV2{
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

// GetObjectInfoV2 query DBObjectV2 by txHash, and convert to types.ObjectInfo.
func (jmi *JobMetaImpl) GetObjectInfoV2(objectID uint64) (*ptypes.ObjectInfo, error) {
	var (
		result         *gorm.DB
		queryCondition *DBObjectV2
		queryReturn    DBObjectV2
	)

	// If the primary key is a number, the query will be written as follows:
	queryCondition = &DBObjectV2{
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

// GetJobContextV2 query DBJobV2 by jobID, and convert to types.JobContext.
func (jmi *JobMetaImpl) GetJobContextV2(jobID uint64) (*ptypes.JobContext, error) {
	var (
		result      *gorm.DB
		queryReturn DBJobV2
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

// SetUploadPayloadJobStateV2 update DBJobV2 record's state.
func (jmi *JobMetaImpl) SetUploadPayloadJobStateV2(jobId uint64, jobState string, timestampSec int64) error {
	var (
		result         *gorm.DB
		queryCondition *DBJobV2
		updateFields   *DBJobV2
	)

	queryCondition = &DBJobV2{
		JobID: jobId,
	}
	updateFields = &DBJobV2{
		JobState:   uint32(ptypes.JobState_value[jobState]),
		ModifyTime: time.Unix(timestampSec, 0),
	}
	result = jmi.db.Model(queryCondition).Updates(updateFields)
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("update job record's state failed, %s", result.Error)
	}
	return nil
}

// SetUploadPayloadJobJobErrorV2 update DBJobV2 record's state and err.
func (jmi *JobMetaImpl) SetUploadPayloadJobJobErrorV2(jobID uint64, jobState string, jobErr string, timestampSec int64) error {
	var (
		result         *gorm.DB
		queryCondition *DBJobV2
		updateFields   *DBJobV2
	)

	queryCondition = &DBJobV2{
		JobID: jobID,
	}
	updateFields = &DBJobV2{
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

// GetPrimaryJobV2 query DBPieceJobV2 by objectID and primary type, and convert to PieceJob.
func (jmi *JobMetaImpl) GetPrimaryJobV2(objectId uint64) ([]*jobdb.PieceJob, error) {
	var (
		result       *gorm.DB
		queryReturns []DBPieceJobV2
		pieceJobs    []*jobdb.PieceJob
	)

	result = jmi.db.
		Where("object_id = ? AND piece_type = ?", objectId, ptypes.JobType_JOB_TYPE_UPLOAD_PRIMARY).
		Find(&queryReturns)
	if result.Error != nil {
		return pieceJobs, fmt.Errorf("select primary piece jobs failed, %s", result.Error)
	}
	for _, job := range queryReturns {
		pieceJobs = append(pieceJobs, &jobdb.PieceJob{
			PieceId:         job.PieceIdx,
			Checksum:        [][]byte{[]byte(job.IntegrityHash)},
			StorageProvider: job.StorageProvider})
	}
	return pieceJobs, nil
}

// SetPrimaryPieceJobDoneV2 create primary DBPieceJobV2 record.
func (jmi *JobMetaImpl) SetPrimaryPieceJobDoneV2(objectID uint64, pj *jobdb.PieceJob) error {
	var (
		result               *gorm.DB
		insertPieceJobRecord *DBPieceJobV2
	)

	insertPieceJobRecord = &DBPieceJobV2{
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

// SetSecondaryPieceJobDoneV2 create secondary DBPieceJobV2 record.
func (jmi *JobMetaImpl) SetSecondaryPieceJobDoneV2(objectID uint64, pj *jobdb.PieceJob) error {
	var (
		result               *gorm.DB
		insertPieceJobRecord *DBPieceJobV2
	)

	insertPieceJobRecord = &DBPieceJobV2{
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

// GetSecondaryJobV2 query DBPieceJobV2 by objectID and secondary type, and convert to PieceJob.
func (jmi *JobMetaImpl) GetSecondaryJobV2(objectID uint64) ([]*jobdb.PieceJob, error) {
	var (
		result       *gorm.DB
		queryReturns []DBPieceJobV2
		pieceJobs    []*jobdb.PieceJob
	)

	result = jmi.db.
		Where("object_id = ? AND piece_type = ?", objectID, ptypes.JobType_JOB_TYPE_UPLOAD_SECONDARY_EC).
		Find(&queryReturns)
	if result.Error != nil {
		return pieceJobs, fmt.Errorf("select secondary piece jobs failed, %s", result.Error)
	}
	for _, job := range queryReturns {
		pieceJobs = append(pieceJobs, &jobdb.PieceJob{
			PieceId:         job.PieceIdx,
			Checksum:        [][]byte{[]byte(job.IntegrityHash)},
			StorageProvider: job.StorageProvider})
	}
	return pieceJobs, nil
}

// ScanObjectInfoV2 query scan DBObjectV2, and convert to ObjectInfo.
func (jmi *JobMetaImpl) ScanObjectInfoV2(offset int, limit int) ([]*ptypes.ObjectInfo, error) {
	var (
		result       *gorm.DB
		queryReturns []DBObjectV2
		objects      []*ptypes.ObjectInfo
	)

	result = jmi.db.Limit(limit).Offset(offset).Find(&queryReturns)
	if result.Error != nil {
		return objects, fmt.Errorf("select primary piece jobs failed, %s", result.Error)
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
