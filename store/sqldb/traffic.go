package sqldb

import (
	"errors"
	"fmt"
	"time"

	merrors "github.com/bnb-chain/greenfield-storage-provider/model/errors"
	"gorm.io/gorm"
)

// CheckQuotaAndAddReadRecord check current quota, and add read record
func (s *SpDBImpl) CheckQuotaAndAddReadRecord(record *ReadRecord, quota *BucketQuota) error {
	yearMonth := TimeToYearMonth(TimestampUsToTime(record.ReadTimestampUs))
	bucketTraffic, err := s.GetBucketTraffic(record.BucketID, yearMonth)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
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
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("failed to insert bucket traffic table: %s", result.Error)
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
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("failed to update bucket traffic table: %s", result.Error)
		}
		bucketTraffic.ReadQuotaSize = quota.ReadQuotaSize
	}

	// check quota
	if bucketTraffic.ReadConsumedSize+record.ReadSize > quota.ReadQuotaSize {
		return merrors.ErrCheckQuotaEnough
	}

	// update bucket traffic
	result := s.db.Model(&BucketTrafficTable{}).
		Where("bucket_id = ? and month = ?", bucketTraffic.BucketID, bucketTraffic.YearMonth).
		Updates(BucketTrafficTable{
			ReadConsumedSize: bucketTraffic.ReadConsumedSize + record.ReadSize,
			ModifiedTime:     time.Now(),
		})
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("failed to update bucket traffic table: %s", result.Error)
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
		return fmt.Errorf("failed to insert read record table: %s", result.Error)
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
		return nil, result.Error
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query bucket traffic table: %s", result.Error)
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
		return nil, result.Error
	}
	if result.Error != nil {
		return records, fmt.Errorf("failed to read record table: %s", result.Error)
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
		return records, fmt.Errorf("failed to query read record table: %s", result.Error)
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
		return records, fmt.Errorf("failed to query read record table: %s", result.Error)
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
		return records, fmt.Errorf("failed to query read record table: %s", result.Error)
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
