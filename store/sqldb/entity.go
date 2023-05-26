package sqldb

import (
	"time"
)

//// IntegrityMeta defines the payload integrity hash and piece checksum with objectID
//type IntegrityMeta struct {
//	ObjectID      uint64
//	Checksum      [][]byte
//	IntegrityHash []byte
//	Signature     []byte
//}

// SpAddressType identify address type of SP
//type SpAddressType int32
//
//const (
//	OperatorAddressType SpAddressType = iota + 1
//	FundingAddressType
//	SealAddressType
//	ApprovalAddressType
//)

// BucketQuota defines read quota of a bucket
//type BucketQuota struct {
//	ReadQuotaSize uint64
//}

// BucketTraffic is record traffic by year and month
//type BucketTraffic struct {
//	BucketID         uint64
//	YearMonth        string // YearMonth is traffic's YearMonth, format "2023-02"
//	BucketName       string
//	ReadConsumedSize uint64
//	ReadQuotaSize    uint64
//	ModifyTime       int64
//}

// TrafficTimeRange is used by query, return records in [StartTimestampUs, EndTimestampUs)
//type TrafficTimeRange struct {
//	StartTimestampUs int64
//	EndTimestampUs   int64
//	LimitNum         int // is unlimited if LimitNum <= 0
//}

// ReadRecord defines a read request record, will decrease the bucket read quota
//type ReadRecord struct {
//	BucketID        uint64
//	ObjectID        uint64
//	UserAddress     string
//	BucketName      string
//	ObjectName      string
//	ReadSize        uint64
//	ReadTimestampUs int64
//}

// GetCurrentYearMonth get current year and month
func GetCurrentYearMonth() string {
	return TimeToYearMonth(time.Now())
}

// GetCurrentUnixTime return a second timestamp
func GetCurrentUnixTime() int64 {
	return time.Now().Unix()
}

// GetCurrentTimestampUs return a microsecond timestamp
func GetCurrentTimestampUs() int64 {
	return time.Now().UnixMicro()
}

// TimestampUsToTime convert a microsecond timestamp to time.Time
func TimestampUsToTime(ts int64) time.Time {
	tUnix := ts / int64(time.Millisecond)
	tUnixNanoRemainder := (ts % int64(time.Millisecond)) * int64(time.Microsecond)
	return time.Unix(tUnix, tUnixNanoRemainder)
}

// TimestampSecToTime convert a second timestamp to time.Time
func TimestampSecToTime(timeUnix int64) time.Time {
	return time.Unix(timeUnix, 0)
}

// TimeToYearMonth convent time.Time to YYYY-MM string
func TimeToYearMonth(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")[0:7]
}
