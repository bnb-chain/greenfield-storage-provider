package spdb

type TrafficDB interface {
	CheckQuotaAndAddReadRecord(record *ReadRecord, quota *BucketQuota) error
	GetBucketTraffic(bucketID uint64, yearMonth string) (*BucketTraffic, error)
	GetReadRecord(timeRange *TrafficTimeRange) ([]*ReadRecord, error)
	GetBucketReadRecord(bucketID uint64, timeRange *TrafficTimeRange) ([]*ReadRecord, error)
	GetObjectReadRecord(objectID uint64, timeRange *TrafficTimeRange) ([]*ReadRecord, error)
	GetUserReadRecord(userAddress string, timeRange *TrafficTimeRange) ([]*ReadRecord, error)
}

type ReadRecord struct {
	BucketID        uint64
	ObjectID        uint64
	UserAddress     string
	BucketName      string
	ObjectName      string
	ReadSize        uint64
	ReadTimestampUs int64
}

type BucketQuota struct {
	ReadQuotaSize uint64
}

type BucketTraffic struct {
	BucketID         uint64
	YearMonth        string // YearMonth is traffic's YearMonth, format "2023-02"
	BucketName       string
	ReadConsumedSize uint64
	ReadQuotaSize    uint64
	ModifyTime       int64
}

type TrafficTimeRange struct {
	StartTimestampUs int64
	EndTimestampUs   int64
	LimitNum         int // is unlimited if LimitNum <= 0
}
