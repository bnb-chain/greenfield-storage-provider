package trafficsql

import (
	"fmt"
	"time"

	"github.com/bnb-chain/greenfield-storage-provider/store/config"
	"github.com/bnb-chain/greenfield-storage-provider/store/spdb"
	"github.com/bnb-chain/greenfield-storage-provider/util/log"
	"gorm.io/gorm"
)

var _ spdb.TrafficDB = &TrafficSQLDB{}

// TrafficSQLDB is an implement of TrafficDB interface
type TrafficSQLDB struct {
	db *gorm.DB
}

// NewTrafficSQLDB return a database instance
func NewTrafficSQLDB(config *config.SqlDBConfig) (*TrafficSQLDB, error) {
	db, err := InitDB(config)
	if err != nil {
		return nil, err
	}
	return &TrafficSQLDB{db: db}, nil
}

// CheckQuotaAndAddReadRecord check current quota, and add read record
func (t *TrafficSQLDB) CheckQuotaAndAddReadRecord(record *spdb.ReadRecord, quota *spdb.BucketQuota) error {
	var (
		result        *gorm.DB
		bucketTraffic *spdb.BucketTraffic
	)

	yearMonth := spdb.Time2YearMonth(spdb.TimeUnix2Time(record.ReadTime))
	bucketTraffic, err := t.GetBucketTraffic(record.BucketID, yearMonth)
	if err != nil {
		return err
	}
	if bucketTraffic == nil {
		// insert, if not existed
		insertBucketTraffic := &DBBucketTraffic{
			BucketID:      record.BucketID,
			YearMonth:     yearMonth,
			BucketName:    record.BucketName,
			ReadCostSize:  0,
			ReadQuotaSize: quota.ReadQuotaSize,
			ModifyTime:    time.Now(),
		}
		result = t.db.Create(insertBucketTraffic)
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("insert bucket traffic failed, %s", result.Error)
		}
		log.Infow("insert bucket traffic to db", "bucket_traffic", insertBucketTraffic)
		bucketTraffic.BucketID = insertBucketTraffic.BucketID
		bucketTraffic.YearMonth = insertBucketTraffic.YearMonth
		bucketTraffic.BucketName = insertBucketTraffic.BucketName
		bucketTraffic.ReadCostSize = insertBucketTraffic.ReadCostSize
		bucketTraffic.ReadQuotaSize = insertBucketTraffic.ReadQuotaSize
	}
	if bucketTraffic.ReadQuotaSize != quota.ReadQuotaSize {
		// update if chain quota has changed
		result = t.db.Model(&DBBucketTraffic{}).
			Where("bucket_id = ? and year_month = ?", bucketTraffic.BucketID, bucketTraffic.YearMonth).
			Updates(DBBucketTraffic{
				ReadQuotaSize: quota.ReadQuotaSize,
				ModifyTime:    time.Now(),
			})
		if result.Error != nil || result.RowsAffected != 1 {
			return fmt.Errorf("update update bucket traffic failed, %s", result.Error)
		}
		log.Infow("update bucket traffic quota", "from",
			bucketTraffic.ReadQuotaSize, "to", quota.ReadQuotaSize)
		bucketTraffic.ReadQuotaSize = quota.ReadQuotaSize
	}

	// check quota
	if bucketTraffic.ReadCostSize+record.ReadSize > bucketTraffic.ReadQuotaSize {
		log.Infow("failed to check quota", "bucket_traffic", bucketTraffic, "read_record", record)
		// TODO: check this error in outside
		return fmt.Errorf("failed to check quota")
	}

	// update bucket traffic
	result = t.db.Model(&DBBucketTraffic{}).
		Where("bucket_id = ? and year_month = ?", bucketTraffic.BucketID, bucketTraffic.YearMonth).
		Updates(DBBucketTraffic{
			ReadCostSize: bucketTraffic.ReadCostSize + record.ReadSize,
			ModifyTime:   time.Now(),
		})
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("update bucket traffic failed, %s", result.Error)
	}
	log.Infow("update bucket traffic cost size", "from", bucketTraffic.ReadCostSize,
		"to", bucketTraffic.ReadCostSize+record.ReadSize)

	// add read record
	insertReadRecord := &DBReadRecord{
		BucketID:    record.BucketID,
		ObjectID:    record.ObjectID,
		UserAddress: record.UserAddress,
		ReadTime:    record.ReadTime,
		BucketName:  record.BucketName,
		ObjectName:  record.ObjectName,
		ReadSize:    record.ReadSize,
	}
	result = t.db.Create(insertReadRecord)
	if result.Error != nil || result.RowsAffected != 1 {
		return fmt.Errorf("insert read record failed, %s", result.Error)
	}
	log.Infow("insert read record to db", "bucket_traffic", insertReadRecord)
	return nil
}

// GetBucketTraffic return bucket traffic info
func (t *TrafficSQLDB) GetBucketTraffic(bucketID uint64, yearMonth string) (*spdb.BucketTraffic, error) {
	var (
		result      *gorm.DB
		queryReturn DBBucketTraffic
	)

	result = t.db.First(&queryReturn, "bucket_id = ? and year_month = ?", bucketID, yearMonth)
	if result.Error == gorm.ErrRecordNotFound {
		// not found
		return nil, nil
	}
	if result.Error != nil {
		return nil, fmt.Errorf("select bucket traffic failed, %s", result.Error)
	}
	return &spdb.BucketTraffic{
		BucketID:      queryReturn.BucketID,
		YearMonth:     queryReturn.YearMonth,
		BucketName:    queryReturn.BucketName,
		ReadCostSize:  queryReturn.ReadCostSize,
		ReadQuotaSize: queryReturn.ReadQuotaSize,
		ModifyTime:    queryReturn.ModifyTime.Unix(),
	}, nil
}

// GetReadRecord return record list by time range
func (t *TrafficSQLDB) GetReadRecord(timeRange *spdb.TrafficTimeRange) ([]*spdb.ReadRecord, error) {
	var (
		result       *gorm.DB
		records      []*spdb.ReadRecord
		queryReturns []DBReadRecord
	)

	if timeRange.LimitNum <= 0 {
		result = t.db.Where("read_time >= ? and read_time < ?", timeRange.StartTime, timeRange.EndTime).
			Find(&queryReturns)
	} else {
		result = t.db.Where("read_time >= ? and read_time < ?", timeRange.StartTime, timeRange.EndTime).
			Limit(timeRange.LimitNum).Find(&queryReturns)
	}
	if result.Error != nil {
		return records, fmt.Errorf("select read records failed, %s", result.Error)
	}
	for _, record := range queryReturns {
		records = append(records, &spdb.ReadRecord{
			BucketID:    record.BucketID,
			ObjectID:    record.ObjectID,
			UserAddress: record.UserAddress,
			BucketName:  record.BucketName,
			ObjectName:  record.ObjectName,
			ReadSize:    record.ReadSize,
			ReadTime:    record.ReadTime,
		})
	}
	return records, nil
}

// GetBucketReadRecord return bucket record list by time range
func (t *TrafficSQLDB) GetBucketReadRecord(bucketID uint64, timeRange *spdb.TrafficTimeRange) ([]*spdb.ReadRecord, error) {
	var (
		result       *gorm.DB
		records      []*spdb.ReadRecord
		queryReturns []DBReadRecord
	)

	if timeRange.LimitNum <= 0 {
		result = t.db.Where("read_time >= ? and read_time < ? and bucket_id = ?",
			timeRange.StartTime, timeRange.EndTime, bucketID).
			Find(&queryReturns)
	} else {
		result = t.db.Where("read_time >= ? and read_time < ? and bucket_id = ?",
			timeRange.StartTime, timeRange.EndTime, bucketID).
			Limit(timeRange.LimitNum).Find(&queryReturns)
	}
	if result.Error != nil {
		return records, fmt.Errorf("select bucket read records failed, %s", result.Error)
	}
	for _, record := range queryReturns {
		records = append(records, &spdb.ReadRecord{
			BucketID:    record.BucketID,
			ObjectID:    record.ObjectID,
			UserAddress: record.UserAddress,
			BucketName:  record.BucketName,
			ObjectName:  record.ObjectName,
			ReadSize:    record.ReadSize,
			ReadTime:    record.ReadTime,
		})
	}
	return records, nil
}

// GetObjectReadRecord return object record list by time range
func (t *TrafficSQLDB) GetObjectReadRecord(objectID uint64, timeRange *spdb.TrafficTimeRange) ([]*spdb.ReadRecord, error) {
	var (
		result       *gorm.DB
		records      []*spdb.ReadRecord
		queryReturns []DBReadRecord
	)

	if timeRange.LimitNum <= 0 {
		result = t.db.Where("read_time >= ? and read_time < ? and object_id = ?",
			timeRange.StartTime, timeRange.EndTime, objectID).
			Find(&queryReturns)
	} else {
		result = t.db.Where("read_time >= ? and read_time < ? and object_id = ?",
			timeRange.StartTime, timeRange.EndTime, objectID).
			Limit(timeRange.LimitNum).Find(&queryReturns)
	}
	if result.Error != nil {
		return records, fmt.Errorf("select object read records failed, %s", result.Error)
	}
	for _, record := range queryReturns {
		records = append(records, &spdb.ReadRecord{
			BucketID:    record.BucketID,
			ObjectID:    record.ObjectID,
			UserAddress: record.UserAddress,
			BucketName:  record.BucketName,
			ObjectName:  record.ObjectName,
			ReadSize:    record.ReadSize,
			ReadTime:    record.ReadTime,
		})
	}
	return records, nil
}

// GetUserReadRecord return user record list by time range
func (t *TrafficSQLDB) GetUserReadRecord(userAddress string, timeRange *spdb.TrafficTimeRange) ([]*spdb.ReadRecord, error) {
	var (
		result       *gorm.DB
		records      []*spdb.ReadRecord
		queryReturns []DBReadRecord
	)

	if timeRange.LimitNum <= 0 {
		result = t.db.Where("read_time >= ? and read_time < ? and user_address = ?",
			timeRange.StartTime, timeRange.EndTime, userAddress).
			Find(&queryReturns)
	} else {
		result = t.db.Where("read_time >= ? and read_time < ? and user_address = ?",
			timeRange.StartTime, timeRange.EndTime, userAddress).
			Limit(timeRange.LimitNum).Find(&queryReturns)
	}
	if result.Error != nil {
		return records, fmt.Errorf("select user read records failed, %s", result.Error)
	}
	for _, record := range queryReturns {
		records = append(records, &spdb.ReadRecord{
			BucketID:    record.BucketID,
			ObjectID:    record.ObjectID,
			UserAddress: record.UserAddress,
			BucketName:  record.BucketName,
			ObjectName:  record.ObjectName,
			ReadSize:    record.ReadSize,
			ReadTime:    record.ReadTime,
		})
	}
	return records, nil
}
