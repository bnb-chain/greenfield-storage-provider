package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	types "github.com/bnb-chain/inscription-storage-provider/pkg/types/v1"
)

// JobMetaImpl is an implement of JobDB interface
type JobMetaImpl struct {
	db *gorm.DB
}

func NewJobMetaImpl() (*JobMetaImpl, error) {
	db, err := InitDB(DefaultDBOption)
	if err != nil {
		return nil, err
	}
	return &JobMetaImpl{db: db}, nil
}

// CreateUploadPayloadJob create DBJob record and DBObject record, Use JobID field for association.
func (jmi *JobMetaImpl) CreateUploadPayloadJob(txHash []byte, info *types.ObjectInfo) error {
	var (
		result             *gorm.DB
		insertJobRecord    *DBJob
		insertObjectRecord *DBObject
	)

	insertJobRecord = &DBJob{
		JobType:    uint32(types.JobType_JOB_TYPE_CREATE_OBJECT),
		JobState:   uint32(types.JobState_JOB_STATE_CREATE_OBJECT_INIT),
		CreateTime: time.Now(),
		ModifyTime: time.Now(),
	}
	result = jmi.db.Create(insertJobRecord)
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("insert job record failed, %s", result.Error)
	}
	insertObjectRecord = &DBObject{
		CreateHash:     string(txHash),
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
		return fmt.Errorf("insert object record failed, %s", result.Error)
	}
	return nil
}

// SetObjectCreateHeight update DBObject record's height.
func (jmi *JobMetaImpl) SetObjectCreateHeight(txHash []byte, height uint64) error {
	var (
		result         *gorm.DB
		queryCondition *DBObject
		updateFields   *DBObject
	)

	queryCondition = &DBObject{
		CreateHash: string(txHash),
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
func (jmi *JobMetaImpl) GetObjectInfo(txHash []byte) (*types.ObjectInfo, error) {
	var (
		result      *gorm.DB
		queryReturn DBObject
	)

	// If the primary key is a string, the query will be written as follows:
	result = jmi.db.First(&queryReturn, "create_hash = ?", string(txHash))
	if result.Error != nil {
		return nil, fmt.Errorf("select object record's failed, %s", result.Error)
	}

	return &types.ObjectInfo{
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
		TxHash:         []byte(queryReturn.CreateHash),
		RedundancyType: types.RedundancyType(queryReturn.RedundancyType),
		SecondarySps:   nil, // todo: how to fill
	}, nil
}

// ScanObjectInfo query scan DBObject, and convert to ObjectInfo.
func (jmi *JobMetaImpl) ScanObjectInfo(offset int, limit int) ([]*types.ObjectInfo, error) {
	var (
		result       *gorm.DB
		queryReturns []DBObject
		objects      []*types.ObjectInfo
	)

	result = jmi.db.Limit(limit).Offset(offset).Find(&queryReturns)
	if result.Error != nil {
		return objects, fmt.Errorf("select primary piece jobs failed, %s", result.Error)
	}
	for _, object := range queryReturns {
		objects = append(objects, &types.ObjectInfo{
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
			TxHash:         []byte(object.CreateHash),
			RedundancyType: types.RedundancyType(object.RedundancyType),
			SecondarySps:   nil, // todo: how to fill
		})
	}
	return objects, nil
}

// GetJobContext query DBJob by jobID, and convert to types.JobContext.
func (jmi *JobMetaImpl) GetJobContext(jobId uint64) (*types.JobContext, error) {
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
	return &types.JobContext{
		JobId:      queryReturn.JobID,
		JobType:    types.JobType(queryReturn.JobType),
		JobState:   types.JobState(queryReturn.JobState),
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
		JobState:   uint32(types.JobState_value[jobState]),
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
		JobState:   uint32(types.JobState_value[jobState]),
		ModifyTime: time.Unix(timestampSec, 0),
		JobErr:     jobErr,
	}
	result = jmi.db.Model(queryCondition).Updates(updateFields)
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("update job record's state and err failed, %s", result.Error)
	}
	return nil
}

type PieceJob struct {
	PieceId         uint32
	Checksum        []byte
	StorageProvider string
}

// SetPrimaryPieceJobDone create primary DBPieceJob record.
func (jmi *JobMetaImpl) SetPrimaryPieceJobDone(txHash []byte, pj *PieceJob) error {
	var (
		result               *gorm.DB
		insertPieceJobRecord *DBPieceJob
	)

	insertPieceJobRecord = &DBPieceJob{
		CreateHash:      string(txHash),
		PieceType:       uint32(types.JobType_JOB_TYPE_UPLOAD_PRIMARY),
		PieceIdx:        pj.PieceId,
		CheckSum:        string(pj.Checksum),
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
func (jmi *JobMetaImpl) GetPrimaryJob(txHash []byte) ([]*PieceJob, error) {
	var (
		result       *gorm.DB
		queryReturns []DBPieceJob
		pieceJobs    []*PieceJob
	)

	result = jmi.db.
		Where("create_hash = ? AND piece_type = ?", string(txHash), types.JobType_JOB_TYPE_UPLOAD_PRIMARY).
		Find(&queryReturns)
	if result.Error != nil {
		return pieceJobs, fmt.Errorf("select primary piece jobs failed, %s", result.Error)
	}
	for _, job := range queryReturns {
		pieceJobs = append(pieceJobs, &PieceJob{
			PieceId:         job.PieceIdx,
			Checksum:        []byte(job.IntegrityHash),
			StorageProvider: job.StorageProvider})
	}
	return pieceJobs, nil
}

// SetSecondaryPieceJobDone create secondary DBPieceJob record.
func (jmi *JobMetaImpl) SetSecondaryPieceJobDone(txHash []byte, pj *PieceJob) error {
	var (
		result               *gorm.DB
		insertPieceJobRecord *DBPieceJob
	)

	insertPieceJobRecord = &DBPieceJob{
		CreateHash:      string(txHash),
		PieceType:       uint32(types.JobType_JOB_TYPE_UPLOAD_SECONDARY_EC),
		PieceIdx:        pj.PieceId,
		CheckSum:        string(pj.Checksum),
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
func (jmi *JobMetaImpl) GetSecondaryJob(txHash []byte) ([]*PieceJob, error) {
	var (
		result       *gorm.DB
		queryReturns []DBPieceJob
		pieceJobs    []*PieceJob
	)

	result = jmi.db.
		Where("create_hash = ? AND piece_type = ?", string(txHash), types.JobType_JOB_TYPE_UPLOAD_SECONDARY_EC).
		Find(&queryReturns)
	if result.Error != nil {
		return pieceJobs, fmt.Errorf("select secondary piece jobs failed, %s", result.Error)
	}
	for _, job := range queryReturns {
		pieceJobs = append(pieceJobs, &PieceJob{
			PieceId:         job.PieceIdx,
			Checksum:        []byte(job.IntegrityHash),
			StorageProvider: job.StorageProvider})
	}
	return pieceJobs, nil
}
