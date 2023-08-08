package sqldb

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
)

const (
	// SPDBSuccessCheckQuotaAndAddReadRecord defines the metrics label of successfully check and add read record
	SPDBSuccessCheckQuotaAndAddReadRecord = "check_and_add_read_record_success"
	// SPDBFailureCheckQuotaAndAddReadRecord defines the metrics label of unsuccessfully check and add read record
	SPDBFailureCheckQuotaAndAddReadRecord = "check_and_add_read_record_failure"
	// SPDBSuccessGetBucketTraffic defines the metrics label of successfully get bucket traffic
	SPDBSuccessGetBucketTraffic = "get_bucket_traffic_success"
	// SPDBFailureGetBucketTraffic defines the metrics label of unsuccessfully get bucket traffic
	SPDBFailureGetBucketTraffic = "get_bucket_traffic_failure"
	// SPDBSuccessGetReadRecord defines the metrics label of successfully get read record
	SPDBSuccessGetReadRecord = "get_read_record_success"
	// SPDBFailureGetReadRecord defines the metrics label of unsuccessfully get read record
	SPDBFailureGetReadRecord = "get_read_record_failure"
	// SPDBSuccessGetBucketReadRecord defines the metrics label of successfully get bucket read record
	SPDBSuccessGetBucketReadRecord = "get_bucket_read_record_success"
	// SPDBFailureGetBucketReadRecord defines the metrics label of unsuccessfully get bucket read record
	SPDBFailureGetBucketReadRecord = "get_bucket_read_record_failure"
	// SPDBSuccessGetObjectReadRecord defines the metrics label of successfully get object read record
	SPDBSuccessGetObjectReadRecord = "get_object_read_record_success"
	// SPDBFailureGetObjectReadRecord defines the metrics label of unsuccessfully get object read record
	SPDBFailureGetObjectReadRecord = "get_object_read_record_failure"
	// SPDBSuccessGetUserReadRecord defines the metrics label of successfully get user read record
	SPDBSuccessGetUserReadRecord = "get_user_read_record_success"
	// SPDBFailureGetUserReadRecord defines the metrics label of unsuccessfully get user read record
	SPDBFailureGetUserReadRecord = "get_user_read_record_failure"
)

// CheckQuotaAndAddReadRecord check current quota, and add read record
// TODO: Traffic statistics may be inaccurate in extreme cases, optimize it in the future
func (s *SpDBImpl) CheckQuotaAndAddReadRecord(record *corespdb.ReadRecord, quota *corespdb.BucketQuota) (err error) {
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureCheckQuotaAndAddReadRecord).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureCheckQuotaAndAddReadRecord).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessCheckQuotaAndAddReadRecord).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessCheckQuotaAndAddReadRecord).Observe(
			time.Since(startTime).Seconds())
	}()

	bucketTraffic, err := s.GetBucketTraffic(record.BucketID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	// bucket traffic table should be initialized before checking meta
	if bucketTraffic == nil {
		err = fmt.Errorf("failed to get bucket traffic table")
		return err
	}

	// the ChargedQuotaSize
	if bucketTraffic.ChargedQuotaSize != quota.ChargedQuotaSize {
		// update if chain quota has changed
		result := s.db.Model(&BucketTrafficTable{}).
			Where("bucket_id = ?", bucketTraffic.BucketID).
			Updates(BucketTrafficTable{
				ChargedQuotaSize: quota.ChargedQuotaSize,
				ModifiedTime:     time.Now(),
			})
		if result.Error != nil {
			err = fmt.Errorf("failed to update bucket traffic table: %s", result.Error)
			return err
		}
		// TODO:: change the transaction way to update
		if result.RowsAffected != 1 {
			log.Infow("update traffic", "RowsAffected", result.RowsAffected, "record", record, "quota", quota)
		}
		bucketTraffic.ChargedQuotaSize = quota.ChargedQuotaSize
		log.Infow("updated charged quota", "db quota:", quota.ChargedQuotaSize)
	}

	recordQuotaCost := record.ReadSize
	needCheckChainQuota := true
	freeQuotaRemain := bucketTraffic.FreeQuotaSize - bucketTraffic.FreeQuotaConsumedSize
	// if remain free quota more than 0, consume free quota first
	if freeQuotaRemain > 0 && recordQuotaCost < freeQuotaRemain {
		// if free quota is enough, no need to check charged quota
		bucketTraffic.ReadConsumedSize += recordQuotaCost
		bucketTraffic.FreeQuotaConsumedSize += recordQuotaCost
		needCheckChainQuota = false
	}
	// if free quota is not enough, check the charged quota
	if needCheckChainQuota {
		if bucketTraffic.ReadConsumedSize+recordQuotaCost > bucketTraffic.ChargedQuotaSize+bucketTraffic.FreeQuotaSize {
			return ErrCheckQuotaEnough
		}
		bucketTraffic.ReadConsumedSize += recordQuotaCost
		if freeQuotaRemain > 0 {
			bucketTraffic.FreeQuotaConsumedSize += freeQuotaRemain
		}
	}

	// update bucket traffic
	result := s.db.Model(&BucketTrafficTable{}).
		Where("bucket_id = ?", bucketTraffic.BucketID).
		Updates(BucketTrafficTable{
			ReadConsumedSize:      bucketTraffic.ReadConsumedSize,
			FreeQuotaConsumedSize: bucketTraffic.FreeQuotaConsumedSize,
			ModifiedTime:          time.Now(),
		})
	if result.Error != nil {
		err = fmt.Errorf("failed to update bucket traffic table: %s", result.Error)
		return err
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
		err = fmt.Errorf("failed to insert read record table: %s", result.Error)
		return err
	}
	return nil
}

func (s *SpDBImpl) InitBucketTraffic(bucketID uint64, bucketName string, quota *corespdb.BucketQuota) error {
	insertBucketTraffic := &BucketTrafficTable{
		BucketID:              bucketID,
		FreeQuotaSize:         quota.FreeQuotaSize,
		FreeQuotaConsumedSize: 0,
		BucketName:            bucketName,
		ReadConsumedSize:      0,
		ChargedQuotaSize:      quota.ChargedQuotaSize,
		ModifiedTime:          time.Now(),
	}
	result := s.db.Create(insertBucketTraffic)
	if result.Error != nil && MysqlErrCode(result.Error) != ErrDuplicateEntryCode {
		err := fmt.Errorf("failed to create bucket traffic table: %s", result.Error)
		return err
	}

	return nil
}

// GetBucketTraffic return bucket traffic info
func (s *SpDBImpl) GetBucketTraffic(bucketID uint64) (traffic *corespdb.BucketTraffic, err error) {
	var (
		result      *gorm.DB
		queryReturn BucketTrafficTable
	)

	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureGetBucketTraffic).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureGetBucketTraffic).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessGetBucketTraffic).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessGetBucketTraffic).Observe(
			time.Since(startTime).Seconds())
	}()

	result = s.db.Where("bucket_id = ?", bucketID).First(&queryReturn)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		err = result.Error
		return nil, err
	}
	if result.Error != nil {
		err = fmt.Errorf("failed to query bucket traffic table: %s", result.Error)
		return nil, err
	}
	return &corespdb.BucketTraffic{
		BucketID:              queryReturn.BucketID,
		FreeQuotaSize:         queryReturn.FreeQuotaSize,
		FreeQuotaConsumedSize: queryReturn.FreeQuotaConsumedSize,
		BucketName:            queryReturn.BucketName,
		ReadConsumedSize:      queryReturn.ReadConsumedSize,
		ChargedQuotaSize:      queryReturn.ChargedQuotaSize,
		ModifyTime:            queryReturn.ModifiedTime.Unix(),
	}, nil
}

// GetReadRecord return record list by time range
func (s *SpDBImpl) GetReadRecord(timeRange *corespdb.TrafficTimeRange) (records []*corespdb.ReadRecord, err error) {
	var (
		result       *gorm.DB
		queryReturns []ReadRecordTable
	)
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureGetReadRecord).Inc()
			metrics.SPDBTime.WithLabelValues(SPDBFailureGetReadRecord).Observe(
				time.Since(startTime).Seconds())
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessGetReadRecord).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessGetReadRecord).Observe(
			time.Since(startTime).Seconds())
	}()

	if timeRange.LimitNum <= 0 {
		result = s.db.Where("read_timestamp_us >= ? and read_timestamp_us < ?", timeRange.StartTimestampUs, timeRange.EndTimestampUs).
			Find(&queryReturns)
	} else {
		result = s.db.Where("read_timestamp_us >= ? and read_timestamp_us < ?", timeRange.StartTimestampUs, timeRange.EndTimestampUs).
			Limit(timeRange.LimitNum).Find(&queryReturns)
	}
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		err = result.Error
		return nil, err
	}
	if result.Error != nil {
		err = fmt.Errorf("failed to read record table: %s", result.Error)
		return records, err
	}
	for _, record := range queryReturns {
		records = append(records, &corespdb.ReadRecord{
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
func (s *SpDBImpl) GetBucketReadRecord(bucketID uint64, timeRange *corespdb.TrafficTimeRange) (records []*corespdb.ReadRecord, err error) {
	var (
		result       *gorm.DB
		queryReturns []ReadRecordTable
	)
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureGetBucketReadRecord).Inc()
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessGetBucketReadRecord).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessGetBucketReadRecord).Observe(
			time.Since(startTime).Seconds())
	}()

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
		err = fmt.Errorf("failed to query read record table: %s", result.Error)
		return records, err
	}
	for _, record := range queryReturns {
		records = append(records, &corespdb.ReadRecord{
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
func (s *SpDBImpl) GetObjectReadRecord(objectID uint64, timeRange *corespdb.TrafficTimeRange) (records []*corespdb.ReadRecord, err error) {
	var (
		result       *gorm.DB
		queryReturns []ReadRecordTable
	)
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureGetObjectReadRecord).Inc()
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessGetObjectReadRecord).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessGetObjectReadRecord).Observe(
			time.Since(startTime).Seconds())
	}()

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
		err = fmt.Errorf("failed to query read record table: %s", result.Error)
		return records, err
	}
	for _, record := range queryReturns {
		records = append(records, &corespdb.ReadRecord{
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
func (s *SpDBImpl) GetUserReadRecord(userAddress string, timeRange *corespdb.TrafficTimeRange) (records []*corespdb.ReadRecord, err error) {
	var (
		result       *gorm.DB
		queryReturns []ReadRecordTable
	)
	startTime := time.Now()
	defer func() {
		if err != nil {
			metrics.SPDBCounter.WithLabelValues(SPDBFailureGetUserReadRecord).Inc()
			return
		}
		metrics.SPDBCounter.WithLabelValues(SPDBSuccessGetUserReadRecord).Inc()
		metrics.SPDBTime.WithLabelValues(SPDBSuccessGetUserReadRecord).Observe(
			time.Since(startTime).Seconds())
	}()

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
		err = fmt.Errorf("failed to query read record table: %s", result.Error)
		return records, err
	}
	for _, record := range queryReturns {
		records = append(records, &corespdb.ReadRecord{
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
