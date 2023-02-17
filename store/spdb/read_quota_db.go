package spdb

type BucketReadQuota struct {
	BucketID      uint64 // primary index
	BucketName    string
	CostReadQuota int32 // cleared at the beginning of the month
	ModifyTime    int64
}

type BucketReadRecord struct {
	BucketID    uint64 // primary index
	BucketName  string
	UserAddress string
	ObjectName  string
	ObjectID    uint64
	ReadSize    int64
	ReadTime    int64
}

type ReadQuotaDB interface {
	AddBucket(quota *BucketReadQuota) error
	UpdateBucketQuota(quota *BucketReadQuota) error

	AddBucketReadRecord(record *BucketReadRecord) error
	GetBucketReadRecord(bucketID uint64) (*BucketReadRecord, error)
	ListBucketReadRecord(offset, length int) ([]*BucketReadRecord, error)
}
