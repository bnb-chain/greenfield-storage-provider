package sqldb

import "time"

// IntegrityMeta defines the payload integrity hash and piece checksum with objectID
type IntegrityMeta struct {
	ObjectID      uint64
	Checksum      [][]byte
	IntegrityHash []byte
	Signature     []byte
}

// SpAddressType identify SP's address type
type SpAddressType int32

const (
	OperatorAddressType SpAddressType = iota + 1
	FundingAddressType
	SealAddressType
	ApprovalAddressType
)

// BucketQuota defines read quota of a bucket
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

// ReadRecord defines a read request record, will decrease the bucket read quota
type ReadRecord struct {
	BucketID    uint64
	ObjectID    uint64
	UserAddress string
	BucketName  string
	ObjectName  string
	ReadSize    int64
	ReadTime    int64
}

// GetCurrentYearMonth get current year and month
func GetCurrentYearMonth() string {
	return TimeToYearMonth(time.Now())
}

// GetCurrentUnixTime return a second timestamp
func GetCurrentUnixTime() int64 {
	return time.Now().Unix()
}

// TimeUnixToTime convert a second timestamp to time.Time
func TimeUnixToTime(timeUnix int64) time.Time {
	return time.Unix(timeUnix, 0)
}

// TimeToYearMonth convent time.Time to YYYY-MM string
func TimeToYearMonth(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")[0:7]
}
