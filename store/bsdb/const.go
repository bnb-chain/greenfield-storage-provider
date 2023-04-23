package bsdb

// define metadata query statement
const (
	// DeletedObjectsDefaultSize defines the default size of ListDeletedObjectsByBlockNumberRange response
	DeletedObjectsDefaultSize = 1000
	// ExpiredBucketsDefaultSize defines the default size of ListExpiredBucketsBySp response
	ExpiredBucketsDefaultSize = 1000
	// ListObjectsDefaultMaxKeys defines the default size of ListObjectsByBucketName response
	ListObjectsDefaultMaxKeys = 50
	// GetUserBucketsLimitSize defines the default limit for the number of buckets in any given account is 100
	GetUserBucketsLimitSize = 100
)

// define table name constant of block syncer db
const (
	// BucketTableName defines the name of bucket table
	BucketTableName = "buckets"
	// ObjectTableName defines the name of object table
	ObjectTableName = "objects"
	// EpochTableName defines the name of epoch table
	EpochTableName = "epoch"
	// PermissionTableName defines the name of permission table
	PermissionTableName = "permission"
	// StatementTableName defines the name of statement table
	StatementTableName = "statements"
	// GroupTableName defines the name of group table
	GroupTableName = "groups"
)
