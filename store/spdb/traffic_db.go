package spdb

import "time"

// BucketQuota is a quota config from chain
type BucketQuota struct {
	ReadQuotaSize int64
}

// BucketTraffic is record traffic by year and month
type BucketTraffic struct {
	BucketID      uint64
	YearMonth     string // YearMonth is traffic's YearMonth, format "2023-02"
	BucketName    string
	ReadCostSize  int64
	ReadQuotaSize int64
	ModifyTime    int64
}

// TrafficTimeRange is used by query, return records in [StartTime, EndTime)
type TrafficTimeRange struct {
	StartTime int64
	EndTime   int64
	LimitNum  int // is unlimited if LimitNum <= 0
}

// ReadRecord is a read request
type ReadRecord struct {
	BucketID    uint64
	ObjectID    uint64
	UserAddress string
	BucketName  string
	ObjectName  string
	ReadSize    int64
	ReadTime    int64
}

func GetCurrentYearMonth() string {
	return Time2YearMonth(time.Now())
}

// GetNowTimeUnix return a second timestamp
func GetNowTimeUnix() int64 {
	return time.Now().Unix()
}

// TimeUnix2Time convent a second timestamp to time.Time
func TimeUnix2Time(timeUnix int64) time.Time {
	return time.Unix(timeUnix, 0)
}

// Time2YearMonth convent time.Time to YYYY-MM string
func Time2YearMonth(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")[0:6]
}

// TrafficDB define a series of traffic interfaces
type TrafficDB interface {
	// CheckQuotaAndAddReadRecord create bucket traffic firstly if bucket is not existed,
	// and check whether the added traffic record exceeds the quota, if it exceeds the quota,
	// it will return error, Otherwise, add a record and return nil.
	CheckQuotaAndAddReadRecord(record *ReadRecord, quota *BucketQuota) error

	// GetBucketTraffic return bucket traffic info,
	// notice maybe return (nil, nil) while there is no bucket traffic
	GetBucketTraffic(bucketID uint64, yearMonth string) (*BucketTraffic, error)

	// GetReadRecord return record list by time range
	GetReadRecord(timeRange *TrafficTimeRange) ([]*ReadRecord, error)

	// GetBucketReadRecord return bucket record list by time range
	GetBucketReadRecord(bucketID uint64, timeRange *TrafficTimeRange) ([]*ReadRecord, error)

	// GetObjectReadRecord return object record list by time range
	GetObjectReadRecord(objectID uint64, timeRange *TrafficTimeRange) ([]*ReadRecord, error)

	// GetUserReadRecord return user record list by time range
	GetUserReadRecord(userAddress string, timeRange *TrafficTimeRange) ([]*ReadRecord, error)
}
