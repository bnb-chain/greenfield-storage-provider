package sqldb

import (
	"context"
	"errors"
	"fmt"
	"time"

	corespdb "github.com/bnb-chain/greenfield-storage-provider/core/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
	"github.com/bnb-chain/greenfield-storage-provider/pkg/metrics"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	err = s.updateConsumedQuota(record, quota)
	if err != nil {
		log.Errorw("failed to commit the transaction of updating bucketTraffic table, ", "error", err)
		return err
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
	result := s.db.Create(insertReadRecord)
	if result.Error != nil || result.RowsAffected != 1 {
		err = fmt.Errorf("failed to insert read record table: %s", result.Error)
		return err
	}
	return nil
}

// getUpdatedConsumedQuota compute the updated quota of traffic table by the incoming read cost and the newest record.
// it returns the updated consumed free quota,consumed charged quota and remained free quota
func getUpdatedConsumedQuota(recordQuotaCost, freeQuotaRemain, consumeFreeQuota, consumeChargedQuota, chargedQuota uint64) (uint64, uint64, uint64, bool, error) {
	// if remain free quota enough, just consume free quota
	needUpdateFreeQuota := false
	if recordQuotaCost < freeQuotaRemain {
		needUpdateFreeQuota = true
		consumeFreeQuota += recordQuotaCost
		freeQuotaRemain -= recordQuotaCost
	} else {
		// if free remain quota exist, consume all the free remain quota first
		if freeQuotaRemain > 0 {
			if freeQuotaRemain+chargedQuota < recordQuotaCost {
				return 0, 0, 0, false, ErrCheckQuotaEnough
			}
			needUpdateFreeQuota = true
			consumeFreeQuota += freeQuotaRemain
			// update the consumed charge quota by remained free quota
			// if read cost 5G, and the remained free quota is 2G, consumed charge quota should be 3G and remained free quota should be 0
			consumeQuota := recordQuotaCost - freeQuotaRemain
			consumeChargedQuota += consumeQuota
			freeQuotaRemain = uint64(0)
			log.CtxDebugw(context.Background(), "free quota has been exhausted", "consumed", consumeFreeQuota, "remained", freeQuotaRemain)
		} else {
			// free remain quota is zero, no need to consider the free quota
			// the consumeChargedQuota plus record cost need to be more than total charged quota
			if chargedQuota < consumeChargedQuota+recordQuotaCost {
				return 0, 0, 0, false, ErrCheckQuotaEnough
			}
			consumeChargedQuota += recordQuotaCost
		}
	}
	return consumeFreeQuota, consumeChargedQuota, freeQuotaRemain, needUpdateFreeQuota, nil
}

// updateConsumedQuota update the consumed quota of BucketTraffic table in the transaction way
func (s *SpDBImpl) updateConsumedQuota(record *corespdb.ReadRecord, quota *corespdb.BucketQuota) error {
	yearMonth := TimeToYearMonth(TimestampUsToTime(record.ReadTimestampUs))
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var bucketTraffic BucketTrafficTable
		var err error
		if err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("bucket_id = ? and month = ?", record.BucketID, yearMonth).First(&bucketTraffic).Error; err != nil {
			return fmt.Errorf("failed to query bucket traffic table: %v", err)
		}

		// if charged quota changed, update the new value
		if bucketTraffic.ChargedQuotaSize != quota.ChargedQuotaSize {
			result := tx.Model(&bucketTraffic).
				Updates(BucketTrafficTable{
					ChargedQuotaSize: quota.ChargedQuotaSize,
					ModifiedTime:     time.Now(),
				})

			if result.Error != nil {
				return fmt.Errorf("failed to update bucket traffic table: %s", result.Error)
			}
			if result.RowsAffected != 1 {
				return fmt.Errorf("update traffic of %s has affected more than one rows %d, "+
					"update charged quota %d", bucketTraffic.BucketName, result.RowsAffected, quota.ChargedQuotaSize)
			}
			log.CtxDebugw(context.Background(), "updated quota", "charged quota", quota.ChargedQuotaSize)
		}

		// compute the new consumed quota size to be updated by the newest record and the read cost size
		updatedConsumedFreeQuota, updatedConsumedChargedQuota, updatedRemainedFreeQuota, needUpdateFreeQuota, err := getUpdatedConsumedQuota(record.ReadSize,
			bucketTraffic.FreeQuotaSize, bucketTraffic.FreeQuotaConsumedSize,
			bucketTraffic.ReadConsumedSize, quota.ChargedQuotaSize)
		if err != nil {
			return err
		}

		if needUpdateFreeQuota {
			// it is needed to add select items if you need to update a value to zero in gorm db
			err = tx.Model(&bucketTraffic).
				Select("read_consumed_size", "free_quota_consumed_size", "free_quota_size", "modified_time").Updates(BucketTrafficTable{
				ReadConsumedSize:      updatedConsumedChargedQuota,
				FreeQuotaConsumedSize: updatedConsumedFreeQuota,
				FreeQuotaSize:         updatedRemainedFreeQuota,
				ModifiedTime:          time.Now(),
			}).Error
		} else {
			err = tx.Model(&bucketTraffic).Updates(BucketTrafficTable{
				ReadConsumedSize: updatedConsumedChargedQuota,
				ModifiedTime:     time.Now(),
			}).Error
		}
		if err != nil {
			return fmt.Errorf("failed to update bucket traffic table: %v", err)
		}

		return nil
	})

	if err != nil {
		log.CtxErrorw(context.Background(), "updated quota transaction fail", "error", err)
	}
	return err
}

// InitBucketTraffic init the bucket traffic table
func (s *SpDBImpl) InitBucketTraffic(record *corespdb.ReadRecord, quota *corespdb.BucketQuota) error {
	bucketID := record.BucketID
	bucketName := record.BucketName
	yearMonth := TimestampYearMonth(record.ReadTimestampUs)
	// if not created, init the bucket id in transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var insertBucketTraffic *BucketTrafficTable
		var bucketTraffic BucketTrafficTable
		result := s.db.Where("bucket_id = ?", bucketID).First(&bucketTraffic)
		if result.Error != nil {
			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return result.Error
			} else {
				// If the record of this bucket id does not exist, then the free quota consumed is initialized to 0
				insertBucketTraffic = &BucketTrafficTable{
					BucketID:              bucketID,
					Month:                 yearMonth,
					FreeQuotaSize:         quota.FreeQuotaSize,
					FreeQuotaConsumedSize: 0,
					BucketName:            bucketName,
					ReadConsumedSize:      0,
					ChargedQuotaSize:      quota.ChargedQuotaSize,
					ModifiedTime:          time.Now(),
				}
			}
		} else {
			// If the record of this bucket id already exist, then read the record of the newest month
			// and use the free quota consumed of this record to init free quota item
			var newestTraffic BucketTrafficTable
			queryErr := s.db.Where("bucket_id = ?", bucketID).Order("month DESC").Limit(1).Find(&newestTraffic).Error
			if queryErr != nil {
				return queryErr
			}

			insertBucketTraffic = &BucketTrafficTable{
				BucketID:              bucketID,
				Month:                 yearMonth,
				FreeQuotaSize:         newestTraffic.FreeQuotaSize,
				FreeQuotaConsumedSize: 0,
				BucketName:            bucketName,
				ReadConsumedSize:      0,
				ChargedQuotaSize:      quota.ChargedQuotaSize,
				ModifiedTime:          time.Now(),
			}
		}
		result = tx.Create(insertBucketTraffic)
		if result.Error != nil && MysqlErrCode(result.Error) != ErrDuplicateEntryCode {
			return fmt.Errorf("failed to create bucket traffic table: %s", result.Error)
		}

		return nil
	})

	if err != nil {
		log.CtxErrorw(context.Background(), "init traffic table error ", "bucket name", bucketName, "error", err)
	}
	return err
}

// GetBucketTraffic return bucket traffic info by the year and month info
// year_month is the query bucket quota's month, like "2023-03"
func (s *SpDBImpl) GetBucketTraffic(bucketID uint64, yearMonth string) (traffic *corespdb.BucketTraffic, err error) {
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

	result = s.db.Where("bucket_id = ? and month = ?", bucketID, yearMonth).First(&queryReturn)
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
		YearMonth:             queryReturn.Month,
		FreeQuotaSize:         queryReturn.FreeQuotaSize,
		FreeQuotaConsumedSize: queryReturn.FreeQuotaConsumedSize,
		BucketName:            queryReturn.BucketName,
		ReadConsumedSize:      queryReturn.ReadConsumedSize,
		ChargedQuotaSize:      queryReturn.ChargedQuotaSize,
		ModifyTime:            queryReturn.ModifiedTime.Unix(),
	}, nil
}

// UpdateExtraQuota update the read consumed quota and free consumed quota in traffic db with the extra quota
func (s *SpDBImpl) UpdateExtraQuota(bucketID, extraQuota uint64, yearMonth string) error {
	log.CtxErrorw(context.Background(), "begin to update extra quota for traffic db", "extra quota", extraQuota)
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var bucketTraffic BucketTrafficTable
		var err error
		yearMonthOfNow := TimestampYearMonth(GetCurrentTimestampUs())
		var extraUpdateOnNextMonth bool

		// In most cases, the month of extra quota compensation should be the same as the current month,
		// if not, the current month is the next month of the compensation month
		if IsNextMonth(yearMonthOfNow, yearMonth) {
			err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("bucket_id = ? and month = ?", bucketID, yearMonthOfNow).First(&bucketTraffic).Error
			if err != nil {
				// if it is a new month but the record has not been inserted, just update the extra quota to the old month
				// the new month will init by the newest old month record.
				if errors.Is(err, gorm.ErrRecordNotFound) {
					extraUpdateOnNextMonth = false
				}
			} else {
				extraUpdateOnNextMonth = true
			}
		} else {
			if yearMonthOfNow != yearMonth {
				return fmt.Errorf("the month of record of traffic table is invalid %s", yearMonth)
			}
			if err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("bucket_id = ? and month = ?", bucketID, yearMonth).First(&bucketTraffic).Error; err != nil {
				return fmt.Errorf("failed to query bucket traffic table: %v", err)
			}
		}

		consumedFreeQuota := bucketTraffic.FreeQuotaConsumedSize
		remainedFreeQuota := bucketTraffic.FreeQuotaSize
		log.CtxDebugw(context.Background(), "updated read consumed quota with extra quota:", "month", yearMonth, "consumed:", bucketTraffic.ReadConsumedSize,
			"remained free quota", remainedFreeQuota, "extra", extraQuota)

		if extraUpdateOnNextMonth {
			// if the extra quota generate on the different month, needed to add the extra quota to free quota
			// for example, the extra quota was generated at the last second on August 31, and the action of replenishing the quota began on September 1.
			// At this time, the latest quota record of traffic table is in September instead of August.
			// This situation is very unlikely to happen, in this case, the quota will be restored to the free quota.
			err = tx.Model(&bucketTraffic).
				Updates(BucketTrafficTable{
					FreeQuotaSize: bucketTraffic.FreeQuotaSize + extraQuota,
					ModifiedTime:  time.Now(),
				}).Error
		} else {
			// if the free quota has not exhaust even after consumed extra quota or the consumed charged quota is still zero,
			// the consumed free quota should be updated
			consumedChargeQuota := bucketTraffic.ReadConsumedSize
			if bucketTraffic.FreeQuotaSize > 0 || (bucketTraffic.FreeQuotaSize == 0 && consumedChargeQuota == 0) {
				err = tx.Model(&bucketTraffic).Select("free_quota_consumed_size", "free_quota_size", "modified_time").
					Updates(BucketTrafficTable{
						FreeQuotaConsumedSize: consumedFreeQuota - extraQuota,
						FreeQuotaSize:         remainedFreeQuota + extraQuota,
						ModifiedTime:          time.Now(),
					}).Error
			} else {
				// if consumed charge quota is more than extra quota, excess quota needs to be deducted from the consumed charge quota
				if consumedChargeQuota > extraQuota {
					err = tx.Model(&bucketTraffic).Select("read_consumed_size", "modified_time").
						Updates(BucketTrafficTable{
							ReadConsumedSize: consumedChargeQuota - extraQuota,
							ModifiedTime:     time.Now(),
						}).Error
				} else {
					extraFreeQuota := extraQuota - consumedChargeQuota
					// it is needed to add select items if you need to update a value to zero in gorm db
					err = tx.Model(&bucketTraffic).
						Select("read_consumed_size", "free_quota_consumed_size", "free_quota_size", "modified_time").Updates(BucketTrafficTable{
						ReadConsumedSize:      uint64(0),
						FreeQuotaConsumedSize: consumedFreeQuota - extraFreeQuota,
						FreeQuotaSize:         remainedFreeQuota + extraFreeQuota,
						ModifiedTime:          time.Now(),
					}).Error
				}
			}
		}

		if err != nil {
			return fmt.Errorf("failed to update bucket traffic table: %v", err)
		}

		return nil
	})

	if err != nil {
		log.CtxErrorw(context.Background(), "failed to update the table by extra quota ", "bucket id", bucketID, "error", err)
	}

	return err
}

// GetLatestBucketTraffic return the latest bucket traffic info of the bucket
func (s *SpDBImpl) GetLatestBucketTraffic(bucketID uint64) (traffic *corespdb.BucketTraffic, err error) {
	var queryReturn BucketTrafficTable

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

	err = s.db.Where("bucket_id = ?", bucketID).Order("month DESC").Limit(1).Find(&queryReturn).Error
	if err != nil {
		return nil, err
	}

	return &corespdb.BucketTraffic{
		BucketID:              queryReturn.BucketID,
		YearMonth:             queryReturn.Month,
		FreeQuotaSize:         queryReturn.FreeQuotaSize,
		FreeQuotaConsumedSize: queryReturn.FreeQuotaConsumedSize,
		BucketName:            queryReturn.BucketName,
		ReadConsumedSize:      queryReturn.ReadConsumedSize,
		ChargedQuotaSize:      queryReturn.ChargedQuotaSize,
		ModifyTime:            queryReturn.ModifiedTime.Unix(),
	}, nil
}

// UpdateBucketTraffic update the bucket traffic in traffic db with the new traffic
func (s *SpDBImpl) UpdateBucketTraffic(bucketID uint64, update *corespdb.BucketTraffic) (err error) {
	var (
		result      *gorm.DB
		queryReturn BucketTrafficTable
		needInsert  = false
	)

	result = s.db.Where("bucket_id = ? and month = ?", bucketID, update.YearMonth).First(&queryReturn)

	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	if result.Error != nil {
		needInsert = errors.Is(result.Error, gorm.ErrRecordNotFound)
	}

	updateRecord := &BucketTrafficTable{
		BucketID:              update.BucketID,
		Month:                 update.YearMonth,
		FreeQuotaSize:         update.FreeQuotaSize,
		FreeQuotaConsumedSize: update.FreeQuotaConsumedSize,
		BucketName:            update.BucketName,
		ReadConsumedSize:      update.ReadConsumedSize,
		ChargedQuotaSize:      update.ChargedQuotaSize,
		ModifiedTime:          time.Now(),
	}

	if needInsert {
		result = s.db.Create(updateRecord)
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("failed to insert record in bucket traffic table: %s", result.Error)
		}
	} else { // update
		result = s.db.Model(&BucketTrafficTable{}).
			Where("bucket_id = ? and month = ?", bucketID, update.YearMonth).Updates(updateRecord)
		if result.Error != nil {
			return fmt.Errorf("failed to update record in bucket traffic table: %s", result.Error)
		}
	}
	return nil
}

// DeleteAllBucketTrafficExpired update the bucket traffic in traffic db with the new traffic
func (s *SpDBImpl) DeleteAllBucketTrafficExpired(yearMonth string) (err error) {
	result := s.db.Where("month < ?", yearMonth).Delete(&BucketTrafficTable{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete bucket traffic record in bucket traffic table: %s, year_month:%s", result.Error, yearMonth)
	}
	return nil
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

// DeleteAllReadRecordExpired delete all read record before ts(ts is UnixMicro)
func (s *SpDBImpl) DeleteAllReadRecordExpired(ts uint64) (err error) {
	result := s.db.Where("read_timestamp_us < ?", ts).Delete(&ReadRecordTable{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete read record in read record table: %s, ts:%d", result.Error, ts)
	}
	return nil
}
