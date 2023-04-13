package sqldb

import (
	"errors"
	"time"

	"gorm.io/gorm"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	errorstypes "github.com/bnb-chain/greenfield-storage-provider/pkg/errors/types"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

// CheckQuotaAndAddReadRecord check current quota, and add read record
// TODO: Traffic statistics may be inaccurate in extreme cases, optimize it in the future
func (s *SpDBImpl) CheckQuotaAndAddReadRecord(record *ReadRecord, quota *BucketQuota) error {
	yearMonth := TimeToYearMonth(TimestampUsToTime(record.ReadTimestampUs))
	bucketTraffic, err := s.GetBucketTraffic(record.BucketID, yearMonth)
	if err != nil && errorstypes.Code(err) != merrors.DBRecordNotFoundErrCode {
		return err
	}
	if bucketTraffic == nil {
		// insert, if not existed
		insertBucketTraffic := &BucketTrafficTable{
			BucketID:         record.BucketID,
			Month:            yearMonth,
			BucketName:       record.BucketName,
			ReadConsumedSize: 0,
			ReadQuotaSize:    quota.ReadQuotaSize,
			ModifiedTime:     time.Now(),
		}
		result := s.db.Create(insertBucketTraffic)
		if result.Error != nil {
			return errorstypes.Error(merrors.DBInsertInBucketTrafficTableErrCode, result.Error.Error())
		}
		if result.RowsAffected != 1 {
			log.Infow("insert traffic", "RowsAffected", result.RowsAffected, "record", record, "quota", quota)
		}
		bucketTraffic = &BucketTraffic{
			BucketID:         insertBucketTraffic.BucketID,
			YearMonth:        insertBucketTraffic.Month,
			BucketName:       insertBucketTraffic.BucketName,
			ReadConsumedSize: insertBucketTraffic.ReadConsumedSize,
			ReadQuotaSize:    insertBucketTraffic.ReadQuotaSize,
		}
	}
	if bucketTraffic.ReadQuotaSize != quota.ReadQuotaSize {
		// update if chain quota has changed
		result := s.db.Model(&BucketTrafficTable{}).
			Where("bucket_id = ? and month = ?", bucketTraffic.BucketID, bucketTraffic.YearMonth).
			Updates(BucketTrafficTable{
				ReadQuotaSize: quota.ReadQuotaSize,
				ModifiedTime:  time.Now(),
			})
		if result.Error != nil {
			return errorstypes.Error(merrors.DBUpdateInBucketTrafficTableErrCode, result.Error.Error())
		}
		if result.RowsAffected != 1 {
			log.Infow("update traffic", "RowsAffected", result.RowsAffected, "record", record, "quota", quota)
		}
		bucketTraffic.ReadQuotaSize = quota.ReadQuotaSize
	}

	// check quota
	if bucketTraffic.ReadConsumedSize+record.ReadSize > quota.ReadQuotaSize {
		return errorstypes.Error(merrors.DBQuotaNotEnoughErrCode, merrors.ErrCheckQuotaEnough.Error())
	}

	// update bucket traffic
	result := s.db.Model(&BucketTrafficTable{}).
		Where("bucket_id = ? and month = ?", bucketTraffic.BucketID, bucketTraffic.YearMonth).
		Updates(BucketTrafficTable{
			ReadConsumedSize: bucketTraffic.ReadConsumedSize + record.ReadSize,
			ModifiedTime:     time.Now(),
		})
	if result.Error != nil {
		return errorstypes.Error(merrors.DBUpdateInBucketTrafficTableErrCode, result.Error.Error())
	}
	if result.RowsAffected != 1 {
		log.Infow("update traffic", "RowsAffected", result.RowsAffected, "record", record, "quota", quota)
	}

	// add read record
	insertReadRecord := &ReadRecordTable{
		BucketID:        record.BucketID,
		ObjectID:        record.ObjectID,
		UserAddress:     record.UserAddress,
		ReadTimestampUs: record.ReadTimestampUs,
		BucketName:      record.BucketName,
		ObjectName:      record.ObjectName,
		ReadSize:        record.ReadSize,
	}
	result = s.db.Create(insertReadRecord)
	if result.Error != nil || result.RowsAffected != 1 {
		return errorstypes.Error(merrors.DBInsertInReadRecordTableErrCode, result.Error.Error())
	}
	return nil
}

// GetBucketTraffic return bucket traffic info
func (s *SpDBImpl) GetBucketTraffic(bucketID uint64, yearMonth string) (*BucketTraffic, error) {
	var (
		result      *gorm.DB
		queryReturn BucketTrafficTable
	)

	result = s.db.Where("bucket_id = ? and month = ?", bucketID, yearMonth).First(&queryReturn)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errorstypes.Error(merrors.DBRecordNotFoundErrCode, result.Error.Error())
	}
	if result.Error != nil {
		return nil, errorstypes.Error(merrors.DBQueryInBucketTrafficTableErrCode, result.Error.Error())
	}
	return &BucketTraffic{
		BucketID:         queryReturn.BucketID,
		YearMonth:        queryReturn.Month,
		BucketName:       queryReturn.BucketName,
		ReadConsumedSize: queryReturn.ReadConsumedSize,
		ReadQuotaSize:    queryReturn.ReadQuotaSize,
		ModifyTime:       queryReturn.ModifiedTime.Unix(),
	}, nil
}

// GetReadRecord return record list by time range
func (s *SpDBImpl) GetReadRecord(timeRange *TrafficTimeRange) ([]*ReadRecord, error) {
	var (
		result       *gorm.DB
		records      []*ReadRecord
		queryReturns []ReadRecordTable
	)

	if timeRange.LimitNum <= 0 {
		result = s.db.Where("read_timestamp_us >= ? and read_timestamp_us < ?", timeRange.StartTimestampUs, timeRange.EndTimestampUs).
			Find(&queryReturns)
	} else {
		result = s.db.Where("read_timestamp_us >= ? and read_timestamp_us < ?", timeRange.StartTimestampUs, timeRange.EndTimestampUs).
			Limit(timeRange.LimitNum).Find(&queryReturns)
	}
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errorstypes.Error(merrors.DBRecordNotFoundErrCode, result.Error.Error())
	}
	if result.Error != nil {
		return records, errorstypes.Error(merrors.DBQueryInReadRecordTableErrCode, result.Error.Error())
	}
	for _, record := range queryReturns {
		records = append(records, &ReadRecord{
			BucketID:        record.BucketID,
			ObjectID:        record.ObjectID,
			UserAddress:     record.UserAddress,
			BucketName:      record.BucketName,
			ObjectName:      record.ObjectName,
			ReadSize:        record.ReadSize,
			ReadTimestampUs: record.ReadTimestampUs,
		})
	}
	return records, nil
}

// GetBucketReadRecord return bucket record list by time range
func (s *SpDBImpl) GetBucketReadRecord(bucketID uint64, timeRange *TrafficTimeRange) ([]*ReadRecord, error) {
	var (
		result       *gorm.DB
		records      []*ReadRecord
		queryReturns []ReadRecordTable
	)

	if timeRange.LimitNum <= 0 {
		result = s.db.Where("read_timestamp_us >= ? and read_timestamp_us < ? and bucket_id = ?",
			timeRange.StartTimestampUs, timeRange.EndTimestampUs, bucketID).
			Find(&queryReturns)
	} else {
		result = s.db.Where("read_timestamp_us >= ? and read_timestamp_us < ? and bucket_id = ?",
			timeRange.StartTimestampUs, timeRange.EndTimestampUs, bucketID).
			Limit(timeRange.LimitNum).Find(&queryReturns)
	}
	if result.Error != nil {
		return records, errorstypes.Error(merrors.DBQueryInReadRecordTableErrCode, result.Error.Error())
	}
	for _, record := range queryReturns {
		records = append(records, &ReadRecord{
			BucketID:        record.BucketID,
			ObjectID:        record.ObjectID,
			UserAddress:     record.UserAddress,
			BucketName:      record.BucketName,
			ObjectName:      record.ObjectName,
			ReadSize:        record.ReadSize,
			ReadTimestampUs: record.ReadTimestampUs,
		})
	}
	return records, nil
}

// GetObjectReadRecord return object record list by time range
func (s *SpDBImpl) GetObjectReadRecord(objectID uint64, timeRange *TrafficTimeRange) ([]*ReadRecord, error) {
	var (
		result       *gorm.DB
		records      []*ReadRecord
		queryReturns []ReadRecordTable
	)

	if timeRange.LimitNum <= 0 {
		result = s.db.Where("read_timestamp_us >= ? and read_timestamp_us < ? and object_id = ?",
			timeRange.StartTimestampUs, timeRange.EndTimestampUs, objectID).
			Find(&queryReturns)
	} else {
		result = s.db.Where("read_timestamp_us >= ? and read_timestamp_us < ? and object_id = ?",
			timeRange.StartTimestampUs, timeRange.EndTimestampUs, objectID).
			Limit(timeRange.LimitNum).Find(&queryReturns)
	}
	if result.Error != nil {
		return records, errorstypes.Error(merrors.DBQueryInReadRecordTableErrCode, result.Error.Error())
	}
	for _, record := range queryReturns {
		records = append(records, &ReadRecord{
			BucketID:        record.BucketID,
			ObjectID:        record.ObjectID,
			UserAddress:     record.UserAddress,
			BucketName:      record.BucketName,
			ObjectName:      record.ObjectName,
			ReadSize:        record.ReadSize,
			ReadTimestampUs: record.ReadTimestampUs,
		})
	}
	return records, nil
}

// GetUserReadRecord return user record list by time range
func (s *SpDBImpl) GetUserReadRecord(userAddress string, timeRange *TrafficTimeRange) ([]*ReadRecord, error) {
	var (
		result       *gorm.DB
		records      []*ReadRecord
		queryReturns []ReadRecordTable
	)

	if timeRange.LimitNum <= 0 {
		result = s.db.Where("read_timestamp_us >= ? and read_timestamp_us < ? and user_address = ?",
			timeRange.StartTimestampUs, timeRange.EndTimestampUs, userAddress).
			Find(&queryReturns)
	} else {
		result = s.db.Where("read_timestamp_us >= ? and read_timestamp_us < ? and user_address = ?",
			timeRange.StartTimestampUs, timeRange.EndTimestampUs, userAddress).
			Limit(timeRange.LimitNum).Find(&queryReturns)
	}
	if result.Error != nil {
		return records, errorstypes.Error(merrors.DBQueryInReadRecordTableErrCode, result.Error.Error())
	}
	for _, record := range queryReturns {
		records = append(records, &ReadRecord{
			BucketID:        record.BucketID,
			ObjectID:        record.ObjectID,
			UserAddress:     record.UserAddress,
			BucketName:      record.BucketName,
			ObjectName:      record.ObjectName,
			ReadSize:        record.ReadSize,
			ReadTimestampUs: record.ReadTimestampUs,
		})
	}
	return records, nil
}
