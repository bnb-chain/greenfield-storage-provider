package spdb

// ReadRecord defines a read request record, will decrease the bucket read quota
type ReadRecord struct {
	BucketID        uint64
	ObjectID        uint64
	UserAddress     string
	BucketName      string
	ObjectName      string
	ReadSize        uint64
	ReadTimestampUs int64
}

// BucketQuota defines read quota of a bucket
type BucketQuota struct {
	ReadQuotaSize uint64
}

// BucketTraffic is record traffic by year and month
type BucketTraffic struct {
	BucketID         uint64
	YearMonth        string // YearMonth is traffic's YearMonth, format "2023-02"
	BucketName       string
	ReadConsumedSize uint64
	ReadQuotaSize    uint64
	ModifyTime       int64
}

// TrafficTimeRange is used by query, return records in [StartTimestampUs, EndTimestampUs)
type TrafficTimeRange struct {
	StartTimestampUs int64
	EndTimestampUs   int64
	LimitNum         int // is unlimited if LimitNum <= 0
}
