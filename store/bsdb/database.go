package bsdb

import (
	"github.com/ethereum/go-ethereum/common"
)

// Metadata contains all the methods required by block syncer db database
type Metadata interface {
	// GetUserBuckets get buckets info by a user address
	GetUserBuckets(accountID common.Address) ([]*Bucket, error)
	// GetUserBucketsCount get buckets count by a user address
	GetUserBucketsCount(accountID common.Address) (int64, error)
	// GetBucketByName get buckets info by a bucket name
	GetBucketByName(bucketName string, isFullList bool) (*Bucket, error)
	// GetBucketByID get buckets info by by a bucket id
	GetBucketByID(bucketID int64, isFullList bool) (*Bucket, error)
	// GetLatestBlockNumber get current latest block number
	GetLatestBlockNumber() (int64, error)
	// GetPaymentByBucketName get bucket payment info by a bucket name
	GetPaymentByBucketName(bucketName string, isFullList bool) (*StreamRecord, error)
	// GetPaymentByBucketID get bucket payment info by a bucket id
	GetPaymentByBucketID(bucketID int64, isFullList bool) (*StreamRecord, error)
	// GetPaymentByPaymentAddress get bucket payment info by a payment address
	GetPaymentByPaymentAddress(address common.Address) (*StreamRecord, error)
	// ListObjectsByBucketName list objects info by a bucket name
	ListObjectsByBucketName(bucketName string) ([]*Object, error)
	// ListDeletedObjectsByBlockNumberRange list deleted objects info by a block number range
	ListDeletedObjectsByBlockNumberRange(startBlockNumber int64, endBlockNumber int64, isFullList bool) ([]*Object, error)
	// GetObjectByName get object info by an object name
	GetObjectByName(objectName string, bucketName string, isFullList bool) (*Object, error)
}

// BSDB contains all the methods required by block syncer database
type BSDB interface {
	Metadata
}
